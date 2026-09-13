package webhook

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/webhooks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/webhooks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/webhooks/{webhookID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/webhooks/{webhookID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/webhooks/{webhookID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/webhooks/{webhookID}/deliveries", auth.RequireUser(utils.NotImplemented))
}
