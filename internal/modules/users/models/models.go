package models

import (
	"encoding/json"
	"time"
)

type Profile struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	DisplayName     string  `json:"display_name"`
	GivenName       string  `json:"given_name"`
	FamilyName      string  `json:"family_name"`
	Username        *string `json:"username,omitempty"`
	Locale          string  `json:"locale"`
	Timezone        string  `json:"timezone"`
	WeekStartsOn    int16   `json:"week_starts_on"`
	DateFormat      string  `json:"date_format"`
	TimeFormat      string  `json:"time_format"`
	Theme           string  `json:"theme"`
	Bio             string  `json:"bio"`
	AvatarObjectKey *string `json:"avatar_object_key,omitempty"`
}

type UpdateProfileRequest struct {
	DisplayName  *string `json:"display_name"`
	GivenName    *string `json:"given_name"`
	FamilyName   *string `json:"family_name"`
	Username     *string `json:"username"`
	Locale       *string `json:"locale"`
	Timezone     *string `json:"timezone"`
	WeekStartsOn *int16  `json:"week_starts_on"`
	DateFormat   *string `json:"date_format"`
	TimeFormat   *string `json:"time_format"`
	Theme        *string `json:"theme"`
	Bio          *string `json:"bio"`
}

type Preferences struct {
	Editor        json.RawMessage `json:"editor"`
	Notifications json.RawMessage `json:"notifications"`
	Accessibility json.RawMessage `json:"accessibility"`
	Shortcuts     json.RawMessage `json:"shortcuts"`
	Extras        json.RawMessage `json:"extras"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type UpdatePreferencesRequest struct {
	Editor        json.RawMessage `json:"editor"`
	Notifications json.RawMessage `json:"notifications"`
	Accessibility json.RawMessage `json:"accessibility"`
	Shortcuts     json.RawMessage `json:"shortcuts"`
	Extras        json.RawMessage `json:"extras"`
}

type Session struct {
	ID         string    `json:"id"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
	Country    string    `json:"country"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	Current    bool      `json:"current"`
}
