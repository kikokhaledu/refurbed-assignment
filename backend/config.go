package main

import (
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type serverConfig struct {
	Host            string
	Port            int
	DataDir         string
	CacheTTL        time.Duration
	CORSAllowOrigin string
}

func loadServerConfig() serverConfig {
	cacheTTLSeconds := envInt("BACKEND_CACHE_TTL_SECONDS", DefaultCacheTTLSeconds)
	if cacheTTLSeconds <= 0 {
		cacheTTLSeconds = DefaultCacheTTLSeconds
	}

	return serverConfig{
		Host:            envString("BACKEND_HOST", DefaultBackendHost),
		Port:            envInt("BACKEND_PORT", DefaultBackendPort),
		DataDir:         envString("BACKEND_DATA_DIR", DefaultBackendDataDir),
		CacheTTL:        time.Duration(cacheTTLSeconds) * time.Second,
		CORSAllowOrigin: envString("BACKEND_CORS_ALLOW_ORIGIN", DefaultCORSAllowOrigin),
	}
}

func (c serverConfig) Address() string {
	host := strings.TrimSpace(c.Host)
	if host == "" {
		host = DefaultBackendHost
	}

	port := c.Port
	if port <= 0 {
		port = DefaultBackendPort
	}

	return net.JoinHostPort(host, strconv.Itoa(port))
}

func (c serverConfig) LogAddress() string {
	host := strings.TrimSpace(c.Host)
	if host == "" || host == DefaultBackendHost || host == "::" {
		host = "localhost"
	}

	port := c.Port
	if port <= 0 {
		port = DefaultBackendPort
	}

	return net.JoinHostPort(host, strconv.Itoa(port))
}

func envString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func envInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
