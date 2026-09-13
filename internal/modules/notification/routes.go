package notification

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/notifications", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/notifications/read", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/notifications/{notificationID}/read", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/activity", auth.RequireUser(utils.NotImplemented))
}
