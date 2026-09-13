package models

import (
	"encoding/json"
	"time"

	"github.com/kashifxyz/flow-server/internal/graph"
)

type Page struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	SpaceID     *string         `json:"space_id,omitempty"`
	ParentID    *string         `json:"parent_id,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        *string         `json:"icon,omitempty"`
	Rank        string          `json:"rank"`
	Props       json.RawMessage `json:"props"`
	Version     int64           `json:"version"`
	ArchivedAt  *time.Time      `json:"archived_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func FromGraph(n graph.Node) Page {
	p := Page{
		ID:          n.ID.String(),
		WorkspaceID: n.WorkspaceID.String(),
		Title:       n.Title,
		Description: n.Description,
		Icon:        n.Icon,
		Rank:        n.Rank,
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
	if n.ParentID != nil {
		s := n.ParentID.String()
		p.ParentID = &s
	}
	return p
}

type CreateRequest struct {
	ParentID string          `json:"parent_id"`
	SpaceID  string          `json:"space_id"`
	Title    string          `json:"title"`
	Icon     string          `json:"icon"`
	Props    json.RawMessage `json:"props"`
}

type UpdateRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	Icon        *string         `json:"icon"`
	Props       json.RawMessage `json:"props"`
}

type MoveRequest struct {
	ParentID *string `json:"parent_id"`
	Rank     string  `json:"rank"`
}
