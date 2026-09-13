package search

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/search/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/search/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/search", auth.RequireUser(h.Search))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/saved-searches", auth.RequireUser(h.ListSavedSearches))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/saved-searches", auth.RequireUser(h.CreateSavedSearch))
	mux.HandleFunc("DELETE /api/v1/saved-searches/{searchID}", auth.RequireUser(h.DeleteSavedSearch))
}
