package workspace

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/usage", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/members", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/members", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}/members/{userID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/members/{userID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/invites", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/invites", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/invites/{inviteID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/invites/{token}/accept", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/roles", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/roles", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}/roles/{roleID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/roles/{roleID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/groups", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/groups", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}/groups/{groupID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/groups/{groupID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/groups/{groupID}/members", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/groups/{groupID}/members", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/groups/{groupID}/members/{userID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/domains", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/domains", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/domains/{domainID}", auth.RequireUser(utils.NotImplemented))
}
