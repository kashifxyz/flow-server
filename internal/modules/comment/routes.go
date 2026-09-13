package comment

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/comment/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/comment/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/nodes/{nodeID}/discussions", auth.RequireUser(h.ListDiscussions))
	mux.HandleFunc("POST /api/v1/nodes/{nodeID}/discussions", auth.RequireUser(h.CreateDiscussion))
	mux.HandleFunc("GET /api/v1/discussions/{discussionID}", auth.RequireUser(h.GetDiscussion))
	mux.HandleFunc("POST /api/v1/discussions/{discussionID}/resolve", auth.RequireUser(h.ResolveDiscussion))
	mux.HandleFunc("GET /api/v1/discussions/{discussionID}/comments", auth.RequireUser(h.ListComments))
	mux.HandleFunc("POST /api/v1/discussions/{discussionID}/comments", auth.RequireUser(h.CreateComment))
	mux.HandleFunc("PATCH /api/v1/comments/{commentID}", auth.RequireUser(h.UpdateComment))
	mux.HandleFunc("DELETE /api/v1/comments/{commentID}", auth.RequireUser(h.DeleteComment))
	mux.HandleFunc("POST /api/v1/comments/{commentID}/reactions", auth.RequireUser(h.AddReaction))
	mux.HandleFunc("DELETE /api/v1/comments/{commentID}/reactions/{reactionID}", auth.RequireUser(h.RemoveReaction))
}
