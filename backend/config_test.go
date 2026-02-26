package main

import (
	"testing"
	"time"
)

func TestLoadServerConfig_Defaults(t *testing.T) {
	t.Setenv("BACKEND_HOST", "")
	t.Setenv("BACKEND_PORT", "")
	t.Setenv("BACKEND_DATA_DIR", "")
	t.Setenv("BACKEND_CACHE_TTL_SECONDS", "")
	t.Setenv("BACKEND_CORS_ALLOW_ORIGIN", "")

	config := loadServerConfig()

	if config.Host != "0.0.0.0" {
		t.Fatalf("expected default host 0.0.0.0, got %q", config.Host)
	}
	if config.Port != 8080 {
		t.Fatalf("expected default port 8080, got %d", config.Port)
	}
	if config.DataDir != "data" {
		t.Fatalf("expected default data dir data, got %q", config.DataDir)
	}
	if config.CacheTTL != 30*time.Second {
		t.Fatalf("expected default cache ttl 30s, got %s", config.CacheTTL)
	}
	if config.CORSAllowOrigin != "*" {
		t.Fatalf("expected default CORS origin *, got %q", config.CORSAllowOrigin)
	}
}

func TestLoadServerConfig_Overrides(t *testing.T) {
	t.Setenv("BACKEND_HOST", "127.0.0.1")
	t.Setenv("BACKEND_PORT", "9090")
	t.Setenv("BACKEND_DATA_DIR", "fixtures")
	t.Setenv("BACKEND_CACHE_TTL_SECONDS", "45")
	t.Setenv("BACKEND_CORS_ALLOW_ORIGIN", "http://localhost:5173")

	config := loadServerConfig()

	if config.Host != "127.0.0.1" {
		t.Fatalf("expected host override, got %q", config.Host)
	}
	if config.Port != 9090 {
		t.Fatalf("expected port override 9090, got %d", config.Port)
	}
	if config.DataDir != "fixtures" {
		t.Fatalf("expected data dir override fixtures, got %q", config.DataDir)
	}
	if config.CacheTTL != 45*time.Second {
		t.Fatalf("expected cache ttl override 45s, got %s", config.CacheTTL)
	}
	if config.CORSAllowOrigin != "http://localhost:5173" {
		t.Fatalf("expected cors origin override, got %q", config.CORSAllowOrigin)
	}
}

func TestLoadServerConfig_InvalidNumbersFallback(t *testing.T) {
	t.Setenv("BACKEND_PORT", "not-a-number")
	t.Setenv("BACKEND_CACHE_TTL_SECONDS", "-5")

	config := loadServerConfig()

	if config.Port != 8080 {
		t.Fatalf("expected invalid port to fallback to 8080, got %d", config.Port)
	}
	if config.CacheTTL != 30*time.Second {
		t.Fatalf("expected invalid ttl to fallback to 30s, got %s", config.CacheTTL)
	}
}

func TestServerConfigAddressNormalization(t *testing.T) {
	config := serverConfig{Host: "", Port: 0}
	if got := config.Address(); got != "0.0.0.0:8080" {
		t.Fatalf("expected normalized default address 0.0.0.0:8080, got %q", got)
	}
	if got := config.LogAddress(); got != "localhost:8080" {
		t.Fatalf("expected normalized default log address localhost:8080, got %q", got)
	}

	config = serverConfig{Host: "::", Port: 9090}
	if got := config.LogAddress(); got != "localhost:9090" {
		t.Fatalf("expected wildcard v6 host to normalize to localhost in logs, got %q", got)
	}
}
