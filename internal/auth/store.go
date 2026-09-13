package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type execer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

type store struct {
	db *pgxpool.Pool
}

func (s store) withTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.DisplayName,
		&u.EmailVerifiedAt,
		&u.FailedLogins,
		&u.LockedUntil,
		&u.DisabledAt,
		&u.DeletedAt,
	)
	return u, err
}

func (s store) insertUser(ctx context.Context, tx pgx.Tx, u User, params []byte) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO users (
			id, email, password_hash, password_algo, password_params,
			display_name, password_updated_at, updated_at
		) VALUES ($1, $2, $3, 'argon2id', $4, $5, now(), now())
	`, u.ID, u.Email, u.PasswordHash, params, u.DisplayName)
	return err
}

func (s store) insertPreferences(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	_, err := tx.Exec(ctx, `INSERT INTO user_preferences (user_id) VALUES ($1)`, userID)
	return err
}

func (s store) insertPasswordHistory(ctx context.Context, q execer, userID uuid.UUID, hash string) error {
	_, err := q.Exec(ctx, `
		INSERT INTO password_history (user_id, password_hash, password_algo)
		VALUES ($1, $2, 'argon2id')
	`, userID, hash)
	return err
}

func (s store) insertAuthToken(ctx context.Context, q execer, userID uuid.UUID, purpose, tokenHash string, expires time.Time, meta RequestMeta) error {
	_, err := q.Exec(ctx, `
		INSERT INTO auth_tokens (user_id, purpose, token_hash, expires_at, request_ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, purpose, tokenHash, expires, inetArg(meta.IP), nullIfEmpty(meta.UserAgent))
	return err
}

func (s store) insertAuthEvent(ctx context.Context, q execer, userID *uuid.UUID, sessionID *uuid.UUID, eventType string, meta RequestMeta) error {
	_, err := q.Exec(ctx, `
		INSERT INTO auth_events (user_id, session_id, event_type, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, sessionID, eventType, inetArg(meta.IP), nullIfEmpty(meta.UserAgent))
	return err
}

func (s store) insertLoginAttempt(ctx context.Context, userID *uuid.UUID, email string, success bool, reason string, meta RequestMeta) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO login_attempts (user_id, email, ip, user_agent, success, failure_reason)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, email, inetArg(meta.IP), nullIfEmpty(meta.UserAgent), success, nullIfEmpty(reason))
	return err
}

func (s store) findUserByEmail(ctx context.Context, email string) (User, error) {
	u, err := scanUser(s.db.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, email_verified_at,
		       failed_login_count, locked_until, disabled_at, deleted_at
		FROM users
		WHERE lower(email) = $1 AND deleted_at IS NULL
	`, email))
	return u, err
}

func (s store) findUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	u, err := scanUser(s.db.QueryRow(ctx, `
		SELECT id, email, password_hash, display_name, email_verified_at,
		       failed_login_count, locked_until, disabled_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id))
	return u, err
}

func (s store) insertSession(ctx context.Context, sess Session, rawIP, ua string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO sessions (
			id, user_id, token_hash, csrf_hash, expires_at, idle_timeout_at, user_agent, ip
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, sess.ID, sess.UserID, sess.TokenHash, sess.CSRFHash, sess.ExpiresAt, sess.IdleTimeoutAt, nullIfEmpty(ua), inetArg(rawIP))
	return err
}

func (s store) findSessionByTokenHash(ctx context.Context, tokenHash string) (Session, User, error) {
	var sess Session
	var u User
	err := s.db.QueryRow(ctx, `
		SELECT s.id, s.user_id, s.token_hash, COALESCE(s.csrf_hash, ''), s.expires_at,
		       s.idle_timeout_at, s.revoked_at,
		       u.id, u.email, u.password_hash, u.display_name, u.email_verified_at,
		       u.failed_login_count, u.locked_until, u.disabled_at, u.deleted_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
	`, tokenHash).Scan(
		&sess.ID, &sess.UserID, &sess.TokenHash, &sess.CSRFHash, &sess.ExpiresAt,
		&sess.IdleTimeoutAt, &sess.RevokedAt,
		&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.EmailVerifiedAt,
		&u.FailedLogins, &u.LockedUntil, &u.DisabledAt, &u.DeletedAt,
	)
	return sess, u, err
}

func (s store) touchSession(ctx context.Context, id uuid.UUID, idleUntil time.Time) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET last_seen_at = now(), idle_timeout_at = $2 WHERE id = $1
	`, id, idleUntil)
	return err
}

