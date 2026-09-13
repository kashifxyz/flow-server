package node

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/nodes/{nodeID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/move", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/archive", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/restore", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/trash", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/permissions", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PUT /api/v1/nodes/{nodeID}/permissions", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/star", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/star", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/pin", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/pin", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/watch", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/watch", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/visit", auth.RequireUser(utils.NotImplemented))
}
