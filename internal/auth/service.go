package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/pkg/mail"
	"github.com/rs/zerolog"
)

const (
	PurposeEmailVerify   = "email_verification"
	PurposePasswordReset = "password_reset"

	lockAfterFailures = 8
	passwordHistoryN  = 5
)

var passwordParamsJSON = []byte(`{"m":65536,"t":1,"p":4}`)

type Service struct {
	store      store
	mail       mail.Sender
	log        zerolog.Logger
	publicURL  string
	sessionTTL time.Duration
	idleTTL    time.Duration
	verifyTTL  time.Duration
	resetTTL   time.Duration
	lockFor    time.Duration
	dummyHash  string
}

func NewService(db *pgxpool.Pool, sender mail.Sender, publicURL string, log zerolog.Logger) (*Service, error) {
	dummy, err := HashPassword("flow-timing-dummy")
	if err != nil {
		return nil, fmt.Errorf("dummy password hash: %w", err)
	}
	return &Service{
		store:      store{db: db},
		mail:       sender,
		log:        log,
		publicURL:  publicURL,
		sessionTTL: 30 * 24 * time.Hour,
		idleTTL:    14 * 24 * time.Hour,
		verifyTTL:  24 * time.Hour,
		resetTTL:   time.Hour,
		lockFor:    15 * time.Minute,
		dummyHash:  dummy,
	}, nil
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type LoginInput struct {
	Email    string
	Password string
}

type IssuedSession struct {
	SessionToken string
	CSRFToken    string
	TTL          time.Duration
	User         User
}

func (s *Service) Register(ctx context.Context, in RegisterInput, meta RequestMeta) error {
	email, err := NormalizeEmail(in.Email)
	if err != nil {
		return err
	}
	if err := ValidatePassword(in.Password); err != nil {
		return err
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return err
	}
	uid, err := utils.NewID()
	if err != nil {
		return err
	}
	user := User{
		ID:           uid,
		Email:        email,
		PasswordHash: hash,
		DisplayName:  NormalizeDisplayName(in.DisplayName, email),
	}
	raw, err := RandomToken()
	if err != nil {
		return err
	}
	tokenHash := HashToken(raw)
	expires := time.Now().Add(s.verifyTTL)

	err = s.store.withTx(ctx, func(tx pgx.Tx) error {
		if err := s.store.insertUser(ctx, tx, user, passwordParamsJSON); err != nil {
			return err
		}
		if err := s.store.insertPreferences(ctx, tx, user.ID); err != nil {
			return err
		}
		if err := s.store.insertPasswordHistory(ctx, tx, user.ID, hash); err != nil {
			return err
		}
		if err := s.store.insertAuthToken(ctx, tx, user.ID, PurposeEmailVerify, tokenHash, expires, meta); err != nil {
			return err
		}
		idCopy := user.ID
		return s.store.insertAuthEvent(ctx, tx, &idCopy, nil, "registered", meta)
	})
	if err != nil {
		if isUniqueViolation(err) {
			return ErrEmailTaken
		}
		return err
	}
	s.sendMail(ctx, email, mail.TemplateVerifyEmail, "/verify-email", raw, "24 hours")
	return nil
}

func (s *Service) Login(ctx context.Context, in LoginInput, meta RequestMeta) (IssuedSession, error) {
	var zero IssuedSession
	email, err := NormalizeEmail(in.Email)
	if err != nil {
		return zero, err
	}
	if in.Password == "" {
		return zero, ErrInvalidRequest
	}

	user, err := s.store.findUserByEmail(ctx, email)
	missing := errors.Is(err, pgx.ErrNoRows)
	if err != nil && !missing {
		return zero, err
	}

	hash := s.dummyHash
	if !missing {
		hash = user.PasswordHash
	}
	ok, cmpErr := ComparePassword(hash, in.Password)
	if cmpErr != nil {
		return zero, cmpErr
	}

	if missing || !ok {
		if !missing {
			_ = s.store.recordLoginFailure(ctx, user.ID, lockAfterFailures, s.lockFor)
			uid := user.ID
			_ = s.store.insertLoginAttempt(ctx, &uid, email, false, "invalid_password", meta)
			_ = s.store.insertAuthEvent(ctx, s.store.db, &uid, nil, "login_failed", meta)
		} else {
			_ = s.store.insertLoginAttempt(ctx, nil, email, false, "unknown_email", meta)
		}
		return zero, ErrInvalidCredentials
	}

	now := time.Now()
	if user.DisabledAt != nil {
		uid := user.ID
		_ = s.store.insertLoginAttempt(ctx, &uid, email, false, "disabled", meta)
		return zero, ErrAccountDisabled
	}
	if user.LockedUntil != nil && user.LockedUntil.After(now) {
		uid := user.ID
		_ = s.store.insertLoginAttempt(ctx, &uid, email, false, "locked", meta)
		return zero, ErrAccountLocked
	}
	if user.EmailVerifiedAt == nil {
		uid := user.ID
		_ = s.store.insertLoginAttempt(ctx, &uid, email, false, "unverified", meta)
		return zero, ErrEmailUnverified
	}

	issued, sessID, err := s.createSession(ctx, user, meta)
	if err != nil {
		return zero, err
	}
	if err := s.store.recordLoginSuccess(ctx, user.ID, meta.IP); err != nil {
		return zero, err
	}
	uid := user.ID
	_ = s.store.insertLoginAttempt(ctx, &uid, email, true, "", meta)
	_ = s.store.insertAuthEvent(ctx, s.store.db, &uid, &sessID, "logged_in", meta)
	return issued, nil
}

func (s *Service) createSession(ctx context.Context, user User, meta RequestMeta) (IssuedSession, uuid.UUID, error) {
	var zero IssuedSession
	sessToken, err := RandomToken()
	if err != nil {
		return zero, uuid.Nil, err
	}
	csrfToken, err := RandomToken()
	if err != nil {
		return zero, uuid.Nil, err
	}
	sessID, err := utils.NewID()
	if err != nil {
		return zero, uuid.Nil, err
	}
	now := time.Now()
	idle := now.Add(s.idleTTL)
	sess := Session{
		ID:            sessID,
		UserID:        user.ID,
		TokenHash:     HashToken(sessToken),
		CSRFHash:      HashToken(csrfToken),
		ExpiresAt:     now.Add(s.sessionTTL),
		IdleTimeoutAt: &idle,
	}
	if err := s.store.insertSession(ctx, sess, meta.IP, meta.UserAgent); err != nil {
		return zero, uuid.Nil, err
	}
	return IssuedSession{
		SessionToken: sessToken,
		CSRFToken:    csrfToken,
		TTL:          s.sessionTTL,
		User:         user,
	}, sessID, nil
}

func (s *Service) Logout(ctx context.Context, sess Session, meta RequestMeta) error {
	if err := s.store.revokeSession(ctx, sess.ID, "logout"); err != nil {
		return err
	}
	uid := sess.UserID
	sid := sess.ID
	return s.store.insertAuthEvent(ctx, s.store.db, &uid, &sid, "logged_out", meta)
}

func (s *Service) LookupSession(ctx context.Context, rawToken string) (Session, User, error) {
	var zero Session
	var zu User
	if rawToken == "" {
		return zero, zu, ErrUnauthenticated
	}
	sess, user, err := s.store.findSessionByTokenHash(ctx, HashToken(rawToken))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, zu, ErrUnauthenticated
		}
		return zero, zu, err
	}
	now := time.Now()
	if sess.RevokedAt != nil || sess.ExpiresAt.Before(now) {
		return zero, zu, ErrUnauthenticated
	}
	if sess.IdleTimeoutAt != nil && sess.IdleTimeoutAt.Before(now) {
		_ = s.store.revokeSession(ctx, sess.ID, "idle_timeout")
		return zero, zu, ErrUnauthenticated
	}
	if user.DisabledAt != nil || user.DeletedAt != nil {
		_ = s.store.revokeSession(ctx, sess.ID, "account_unavailable")
		return zero, zu, ErrUnauthenticated
	}
	idle := now.Add(s.idleTTL)
	_ = s.store.touchSession(ctx, sess.ID, idle)
	return sess, user, nil
}

