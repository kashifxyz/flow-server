package page

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/pages", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/pages", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/pages/{pageID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/pages/{pageID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/pages/{pageID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/pages/{pageID}/duplicate", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/pages/{pageID}/move", auth.RequireUser(utils.NotImplemented))
}
