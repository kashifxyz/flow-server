package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/node/models"
	"github.com/kashifxyz/flow-server/internal/modules/node/services"
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Get(r.Context(), nodeID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Update(r.Context(), nodeID, actor(r).ID, in)
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

func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.MoveRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Move(r.Context(), nodeID, actor(r).ID, in); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Archive(r.Context(), nodeID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Restore(r.Context(), nodeID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTrash(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListTrash(r.Context(), wsID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListPermissions(r.Context(), nodeID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) SetPermissions(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.SetPermissionsRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.SetPermissions(r.Context(), nodeID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Star(w http.ResponseWriter, r *http.Request)    { h.engage(w, r, h.Svc.Star) }
func (h *Handler) Unstar(w http.ResponseWriter, r *http.Request)  { h.engage(w, r, h.Svc.Unstar) }
func (h *Handler) Pin(w http.ResponseWriter, r *http.Request)     { h.engage(w, r, h.Svc.Pin) }
func (h *Handler) Unpin(w http.ResponseWriter, r *http.Request)   { h.engage(w, r, h.Svc.Unpin) }
func (h *Handler) Watch(w http.ResponseWriter, r *http.Request)   { h.engage(w, r, h.Svc.Watch) }
func (h *Handler) Unwatch(w http.ResponseWriter, r *http.Request) { h.engage(w, r, h.Svc.Unwatch) }
func (h *Handler) Visit(w http.ResponseWriter, r *http.Request)   { h.engage(w, r, h.Svc.Visit) }

func (h *Handler) engage(w http.ResponseWriter, r *http.Request, fn func(context.Context, uuid.UUID, uuid.UUID) error) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := fn(r.Context(), nodeID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
