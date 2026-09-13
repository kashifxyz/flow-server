package automation

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/automation/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/automation/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/automations", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/automations", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/automations/{automationID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/automations/{automationID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/automations/{automationID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("POST /api/v1/automations/{automationID}/enable", auth.RequireUser(h.Enable))
	mux.HandleFunc("POST /api/v1/automations/{automationID}/disable", auth.RequireUser(h.Disable))
	mux.HandleFunc("POST /api/v1/automations/{automationID}/run", auth.RequireUser(h.Run))
	mux.HandleFunc("GET /api/v1/automations/{automationID}/runs", auth.RequireUser(h.ListRuns))
}