type SessionSummary struct {
	ID         uuid.UUID
	UserAgent  string
	IP         string
	Country    string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
}

// ListSessions returns userID's active (non-revoked, unexpired) sessions,
// most recently active first.
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID) ([]SessionSummary, error) {
	return s.store.listSessions(ctx, userID)
}

// RevokeSession revokes sessionID only if it belongs to userID.
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	ok, err := s.store.revokeSessionForUser(ctx, sessionID, userID, "user_revoked")
	if err != nil {
		return err
	}
	if !ok {
		return ErrSessionNotFound
	}
	return nil
}

func (s *Service) ValidCSRF(sess Session, headerToken, cookieToken string) bool {
	if headerToken == "" || cookieToken == "" || headerToken != cookieToken {
		return false
	}
	return EqualHash(HashToken(headerToken), sess.CSRFHash)
}

func (s *Service) VerifyEmail(ctx context.Context, rawToken string, meta RequestMeta) error {
	tok, user, err := s.loadUsableToken(ctx, rawToken, PurposeEmailVerify)
	if err != nil {
		return err
	}
	if err := s.store.markEmailVerified(ctx, user.ID); err != nil {
		return err
	}
	if err := s.store.consumeAuthToken(ctx, tok.ID); err != nil {
		return err
	}
	_ = s.store.revokeAuthTokens(ctx, user.ID, PurposeEmailVerify)
	uid := user.ID
	return s.store.insertAuthEvent(ctx, s.store.db, &uid, nil, "email_verified", meta)
}

