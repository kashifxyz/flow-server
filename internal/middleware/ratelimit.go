package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// RateLimit throttles requests to next by client IP using a fixed-window
// counter in Redis, keyed by name so different endpoints get independent
// budgets. It exists for anti-automation-sensitive, pre-authentication
// endpoints (register, login, forgot-password, resend-verification) that the
// per-account lockout in auth.Service.Login does not cover on its own.
//
// If rdb is nil (Redis disabled) requests pass through unthrottled — there is
// nothing to count against — matching how the rest of the codebase treats an
// optional Redis dependency (see internal/cache, internal/modules/health).
func RateLimit(rdb *redis.Client, log zerolog.Logger, name string, limit int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if rdb == nil {
				next(w, r)
				return
			}
			ip := auth.MetaFromRequest(r).IP
			if ip == "" {
				next(w, r)
				return
			}
			ctx := r.Context()
			key := "ratelimit:" + name + ":" + ip
			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				log.Error().Err(err).Str("key", key).Msg("rate limit check failed; allowing request")
				next(w, r)
				return
			}
			if count == 1 {
				if err := rdb.Expire(ctx, key, window).Err(); err != nil {
					log.Error().Err(err).Str("key", key).Msg("rate limit expire failed")
				}
			}
			if count > int64(limit) {
				if ttl, err := rdb.TTL(ctx, key).Result(); err == nil && ttl > 0 {
					w.Header().Set("Retry-After", strconv.Itoa(int(ttl.Seconds())))
				}
				utils.JSONError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests. Try again later.")
				return
			}
			next(w, r)
		}
	}
}
