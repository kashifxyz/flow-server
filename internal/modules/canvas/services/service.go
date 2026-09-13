package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/canvas/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	nodeTypeCanvas    = "canvas"
	nodeTypeObject    = "object"
	nodeTypeConnector = "connector"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Canvas, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	nodes, err := graph.ListByWorkspaceType(ctx, s.DB, workspaceID, nodeTypeCanvas)
	if err != nil {
		return nil, err
	}
	out := make([]models.Canvas, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, models.CanvasFromGraph(n))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateCanvasRequest) (models.Canvas, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Canvas{}, err
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "Untitled canvas"
	}
	insert := graph.Insert{WorkspaceID: workspaceID, Type: nodeTypeCanvas, Title: title, CreatedBy: userID}
	if in.SpaceID != "" {
		id, err := uuid.Parse(in.SpaceID)
		if err != nil {
			return models.Canvas{}, httperr.ErrInvalid
		}
		insert.SpaceID = &id
	}
	var n graph.Node
	err := withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		n, err = graph.Create(ctx, tx, insert)
		return err
	})
	if err != nil {
		return models.Canvas{}, err
	}
	return models.CanvasFromGraph(n), nil
}

func (s *Service) Get(ctx context.Context, canvasID, userID uuid.UUID) (models.Canvas, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID); err != nil {
		return models.Canvas{}, err
	}
	n, err := graph.GetOfType(ctx, s.DB, canvasID, nodeTypeCanvas)
	if err != nil {
		return models.Canvas{}, err
	}
	return models.CanvasFromGraph(n), nil
}

func (s *Service) Update(ctx context.Context, canvasID, userID uuid.UUID, in models.UpdateCanvasRequest) (models.Canvas, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID); err != nil {
		return models.Canvas{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, canvasID, nodeTypeCanvas); err != nil {
		return models.Canvas{}, err
	}
	n, err := graph.Update(ctx, s.DB, canvasID, userID, graph.Patch{Title: in.Title, Description: in.Description})
	if err != nil {
		return models.Canvas{}, err
	}
	return models.CanvasFromGraph(n), nil
}

func (s *Service) Delete(ctx context.Context, canvasID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, canvasID, nodeTypeCanvas); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, canvasID, userID, 0)
	})
}

func (s *Service) ListObjects(ctx context.Context, canvasID, userID uuid.UUID) ([]models.Object, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT o.id, o.canvas_id, o.object_kind, n.title, o.x, o.y, o.width, o.height, o.rotation, o.z_index,
		       o.locked, o.visible, o.opacity, o.style, n.created_at, n.updated_at
		FROM canvas_objects o JOIN nodes n ON n.id = o.id
		WHERE o.canvas_id = $1 AND n.deleted_at IS NULL
		ORDER BY o.z_index
	`, canvasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Object, 0)
	for rows.Next() {
		o, err := scanObjectWithTitle(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Service) CreateObject(ctx context.Context, canvasID, userID uuid.UUID, in models.CreateObjectRequest) (models.Object, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID)
	if err != nil {
		return models.Object{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, canvasID, nodeTypeCanvas); err != nil {
		return models.Object{}, err
	}
	kind := in.ObjectKind
	if kind == "" {
		kind = "shape"
	}
	if !models.ObjectKinds[kind] {
		return models.Object{}, httperr.ErrInvalid
	}
	style := in.Style
	if len(style) == 0 {
		style = []byte(`{}`)
	}
	var o models.Object
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			ParentID:    &canvasID,
			Type:        nodeTypeObject,
			Title:       in.Title,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO canvas_objects (id, workspace_id, canvas_id, object_kind, x, y, width, height, rotation, z_index, style)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, n.ID, workspaceID, canvasID, kind, in.X, in.Y, in.Width, in.Height, in.Rotation, in.ZIndex, style); err != nil {
			return err
		}
		o = models.Object{
			ID: n.ID.String(), CanvasID: canvasID.String(), Kind: kind, Title: in.Title,
			X: in.X, Y: in.Y, Width: in.Width, Height: in.Height, Rotation: in.Rotation, ZIndex: in.ZIndex,
			Visible: true, Opacity: 1, Style: style, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
		}
		return nil
	})
	return o, err
}

func (s *Service) UpdateObject(ctx context.Context, objectID, userID uuid.UUID, in models.UpdateObjectRequest) (models.Object, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, objectID); err != nil {
		return models.Object{}, err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE canvas_objects SET
			x = COALESCE($2, x), y = COALESCE($3, y),
			width = COALESCE($4, width), height = COALESCE($5, height),
			rotation = COALESCE($6, rotation), z_index = COALESCE($7, z_index),
			locked = COALESCE($8, locked), visible = COALESCE($9, visible),
			opacity = COALESCE($10, opacity), style = COALESCE($11, style)
		WHERE id = $1
	`, objectID, in.X, in.Y, in.Width, in.Height, in.Rotation, in.ZIndex, in.Locked, in.Visible, in.Opacity, in.Style)
	if err != nil {
		return models.Object{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Object{}, httperr.ErrNotFound
	}
	if in.Title != nil {
		if _, err := s.DB.Exec(ctx, `UPDATE nodes SET title = $2, updated_at = now() WHERE id = $1`, objectID, *in.Title); err != nil {
			return models.Object{}, err
		}
	}
	return s.getObject(ctx, objectID)
}

