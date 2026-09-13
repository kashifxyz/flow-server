package routes

import (
	"net/http"

	"github.com/kashifxyz/flow-server/cmd/web"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/config"
	"github.com/kashifxyz/flow-server/internal/middleware"
	"github.com/kashifxyz/flow-server/internal/modules/automation"
	"github.com/kashifxyz/flow-server/internal/modules/block"
	"github.com/kashifxyz/flow-server/internal/modules/canvas"
	"github.com/kashifxyz/flow-server/internal/modules/collaboration"
	"github.com/kashifxyz/flow-server/internal/modules/comment"
	flowdb "github.com/kashifxyz/flow-server/internal/modules/database"
	"github.com/kashifxyz/flow-server/internal/modules/file"
	"github.com/kashifxyz/flow-server/internal/modules/health"
	"github.com/kashifxyz/flow-server/internal/modules/node"
	"github.com/kashifxyz/flow-server/internal/modules/notification"
	"github.com/kashifxyz/flow-server/internal/modules/page"
	"github.com/kashifxyz/flow-server/internal/modules/project"
	"github.com/kashifxyz/flow-server/internal/modules/record"
	"github.com/kashifxyz/flow-server/internal/modules/search"
	"github.com/kashifxyz/flow-server/internal/modules/sharing"
	"github.com/kashifxyz/flow-server/internal/modules/space"
	"github.com/kashifxyz/flow-server/internal/modules/task"
	"github.com/kashifxyz/flow-server/internal/modules/users"
	"github.com/kashifxyz/flow-server/internal/modules/view"
	"github.com/kashifxyz/flow-server/internal/modules/webhook"
	"github.com/kashifxyz/flow-server/internal/modules/workspace"
	"github.com/rs/zerolog"
)

type Deps struct {
	Config   config.Config
	Log      zerolog.Logger
	Sessions *auth.Sessions
	Health   health.Module
	Users    users.Module
}

func New(deps Deps) http.Handler {
	mux := http.NewServeMux()
	health.Register(mux, deps.Health)
	users.Register(mux, deps.Users)
	workspace.Register(mux)
	space.Register(mux)
	node.Register(mux)
	page.Register(mux)
	block.Register(mux)
	flowdb.Register(mux)
	record.Register(mux)
	view.Register(mux)
	canvas.Register(mux)
	project.Register(mux)
	task.Register(mux)
	file.Register(mux)
	comment.Register(mux)
	notification.Register(mux)
	collaboration.Register(mux)
	search.Register(mux)
	automation.Register(mux)
	sharing.Register(mux)
	webhook.Register(mux)
	web.Register(mux)

	var h http.Handler = mux
	if deps.Sessions != nil {
		h = deps.Sessions.Middleware(h)
	}
	h = middleware.AccessLog(deps.Log)(h)
	h = middleware.Recover(deps.Log)(h)
	h = middleware.CORS(deps.Config.CORSOrigins)(h)
	h = middleware.RequestID(h)
	return h
}
