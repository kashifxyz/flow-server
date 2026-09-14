package auth

import (
	"net/http"
	"time"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/middleware"
	"github.com/kashifxyz/flow-server/internal/modules/auth/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/auth/services"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Module struct {
	Auth          *auth.Service
	Redis         *redis.Client
	SecureCookies bool
	Log           zerolog.Logger
}

// Rate limits for the endpoints with no other anti-automation control: the
// per-account lockout in auth.Service.Login only kicks in on repeated
// failures against one known email, so these guard registration, credential
// stuffing spread across many emails, and email-bombing via password-reset /
// verification-resend requests.
const (
	registerLimit            = 10
	registerWindow           = time.Hour
	loginLimit               = 20
	loginWindow              = 5 * time.Minute
	forgotPasswordLimit      = 5
	forgotPasswordWindow     = time.Hour
	resendVerificationLimit  = 5
	resendVerificationWindow = time.Hour
)

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.Auth), m.SecureCookies, m.Log)
	limit := func(name string, n int, window time.Duration, next http.HandlerFunc) http.HandlerFunc {
		return middleware.RateLimit(m.Redis, m.Log, name, n, window)(next)
	}
	mux.HandleFunc("POST /api/v1/auth/register", limit("register", registerLimit, registerWindow, h.Register))
	mux.HandleFunc("POST /api/v1/auth/login", limit("login", loginLimit, loginWindow, h.Login))
	mux.HandleFunc("POST /api/v1/auth/logout", auth.RequireUser(h.Logout))
	mux.HandleFunc("GET /api/v1/auth/me", auth.RequireUser(h.Me))
	mux.HandleFunc("POST /api/v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/resend-verification", limit("resend-verification", resendVerificationLimit, resendVerificationWindow, h.ResendVerification))
	mux.HandleFunc("POST /api/v1/auth/forgot-password", limit("forgot-password", forgotPasswordLimit, forgotPasswordWindow, h.ForgotPassword))
	mux.HandleFunc("POST /api/v1/auth/reset-password", h.ResetPassword)
}
