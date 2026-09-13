package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/cache"
	"github.com/kashifxyz/flow-server/internal/database"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Handler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
	Log   zerolog.Logger
}

func New(db *pgxpool.Pool, rdb *redis.Client, log zerolog.Logger) *Handler {
	return &Handler{DB: db, Redis: rdb, Log: log}
}

func (h *Handler) Live(w http.ResponseWriter, _ *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	body := map[string]string{"status": "ready", "postgres": "ok"}
	if err := database.Ping(ctx, h.DB); err != nil {
		h.Log.Error().Err(err).Msg("readiness postgres failed")
		utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status":   "unavailable",
			"postgres": "error",
		})
		return
	}
	if h.Redis != nil {
		if err := cache.Ping(ctx, h.Redis); err != nil {
			h.Log.Error().Err(err).Msg("readiness redis failed")
			utils.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":   "unavailable",
				"postgres": "ok",
				"redis":    "error",
			})
			return
		}
		body["redis"] = "ok"
	} else {
		body["redis"] = "disabled"
	}
	utils.WriteJSON(w, http.StatusOK, body)
}
