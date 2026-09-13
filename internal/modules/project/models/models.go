package models

import (
	"encoding/json"
	"time"

	"github.com/kashifxyz/flow-server/internal/graph"
)

type Project struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	SpaceID     *string         `json:"space_id,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        *string         `json:"icon,omitempty"`
	Props       json.RawMessage `json:"props"`
	Version     int64           `json:"version"`
	ArchivedAt  *time.Time      `json:"archived_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func FromGraph(n graph.Node) Project {
	p := Project{
		ID:          n.ID.String(),
		WorkspaceID: n.WorkspaceID.String(),
		Title:       n.Title,
		Description: n.Description,
		Icon:        n.Icon,
		Props:       n.Props,
		Version:     n.Version,
		ArchivedAt:  n.ArchivedAt,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	if n.SpaceID != nil {
		s := n.SpaceID.String()
		p.SpaceID = &s
	}
	return p
}

type CreateRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	SpaceID     string          `json:"space_id"`
	Icon        string          `json:"icon"`
	Props       json.RawMessage `json:"props"`
}

type UpdateRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	Icon        *string         `json:"icon"`
	Props       json.RawMessage `json:"props"`
}
