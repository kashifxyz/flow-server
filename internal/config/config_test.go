package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("FLOW_MODE", "dev")
	t.Setenv("FLOW_URL", "")
	t.Setenv("FLOW_ENV", "")
	t.Setenv("FLOW_HTTP_ADDR", "")
	t.Setenv("FLOW_DATABASE_URL", "")
	t.Setenv("FLOW_SESSION_SECRET", "")
	t.Setenv("FLOW_CORS_ORIGINS", "")
	t.Setenv("FLOW_REDIS_URL", "redis://127.0.0.1:6379")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Mode != ModeDev {
		t.Fatalf("mode = %q", cfg.Mode)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("addr = %q", cfg.HTTPAddr)
	}
	if cfg.URL != "http://localhost:5173" {
		t.Fatalf("url = %q", cfg.URL)
	}
	if cfg.RedisURL != "redis://127.0.0.1:6379" {
		t.Fatalf("redis = %q", cfg.RedisURL)
	}
}

func TestProductionRequiresSecret(t *testing.T) {
	t.Setenv("FLOW_MODE", "prod")
	t.Setenv("FLOW_URL", "https://flow.sbkashif.com")
	t.Setenv("FLOW_SESSION_SECRET", "")
	t.Setenv("FLOW_CORS_ORIGINS", "")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDotEnvDoesNotOverrideEnv(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(".env", []byte("FLOW_HTTP_ADDR=:9999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FLOW_HTTP_ADDR", ":8081")
	if err := loadDotEnv(".env"); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv("FLOW_HTTP_ADDR"); got != ":8081" {
		t.Fatalf("got %q", got)
	}
}

func TestApplyFlags(t *testing.T) {
	t.Setenv("FLOW_MODE", "dev")
	t.Setenv("FLOW_URL", "http://localhost:5173")
	t.Setenv("FLOW_SESSION_SECRET", "test-secret")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyFlags(&cfg, []string{
		"-addr", ":9090",
		"-mode", "prod",
		"-url", "https://flow.sbkashif.com",
	}, os.Stderr); err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("addr = %q", cfg.HTTPAddr)
	}
	if cfg.Mode != ModeProd || cfg.URL != "https://flow.sbkashif.com" {
		t.Fatalf("mode=%q url=%q", cfg.Mode, cfg.URL)
	}
}
