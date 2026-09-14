package auth

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type ctxKey int

const (
	ctxUser ctxKey = iota
	ctxSession
)

type User struct {
	ID              uuid.UUID
	Email           string
	DisplayName     string
	PasswordHash    string
	EmailVerifiedAt *time.Time
	FailedLogins    int
	LockedUntil     *time.Time
	DisabledAt      *time.Time
	DeletedAt       *time.Time
}

func (u User) EmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

type Session struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TokenHash     string
	CSRFHash      string
	ExpiresAt     time.Time
	IdleTimeoutAt *time.Time
	RevokedAt     *time.Time
}

type RequestMeta struct {
	IP        string
	UserAgent string
}

func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, ctxUser, u)
}

func WithSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, ctxSession, s)
}

func UserFrom(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(ctxUser).(User)
	return u, ok
}

func SessionFrom(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(ctxSession).(Session)
	return s, ok
}

func MetaFromRequest(r *http.Request) RequestMeta {
	return RequestMeta{
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	}
}

var trustedProxies atomic.Pointer[[]netip.Prefix]

// SetTrustedProxies configures the reverse-proxy CIDR ranges clientIP is
// willing to trust X-Forwarded-For from. Call it once at startup (e.g. from
// config); an empty/unset list (the default) means no proxy is trusted and
// X-Forwarded-For is never honored, since it is otherwise trivially spoofable
// by any client and would poison login_attempts/auth_events with fake IPs.
func SetTrustedProxies(prefixes []netip.Prefix) {
	cp := append([]netip.Prefix(nil), prefixes...)
	trustedProxies.Store(&cp)
}

func isTrustedProxy(host string, proxies []netip.Prefix) bool {
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	for _, p := range proxies {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	proxies := trustedProxies.Load()
	if proxies != nil && len(*proxies) > 0 && isTrustedProxy(host, *proxies) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			hops := strings.Split(xff, ",")
			for i := len(hops) - 1; i >= 0; i-- {
				candidate := strings.TrimSpace(hops[i])
				if candidate == "" {
					continue
				}
				// Skip hops that are themselves trusted proxies and keep walking
				// left until we reach the right-most hop the proxy chain didn't
				// vouch for — that is the real, un-spoofable client address.
				if isTrustedProxy(candidate, *proxies) {
					continue
				}
				return candidate
			}
		}
	}
	return host
}

func inetArg(ip string) any {
	if ip == "" {
		return nil
	}
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return nil
	}
	return addr
}
