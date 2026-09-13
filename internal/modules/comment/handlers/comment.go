package handlers

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/comment/models"
	"github.com/kashifxyz/flow-server/internal/modules/comment/services"
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

func (h *Handler) ListDiscussions(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListDiscussions(r.Context(), nodeID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateDiscussion(w http.ResponseWriter, r *http.Request) {
	nodeID, err := utils.PathID(r, "nodeID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateDiscussionRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateDiscussion(r.Context(), nodeID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) GetDiscussion(w http.ResponseWriter, r *http.Request) {
	discussionID, err := utils.PathID(r, "discussionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.GetDiscussion(r.Context(), discussionID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ResolveDiscussion(w http.ResponseWriter, r *http.Request) {
	discussionID, err := utils.PathID(r, "discussionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ResolveDiscussion(r.Context(), discussionID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	discussionID, err := utils.PathID(r, "discussionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.ListComments(r.Context(), discussionID, actor(r).ID)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	discussionID, err := utils.PathID(r, "discussionID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateCommentRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.CreateComment(r.Context(), discussionID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := utils.PathID(r, "commentID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.UpdateCommentRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.UpdateComment(r.Context(), commentID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := utils.PathID(r, "commentID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	if err := h.Svc.DeleteComment(r.Context(), commentID, actor(r).ID); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AddReaction(w http.ResponseWriter, r *http.Request) {
	commentID, err := utils.PathID(r, "commentID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	var in models.CreateReactionRequest
	if err := utils.DecodeJSON(w, r, &in); err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	out, err := h.Svc.AddReaction(r.Context(), commentID, actor(r).ID, in)
	if err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, out)
}

func (h *Handler) RemoveReaction(w http.ResponseWriter, r *http.Request) {
	commentID, err := utils.PathID(r, "commentID")
	if err != nil {
		httperr.Write(w, h.Log, httperr.ErrInvalid)
		return
	}
	emoji := r.PathValue("reactionID")
	if err := h.Svc.RemoveReaction(r.Context(), commentID, actor(r).ID, emoji); err != nil {
		httperr.Write(w, h.Log, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
