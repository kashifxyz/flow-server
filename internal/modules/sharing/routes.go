package sharing

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/sharing/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/sharing/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/share-links", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/share-links", auth.RequireUser(h.Create))
	mux.HandleFunc("PATCH /api/v1/share-links/{linkID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/share-links/{linkID}", auth.RequireUser(h.Revoke))
	mux.HandleFunc("GET /api/v1/public/{token}", h.ResolvePublic)
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/publish", auth.RequireUser(h.Publish))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/publish", auth.RequireUser(h.Unpublish))
}
