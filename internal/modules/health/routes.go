package health

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/modules/health/handlers"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Module struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
	Log   zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(m.DB, m.Redis, m.Log)
	mux.HandleFunc("GET /healthz", h.Live)
	mux.HandleFunc("GET /readyz", h.Ready)
}
