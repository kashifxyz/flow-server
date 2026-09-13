package database

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/database/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/database/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/databases", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/databases", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/databases/{databaseID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/databases/{databaseID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/databases/{databaseID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("GET /api/v1/databases/{databaseID}/fields", auth.RequireUser(h.ListFields))
	mux.HandleFunc("POST /api/v1/databases/{databaseID}/fields", auth.RequireUser(h.CreateField))
	mux.HandleFunc("PATCH /api/v1/fields/{fieldID}", auth.RequireUser(h.UpdateField))
	mux.HandleFunc("DELETE /api/v1/fields/{fieldID}", auth.RequireUser(h.DeleteField))
	mux.HandleFunc("GET /api/v1/fields/{fieldID}/options", auth.RequireUser(h.ListFieldOptions))
	mux.HandleFunc("POST /api/v1/fields/{fieldID}/options", auth.RequireUser(h.CreateFieldOption))
	mux.HandleFunc("PATCH /api/v1/field-options/{optionID}", auth.RequireUser(h.UpdateFieldOption))
	mux.HandleFunc("DELETE /api/v1/field-options/{optionID}", auth.RequireUser(h.DeleteFieldOption))
}
