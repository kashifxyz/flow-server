package models

import (
	"encoding/json"
	"time"
)

type Discussion struct {
	ID              string          `json:"id"`
	NodeID          string          `json:"node_id"`
	SelectionAnchor json.RawMessage `json:"selection_anchor,omitempty"`
	CommentCount    int             `json:"comment_count"`
	ResolvedAt      *time.Time      `json:"resolved_at,omitempty"`
	ResolvedBy      *string         `json:"resolved_by,omitempty"`
	CreatedBy       *string         `json:"created_by,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CreateDiscussionRequest struct {
	SelectionAnchor json.RawMessage `json:"selection_anchor"`
}

type Comment struct {
	ID           string          `json:"id"`
	DiscussionID string          `json:"discussion_id"`
	NodeID       string          `json:"node_id"`
	ParentID     *string         `json:"parent_id,omitempty"`
	Body         string          `json:"body"`
	RichText     json.RawMessage `json:"rich_text"`
	BodyFormat   string          `json:"body_format"`
	CreatedBy    *string         `json:"created_by,omitempty"`
	EditedAt     *time.Time      `json:"edited_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type CreateCommentRequest struct {
	Body     string          `json:"body"`
	RichText json.RawMessage `json:"rich_text"`
	ParentID string          `json:"parent_id"`
}

type UpdateCommentRequest struct {
	Body     *string         `json:"body"`
	RichText json.RawMessage `json:"rich_text"`
}

type Reaction struct {
	CommentID string    `json:"comment_id"`
	UserID    string    `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReactionRequest struct {
	Emoji string `json:"emoji"`
}
