package users

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/users/services"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/rs/zerolog"
)

type Module struct {
	Auth          *auth.Service
	SecureCookies bool
	Log           zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.Auth), m.SecureCookies, m.Log)
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", auth.RequireUser(h.Logout))
	mux.HandleFunc("GET /api/v1/auth/me", auth.RequireUser(h.Me))
	mux.HandleFunc("POST /api/v1/auth/verify-email", h.VerifyEmail)
	mux.HandleFunc("POST /api/v1/auth/resend-verification", h.ResendVerification)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", h.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/reset-password", h.ResetPassword)
	mux.HandleFunc("PATCH /api/v1/users/me", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/users/me/preferences", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/users/me/preferences", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/users/me/sessions", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/users/me/sessions/{sessionID}", auth.RequireUser(utils.NotImplemented))
}
