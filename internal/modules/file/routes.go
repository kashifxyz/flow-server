package file

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/file/handlers"
	"github.com/kashifxyz/flow-server/internal/modules/file/services"
	"github.com/rs/zerolog"
)

type Module struct {
	DB     *pgxpool.Pool
	S3     *s3.Client
	Bucket string
	Log    zerolog.Logger
}

func Register(mux *http.ServeMux, m Module) {
	h := handlers.New(services.New(m.DB, m.S3, m.Bucket), m.Log)

	mux.HandleFunc("POST /api/v1/workspaces/{workspaceID}/files/uploads", auth.RequireUser(h.CreateUpload))
	mux.HandleFunc("POST /api/v1/files/uploads/{uploadID}/complete", auth.RequireUser(h.CompleteUpload))
	mux.HandleFunc("GET /api/v1/files/{fileID}", auth.RequireUser(h.Get))
	mux.HandleFunc("PATCH /api/v1/files/{fileID}", auth.RequireUser(h.Update))
	mux.HandleFunc("DELETE /api/v1/files/{fileID}", auth.RequireUser(h.Delete))
	mux.HandleFunc("GET /api/v1/files/{fileID}/download", auth.RequireUser(h.Download))
}
