package handlers

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/database/models"
	"github.com/kashifxyz/flow-server/internal/modules/database/services"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.List(r.Context(), wsID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateDatabaseRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Create(r.Context(), wsID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	dbID, err := utils.PathID(r, "databaseID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Get(r.Context(), dbID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	dbID, err := utils.PathID(r, "databaseID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateDatabaseRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Update(r.Context(), dbID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	dbID, err := utils.PathID(r, "databaseID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Delete(r.Context(), dbID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListFields(w http.ResponseWriter, r *http.Request) {
	dbID, err := utils.PathID(r, "databaseID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListFields(r.Context(), dbID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateField(w http.ResponseWriter, r *http.Request) {
	dbID, err := utils.PathID(r, "databaseID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateFieldRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateField(r.Context(), dbID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) UpdateField(w http.ResponseWriter, r *http.Request) {
	fieldID, err := utils.PathID(r, "fieldID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateFieldRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateField(r.Context(), fieldID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) DeleteField(w http.ResponseWriter, r *http.Request) {
	fieldID, err := utils.PathID(r, "fieldID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.DeleteField(r.Context(), fieldID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListFieldOptions(w http.ResponseWriter, r *http.Request) {
	fieldID, err := utils.PathID(r, "fieldID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListFieldOptions(r.Context(), fieldID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateFieldOption(w http.ResponseWriter, r *http.Request) {
	fieldID, err := utils.PathID(r, "fieldID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateFieldOptionRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateFieldOption(r.Context(), fieldID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) UpdateFieldOption(w http.ResponseWriter, r *http.Request) {
	optionID, err := utils.PathID(r, "optionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateFieldOptionRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateFieldOption(r.Context(), optionID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) DeleteFieldOption(w http.ResponseWriter, r *http.Request) {
	optionID, err := utils.PathID(r, "optionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.DeleteFieldOption(r.Context(), optionID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
