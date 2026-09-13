package comment

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestListDiscussionsRequiresAuth(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Module{Log: zerolog.Nop()})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nodes/00000000-0000-7000-8000-000000000000/discussions", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}
