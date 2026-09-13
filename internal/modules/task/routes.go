package task

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/task/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/task/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/projects/{projectID}/tasks", auth.RequireUser(h.ListByProject))
	mux.HandleFunc("POST /api/v1/projects/{projectID}/tasks", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/tasks", auth.RequireUser(h.ListByWorkspace))
	mux.HandleFunc("GET /api/v1/tasks/{taskID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/tasks/{taskID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/tasks/{taskID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("POST /api/v1/tasks/{taskID}/move", auth.RequireUser(h.Move))
}
