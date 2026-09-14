package workspace

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/workspace/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/workspace/services"
	"github.com/kashifxyz/flow-server/pkg/mail"
	"github.com/rs/zerolog"
)

type Module struct {
	DB        *pgxpool.Pool
	Mail      mail.Sender
	PublicURL string
	Log       zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB, m.Mail, m.PublicURL), m.Log)

	mux.HandleFunc("GET /api/v1/workspaces", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/workspaces", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/usage", auth.RequireUser(h.Usage))

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/members", auth.RequireUser(h.ListMembers))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/members", auth.RequireUser(h.AddMember))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}/members/{userID}", auth.RequireUser(h.UpdateMember))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/members/{userID}", auth.RequireUser(h.RemoveMember))

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/invites", auth.RequireUser(h.ListInvites))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/invites", auth.RequireUser(h.CreateInvite))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/invites/{inviteID}", auth.RequireUser(h.RevokeInvite))
	mux.HandleFunc("POST /api/v1/invites/{token}/accept", auth.RequireUser(h.AcceptInvite))
	mux.HandleFunc("GET /api/v1/invites/mine", auth.RequireUser(h.ListMyInvites))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/invites/{inviteID}/accept", auth.RequireUser(h.AcceptInviteByID))

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/roles", auth.RequireUser(h.ListRoles))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/roles", auth.RequireUser(h.CreateRole))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}/roles/{roleID}", auth.RequireUser(h.UpdateRole))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/roles/{roleID}", auth.RequireUser(h.DeleteRole))

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/groups", auth.RequireUser(h.ListGroups))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/groups", auth.RequireUser(h.CreateGroup))
	mux.HandleFunc("PATCH /api/v1/workspaces/{workspaceID}/groups/{groupID}", auth.RequireUser(h.UpdateGroup))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/groups/{groupID}", auth.RequireUser(h.DeleteGroup))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/groups/{groupID}/members", auth.RequireUser(h.ListGroupMembers))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/groups/{groupID}/members", auth.RequireUser(h.AddGroupMember))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/groups/{groupID}/members/{userID}", auth.RequireUser(h.RemoveGroupMember))

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/domains", auth.RequireUser(h.ListDomains))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/domains", auth.RequireUser(h.CreateDomain))
	mux.HandleFunc("DELETE /api/v1/workspaces/{workspaceID}/domains/{domainID}", auth.RequireUser(h.DeleteDomain))
}
