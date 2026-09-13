package models

import (
	"encoding/json"
	"time"

	"github.com/kashifxyz/flow-server/internal/graph"
)

type Task struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	ProjectID   *string         `json:"project_id,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Rank        string          `json:"rank"`
	Props       json.RawMessage `json:"props"`
	Version     int64           `json:"version"`
	ArchivedAt  *time.Time      `json:"archived_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func FromGraph(n graph.Node) Task {
	t := Task{
		ID:          n.ID.String(),
		WorkspaceID: n.WorkspaceID.String(),
		Title:       n.Title,
		Description: n.Description,
		Rank:        n.Rank,
		Props:       n.Props,
		Version:     n.Version,
		ArchivedAt:  n.ArchivedAt,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	if n.ParentID != nil {
		s := n.ParentID.String()
		t.ProjectID = &s
	}
	return t
}

type CreateRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Rank        string          `json:"rank"`
	Props       json.RawMessage `json:"props"`
}

type UpdateRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	Props       json.RawMessage `json:"props"`
	Rank        *string         `json:"rank"`
}

type MoveRequest struct {
	ProjectID *string `json:"project_id"`
	Rank      string  `json:"rank"`
}
