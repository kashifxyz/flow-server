package handlers

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration/models"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration/services"
	"github.com/kashifxyz/flow-server/internal/realtime"
	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/kashifxyz/flow-server/internal/utils/httperr"
	"github.com/rs/zerolog"
)

type Handler struct {
	Svc     *services.Service
	Hub     *realtime.Hub
	Origins []string
	Log     zerolog.Logger
}

func New(svc *services.Service, hub *realtime.Hub, origins []string, log zerolog.Logger) *Handler {
	return &Handler{Svc: svc, Hub: hub, Origins: origins, Log: log}
}

func actor(r *http.Request) (u auth.User) {
	u, _ = auth.UserFrom(r.Context())
	return
}

func (h *Handler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListByWorkspace(r.Context(), wsID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListByNode(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListByNode(r.Context(), nodeID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.PutPresenceRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Put(r.Context(), nodeID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Delete(r.Context(), nodeID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
