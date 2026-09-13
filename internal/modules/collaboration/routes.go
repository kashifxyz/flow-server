package collaboration

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration/services"
	"github.com/kashifxyz/flow-server/internal/realtime"
	"github.com/rs/zerolog"
)

type Module struct {
	DB      *pgxpool.Pool
	Hub     *realtime.Hub
	Origins []string
	Log     zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Hub, m.Origins, m.Log)

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/presence", auth.RequireUser(h.ListByWorkspace))
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/presence", auth.RequireUser(h.ListByNode))
	mux.HandleFunc("PUT /api/v1/nodes/{nodeID}/presence", auth.RequireUser(h.Put))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/presence", auth.RequireUser(h.Delete))
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/ws", auth.RequireUser(h.ServeWS))
}
