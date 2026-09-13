package canvas

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/utils"
)

func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/canvases", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/canvases", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/canvases/{canvasID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/canvases/{canvasID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}/objects", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/canvases/{canvasID}/objects", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/canvas-objects/{objectID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/canvas-objects/{objectID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}/connectors", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/canvases/{canvasID}/connectors", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PATCH /api/v1/canvas-connectors/{connectorID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("DELETE /api/v1/canvas-connectors/{connectorID}", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("PUT /api/v1/canvases/{canvasID}/camera", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}/tags", auth.RequireUser(utils.NotImplemented))
	mux.HandleFunc("POST /api/v1/canvases/{canvasID}/tags", auth.RequireUser(utils.NotImplemented))
}
