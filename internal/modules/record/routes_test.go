package record

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestGetRequiresAuth(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Module{Log: zerolog.Nop()})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/records/00000000-0000-7000-8000-000000000000", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}
