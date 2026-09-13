package models

import (
	"encoding/json"
	"time"
)

type Automation struct {
	ID          string          `json:"id"`
	WorkspaceID string          `json:"workspace_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Enabled     bool            `json:"enabled"`
	TriggerType string          `json:"trigger_type"`
	Trigger     json.RawMessage `json:"trigger"`
	Conditions  json.RawMessage `json:"conditions"`
	LastRunAt   *time.Time      `json:"last_run_at,omitempty"`
	LastStatus  *string         `json:"last_status,omitempty"`
	RunCount    int64           `json:"run_count"`
	Version     int             `json:"version"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Step struct {
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config"`
	Rank   int             `json:"rank"`
}

type CreateAutomationRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TriggerType string          `json:"trigger_type"`
	Trigger     json.RawMessage `json:"trigger"`
	Conditions  json.RawMessage `json:"conditions"`
	Steps       []Step          `json:"steps"`
}

type UpdateAutomationRequest struct {
	Name        *string         `json:"name"`
	Description *string         `json:"description"`
	Trigger     json.RawMessage `json:"trigger"`
	Conditions  json.RawMessage `json:"conditions"`
}

type Run struct {
	ID             string          `json:"id"`
	AutomationID   string          `json:"automation_id"`
	Status         string          `json:"status"`
	TriggerPayload json.RawMessage `json:"trigger_payload"`
	Error          *string         `json:"error,omitempty"`
	StartedAt      *time.Time      `json:"started_at,omitempty"`
	FinishedAt     *time.Time      `json:"finished_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

type RunRequest struct {
	Payload json.RawMessage `json:"payload"`
}
