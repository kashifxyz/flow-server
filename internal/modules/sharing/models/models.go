package models

import "time"

var Roles = map[string]bool{"view": true, "comment": true, "edit": true}

type ShareLink struct {
	ID                string     `json:"id"`
	NodeID            string     `json:"node_id"`
	Role              string     `json:"role"`
	PasswordProtected bool       `json:"password_protected"`
	IncludeSubtree    bool       `json:"include_subtree"`
	AllowDuplicate    bool       `json:"allow_duplicate"`
	MaxUses           *int       `json:"max_uses,omitempty"`
	UseCount          int        `json:"use_count"`
	Watermark         bool       `json:"watermark"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type CreateShareLinkRequest struct {
	Role           string     `json:"role"`
	Password       string     `json:"password"`
	IncludeSubtree bool       `json:"include_subtree"`
	AllowDuplicate bool       `json:"allow_duplicate"`
	MaxUses        *int       `json:"max_uses"`
	Watermark      bool       `json:"watermark"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

// CreateShareLinkResponse embeds the ShareLink and includes the raw token exactly once.
type CreateShareLinkResponse struct {
	ShareLink
	Token string `json:"token"`
}

type UpdateShareLinkRequest struct {
	Role           *string    `json:"role"`
	IncludeSubtree *bool      `json:"include_subtree"`
	AllowDuplicate *bool      `json:"allow_duplicate"`
	MaxUses        *int       `json:"max_uses"`
	Watermark      *bool      `json:"watermark"`
	ExpiresAt      *time.Time `json:"expires_at"`
}

type PublicResolveResponse struct {
	NodeID string `json:"node_id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Role   string `json:"role"`
}

type PublishRequest struct {
	Slug    string `json:"slug"`
	Indexed bool   `json:"indexed"`
}

type PublishedPage struct {
	NodeID    string    `json:"node_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Indexed   bool      `json:"indexed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