func (s store) revokeSession(ctx context.Context, id uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now(), revoke_reason = $2 WHERE id = $1 AND revoked_at IS NULL
	`, id, reason)
	return err
}

func (s store) revokeUserSessions(ctx context.Context, userID uuid.UUID, reason string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now(), revoke_reason = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID, reason)
	return err
}

func (s store) recordLoginSuccess(ctx context.Context, userID uuid.UUID, ip string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET
			last_login_at = now(),
			last_login_ip = $2,
			last_seen_at = now(),
			failed_login_count = 0,
			locked_until = NULL,
			updated_at = now()
		WHERE id = $1
	`, userID, inetArg(ip))
	return err
}

func (s store) recordLoginFailure(ctx context.Context, userID uuid.UUID, lockAfter int, lockFor time.Duration) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET
			failed_login_count = failed_login_count + 1,
			locked_until = CASE
				WHEN failed_login_count + 1 >= $2 THEN now() + $3::interval
				ELSE locked_until
			END,
			updated_at = now()
		WHERE id = $1
	`, userID, lockAfter, fmt.Sprintf("%d seconds", int(lockFor.Seconds())))
	return err
}

func (s store) markEmailVerified(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET email_verified_at = COALESCE(email_verified_at, now()), updated_at = now()
		WHERE id = $1
	`, userID)
	return err
}

func (s store) updatePassword(ctx context.Context, userID uuid.UUID, hash string, params []byte) error {
	_, err := s.db.Exec(ctx, `
		UPDATE users SET
			password_hash = $2,
			password_algo = 'argon2id',
			password_params = $3,
			password_updated_at = now(),
			failed_login_count = 0,
			locked_until = NULL,
			updated_at = now()
		WHERE id = $1
	`, userID, hash, params)
	return err
}

func (s store) recentPasswordHashes(ctx context.Context, userID uuid.UUID, limit int) ([]string, error) {
	rows, err := s.db.Query(ctx, `
		SELECT password_hash FROM password_history
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hashes []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		hashes = append(hashes, h)
	}
	return hashes, rows.Err()
}

type authTokenRow struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Purpose   string
	ExpiresAt time.Time
	Consumed  *time.Time
	Revoked   *time.Time
}

func (s store) findAuthToken(ctx context.Context, tokenHash, purpose string) (authTokenRow, error) {
	var t authTokenRow
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, purpose, expires_at, consumed_at, revoked_at
		FROM auth_tokens
		WHERE token_hash = $1 AND purpose = $2
	`, tokenHash, purpose).Scan(&t.ID, &t.UserID, &t.Purpose, &t.ExpiresAt, &t.Consumed, &t.Revoked)
	return t, err
}

func (s store) consumeAuthToken(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE auth_tokens SET consumed_at = now() WHERE id = $1 AND consumed_at IS NULL AND revoked_at IS NULL
	`, id)
	return err
}

func (s store) revokeAuthTokens(ctx context.Context, userID uuid.UUID, purpose string) error {
	_, err := s.db.Exec(ctx, `
		UPDATE auth_tokens SET revoked_at = now()
		WHERE user_id = $1 AND purpose = $2 AND consumed_at IS NULL AND revoked_at IS NULL
	`, userID, purpose)
	return err
}

func (s store) listSessions(ctx context.Context, userID uuid.UUID) ([]SessionSummary, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, COALESCE(user_agent, ''), COALESCE(host(ip), ''), COALESCE(country, ''), created_at, last_seen_at, expires_at
		FROM sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()
		ORDER BY last_seen_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SessionSummary, 0)
	for rows.Next() {
		var sess SessionSummary
		if err := rows.Scan(&sess.ID, &sess.UserAgent, &sess.IP, &sess.Country, &sess.CreatedAt, &sess.LastSeenAt, &sess.ExpiresAt); err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s store) revokeSessionForUser(ctx context.Context, id, userID uuid.UUID, reason string) (bool, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE sessions SET revoked_at = now(), revoke_reason = $3
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL
	`, id, userID, reason)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
