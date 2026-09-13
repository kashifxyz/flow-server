package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/page/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const nodeType = "page"

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) List(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Page, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	nodes, err := graph.ListByWorkspaceType(ctx, s.DB, workspaceID, nodeType)
	if err != nil {
		return nil, err
	}
	out := make([]models.Page, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, models.FromGraph(n))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, workspaceID, userID uuid.UUID, in models.CreateRequest) (models.Page, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return models.Page{}, err
	}
	insert := graph.Insert{
		WorkspaceID: workspaceID,
		Type:        nodeType,
		Title:       in.Title,
		Props:       in.Props,
		CreatedBy:   userID,
	}
	if in.Icon != "" {
		icon := in.Icon
		insert.Icon = &icon
	}
	if in.SpaceID != "" {
		id, err := uuid.Parse(in.SpaceID)
		if err != nil {
			return models.Page{}, httperr.ErrInvalid
		}
		insert.SpaceID = &id
	}
	if in.ParentID != "" {
		parentID, err := uuid.Parse(in.ParentID)
		if err != nil {
			return models.Page{}, httperr.ErrInvalid
		}
		if _, err := graph.GetOfType(ctx, s.DB, parentID, nodeType); err != nil {
			return models.Page{}, err
		}
		insert.ParentID = &parentID
	}
	var n graph.Node
	err := withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		n, err = graph.Create(ctx, tx, insert)
		return err
	})
	if err != nil {
		return models.Page{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Get(ctx context.Context, pageID, userID uuid.UUID) (models.Page, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return models.Page{}, err
	}
	n, err := graph.GetOfType(ctx, s.DB, pageID, nodeType)
	if err != nil {
		return models.Page{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Update(ctx context.Context, pageID, userID uuid.UUID, in models.UpdateRequest) (models.Page, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return models.Page{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, pageID, nodeType); err != nil {
		return models.Page{}, err
	}
	n, err := graph.Update(ctx, s.DB, pageID, userID, graph.Patch{
		Title:       in.Title,
		Description: in.Description,
		Icon:        in.Icon,
		Props:       in.Props,
	})
	if err != nil {
		return models.Page{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Delete(ctx context.Context, pageID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, pageID, nodeType); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, pageID, userID, 0)
	})
}

func (s *Service) Duplicate(ctx context.Context, pageID, userID uuid.UUID) (models.Page, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return models.Page{}, err
	}
	src, err := graph.GetOfType(ctx, s.DB, pageID, nodeType)
	if err != nil {
		return models.Page{}, err
	}
	var n graph.Node
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		n, err = graph.Duplicate(ctx, tx, src, userID)
		return err
	})
	if err != nil {
		return models.Page{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Move(ctx context.Context, pageID, userID uuid.UUID, in models.MoveRequest) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, pageID, nodeType); err != nil {
		return err
	}
	var parentID *uuid.UUID
	if in.ParentID != nil && *in.ParentID != "" {
		id, err := uuid.Parse(*in.ParentID)
		if err != nil {
			return httperr.ErrInvalid
		}
		parentID = &id
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.Move(ctx, tx, pageID, parentID, in.Rank)
	})
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
