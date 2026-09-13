package users

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/users/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB   *pgxpool.Pool
	Auth *auth.Service
	Log  zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB, m.Auth), m.Log)
	mux.HandleFunc("PATCH /api/v1/users/me", auth.RequireUser(h.UpdateProfile))
	mux.HandleFunc("GET /api/v1/users/me/preferences", auth.RequireUser(h.GetPreferences))
	mux.HandleFunc("PATCH /api/v1/users/me/preferences", auth.RequireUser(h.UpdatePreferences))
	mux.HandleFunc("GET /api/v1/users/me/sessions", auth.RequireUser(h.ListSessions))
	mux.HandleFunc("DELETE /api/v1/users/me/sessions/{sessionID}", auth.RequireUser(h.RevokeSession))
}
