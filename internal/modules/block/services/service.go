package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/access"
	"github.com/kashifxyz/flow-server/internal/graph"
	"github.com/kashifxyz/flow-server/internal/modules/block/models"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
)

const (
	pageType  = "page"
	blockType = "block"
)

type Service struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{DB: db}
}

func (s *Service) ListFlavours(ctx context.Context) ([]models.Flavour, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT flavour, version, node_type, schema, deprecated_at
		FROM block_flavours WHERE deprecated_at IS NULL
		ORDER BY flavour, version DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]models.Flavour, 0)
	for rows.Next() {
		var f models.Flavour
		if err := rows.Scan(&f.Name, &f.Version, &f.NodeType, &f.Schema, &f.DeprecatedAt); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func (s *Service) ListByPage(ctx context.Context, pageID, userID uuid.UUID) ([]models.Block, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return nil, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, pageID, pageType); err != nil {
		return nil, err
	}
	nodes, err := graph.ListChildrenOfType(ctx, s.DB, pageID, blockType)
	if err != nil {
		return nil, err
	}
	out := make([]models.Block, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, models.FromGraph(n))
	}
	return out, nil
}

func (s *Service) Create(ctx context.Context, pageID, userID uuid.UUID, in models.CreateRequest) (models.Block, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return models.Block{}, err
	}
	page, err := graph.GetOfType(ctx, s.DB, pageID, pageType)
	if err != nil {
		return models.Block{}, err
	}
	if in.Flavour == "" {
		return models.Block{}, httperr.ErrInvalid
	}
	flavour := in.Flavour
	rank := in.Rank
	if rank == "" {
		rank = "0"
	}
	var n graph.Node
	err = withTx(ctx, s.DB, func(tx pgx.Tx) error {
		var err error
		n, err = graph.Create(ctx, tx, graph.Insert{
			WorkspaceID: page.WorkspaceID,
			ParentID:    &pageID,
			Type:        blockType,
			Flavour:     &flavour,
			Rank:        rank,
			Props:       in.Props,
			CreatedBy:   userID,
		})
		return err
	})
	if err != nil {
		return models.Block{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Get(ctx context.Context, blockID, userID uuid.UUID) (models.Block, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, blockID); err != nil {
		return models.Block{}, err
	}
	n, err := graph.GetOfType(ctx, s.DB, blockID, blockType)
	if err != nil {
		return models.Block{}, err
	}
	return models.FromGraph(n), nil
}

func (s *Service) Update(ctx context.Context, blockID, userID uuid.UUID, in models.UpdateRequest) (models.Block, error) {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, blockID); err != nil {
		return models.Block{}, err
	}
	if _, err := graph.GetOfType(ctx, s.DB, blockID, blockType); err != nil {
		return models.Block{}, err
	}
	n, err := graph.Update(ctx, s.DB, blockID, userID, graph.Patch{
		Props: in.Props,
		Rank:  in.Rank,
	})
	if err != nil {
		return models.Block{}, err
	}
	if in.Flavour != nil {
		if _, err := s.DB.Exec(ctx, `UPDATE nodes SET flavour = $2, updated_at = now() WHERE id = $1`, blockID, *in.Flavour); err != nil {
			return models.Block{}, err
		}
		n.Flavour = in.Flavour
	}
	return models.FromGraph(n), nil
}

func (s *Service) Delete(ctx context.Context, blockID, userID uuid.UUID) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, blockID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, blockID, blockType); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		return graph.SoftDelete(ctx, tx, blockID, userID, 0)
	})
}

func (s *Service) Reorder(ctx context.Context, pageID, userID uuid.UUID, in models.ReorderRequest) error {
	if _, err := access.RequireNodeMember(ctx, s.DB, userID, pageID); err != nil {
		return err
	}
	if _, err := graph.GetOfType(ctx, s.DB, pageID, pageType); err != nil {
		return err
	}
	return withTx(ctx, s.DB, func(tx pgx.Tx) error {
		for _, item := range in.Order {
			id, err := uuid.Parse(item.ID)
			if err != nil {
				return httperr.ErrInvalid
			}
			tag, err := tx.Exec(ctx, `
				UPDATE nodes SET rank = $3, updated_at = now()
				WHERE id = $1 AND parent_id = $2 AND type = 'block' AND deleted_at IS NULL
			`, id, pageID, item.Rank)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return httperr.ErrNotFound
			}
		}
		return nil
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
