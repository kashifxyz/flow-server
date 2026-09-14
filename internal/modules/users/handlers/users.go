package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/models"
	"github.com/kashifxyz/flow-server/internal/modules/users/services"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
	"github.com/rs/zerolog"
)

type Handler struct {
	Svc *services.Service
	Log zerolog.Logger
}

func New(svc *services.Service, log zerolog.Logger) *Handler {
	return &Handler{Svc: svc, Log: log}
}

func actor(r *http.Request) (u auth.User) {
	u, _ = auth.UserFrom(r.Context())
	return
}

func currentSessionID(r *http.Request) uuid.UUID {
	sess, _ := auth.SessionFrom(r.Context())
	return sess.ID
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	out, err := h.Svc.GetProfile(r.Context(), actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var in models.ChangePasswordRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.ChangePassword(r.Context(), actor(r).ID, currentSessionID(r), in, auth.MetaFromRequest(r)); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.OKResponse{OK: true})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var in models.UpdateProfileRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateProfile(r.Context(), actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	out, err := h.Svc.GetPreferences(r.Context(), actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	var in models.UpdatePreferencesRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdatePreferences(r.Context(), actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	out, err := h.Svc.ListSessions(r.Context(), actor(r).ID, currentSessionID(r))
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := utils.PathID(r, "sessionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.RevokeSession(r.Context(), actor(r).ID, sessionID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
