package models

import (
	"encoding/json"
	"time"

	"github.com/kashifxyz/flow-server/internal/graph"
)

type Block struct {
	ID        string          `json:"id"`
	PageID    string          `json:"page_id"`
	Flavour   *string         `json:"flavour,omitempty"`
	Rank      string          `json:"rank"`
	Props     json.RawMessage `json:"props"`
	Version   int64           `json:"version"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func FromGraph(n graph.Node) Block {
	b := Block{
		ID:        n.ID.String(),
		Flavour:   n.Flavour,
		Rank:      n.Rank,
		Props:     n.Props,
		Version:   n.Version,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
	}
	if n.ParentID != nil {
		b.PageID = n.ParentID.String()
	}
	return b
}

type CreateRequest struct {
	Flavour string          `json:"flavour"`
	Rank    string          `json:"rank"`
	Props   json.RawMessage `json:"props"`
}

type UpdateRequest struct {
	Flavour *string         `json:"flavour"`
	Props   json.RawMessage `json:"props"`
	Rank    *string         `json:"rank"`
}

type ReorderRequest struct {
	Order []ReorderItem `json:"order"`
}

type ReorderItem struct {
	ID   string `json:"id"`
	Rank string `json:"rank"`
}

type Flavour struct {
	Name         string          `json:"flavour"`
	Version      int             `json:"version"`
	NodeType     string          `json:"node_type"`
	Schema       json.RawMessage `json:"schema"`
	DeprecatedAt *time.Time      `json:"deprecated_at,omitempty"`
}
