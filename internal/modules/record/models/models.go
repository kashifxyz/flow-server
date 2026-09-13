package models

import (
	"encoding/json"
	"time"
)

type Record struct {
	ID         string                     `json:"id"`
	DatabaseID string                     `json:"database_id"`
	Title      string                     `json:"title"`
	Rank       string                     `json:"rank"`
	Values     map[string]json.RawMessage `json:"values"`
	CreatedAt  time.Time                  `json:"created_at"`
	UpdatedAt  time.Time                  `json:"updated_at"`
}

type CreateRecordRequest struct {
	Title  string                     `json:"title"`
	Rank   string                     `json:"rank"`
	Values map[string]json.RawMessage `json:"values"`
}

type UpdateRecordRequest struct {
	Title *string `json:"title"`
	Rank  *string `json:"rank"`
}

type SetFieldValueRequest struct {
	Value json.RawMessage `json:"value"`
}

type FieldValueRevision struct {
	ID        string          `json:"id"`
	FieldID   string          `json:"field_id"`
	ActorID   *string         `json:"actor_id,omitempty"`
	Value     json.RawMessage `json:"value"`
	CreatedAt time.Time       `json:"created_at"`
}
