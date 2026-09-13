package auth

import (
	"net/http"
	"time"
)

const (
	SessionCookie = "flow_session"
	CSRFCookie    = "flow_csrf"
	CSRFHeader    = "X-CSRF-Token"
)

type CookieOpts struct {
	Secure bool
	MaxAge time.Duration
}

func SetAuthCookies(w http.ResponseWriter, sessionToken, csrfToken string, opts CookieOpts) {
	maxAge := int(opts.MaxAge.Seconds())
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookie,
		Value:    csrfToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearAuthCookies(w http.ResponseWriter, secure bool) {
	expired := &http.Cookie{
		Path:     "/",
		MaxAge:   -1,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	}
	sid := *expired
	sid.Name = SessionCookie
	sid.HttpOnly = true
	http.SetCookie(w, &sid)
	csrf := *expired
	csrf.Name = CSRFCookie
	http.SetCookie(w, &csrf)
}
