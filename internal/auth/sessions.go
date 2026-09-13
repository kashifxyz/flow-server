package auth

import (
	"net/http"

	"github.com/kashifxyz/flow-server/internal/utils"
	"github.com/rs/zerolog"
)

type Sessions struct {
	Service *Service
	Log     zerolog.Logger
}

func (s *Sessions) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s == nil || s.Service == nil {
			next.ServeHTTP(w, r)
			return
		}
		if c, err := r.Cookie(SessionCookie); err == nil && c.Value != "" {
			sess, user, err := s.Service.LookupSession(r.Context(), c.Value)
			if err == nil {
				ctx := WithUser(r.Context(), user)
				ctx = WithSession(ctx, sess)
				r = r.WithContext(ctx)
			}
		}
		if err := s.checkCSRF(r); err != nil {
			utils.JSONError(w, http.StatusForbidden, "csrf_failed", "Reload the page and try again.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Sessions) checkCSRF(r *http.Request) error {
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return nil
	}
	sess, ok := SessionFrom(r.Context())
	if !ok {
		return nil
	}
	header := r.Header.Get(CSRFHeader)
	cookie, err := r.Cookie(CSRFCookie)
	if err != nil || !s.Service.ValidCSRF(sess, header, cookie.Value) {
		return ErrCSRF
	}
	return nil
}

func RequireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFrom(r.Context()); !ok {
			utils.JSONError(w, http.StatusUnauthorized, "unauthenticated", "Sign in required.")
			return
		}
		next(w, r)
	}
}
