package notification

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/notification/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/notification/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/notifications", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/notifications/read", auth.RequireUser(h.MarkAllRead))
	mux.HandleFunc("POST /api/v1/notifications/{notificationID}/read", auth.RequireUser(h.MarkRead))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/activity", auth.RequireUser(h.ListActivity))
}
