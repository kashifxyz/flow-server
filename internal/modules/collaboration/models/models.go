package models

import (
	"encoding/json"
	"time"
)

type Presence struct {
	UserID     string          `json:"user_id"`
	NodeID     string          `json:"node_id"`
	Selection  json.RawMessage `json:"selection,omitempty"`
	Cursor     json.RawMessage `json:"cursor,omitempty"`
	Viewport   json.RawMessage `json:"viewport,omitempty"`
	LastSeenAt time.Time       `json:"last_seen_at"`
}

type PutPresenceRequest struct {
	Selection json.RawMessage `json:"selection"`
	Cursor    json.RawMessage `json:"cursor"`
	Viewport  json.RawMessage `json:"viewport"`
}
