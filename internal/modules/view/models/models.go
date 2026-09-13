package models

import (
	"encoding/json"
	"time"
)

var ViewTypes = map[string]bool{
	"table": true, "board": true, "calendar": true, "timeline": true,
	"list": true, "gallery": true, "map": true,
}

type View struct {
	ID              string          `json:"id"`
	DatabaseID      string          `json:"database_id"`
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	Filter          json.RawMessage `json:"filter"`
	Sorts           json.RawMessage `json:"sorts"`
	Groups          json.RawMessage `json:"groups"`
	VisibleFieldIDs []string        `json:"visible_field_ids"`
	FrozenFieldIDs  []string        `json:"frozen_field_ids"`
	ColumnWidths    json.RawMessage `json:"column_widths"`
	IsDefault       bool            `json:"is_default"`
	IsPersonal      bool            `json:"is_personal"`
	IsLocked        bool            `json:"is_locked"`
	RowHeight       *string         `json:"row_height,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type CreateViewRequest struct {
	Type       string `json:"type"`
	Name       string `json:"name"`
	IsPersonal bool   `json:"is_personal"`
}

type UpdateViewRequest struct {
	Name            *string         `json:"name"`
	Filter          json.RawMessage `json:"filter"`
	Sorts           json.RawMessage `json:"sorts"`
	Groups          json.RawMessage `json:"groups"`
	VisibleFieldIDs []string        `json:"visible_field_ids"`
	FrozenFieldIDs  []string        `json:"frozen_field_ids"`
	ColumnWidths    json.RawMessage `json:"column_widths"`
	IsLocked        *bool           `json:"is_locked"`
	RowHeight       *string         `json:"row_height"`
}
