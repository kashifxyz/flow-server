package collaboration

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/presence", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/presence", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PUT /api/v1/nodes/{nodeID}/presence", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/presence", auth.RequireUser(utils.NotImplemented))
}
