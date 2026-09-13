package models

import (
	"encoding/json"
	"time"
)

type Webhook struct {
	ID            string     `json:"id"`
	WorkspaceID   string     `json:"workspace_id"`
	Name          string     `json:"name"`
	URL           string     `json:"url"`
	Events        []string   `json:"events"`
	Enabled       bool       `json:"enabled"`
	FailureCount  int        `json:"failure_count"`
	LastSuccessAt *time.Time `json:"last_success_at,omitempty"`
	LastFailureAt *time.Time `json:"last_failure_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CreateWebhookRequest struct {
	Name   string   `json:"name"`
	URL    string   `json:"url"`
	Events []string `json:"events"`
}

// CreateWebhookResponse includes the raw secret exactly once.
type CreateWebhookResponse struct {
	Webhook
	Secret string `json:"secret"`
}

type UpdateWebhookRequest struct {
	Name    *string  `json:"name"`
	URL     *string  `json:"url"`
	Events  []string `json:"events"`
	Enabled *bool    `json:"enabled"`
}

type Delivery struct {
	ID             string          `json:"id"`
	EventType      string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload"`
	Status         string          `json:"status"`
	AttemptCount   int             `json:"attempt_count"`
	ResponseStatus *int            `json:"response_status,omitempty"`
	DeliveredAt    *time.Time      `json:"delivered_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}
