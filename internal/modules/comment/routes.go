package comment

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/discussions", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/discussions", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/discussions/{discussionID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/discussions/{discussionID}/resolve", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/discussions/{discussionID}/comments", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/discussions/{discussionID}/comments", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/comments/{commentID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/comments/{commentID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/comments/{commentID}/reactions", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/comments/{commentID}/reactions/{reactionID}", auth.RequireUser(utils.NotImplemented))
}
