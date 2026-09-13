package auth

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
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

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
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