func (s *Service) DeleteObject(ctx context.Context, objectID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, objectID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, objectID, nodeTypeObject); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, objectID, userID, 0)
	})
}

const connectorCols = `c.id, c.canvas_id, c.start_object_id, c.end_object_id, c.shape, c.start_snap_to, c.end_snap_to, c.captions, c.style, c.created_at, c.updated_at`

func (s *Service) ListConnectors(ctx context.Context, canvasID, userID uuid.UUID) ([]models.Connector, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+connectorCols+` FROM canvas_connectors c WHERE c.canvas_id = $1 ORDER BY c.created_at`, canvasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Connector, 0)
	for rows.Next() {
		c, err := scanConnector(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Service) CreateConnector(ctx context.Context, canvasID, userID uuid.UUID, in models.CreateConnectorRequest) (models.Connector, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID)
	if err != nil {
		return models.Connector{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, canvasID, nodeTypeCanvas); err != nil {
		return models.Connector{}, err
	}
	startID, err := uuid.Parse(in.StartObjectID)
	if err != nil {
		return models.Connector{}, httperr.ErrInvalid
	}
	endID, err := uuid.Parse(in.EndObjectID)
	if err != nil {
		return models.Connector{}, httperr.ErrInvalid
	}
	shape := in.Shape
	if shape == "" {
		shape = "straight"
	}
	startSnap := in.StartSnapTo
	if startSnap == "" {
		startSnap = "auto"
	}
	endSnap := in.EndSnapTo
	if endSnap == "" {
		endSnap = "auto"
	}
	captions := in.Captions
	if len(captions) == 0 {
		captions = []byte(`[]`)
	}
	style := in.Style
	if len(style) == 0 {
		style = []byte(`{}`)
	}
	var c models.Connector
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		n, err := graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			ParentID:    &canvasID,
			Type:        nodeTypeConnector,
			CreatedBy:   userID,
		})
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO canvas_connectors (id, workspace_id, canvas_id, start_object_id, end_object_id, shape, start_snap_to, end_snap_to, captions, style)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`, n.ID, workspaceID, canvasID, startID, endID, shape, startSnap, endSnap, captions, style); err != nil {
			return err
		}
		c = models.Connector{
			ID: n.ID.String(), CanvasID: canvasID.String(), StartObjectID: startID.String(), EndObjectID: endID.String(),
			Shape: shape, StartSnapTo: startSnap, EndSnapTo: endSnap, Captions: captions, Style: style,
			CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt,
		}
		return nil
	})
	return c, err
}

func (s *Service) UpdateConnector(ctx context.Context, connectorID, userID uuid.UUID, in models.UpdateConnectorRequest) (models.Connector, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, connectorID); err != nil {
		return models.Connector{}, err
	}
	tag, err := s.DB.Exec(ctx, `
		UPDATE canvas_connectors SET
			shape = COALESCE($2, shape),
			start_snap_to = COALESCE($3, start_snap_to),
			end_snap_to = COALESCE($4, end_snap_to),
			captions = COALESCE($5, captions),
			style = COALESCE($6, style),
			updated_at = now()
		WHERE id = $1
	`, connectorID, in.Shape, in.StartSnapTo, in.EndSnapTo, in.Captions, in.Style)
	if err != nil {
		return models.Connector{}, err
	}
	if tag.RowsAffected() == 0 {
		return models.Connector{}, httperr.ErrNotFound
	}
	return s.getConnector(ctx, connectorID)
}

func (s *Service) DeleteConnector(ctx context.Context, connectorID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, connectorID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, connectorID, nodeTypeConnector); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, connectorID, userID, 0)
	})
}

func (s *Service) PutCamera(ctx context.Context, canvasID, userID uuid.UUID, in models.PutCameraRequest) (models.Camera, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID)
	if err != nil {
		return models.Camera{}, err
	}
	var cam models.Camera
	err = s.DB.QueryRow(ctx, `
		INSERT INTO canvas_cameras (user_id, canvas_id, workspace_id, x, y, zoom)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, canvas_id) DO UPDATE SET x = EXCLUDED.x, y = EXCLUDED.y, zoom = EXCLUDED.zoom, updated_at = now()
		RETURNING x, y, zoom, updated_at
	`, userID, canvasID, workspaceID, in.X, in.Y, in.Zoom).Scan(&cam.X, &cam.Y, &cam.Zoom, &cam.UpdatedAt)
	return cam, err
}

func (s *Service) ListTags(ctx context.Context, canvasID, userID uuid.UUID) ([]models.Tag, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `
		SELECT id, canvas_id, name, color, created_at FROM canvas_tags WHERE canvas_id = $1 ORDER BY name
	`, canvasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Tag, 0)
	for rows.Next() {
		var t models.Tag
		var id, cid uuid.UUID
		if err := rows.Scan(&id, &cid, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.ID = id.String()
		t.CanvasID = cid.String()
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) CreateTag(ctx context.Context, canvasID, userID uuid.UUID, in models.CreateTagRequest) (models.Tag, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, canvasID)
	if err != nil {
		return models.Tag{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return models.Tag{}, httperr.ErrInvalid
	}
	var t models.Tag
	var id, cid uuid.UUID
	err = s.DB.QueryRow(ctx, `
		INSERT INTO canvas_tags (workspace_id, canvas_id, name, color)
		VALUES ($1, $2, $3, $4)
		RETURNING id, canvas_id, name, color, created_at
	`, workspaceID, canvasID, name, nullIfEmpty(in.Color)).Scan(&id, &cid, &t.Name, &t.Color, &t.CreatedAt)
	t.ID = id.String()
	t.CanvasID = cid.String()
	return t, err
}

func (s *Service) getObject(ctx context.Context, objectID uuid.UUID) (models.Object, error) {
	o, err := scanObjectWithTitle(s.DB.QueryRow(ctx, `
		SELECT o.id, o.canvas_id, o.object_kind, n.title, o.x, o.y, o.width, o.height, o.rotation, o.z_index,
		       o.locked, o.visible, o.opacity, o.style, n.created_at, n.updated_at
		FROM canvas_objects o JOIN nodes n ON n.id = o.id
		WHERE o.id = $1
	`, objectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Object{}, httperr.ErrNotFound
	}
	return o, err
}

func (s *Service) getConnector(ctx context.Context, connectorID uuid.UUID) (models.Connector, error) {
	c, err := scanConnector(s.DB.QueryRow(ctx, `SELECT `+connectorCols+` FROM canvas_connectors c WHERE c.id = $1`, connectorID))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Connector{}, httperr.ErrNotFound
	}
	return c, err
}

func scanObjectWithTitle(row pgx.Row) (models.Object, error) {
	var o models.Object
	var id, canvasID uuid.UUID
	err := row.Scan(&id, &canvasID, &o.Kind, &o.Title, &o.X, &o.Y, &o.Width, &o.Height, &o.Rotation, &o.ZIndex,
		&o.Locked, &o.Visible, &o.Opacity, &o.Style, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return models.Object{}, err
	}
	o.ID = id.String()
	o.CanvasID = canvasID.String()
	return o, nil
}

func scanConnector(row pgx.Row) (models.Connector, error) {
	var c models.Connector
	var id, canvasID, startID, endID uuid.UUID
	err := row.Scan(&id, &canvasID, &startID, &endID, &c.Shape, &c.StartSnapTo, &c.EndSnapTo, &c.Captions, &c.Style, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return models.Connector{}, err
	}
	c.ID = id.String()
	c.CanvasID = canvasID.String()
	c.StartObjectID = startID.String()
	c.EndObjectID = endID.String()
	return c, nil
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
