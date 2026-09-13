package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kashifxyz/flow-server/internal/config"
	"github.com/rs/zerolog"
)

func TestHealthzMounted(t *testing.T) {
	h := New(Deps{
		Config: config.Config{CORSOrigins: []string{"http://localhost:5173"}},
		Log:    zerolog.Nop(),
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("healthz status = %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("home status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") == "" {
		t.Fatal("expected html content type")
	}
}
