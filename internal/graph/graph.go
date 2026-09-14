package graph

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

type Node struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID       `json:"workspace_id"`
	SpaceID     *uuid.UUID      `json:"space_id,omitempty"`
	ParentID    *uuid.UUID      `json:"parent_id,omitempty"`
	Type        string          `json:"type"`
	Subtype     *string         `json:"subtype,omitempty"`
	Flavour     *string         `json:"flavour,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        *string         `json:"icon,omitempty"`
	Rank        string          `json:"rank"`
	Props       json.RawMessage `json:"props"`
	Version     int64           `json:"version"`
	CreatedBy   *uuid.UUID      `json:"created_by,omitempty"`
	ArchivedAt  *time.Time      `json:"archived_at,omitempty"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Insert struct {
	WorkspaceID uuid.UUID
	SpaceID     *uuid.UUID
	ParentID    *uuid.UUID
	Type        string
	Subtype     *string
	Flavour     *string
	Title       string
	Description string
	Icon        *string
	Rank        string
	Props       json.RawMessage
	CreatedBy   uuid.UUID
}

type Patch struct {
	Title       *string
	Description *string
	Icon        *string
	Props       json.RawMessage
	Rank        *string
	SpaceID     *uuid.UUID
}

func Create(ctx context.Context, tx pgx.Tx, in Insert) (Node, error) {
	id, err := utils.NewID()
	if err != nil {
		return Node{}, err
	}
	if in.Rank == "" {
		in.Rank = "0"
	}
	props := in.Props
	if len(props) == 0 {
		props = json.RawMessage(`{}`)
	}
	var rootID uuid.UUID
	if in.ParentID != nil {
		rootID = *in.ParentID
		var parentRoot *uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT COALESCE(root_id, id) FROM nodes WHERE id = $1`, *in.ParentID).Scan(&parentRoot); err == nil && parentRoot != nil {
			rootID = *parentRoot
		}
	} else {
		rootID = id
	}
	n := Node{}
	err = tx.QueryRow(ctx, `
		INSERT INTO nodes (
			id, workspace_id, space_id, parent_id, root_id, created_by, last_edited_by,
			type, subtype, flavour, title, description, icon, rank, props
		) VALUES ($1,$2,$3,$4,$5,$6,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING id, workspace_id, space_id, parent_id, type, subtype, flavour, title, description, icon, rank, props, version,
		          created_by, archived_at, deleted_at, created_at, updated_at
	`, id, in.WorkspaceID, in.SpaceID, in.ParentID, rootID, in.CreatedBy, in.Type, in.Subtype, in.Flavour,
		in.Title, in.Description, in.Icon, in.Rank, props).Scan(
		&n.ID, &n.WorkspaceID, &n.SpaceID, &n.ParentID, &n.Type, &n.Subtype, &n.Flavour, &n.Title, &n.Description,
		&n.Icon, &n.Rank, &n.Props, &n.Version, &n.CreatedBy, &n.ArchivedAt, &n.DeletedAt, &n.CreatedAt, &n.UpdatedAt,
	)
	if err != nil {
		return Node{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO node_closure (ancestor_id, descendant_id, workspace_id, depth)
		VALUES ($1, $1, $2, 0)
	`, n.ID, n.WorkspaceID); err != nil {
		return Node{}, err
	}
	if in.ParentID != nil {
		if _, err := tx.Exec(ctx, `
			INSERT INTO node_closure (ancestor_id, descendant_id, workspace_id, depth)
			SELECT ancestor_id, $1, workspace_id, depth + 1
			FROM node_closure WHERE descendant_id = $2
		`, n.ID, *in.ParentID); err != nil {
			return Node{}, err
		}
		_, _ = tx.Exec(ctx, `UPDATE nodes SET child_count = child_count + 1, updated_at = now() WHERE id = $1`, *in.ParentID)
	}
	return n, nil
}

func Get(ctx context.Context, db *pgxpool.Pool, id uuid.UUID) (Node, error) {
	n, err := scan(db.QueryRow(ctx, selectSQL+` WHERE id = $1 AND purged_at IS NULL`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Node{}, httperr.ErrNotFound
	}
	return n, err
}

func GetOfType(ctx context.Context, db *pgxpool.Pool, id uuid.UUID, nodeType string) (Node, error) {
	n, err := Get(ctx, db, id)
	if err != nil {
		return Node{}, err
	}
	if n.Type != nodeType {
		return Node{}, httperr.ErrNotFound
	}
	return n, nil
}

func ListByWorkspaceType(ctx context.Context, db *pgxpool.Pool, workspaceID uuid.UUID, nodeType string) ([]Node, error) {
	rows, err := db.Query(ctx, selectSQL+`
		WHERE workspace_id = $1 AND type = $2 AND deleted_at IS NULL AND purged_at IS NULL
		ORDER BY rank, created_at
	`, workspaceID, nodeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

func ListChildrenOfType(ctx context.Context, db *pgxpool.Pool, parentID uuid.UUID, nodeType string) ([]Node, error) {
	rows, err := db.Query(ctx, selectSQL+`
		WHERE parent_id = $1 AND type = $2 AND deleted_at IS NULL AND purged_at IS NULL
		ORDER BY rank, created_at
	`, parentID, nodeType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collect(rows)
}

func Update(ctx context.Context, db *pgxpool.Pool, id, editor uuid.UUID, patch Patch) (Node, error) {
	n, err := scan(db.QueryRow(ctx, `
		UPDATE nodes SET
			title = COALESCE($2, title),
			description = COALESCE($3, description),
			icon = COALESCE($4, icon),
			props = COALESCE($5, props),
			rank = COALESCE($6, rank),
			space_id = COALESCE($7, space_id),
			last_edited_by = $8,
			last_edited_at = now(),
			version = version + 1,
			updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND purged_at IS NULL
		RETURNING id, workspace_id, space_id, parent_id, type, subtype, flavour, title, description, icon, rank, props, version,
		          created_by, archived_at, deleted_at, created_at, updated_at
	`, id, patch.Title, patch.Description, patch.Icon, patch.Props, patch.Rank, patch.SpaceID, editor))
	if errors.Is(err, pgx.ErrNoRows) {
		return Node{}, httperr.ErrNotFound
	}
	return n, err
}

func SoftDelete(ctx context.Context, tx pgx.Tx, id, userID uuid.UUID, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	tag, err := tx.Exec(ctx, `
		UPDATE nodes
		SET deleted_at = now(), deleted_by = $2, restore_until = now() + ($3 * interval '1 day'), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND purged_at IS NULL
	`, id, userID, retentionDays)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO trash_items (node_id, workspace_id, deleted_by, restore_until, original_parent_id, original_rank)
		SELECT id, workspace_id, $2, restore_until, parent_id, rank FROM nodes WHERE id = $1
		ON CONFLICT (node_id) DO UPDATE SET deleted_at = now(), deleted_by = EXCLUDED.deleted_by, restore_until = EXCLUDED.restore_until
	`, id, userID)
	return err
}

func Restore(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE nodes SET deleted_at = NULL, deleted_by = NULL, restore_until = NULL, updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL AND purged_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	_, err = tx.Exec(ctx, `DELETE FROM trash_items WHERE node_id = $1`, id)
	return err
}

func Archive(ctx context.Context, db *pgxpool.Pool, id, userID uuid.UUID) error {
	tag, err := db.Exec(ctx, `
		UPDATE nodes SET archived_at = now(), archived_by = $2, updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL AND purged_at IS NULL
	`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func Unarchive(ctx context.Context, db *pgxpool.Pool, id uuid.UUID) error {
	tag, err := db.Exec(ctx, `
		UPDATE nodes SET archived_at = NULL, archived_by = NULL, updated_at = now()
		WHERE id = $1 AND purged_at IS NULL
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	return nil
}

func Move(ctx context.Context, tx pgx.Tx, id uuid.UUID, parentID *uuid.UUID, rank string) error {
	if rank == "" {
		rank = "0"
	}
	_, err := tx.Exec(ctx, `DELETE FROM node_closure WHERE descendant_id = $1 AND ancestor_id <> descendant_id`, id)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `
		UPDATE nodes SET parent_id = $2, rank = $3, updated_at = now(), version = version + 1
		WHERE id = $1 AND deleted_at IS NULL AND purged_at IS NULL
	`, id, parentID, rank)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return httperr.ErrNotFound
	}
	if parentID != nil {
		_, err = tx.Exec(ctx, `
			INSERT INTO node_closure (ancestor_id, descendant_id, workspace_id, depth)
			SELECT ancestor_id, $1, workspace_id, depth + 1
			FROM node_closure WHERE descendant_id = $2
			ON CONFLICT DO NOTHING
		`, id, *parentID)
	}
	return err
}

func Duplicate(ctx context.Context, tx pgx.Tx, src Node, userID uuid.UUID) (Node, error) {
	title := src.Title
	if title != "" {
		title = title + " copy"
	}
	return Create(ctx, tx, Insert{
		WorkspaceID: src.WorkspaceID,
		SpaceID:     src.SpaceID,
		ParentID:    src.ParentID,
		Type:        src.Type,
		Subtype:     src.Subtype,
		Flavour:     src.Flavour,
		Title:       title,
		Description: src.Description,
		Icon:        src.Icon,
		Rank:        src.Rank,
		Props:       src.Props,
		CreatedBy:   userID,
	})
}

const selectSQL = `
	SELECT id, workspace_id, space_id, parent_id, type, subtype, flavour, title, description, icon, rank, props, version,
	       created_by, archived_at, deleted_at, created_at, updated_at
	FROM nodes
`

func scan(row pgx.Row) (Node, error) {
	var n Node
	err := row.Scan(
		&n.ID, &n.WorkspaceID, &n.SpaceID, &n.ParentID, &n.Type, &n.Subtype, &n.Flavour, &n.Title, &n.Description,
		&n.Icon, &n.Rank, &n.Props, &n.Version, &n.CreatedBy, &n.ArchivedAt, &n.DeletedAt, &n.CreatedAt, &n.UpdatedAt,
	)
	return n, err
}

func collect(rows pgx.Rows) ([]Node, error) {
	out := make([]Node, 0)
	for rows.Next() {
		n, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}
