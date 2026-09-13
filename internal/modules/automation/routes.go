package automation

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/automations", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/automations", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/automations/{automationID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/automations/{automationID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/automations/{automationID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/automations/{automationID}/enable", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/automations/{automationID}/disable", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/automations/{automationID}/run", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/automations/{automationID}/runs", auth.RequireUser(utils.NotImplemented))
}
