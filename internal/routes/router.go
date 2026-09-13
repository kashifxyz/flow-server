package routes

import (
	"net/http"

	"github.com/kashifxyz/flow-server/cmd/web"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/config"
	"github.com/kashifxyz/flow-server/internal/middleware"
	authmodule "github.com/kashifxyz/flow-server/internal/modules/auth"
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
	"github.com/kashifxyz/flow-server/internal/realtime"
	"github.com/rs/zerolog"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Deps struct {
	Config        config.Config
	Log           zerolog.Logger
	DB            *pgxpool.Pool
	S3            *s3.Client
	Hub           *realtime.Hub
	Sessions      *auth.Sessions
	AuthService   *auth.Service
	SecureCookies bool
	Health        health.Module
}

func New(deps Deps) http.Handler {
	mux := http.NewServeMux()
	health.Register(mux, deps.Health)
	authmodule.Register(mux, authmodule.Module{Auth: deps.AuthService, SecureCookies: deps.SecureCookies, Log: deps.Log})
	users.Register(mux, users.Module{DB: deps.DB, Auth: deps.AuthService, Log: deps.Log})
	workspace.Register(mux, workspace.Module{DB: deps.DB, Log: deps.Log})
	space.Register(mux, space.Module{DB: deps.DB, Log: deps.Log})
	node.Register(mux, node.Module{DB: deps.DB, Log: deps.Log})
	page.Register(mux, page.Module{DB: deps.DB, Log: deps.Log})
	block.Register(mux, block.Module{DB: deps.DB, Log: deps.Log})
	flowdb.Register(mux, flowdb.Module{DB: deps.DB, Log: deps.Log})
	record.Register(mux, record.Module{DB: deps.DB, Log: deps.Log})
	view.Register(mux, view.Module{DB: deps.DB, Log: deps.Log})
	canvas.Register(mux, canvas.Module{DB: deps.DB, Log: deps.Log})
	project.Register(mux, project.Module{DB: deps.DB, Log: deps.Log})
	task.Register(mux, task.Module{DB: deps.DB, Log: deps.Log})
	file.Register(mux, file.Module{DB: deps.DB, S3: deps.S3, Bucket: deps.Config.S3.Bucket, Log: deps.Log})
	comment.Register(mux, comment.Module{DB: deps.DB, Log: deps.Log})
	notification.Register(mux, notification.Module{DB: deps.DB, Log: deps.Log})
	collaboration.Register(mux, collaboration.Module{DB: deps.DB, Hub: deps.Hub, Origins: deps.Config.CORSOrigins, Log: deps.Log})
	search.Register(mux, search.Module{DB: deps.DB, Log: deps.Log})
	automation.Register(mux, automation.Module{DB: deps.DB, Log: deps.Log})
	sharing.Register(mux, sharing.Module{DB: deps.DB, Log: deps.Log})
	webhook.Register(mux, webhook.Module{DB: deps.DB, Log: deps.Log})
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
