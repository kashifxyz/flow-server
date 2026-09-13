package models

import (
	"encoding/json"
	"time"

	"github.com/kashifxyz/flow-server/internal/graph"
)

type Canvas struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	SpaceID     *string         `json:"space_id,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Props       json.RawMessage `json:"props"`
	Version     int64           `json:"version"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func CanvasFromGraph(n graph.Node) Canvas {
	c := Canvas{
		ID:          n.ID.String(),
		WorkspaceID: n.WorkspaceID.String(),
		Title:       n.Title,
		Description: n.Description,
		Props:       n.Props,
		Version:     n.Version,
		CreatedAt:   n.CreatedAt,
		UpdatedAt:   n.UpdatedAt,
	}
	if n.SpaceID != nil {
		s := n.SpaceID.String()
		c.SpaceID = &s
	}
	return c
}

type CreateCanvasRequest struct {
	Title   string `json:"title"`
	SpaceID string `json:"space_id"`
}

type UpdateCanvasRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

var ObjectKinds = map[string]bool{
	"shape": true, "text": true, "image": true, "embed": true, "sticky_note": true, "card": true,
}

type Object struct {
	ID        string          `json:"id"`
	CanvasID  string          `json:"canvas_id"`
	Kind      string          `json:"object_kind"`
	Title     string          `json:"title"`
	X         float64         `json:"x"`
	Y         float64         `json:"y"`
	Width     float64         `json:"width"`
	Height    float64         `json:"height"`
	Rotation  float64         `json:"rotation"`
	ZIndex    int             `json:"z_index"`
	Locked    bool            `json:"locked"`
	Visible   bool            `json:"visible"`
	Opacity   float64         `json:"opacity"`
	Style     json.RawMessage `json:"style"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateObjectRequest struct {
	ObjectKind string          `json:"object_kind"`
	Title      string          `json:"title"`
	X          float64         `json:"x"`
	Y          float64         `json:"y"`
	Width      float64         `json:"width"`
	Height     float64         `json:"height"`
	Rotation   float64         `json:"rotation"`
	ZIndex     int             `json:"z_index"`
	Style      json.RawMessage `json:"style"`
}

type UpdateObjectRequest struct {
	Title    *string         `json:"title"`
	X        *float64        `json:"x"`
	Y        *float64        `json:"y"`
	Width    *float64        `json:"width"`
	Height   *float64        `json:"height"`
	Rotation *float64        `json:"rotation"`
	ZIndex   *int            `json:"z_index"`
	Locked   *bool           `json:"locked"`
	Visible  *bool           `json:"visible"`
	Opacity  *float64        `json:"opacity"`
	Style    json.RawMessage `json:"style"`
}

type Connector struct {
	ID            string          `json:"id"`
	CanvasID      string          `json:"canvas_id"`
	StartObjectID string          `json:"start_object_id"`
	EndObjectID   string          `json:"end_object_id"`
	Shape         string          `json:"shape"`
	StartSnapTo   string          `json:"start_snap_to"`
	EndSnapTo     string          `json:"end_snap_to"`
	Captions      json.RawMessage `json:"captions"`
	Style         json.RawMessage `json:"style"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type CreateConnectorRequest struct {
	StartObjectID string          `json:"start_object_id"`
	EndObjectID   string          `json:"end_object_id"`
	Shape         string          `json:"shape"`
	StartSnapTo   string          `json:"start_snap_to"`
	EndSnapTo     string          `json:"end_snap_to"`
	Captions      json.RawMessage `json:"captions"`
	Style         json.RawMessage `json:"style"`
}

type UpdateConnectorRequest struct {
	Shape       *string         `json:"shape"`
	StartSnapTo *string         `json:"start_snap_to"`
	EndSnapTo   *string         `json:"end_snap_to"`
	Captions    json.RawMessage `json:"captions"`
	Style       json.RawMessage `json:"style"`
}

type Camera struct {
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Zoom      float64   `json:"zoom"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PutCameraRequest struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

type Tag struct {
	ID        string    `json:"id"`
	CanvasID  string    `json:"canvas_id"`
	Name      string    `json:"name"`
	Color     *string   `json:"color,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
