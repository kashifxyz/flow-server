package services

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/file/models"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
	"github.com/kashifxyz/flow-server/pkg/storage"
)

const (
	nodeTypeFile = "file"
	uploadTTL    = 15 * time.Minute
	downloadTTL  = 10 * time.Minute
	maxSizeBytes = 5 << 30 // 5 GiB sanity cap; real limits belong to storage/proxy config
)

type Service struct {
	DB     *pgxpool.Pool
	S3     *s3.Client
	Bucket string
}

func New(db *pgxpool.Pool, s3Client *s3.Client, bucket string) *Service {
	return &Service{DB: db, S3: s3Client, Bucket: bucket}
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

func sanitizeFilename(name string) string {
	name = path.Base(strings.TrimSpace(name))
	name = unsafeChars.ReplaceAllString(name, "_")
	name = strings.TrimLeft(name, ".")
	if name == "" {
		name = "file"
	}
	return name
}

func (s *Service) CreateUpload(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateUploadRequest) (models.CreateUploadResponse, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.CreateUploadResponse{}, err
	}
	if strings.TrimSpace(in.Filename) == "" || in.MimeType == "" || in.SizeBytes <= 0 || in.SizeBytes > maxSizeBytes {
		return models.CreateUploadResponse{}, httperr.ErrInvalid
	}
	sessionID, err := utils.NewID()
	if err != nil {
		return models.CreateUploadResponse{}, err
	}
	objectKey := fmt.Sprintf("workspaces/%s/uploads/%s/%s", workspaceID, sessionID, sanitizeFilename(in.Filename))
	uploadURL, err := storage.PresignPutURL(ctx, s.S3, s.Bucket, objectKey, in.MimeType, uploadTTL)
	if err != nil {
		return models.CreateUploadResponse{}, err
	}
	expiresAt := time.Now().Add(uploadTTL)
	_, err = s.DB.Exec(ctx, `
		INSERT INTO file_upload_sessions (id, workspace_id, created_by, upload_id, object_key, mime_type, expected_size_bytes, status, expires_at)
		VALUES ($1, $2, $3, $1, $4, $5, $6, 'open', $7)
	`, sessionID, workspaceID, userID, objectKey, in.MimeType, in.SizeBytes, expiresAt)
	if err != nil {
		return models.CreateUploadResponse{}, err
	}
	return models.CreateUploadResponse{
		UploadID:  sessionID.String(),
		ObjectKey: objectKey,
		UploadURL: uploadURL,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) CompleteUpload(ctx context.Context, sessionID, userID uuid.UUID, in models.CompleteUploadRequest) (models.File, error) {
	var workspaceID, createdBy uuid.UUID
	var objectKey, mimeType, status string
	var expiresAt time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT workspace_id, created_by, object_key, mime_type, status, expires_at
		FROM file_upload_sessions WHERE id = $1
	`, sessionID).Scan(&workspaceID, &createdBy, &objectKey, &mimeType, &status, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.File{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.File{}, err
	}
	if createdBy != userID {
		return models.File{}, httperr.ErrForbidden
	}
	if status != "open" || time.Now().After(expiresAt) {
		return models.File{}, httperr.ErrInvalid
	}
	// Never trust the client-declared expected_size_bytes for accounting: confirm
	// the object was actually uploaded and use S3's own record of its size.
	actualSize, err := storage.HeadObject(ctx, s.S3, s.Bucket, objectKey)
	if err != nil {
		return models.File{}, httperr.ErrInvalid
	}
	originalName := path.Base(objectKey)
	var f models.File
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			Type:        nodeTypeFile,
			Title:       originalName,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO files (id, workspace_id, uploader_id, bucket, object_key, original_name, mime_type, size_bytes, checksum, etag, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'ready')
		`, n.ID, workspaceID, userID, s.Bucket, objectKey, originalName, mimeType, actualSize,
			nullIfEmpty(in.Checksum), nullIfEmpty(in.Etag)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE file_upload_sessions SET status = 'completed', completed_at = now(), file_id = $2 WHERE id = $1
		`, sessionID, n.ID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			UPDATE workspaces SET storage_used_bytes = storage_used_bytes + $2, updated_at = now() WHERE id = $1
		`, workspaceID, actualSize); err != nil {
			return err
		}
		f = models.File{
			ID: n.ID.String(), WorkspaceID: workspaceID.String(), OriginalName: originalName,
			MimeType: mimeType, SizeBytes: actualSize, Status: "ready", CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
		}
		return nil
	})
	return f, err
}

