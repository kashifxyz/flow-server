package database

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/databases", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/databases", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/databases/{databaseID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/databases/{databaseID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/databases/{databaseID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/databases/{databaseID}/fields", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/databases/{databaseID}/fields", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/fields/{fieldID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/fields/{fieldID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/fields/{fieldID}/options", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/fields/{fieldID}/options", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/field-options/{optionID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/field-options/{optionID}", auth.RequireUser(utils.NotImplemented))
}
