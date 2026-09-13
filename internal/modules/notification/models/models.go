package models

import (
	"encoding/json"
	"time"
)

type Notification struct {
	ID            string          `json:"id"`
	WorkspaceID   string          `json:"workspace_id"`
	ActorID       *string         `json:"actor_id,omitempty"`
	SubjectNodeID *string         `json:"subject_node_id,omitempty"`
	Type          string          `json:"type"`
	Title         string          `json:"title"`
	Body          string          `json:"body"`
	Payload       json.RawMessage `json:"payload"`
	ReadAt        *time.Time      `json:"read_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

type ActivityEvent struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	NodeID      *string         `json:"node_id,omitempty"`
	ActorID     *string         `json:"actor_id,omitempty"`
	EventType   string          `json:"event_type"`
	Summary     string          `json:"summary"`
	Payload     json.RawMessage `json:"payload"`
	CreatedAt   time.Time       `json:"created_at"`
}
