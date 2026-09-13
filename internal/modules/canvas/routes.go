package canvas

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/canvas/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/canvas/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB  *pgxpool.Pool
	Log zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB), m.Log)

	mux.HandleFunc("GET /api/v1/workspaces/{workspaceID}/canvases", auth.RequireUser(h.List))
	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/canvases", auth.RequireUser(h.Create))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/canvases/{canvasID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/canvases/{canvasID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}/objects", auth.RequireUser(h.ListObjects))
	mux.HandleFunc("POST /api/v1/canvases/{canvasID}/objects", auth.RequireUser(h.CreateObject))
	mux.HandleFunc("PATCH /api/v1/canvas-objects/{objectID}", auth.RequireUser(h.UpdateObject))
	mux.HandleFunc("DELETE /api/v1/canvas-objects/{objectID}", auth.RequireUser(h.DeleteObject))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}/connectors", auth.RequireUser(h.ListConnectors))
	mux.HandleFunc("POST /api/v1/canvases/{canvasID}/connectors", auth.RequireUser(h.CreateConnector))
	mux.HandleFunc("PATCH /api/v1/canvas-connectors/{connectorID}", auth.RequireUser(h.UpdateConnector))
	mux.HandleFunc("DELETE /api/v1/canvas-connectors/{connectorID}", auth.RequireUser(h.DeleteConnector))
	mux.HandleFunc("PUT /api/v1/canvases/{canvasID}/camera", auth.RequireUser(h.PutCamera))
	mux.HandleFunc("GET /api/v1/canvases/{canvasID}/tags", auth.RequireUser(h.ListTags))
	mux.HandleFunc("POST /api/v1/canvases/{canvasID}/tags", auth.RequireUser(h.CreateTag))
}
