package block

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/block-flavours", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/pages/{pageID}/blocks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/pages/{pageID}/blocks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/pages/{pageID}/blocks/reorder", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/blocks/{blockID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/blocks/{blockID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/blocks/{blockID}", auth.RequireUser(utils.NotImplemented))
}
