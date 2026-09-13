package handlers

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/block/models"
	"github.com/kashifxyz/flow-server/internal/modules/block/services"
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

func (h *Handler) ListFlavours(w http.ResponseWriter, r *http.Request) {
	out, err := h.Svc.ListFlavours(r.Context())
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListByPage(w http.ResponseWriter, r *http.Request) {
	pageID, err := utils.PathID(r, "pageID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListByPage(r.Context(), pageID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	pageID, err := utils.PathID(r, "pageID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Create(r.Context(), pageID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	pageID, err := utils.PathID(r, "pageID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.ReorderRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Reorder(r.Context(), pageID, actor(r).ID, in); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	blockID, err := utils.PathID(r, "blockID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Get(r.Context(), blockID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	blockID, err := utils.PathID(r, "blockID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Update(r.Context(), blockID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	blockID, err := utils.PathID(r, "blockID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Delete(r.Context(), blockID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
