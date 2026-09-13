package models

import (
	"encoding/json"
	"time"

	"github.com/kashifxyz/flow-server/internal/graph"
)

type Node struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	SpaceID     *string         `json:"space_id,omitempty"`
	ParentID    *string         `json:"parent_id,omitempty"`
	Type        string          `json:"type"`
	Subtype     *string         `json:"subtype,omitempty"`
	Flavour     *string         `json:"flavour,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        *string         `json:"icon,omitempty"`
	Rank        string          `json:"rank"`
	Props       json.RawMessage `json:"props"`
	Version     int64           `json:"version"`
	CreatedBy   *string         `json:"created_by,omitempty"`
	ArchivedAt  *time.Time      `json:"archived_at,omitempty"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func FromGraph(n graph.Node) Node {
	out := Node{
		ID:          n.ID.String(),
		WorkspaceID: n.WorkspaceID.String(),
		Type:        n.Type,
		Subtype:     n.Subtype,
		Flavour:     n.Flavour,
		Title:       n.Title,
		Description: n.Description,
		Icon:        n.Icon,
		Rank:        n.Rank,
		Props:       n.Props,
		Version:     n.Version,
		ArchivedAt:  n.ArchivedAt,
		DeletedAt:   n.DeletedAt,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	if n.SpaceID != nil {
		s := n.SpaceID.String()
		out.SpaceID = &s
	}
	if n.ParentID != nil {
		s := n.ParentID.String()
		out.ParentID = &s
	}
	if n.CreatedBy != nil {
		s := n.CreatedBy.String()
		out.CreatedBy = &s
	}
	return out
}

type UpdateRequest struct {
	Title       *string         `json:"title"`
	Description *string         `json:"description"`
	Icon        *string         `json:"icon"`
	Props       json.RawMessage `json:"props"`
	Rank        *string         `json:"rank"`
	SpaceID     *string         `json:"space_id"`
}

type MoveRequest struct {
	ParentID *string `json:"parent_id"`
	Rank     string  `json:"rank"`
}

type TrashItem struct {
	NodeID           string    `json:"node_id"`
	Title            string    `json:"title"`
	Type             string    `json:"type"`
	DeletedAt        time.Time `json:"deleted_at"`
	RestoreUntil     time.Time `json:"restore_until"`
	OriginalParentID *string   `json:"original_parent_id,omitempty"`
}

type Permission struct {
	PrincipalType string     `json:"principal_type"`
	PrincipalID   string     `json:"principal_id"`
	Role          string     `json:"role"`
	Inherit       bool       `json:"inherit"`
	PublicRole    *string    `json:"public_role,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

type SetPermissionsRequest struct {
	Permissions []PermissionInput `json:"permissions"`
}

type PermissionInput struct {
	PrincipalType string     `json:"principal_type"`
	PrincipalID   string     `json:"principal_id"`
	Role          string     `json:"role"`
	Inherit       *bool      `json:"inherit"`
	ExpiresAt     *time.Time `json:"expires_at"`
}