func (s *Service) ResendVerification(ctx context.Context, emailRaw string, meta RequestMeta) error {
	email, err := NormalizeEmail(emailRaw)
	if err != nil {
		return err
	}
	user, err := s.store.findUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if user.EmailVerifiedAt != nil || user.DisabledAt != nil {
		return nil
	}
	return s.issueEmailToken(ctx, user, PurposeEmailVerify, mail.TemplateVerifyEmail, "/verify-email", s.verifyTTL, "24 hours", "verification_resent", meta)
}

func (s *Service) ForgotPassword(ctx context.Context, emailRaw string, meta RequestMeta) error {
	email, err := NormalizeEmail(emailRaw)
	if err != nil {
		return err
	}
	user, err := s.store.findUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if user.DisabledAt != nil {
		return nil
	}
	uid := user.ID
	_ = s.store.insertAuthEvent(ctx, s.store.db, &uid, nil, "password_reset_requested", meta)
	return s.issueEmailToken(ctx, user, PurposePasswordReset, mail.TemplateResetPassword, "/reset-password", s.resetTTL, "1 hour", "", meta)
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, password string, meta RequestMeta) error {
	if err := ValidatePassword(password); err != nil {
		return err
	}
	tok, user, err := s.loadUsableToken(ctx, rawToken, PurposePasswordReset)
	if err != nil {
		return err
	}
	reused, err := s.passwordReused(ctx, user, password)
	if err != nil {
		return err
	}
	if reused {
		return ErrPasswordReused
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	if err := s.store.updatePassword(ctx, user.ID, hash, passwordParamsJSON); err != nil {
		return err
	}
	if err := s.store.insertPasswordHistory(ctx, s.store.db, user.ID, hash); err != nil {
		return err
	}
	if err := s.store.consumeAuthToken(ctx, tok.ID); err != nil {
		return err
	}
	_ = s.store.revokeAuthTokens(ctx, user.ID, PurposePasswordReset)
	if err := s.store.revokeUserSessions(ctx, user.ID, "password_reset"); err != nil {
		return err
	}
	uid := user.ID
	return s.store.insertAuthEvent(ctx, s.store.db, &uid, nil, "password_reset", meta)
}

func (s *Service) issueEmailToken(ctx context.Context, user User, purpose, template, path string, ttl time.Duration, expiresLabel, eventType string, meta RequestMeta) error {
	if err := s.store.revokeAuthTokens(ctx, user.ID, purpose); err != nil {
		return err
	}
	raw, err := RandomToken()
	if err != nil {
		return err
	}
	if err := s.store.insertAuthToken(ctx, s.store.db, user.ID, purpose, HashToken(raw), time.Now().Add(ttl), meta); err != nil {
		return err
	}
	if eventType != "" {
		uid := user.ID
		_ = s.store.insertAuthEvent(ctx, s.store.db, &uid, nil, eventType, meta)
	}
	s.sendMail(ctx, user.Email, template, path, raw, expiresLabel)
	return nil
}

func (s *Service) loadUsableToken(ctx context.Context, raw, purpose string) (authTokenRow, User, error) {
	var zero authTokenRow
	var zu User
	if raw == "" {
		return zero, zu, ErrInvalidToken
	}
	tok, err := s.store.findAuthToken(ctx, HashToken(raw), purpose)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, zu, ErrInvalidToken
		}
		return zero, zu, err
	}
	now := time.Now()
	if tok.Consumed != nil || tok.Revoked != nil || tok.ExpiresAt.Before(now) {
		return zero, zu, ErrInvalidToken
	}
	user, err := s.store.findUserByID(ctx, tok.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return zero, zu, ErrInvalidToken
		}
		return zero, zu, err
	}
	if user.DisabledAt != nil {
		return zero, zu, ErrAccountDisabled
	}
	return tok, user, nil
}

func (s *Service) passwordReused(ctx context.Context, user User, password string) (bool, error) {
	ok, err := ComparePassword(user.PasswordHash, password)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	hashes, err := s.store.recentPasswordHashes(ctx, user.ID, passwordHistoryN)
	if err != nil {
		return false, err
	}
	for _, h := range hashes {
		match, err := ComparePassword(h, password)
		if err != nil {
			continue
		}
		if match {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) sendMail(ctx context.Context, to, template, path, raw, expiresLabel string) {
	if s.mail == nil {
		return
	}
	link := s.publicURL + path + "?token=" + url.QueryEscape(raw)
	if err := s.mail.Send(ctx, mail.Message{
		To:       to,
		Template: template,
		Payload: map[string]any{
			"link":    link,
			"expires": expiresLabel,
		},
	}); err != nil {
		s.log.Error().Err(err).Str("template", template).Msg("auth mail send failed")
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
