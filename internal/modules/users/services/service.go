package services

import (
	"context"
	"errors"
	"strings"
	"time"
	_ "time/tzdata" // embed the IANA tzdata so time.LoadLocation works even without an OS copy
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	maxBioLen      = 2000
	maxLocaleLen   = 35 // RFC 5646 caps a full BCP-47 language tag at 35 chars
	maxTimezoneLen = 100
	maxFormatLen   = 40
)

type Service struct {
	DB   *pgxpool.Pool
	Auth *auth.Service
}

func New(db *pgxpool.Pool, authSvc *auth.Service) *Service {
	return &Service{DB: db, Auth: authSvc}
}

var themes = map[string]bool{"system": true, "light": true, "dark": true, "amoled": true}

const profileCols = `id, email, display_name, given_name, family_name, username, locale, timezone, week_starts_on, date_format, time_format, theme, bio, avatar_object_key`

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (models.Profile, error) {
	p, err := scanProfile(s.DB.QueryRow(ctx, `SELECT `+profileCols+` FROM users WHERE id = $1 AND deleted_at IS NULL`, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Profile{}, httperr.ErrNotFound
	}
	return p, err
}

func (s *Service) ChangePassword(ctx context.Context, userID, sessionID uuid.UUID, in models.ChangePasswordRequest, meta auth.RequestMeta) error {
	if in.CurrentPassword == "" || in.NewPassword == "" {
		return httperr.ErrInvalid
	}
	return s.Auth.ChangePassword(ctx, userID, sessionID, in.CurrentPassword, in.NewPassword, meta)
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, in models.UpdateProfileRequest) (models.Profile, error) {
	if in.Theme != nil && !themes[*in.Theme] {
		return models.Profile{}, httperr.ErrInvalid
	}
	if in.WeekStartsOn != nil && (*in.WeekStartsOn < 0 || *in.WeekStartsOn > 6) {
		return models.Profile{}, httperr.ErrInvalid
	}
	if in.Username != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*in.Username))
		if trimmed == "" || len(trimmed) > 39 {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.Username = &trimmed
	}
	if in.DisplayName != nil {
		trimmed := strings.TrimSpace(*in.DisplayName)
		if utf8.RuneCountInString(trimmed) > auth.MaxNameLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.DisplayName = &trimmed
	}
	if in.GivenName != nil {
		trimmed := strings.TrimSpace(*in.GivenName)
		if utf8.RuneCountInString(trimmed) > auth.MaxNameLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.GivenName = &trimmed
	}
	if in.FamilyName != nil {
		trimmed := strings.TrimSpace(*in.FamilyName)
		if utf8.RuneCountInString(trimmed) > auth.MaxNameLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.FamilyName = &trimmed
	}
	if in.Bio != nil {
		trimmed := strings.TrimSpace(*in.Bio)
		if utf8.RuneCountInString(trimmed) > maxBioLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.Bio = &trimmed
	}
	if in.Locale != nil {
		trimmed := strings.TrimSpace(*in.Locale)
		if utf8.RuneCountInString(trimmed) > maxLocaleLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.Locale = &trimmed
	}
	if in.Timezone != nil {
		trimmed := strings.TrimSpace(*in.Timezone)
		if utf8.RuneCountInString(trimmed) > maxTimezoneLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		if trimmed != "" {
			if _, err := time.LoadLocation(trimmed); err != nil {
				return models.Profile{}, httperr.ErrInvalid
			}
		}
		in.Timezone = &trimmed
	}
	if in.DateFormat != nil {
		trimmed := strings.TrimSpace(*in.DateFormat)
		if utf8.RuneCountInString(trimmed) > maxFormatLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.DateFormat = &trimmed
	}
	if in.TimeFormat != nil {
		trimmed := strings.TrimSpace(*in.TimeFormat)
		if utf8.RuneCountInString(trimmed) > maxFormatLen {
			return models.Profile{}, httperr.ErrInvalid
		}
		in.TimeFormat = &trimmed
	}
	p, err := scanProfile(s.DB.QueryRow(ctx, `
		UPDATE users SET
			display_name = COALESCE($2, display_name),
			given_name = COALESCE($3, given_name),
			family_name = COALESCE($4, family_name),
			username = COALESCE($5, username),
			locale = COALESCE($6, locale),
			timezone = COALESCE($7, timezone),
			week_starts_on = COALESCE($8, week_starts_on),
			date_format = COALESCE($9, date_format),
			time_format = COALESCE($10, time_format),
			theme = COALESCE($11, theme),
			bio = COALESCE($12, bio),
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING `+profileCols, userID, in.DisplayName, in.GivenName, in.FamilyName, in.Username,
		in.Locale, in.Timezone, in.WeekStartsOn, in.DateFormat, in.TimeFormat, in.Theme, in.Bio))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Profile{}, httperr.ErrNotFound
	}
	return p, err
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (models.Preferences, error) {
	var p models.Preferences
	err := s.DB.QueryRow(ctx, `
		SELECT editor, notifications, accessibility, shortcuts, extras, updated_at
		FROM user_preferences WHERE user_id = $1
	`, userID).Scan(&p.Editor, &p.Notifications, &p.Accessibility, &p.Shortcuts, &p.Extras, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Preferences{}, httperr.ErrNotFound
	}
	return p, err
}

func (s *Service) UpdatePreferences(ctx context.Context, userID uuid.UUID, in models.UpdatePreferencesRequest) (models.Preferences, error) {
	var p models.Preferences
	err := s.DB.QueryRow(ctx, `
		UPDATE user_preferences SET
			editor = COALESCE($2, editor),
			notifications = COALESCE($3, notifications),
			accessibility = COALESCE($4, accessibility),
			shortcuts = COALESCE($5, shortcuts),
			extras = COALESCE($6, extras),
			updated_at = now()
		WHERE user_id = $1
		RETURNING editor, notifications, accessibility, shortcuts, extras, updated_at
	`, userID, in.Editor, in.Notifications, in.Accessibility, in.Shortcuts, in.Extras).Scan(
		&p.Editor, &p.Notifications, &p.Accessibility, &p.Shortcuts, &p.Extras, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Preferences{}, httperr.ErrNotFound
	}
	return p, err
}

func (s *Service) ListSessions(ctx context.Context, userID, currentSessionID uuid.UUID) ([]models.Session, error) {
	rows, err := s.Auth.ListSessions(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]models.Session, len(rows))
	for i, r := range rows {
		out[i] = models.Session{
			ID:         r.ID.String(),
			UserAgent:  r.UserAgent,
			IP:         r.IP,
			Country:    r.Country,
			CreatedAt:  r.CreatedAt,
			LastSeenAt: r.LastSeenAt,
			ExpiresAt:  r.ExpiresAt,
			Current:    r.ID == currentSessionID,
		}
	}
	return out, nil
}

func (s *Service) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	err := s.Auth.RevokeSession(ctx, userID, sessionID)
	if errors.Is(err, auth.ErrSessionNotFound) {
		return httperr.ErrNotFound
	}
	return err
}

func scanProfile(row pgx.Row) (models.Profile, error) {
	var p models.Profile
	var id uuid.UUID
	err := row.Scan(&id, &p.Email, &p.DisplayName, &p.GivenName, &p.FamilyName, &p.Username,
		&p.Locale, &p.Timezone, &p.WeekStartsOn, &p.DateFormat, &p.TimeFormat, &p.Theme, &p.Bio, &p.AvatarObjectKey)
	if err != nil {
		return models.Profile{}, err
	}
	p.ID = id.String()
	return p, nil
}
