package models

import "time"

type File struct {
	ID           string     `json:"id"`
	WorkspaceID  string     `json:"workspace_id"`
	OriginalName string     `json:"original_name"`
	MimeType     string     `json:"mime_type"`
	SizeBytes    int64      `json:"size_bytes"`
	Status       string     `json:"status"`
	Width        *int       `json:"width,omitempty"`
	Height       *int       `json:"height,omitempty"`
	DurationMs   *int       `json:"duration_ms,omitempty"`
	PageCount    *int       `json:"page_count,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
}

type CreateUploadRequest struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
}

type CreateUploadResponse struct {
	UploadID  string    `json:"upload_id"`
	ObjectKey string    `json:"object_key"`
	UploadURL string    `json:"upload_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CompleteUploadRequest struct {
	Checksum string `json:"checksum"`
	Etag     string `json:"etag"`
}

type UpdateFileRequest struct {
	OriginalName *string `json:"original_name"`
}
