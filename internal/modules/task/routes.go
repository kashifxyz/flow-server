package task

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/projects/{projectID}/tasks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/projects/{projectID}/tasks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/tasks", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/tasks/{taskID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/tasks/{taskID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/tasks/{taskID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/tasks/{taskID}/move", auth.RequireUser(utils.NotImplemented))
}
