package space

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/spaces", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/spaces", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/spaces/{spaceID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/spaces/{spaceID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/spaces/{spaceID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/spaces/{spaceID}/members", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/spaces/{spaceID}/members", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/spaces/{spaceID}/members/{userID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/spaces/{spaceID}/members/{userID}", auth.RequireUser(utils.NotImplemented))
}
