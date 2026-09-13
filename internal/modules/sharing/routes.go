package sharing

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/share-links", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/share-links", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/share-links/{linkID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/share-links/{linkID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/public/{token}", utils.NotImplemented)
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/publish", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/nodes/{nodeID}/publish", auth.RequireUser(utils.NotImplemented))
}
