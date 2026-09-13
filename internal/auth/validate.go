package auth

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	MinPasswordLen = 10
	MaxPasswordLen = 128
	MaxEmailLen    = 254
	MaxNameLen     = 120
)

func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || utf8.RuneCountInString(email) > MaxEmailLen {
		return "", ErrInvalidRequest
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ErrInvalidRequest
	}
	return email, nil
}

func NormalizeDisplayName(raw, email string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		local, _, _ := strings.Cut(email, "@")
		name = local
	}
	if utf8.RuneCountInString(name) > MaxNameLen {
		runes := []rune(name)
		name = string(runes[:MaxNameLen])
	}
	return name
}

func ValidatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < MinPasswordLen || n > MaxPasswordLen {
		return ErrInvalidRequest
	}
	return nil
}
