package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"

	ModeDev  = "dev"
	ModeProd = "prod"
)

type Config struct {
	Mode           string
	Env            string
	URL            string
	HTTPAddr       string
	DatabaseURL    string
	RedisURL       string
	LogLevel       string
	SessionSecret  string
	CORSOrigins    []string
	TrustedProxies []string
	ShutdownWait   time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	Mail           Mail
	S3             S3
}

type Mail struct {
	From         string
	FromName     string
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	Encryption   string
}

func (m Mail) Enabled() bool {
	return m.SMTPHost != "" && m.SMTPUser != "" && m.SMTPPassword != "" && m.From != ""
}

type S3 struct {
	Endpoint     string
	Region       string
	Bucket       string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

func (s S3) Enabled() bool {
	return s.Endpoint != "" && s.Bucket != "" && s.AccessKey != "" && s.SecretKey != ""
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	cfg := Config{
		HTTPAddr:       envOr("FLOW_HTTP_ADDR", ":8080"),
		DatabaseURL:    envOr("FLOW_DATABASE_URL", "postgres://localhost:5432/flow?sslmode=disable"),
		RedisURL:       envOr("FLOW_REDIS_URL", "redis://127.0.0.1:6379"),
		LogLevel:       envOr("FLOW_LOG_LEVEL", "info"),
		SessionSecret:  os.Getenv("FLOW_SESSION_SECRET"),
		TrustedProxies: splitCSV(envOr("FLOW_TRUSTED_PROXIES", "")),
		ShutdownWait:   durationOr("FLOW_SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:    durationOr("FLOW_HTTP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout:   durationOr("FLOW_HTTP_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:    durationOr("FLOW_HTTP_IDLE_TIMEOUT", 60*time.Second),
		MaxHeaderBytes: intOr("FLOW_HTTP_MAX_HEADER_BYTES", 1<<20),
		Mail: Mail{
			From:         envOr("FLOW_MAIL_FROM", ""),
			FromName:     envOr("FLOW_MAIL_FROM_NAME", "Flow"),
			SMTPHost:     envOr("FLOW_MAIL_SMTP_HOST", ""),
			SMTPPort:     intOr("FLOW_MAIL_SMTP_PORT", 465),
			SMTPUser:     envOr("FLOW_MAIL_SMTP_USER", ""),
			SMTPPassword: os.Getenv("FLOW_MAIL_SMTP_PASSWORD"),
			Encryption:   strings.ToLower(envOr("FLOW_MAIL_SMTP_ENCRYPTION", "ssl")),
		},
		S3: S3{
			Endpoint:     envOr("FLOW_S3_ENDPOINT", ""),
			Region:       envOr("FLOW_S3_REGION", "garage"),
			Bucket:       envOr("FLOW_S3_BUCKET", ""),
			AccessKey:    envOr("FLOW_S3_ACCESS_KEY", ""),
			SecretKey:    os.Getenv("FLOW_S3_SECRET_KEY"),
			UsePathStyle: boolOr("FLOW_S3_USE_PATH_STYLE", true),
		},
	}

	if err := applyMode(&cfg, firstNonEmpty(os.Getenv("FLOW_MODE"), os.Getenv("FLOW_ENV"), ModeDev)); err != nil {
		return Config{}, err
	}
	if err := applyURL(&cfg, strings.TrimSpace(os.Getenv("FLOW_URL"))); err != nil {
		return Config{}, err
	}

	if raw := strings.TrimSpace(os.Getenv("FLOW_CORS_ORIGINS")); raw != "" {
		cfg.CORSOrigins = splitCSV(raw)
	}
	cfg.CORSOrigins = mergeOrigins(cfg.CORSOrigins, cfg.URL)

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) IsDev() bool {
	return c.Mode == ModeDev
}

func (c Config) IsProd() bool {
	return c.Mode == ModeProd
}

func (c Config) RedisEnabled() bool {
	return c.RedisURL != ""
}

func (c Config) validate() error {
	switch c.Mode {
	case ModeDev, ModeProd:
	default:
		return fmt.Errorf("FLOW_MODE / -mode must be dev or prod")
	}
	switch c.Env {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		return fmt.Errorf("FLOW_ENV must be development, test, or production")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("FLOW_HTTP_ADDR is required")
	}
	if c.URL == "" {
		return fmt.Errorf("FLOW_URL / -url is required")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("FLOW_DATABASE_URL is required")
	}
	if c.IsProd() && c.SessionSecret == "" {
		return fmt.Errorf("FLOW_SESSION_SECRET is required in prod")
	}
	if c.IsProd() && len(c.CORSOrigins) == 0 {
		return fmt.Errorf("FLOW_CORS_ORIGINS or FLOW_URL is required in prod")
	}
	return nil
}

func applyMode(cfg *Config, raw string) error {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ModeDev, "development":
		cfg.Mode = ModeDev
		if cfg.Env != EnvTest {
			cfg.Env = EnvDevelopment
		}
	case ModeProd, "production":
		cfg.Mode = ModeProd
		cfg.Env = EnvProduction
	case "test":
		cfg.Mode = ModeDev
		cfg.Env = EnvTest
	default:
		return fmt.Errorf("mode must be dev or prod")
	}
	return nil
}

func applyURL(cfg *Config, raw string) error {
	if raw == "" {
		if cfg.IsProd() {
			return fmt.Errorf("url is required in prod")
		}
		raw = "http://localhost:5173"
	}
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return fmt.Errorf("url must be an absolute http(s) origin")
	}
	cfg.URL = raw
	return nil
}

func mergeOrigins(existing []string, url string) []string {
	seen := make(map[string]struct{}, len(existing)+2)
	out := make([]string, 0, len(existing)+2)
	add := func(o string) {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if o == "" {
			return
		}
		if _, ok := seen[o]; ok {
			return
		}
		seen[o] = struct{}{}
		out = append(out, o)
	}
	for _, o := range existing {
		add(o)
	}
	add(url)
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func intOr(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return n
}

func durationOr(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}

func boolOr(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "":
		return fallback
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
