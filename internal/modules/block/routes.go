package block

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/block/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/block/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/block-flavours", auth.RequireUser(h.ListFlavours))
	mux.HandleFunc("GET /api/v1/pages/{pageID}/blocks", auth.RequireUser(h.ListByPage))
	mux.HandleFunc("POST /api/v1/pages/{pageID}/blocks", auth.RequireUser(h.Create))
	mux.HandleFunc("POST /api/v1/pages/{pageID}/blocks/reorder", auth.RequireUser(h.Reorder))
	mux.HandleFunc("GET /api/v1/blocks/{blockID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/blocks/{blockID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/blocks/{blockID}", auth.RequireUser(h.Delete))
}
