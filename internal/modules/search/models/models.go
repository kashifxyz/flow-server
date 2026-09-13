package models

import (
	"encoding/json"
	"time"
)

type Result struct {
	NodeID string  `json:"node_id"`
	Type   string  `json:"type"`
	Title  string  `json:"title"`
	Rank   float64 `json:"rank"`
}

type SavedSearch struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Query     string          `json:"query"`
	Filters   json.RawMessage `json:"filters"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type CreateSavedSearchRequest struct {
	Name    string          `json:"name"`
	Query   string          `json:"query"`
	Filters json.RawMessage `json:"filters"`
}