func (s *Service) Get(ctx context.Context, fileID, userID uuid.UUID) (models.File, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, fileID); err != nil {
		return models.File{}, err
	}
	return s.getFile(ctx, fileID)
}

func (s *Service) Update(ctx context.Context, fileID, userID uuid.UUID, in models.UpdateFileRequest) (models.File, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, fileID); err != nil {
		return models.File{}, err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE files SET original_name = COALESCE($2, original_name), updated_at = now() WHERE id = $1
	`, fileID, in.OriginalName)
	if err != nil {
		return models.File{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.File{}, httperr.ErrNotFound
	}
	if in.OriginalName != nil {
		if _, err := s.DB.Exec(ctx, `UPDATE nodes SET title = $2, updated_at = now() WHERE id = $1`, fileID, *in.OriginalName); err != nil {
			return models.File{}, err
		}
	}
	return s.getFile(ctx, fileID)
}

func (s *Service) Delete(ctx context.Context, fileID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, fileID); err != nil {
		return err
	}
	var workspaceID uuid.UUID
	var objectKey string
	var sizeBytes int64
	err := s.DB.QueryRow(ctx, `SELECT workspace_id, object_key, size_bytes FROM files WHERE id = $1`, fileID).Scan(&workspaceID, &objectKey, &sizeBytes)
	if errors.Is(err, pgx.ErrNoRows) {
		return httperr.ErrNotFound
	}
	if err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		if err := graph.SoftDelete(ctx, tx, fileID, userID, 0); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO file_tombstones (workspace_id, object_key, reason, delete_after)
			VALUES ($1, $2, 'file_deleted', now() + interval '7 days')
			ON CONFLICT (object_key) DO NOTHING
		`, workspaceID, objectKey); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			UPDATE workspaces SET storage_used_bytes = GREATEST(storage_used_bytes - $2, 0), updated_at = now() WHERE id = $1
		`, workspaceID, sizeBytes)
		return err
	})
}

func (s *Service) DownloadURL(ctx context.Context, fileID, userID uuid.UUID) (string, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, fileID); err != nil {
		return "", err
	}
	var objectKey string
	err := s.DB.QueryRow(ctx, `SELECT object_key FROM files WHERE id = $1`, fileID).Scan(&objectKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", httperr.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return storage.PresignGetURL(ctx, s.S3, s.Bucket, objectKey, downloadTTL)
}

func (s *Service) getFile(ctx context.Context, fileID uuid.UUID) (models.File, error) {
	var f models.File
	var id uuid.UUID
	var workspaceID uuid.UUID
	err := s.DB.QueryRow(ctx, `
		SELECT f.id, f.workspace_id, f.original_name, f.mime_type, f.size_bytes, f.status,
		       f.width, f.height, f.duration_ms, f.page_count, f.created_at, f.updated_at
		FROM files f JOIN nodes n ON n.id = f.id
		WHERE f.id = $1 AND n.deleted_at IS NULL
	`, fileID).Scan(&id, &workspaceID, &f.OriginalName, &f.MimeType, &f.SizeBytes, &f.Status,
		&f.Width, &f.Height, &f.DurationMs, &f.PageCount, &f.CreatedAt, &f.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.File{}, httperr.ErrNotFound
	}
	if err != nil {
		return models.File{}, err
	}
	f.ID = id.String()
	f.WorkspaceID = workspaceID.String()
	return f, nil
}

func withTx(ctx context.Context, db *pgxpool.Pool, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
