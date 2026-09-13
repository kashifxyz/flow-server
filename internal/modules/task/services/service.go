package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/task/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	nodeTypeProject = "project"
	nodeTypeTask    = "task"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) ListByProject(ctx context.Context, projectID, userID uuid.UUID) ([]models.Task, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, projectID); err != nil {
		return nil, err
	}
	nodes, err := graph.ListChildrenOfType(ctx, s.DB, projectID, nodeTypeTask)
	if err != nil {
		return nil, err
	}
	return fromGraphList(nodes), nil
}

func (s *Service) ListByWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) ([]models.Task, error) {
	if err := access.RequireMember(ctx, s.DB, workspaceID, userID); err != nil {
		return nil, err
	}
	nodes, err := graph.ListByWorkspaceType(ctx, s.DB, workspaceID, nodeTypeTask)
	if err != nil {
		return nil, err
	}
	return fromGraphList(nodes), nil
}

func (s *Service) Create(ctx context.Context, projectID, userID uuid.UUID, in models.CreateRequest) (models.Task, error) {
	workspaceID, err := access.RequireNodeMember(ctx, s.DB, userID, projectID)
	if err != nil {
		return models.Task{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, projectID, nodeTypeProject); err != nil {
		return models.Task{}, err
	}
	if in.Title == "" {
		return models.Task{}, httperr.ErrInvalid
	}
	rank := in.Rank
	if rank == "" {
		rank = "0"
	}
	var n graph.Node
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		n, err = graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: workspaceID,
			ParentID:    &projectID,
			Type:        nodeTypeTask,
			Title:       in.Title,
			Description: in.Description,
			Rank:        rank,
			Props:       in.Props,
			CreatedBy:   userID,
		})
		return err
	})
	if err != nil {
		return models.Task{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Get(ctx context.Context, taskID, userID uuid.UUID) (models.Task, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, taskID); err != nil {
		return models.Task{}, err
	}
	n, err := graph.GetOfType(ctx, s.DB, taskID, nodeTypeTask)
	if err != nil {
		return models.Task{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Update(ctx context.Context, taskID, userID uuid.UUID, in models.UpdateRequest) (models.Task, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, taskID); err != nil {
		return models.Task{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, taskID, nodeTypeTask); err != nil {
		return models.Task{}, err
	}
	n, err := graph.Update(ctx, s.DB, taskID, userID, graph.Patch{
		Title:       in.Title,
		Description: in.Description,
		Props:       in.Props,
		Rank:        in.Rank,
	})
	if err != nil {
		return models.Task{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Delete(ctx context.Context, taskID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, taskID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, taskID, nodeTypeTask); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, taskID, userID, 0)
	})
}

func (s *Service) Move(ctx context.Context, taskID, userID uuid.UUID, in models.MoveRequest) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, taskID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, taskID, nodeTypeTask); err != nil {
		return err
	}
	var projectID *uuid.UUID
	if in.ProjectID != nil && *in.ProjectID != "" {
		id, err := uuid.Parse(*in.ProjectID)
		if err != nil {
			return httperr.ErrInvalid
		}
		if _, err := graph.GetOfType(ctx, s.DB, id, nodeTypeProject); err != nil {
			return err
		}
		projectID = &id
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.Move(ctx, tx, taskID, projectID, in.Rank)
	})
}

func fromGraphList(nodes []graph.Node) []models.Task {
	out := make([]models.Task, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, models.FromGraph(n))
	}
	return out
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
