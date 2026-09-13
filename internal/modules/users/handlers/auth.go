package handlers

import (
	"net/http"
	"strings"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/models"
	"github.com/kashifxyz/flow-server/internal/modules/users/services"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/rs/zerolog"
)

type Handler struct {
	Users         *services.Service
	SecureCookies bool
	Log           zerolog.Logger
}

func New(users *services.Service, secureCookies bool, log zerolog.Logger) *Handler {
	return &Handler{Users: users, SecureCookies: secureCookies, Log: log}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var in models.RegisterRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		writeError(w, h.Log, auth.ErrInvalidRequest)
		return
	}
	if err := h.Users.Register(r.Context(), in, auth.MetaFromRequest(r)); err != nil {
		writeError(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, models.OK{OK: true})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var in models.LoginRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		writeError(w, h.Log, auth.ErrInvalidRequest)
		return
	}
	issued, err := h.Users.Login(r.Context(), in, auth.MetaFromRequest(r))
	if err != nil {
		writeError(w, h.Log, err)
		return
	}
	auth.SetAuthCookies(w, issued.SessionToken, issued.CSRFToken, auth.CookieOpts{
		Secure: h.secure(r),
		MaxAge: issued.TTL,
	})
	utils.WriteJSON(w, http.StatusOK, models.UserFromAuth(issued.User))
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := auth.SessionFrom(r.Context())
	if err := h.Users.Logout(r.Context(), sess, auth.MetaFromRequest(r)); err != nil {
		writeError(w, h.Log, err)
		return
	}
	auth.ClearAuthCookies(w, h.secure(r))
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFrom(r.Context())
	utils.WriteJSON(w, http.StatusOK, models.UserFromAuth(user))
}

func (h *Handler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var in models.TokenRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		writeError(w, h.Log, auth.ErrInvalidRequest)
		return
	}
	if err := h.Users.VerifyEmail(r.Context(), strings.TrimSpace(in.Token), auth.MetaFromRequest(r)); err != nil {
		writeError(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.OK{OK: true})
}

func (h *Handler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var in models.EmailRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		writeError(w, h.Log, auth.ErrInvalidRequest)
		return
	}
	if err := h.Users.ResendVerification(r.Context(), in.Email, auth.MetaFromRequest(r)); err != nil {
		writeError(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var in models.EmailRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		writeError(w, h.Log, auth.ErrInvalidRequest)
		return
	}
	if err := h.Users.ForgotPassword(r.Context(), in.Email, auth.MetaFromRequest(r)); err != nil {
		writeError(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var in models.ResetPasswordRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		writeError(w, h.Log, auth.ErrInvalidRequest)
		return
	}
	if err := h.Users.ResetPassword(r.Context(), strings.TrimSpace(in.Token), in.Password, auth.MetaFromRequest(r)); err != nil {
		writeError(w, h.Log, err)
		return
	}
	auth.ClearAuthCookies(w, h.secure(r))
	utils.WriteJSON(w, http.StatusOK, models.OK{OK: true})
}

func (h *Handler) secure(r *http.Request) bool {
	if h.SecureCookies {
		return true
	}
	return r.TLS != nil
}
