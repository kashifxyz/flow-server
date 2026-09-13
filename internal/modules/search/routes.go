package search

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/search", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/saved-searches", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/saved-searches", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/saved-searches/{searchID}", auth.RequireUser(utils.NotImplemented))
}
