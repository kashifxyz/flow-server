package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
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
