package record

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/record/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/record/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/databases/{databaseID}/records", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/databases/{databaseID}/records", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/records/{recordID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/records/{recordID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/records/{recordID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("PUT /api/v1/records/{recordID}/fields/{fieldID}", auth.RequireUser(h.SetFieldValue))
	mux.HandleFunc("GET /api/v1/records/{recordID}/revisions", auth.RequireUser(h.ListRevisions))
}
