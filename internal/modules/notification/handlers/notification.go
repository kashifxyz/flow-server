package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/notification/services"
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

func optionalWorkspaceID(r *http.Request) *uuid.UUID {
	raw := r.URL.Query().Get("workspace_id")
	if raw == "" {
		return nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil
	}
	return &id
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	out, err := h.Svc.List(r.Context(), actor(r).ID, optionalWorkspaceID(r))
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	if err := h.Svc.MarkAllRead(r.Context(), actor(r).ID, optionalWorkspaceID(r)); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	notificationID, err := utils.PathID(r, "notificationID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.MarkRead(r.Context(), notificationID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListActivity(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListActivity(r.Context(), wsID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}
