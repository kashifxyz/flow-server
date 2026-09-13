package view

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/view/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/view/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/databases/{databaseID}/views", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/databases/{databaseID}/views", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/views/{viewID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/views/{viewID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/views/{viewID}", auth.RequireUser(h.Delete))
}
