package models

import (
	"encoding/json"
	"time"
)

type Database struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	IsInline       bool      `json:"is_inline"`
	PrimaryFieldID *string   `json:"primary_field_id,omitempty"`
	RowCount       int       `json:"row_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateDatabaseRequest struct {
	Title    string `json:"title"`
	ParentID string `json:"parent_id"`
	SpaceID  string `json:"space_id"`
	IsInline bool   `json:"is_inline"`
}

type UpdateDatabaseRequest struct {
	Title          *string `json:"title"`
	Description    *string `json:"description"`
	PrimaryFieldID *string `json:"primary_field_id"`
}

// FieldTypes is the Go-owned catalog of allowed field types (SCHEMA.md §6.5 / PROJECT-INFO.md §48).
var FieldTypes = map[string]bool{
	"text": true, "number": true, "boolean": true, "date": true, "datetime": true,
	"select": true, "multi_select": true, "person": true, "relation": true,
	"formula": true, "file": true, "url": true, "email": true, "status": true,
}

type Field struct {
	ID          string          `json:"id"`
	DatabaseID  string          `json:"database_id"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Options     json.RawMessage `json:"options"`
	IsPrimary   bool            `json:"is_primary"`
	IsRequired  bool            `json:"is_required"`
	IsUnique    bool            `json:"is_unique"`
	IsComputed  bool            `json:"is_computed"`
	IsHidden    bool            `json:"is_hidden"`
	Formula     *string         `json:"formula,omitempty"`
	Rank        string          `json:"rank"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateFieldRequest struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	Options     json.RawMessage `json:"options"`
	IsRequired  bool            `json:"is_required"`
	IsUnique    bool            `json:"is_unique"`
	IsHidden    bool            `json:"is_hidden"`
	Formula     string          `json:"formula"`
	Rank        string          `json:"rank"`
}

type UpdateFieldRequest struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Options     json.RawMessage `json:"options"`
	IsRequired  *bool           `json:"is_required"`
	IsUnique    *bool           `json:"is_unique"`
	IsHidden    *bool           `json:"is_hidden"`
	Formula     *string         `json:"formula"`
	Rank        *string         `json:"rank"`
}

type FieldOption struct {
	ID         string     `json:"id"`
	FieldID    string     `json:"field_id"`
	Key        string     `json:"key"`
	Label      string     `json:"label"`
	Color      *string    `json:"color,omitempty"`
	Rank       string     `json:"rank"`
	DisabledAt *time.Time `json:"disabled_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateFieldOptionRequest struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Color string `json:"color"`
	Rank  string `json:"rank"`
}

type UpdateFieldOptionRequest struct {
	Label    *string `json:"label"`
	Color    *string `json:"color"`
	Rank     *string `json:"rank"`
	Disabled *bool   `json:"disabled"`
}
