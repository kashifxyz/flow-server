package record

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/databases/{databaseID}/records", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/databases/{databaseID}/records", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/records/{recordID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/records/{recordID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/records/{recordID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PUT /api/v1/records/{recordID}/fields/{fieldID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/records/{recordID}/revisions", auth.RequireUser(utils.NotImplemented))
}
