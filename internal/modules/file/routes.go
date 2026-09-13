package file

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/files/uploads", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/files/uploads/{uploadID}/complete", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/files/{fileID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/files/{fileID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/files/{fileID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/files/{fileID}/download", auth.RequireUser(utils.NotImplemented))
}
