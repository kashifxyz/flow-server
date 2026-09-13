package handlers

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/workspace/models"
	"github.com/kashifxyz/flow-server/internal/modules/workspace/services"
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

func userID(r *http.Request) (u auth.User) {
	u, _ = auth.UserFrom(r.Context())
	return
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	out, err := h.Svc.List(r.Context(), userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var in models.CreateWorkspaceRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Create(r.Context(), userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Get(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateWorkspaceRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Update(r.Context(), wsID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.Delete(r.Context(), wsID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Usage(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.Usage(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListMembers(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.AddMemberRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.AddMember(r.Context(), wsID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	targetID, err := utils.PathID(r, "userID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateMemberRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateMember(r.Context(), wsID, targetID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	targetID, err := utils.PathID(r, "userID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.RemoveMember(r.Context(), wsID, targetID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListInvites(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListInvites(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateInviteRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, token, err := h.Svc.CreateInvite(r.Context(), wsID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, struct {
		models.Invite
		Token string `json:"token"`
	}{out, token})
}

func (h *Handler) RevokeInvite(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	inviteID, err := utils.PathID(r, "inviteID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.RevokeInvite(r.Context(), wsID, inviteID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if err := h.Svc.AcceptInvite(r.Context(), token, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.OKResponse{OK: true})
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListRoles(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateRoleRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateRole(r.Context(), wsID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	roleID, err := utils.PathID(r, "roleID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateRoleRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateRole(r.Context(), wsID, roleID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	roleID, err := utils.PathID(r, "roleID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.DeleteRole(r.Context(), wsID, roleID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListGroups(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListGroups(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateGroupRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateGroup(r.Context(), wsID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	groupID, err := utils.PathID(r, "groupID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateGroupRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateGroup(r.Context(), wsID, groupID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	groupID, err := utils.PathID(r, "groupID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.DeleteGroup(r.Context(), wsID, groupID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	groupID, err := utils.PathID(r, "groupID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListGroupMembers(r.Context(), wsID, groupID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	groupID, err := utils.PathID(r, "groupID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.AddGroupMemberRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.AddGroupMember(r.Context(), wsID, groupID, userID(r).ID, in); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	groupID, err := utils.PathID(r, "groupID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	targetID, err := utils.PathID(r, "userID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.RemoveGroupMember(r.Context(), wsID, groupID, targetID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListDomains(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListDomains(r.Context(), wsID, userID(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateDomain(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateDomainRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateDomain(r.Context(), wsID, userID(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	wsID, err := utils.PathID(r, "workspaceID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	domainID, err := utils.PathID(r, "domainID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.DeleteDomain(r.Context(), wsID, domainID, userID(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
