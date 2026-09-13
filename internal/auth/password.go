package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func ComparePassword(encoded, password string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("unsupported password hash")
	}
	if !strings.HasPrefix(parts[2], "v=") {
		return false, fmt.Errorf("parse hash version")
	}
	var memory uint32
	var timeCost uint32
	var threads uint8
	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		return false, fmt.Errorf("parse hash params")
	}
	m, err := strconv.ParseUint(strings.TrimPrefix(paramParts[0], "m="), 10, 32)
	if err != nil {
		return false, fmt.Errorf("parse memory: %w", err)
	}
	t, err := strconv.ParseUint(strings.TrimPrefix(paramParts[1], "t="), 10, 32)
	if err != nil {
		return false, fmt.Errorf("parse time: %w", err)
	}
	p, err := strconv.ParseUint(strings.TrimPrefix(paramParts[2], "p="), 10, 8)
	if err != nil {
		return false, fmt.Errorf("parse threads: %w", err)
	}
	memory, timeCost, threads = uint32(m), uint32(t), uint8(p)

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}
	got := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(want)))
	if subtle.ConstantTimeCompare(want, got) != 1 {
		return false, nil
	}
	return true, nil
}
