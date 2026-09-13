package users

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
)

func TestUpdateProfileRequiresAuth(t *testing.T) {
	mux := http.NewServeMux()
	Register(mux, Module{Log: zerolog.Nop()})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}
