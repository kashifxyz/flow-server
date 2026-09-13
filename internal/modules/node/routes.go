package node

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/node/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/node/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/nodes/{nodeID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/nodes/{nodeID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/move", auth.RequireUser(h.Move))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/archive", auth.RequireUser(h.Archive))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/restore", auth.RequireUser(h.Restore))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/trash", auth.RequireUser(h.ListTrash))
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/permissions", auth.RequireUser(h.ListPermissions))
	mux.HandleFunc("PUT /api/v1/nodes/{nodeID}/permissions", auth.RequireUser(h.SetPermissions))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/star", auth.RequireUser(h.Star))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/star", auth.RequireUser(h.Unstar))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/pin", auth.RequireUser(h.Pin))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/pin", auth.RequireUser(h.Unpin))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/watch", auth.RequireUser(h.Watch))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/watch", auth.RequireUser(h.Unwatch))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/visit", auth.RequireUser(h.Visit))
}
