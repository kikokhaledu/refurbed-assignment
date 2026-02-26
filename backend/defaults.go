package main

import "time"

const (
	DefaultBackendHost      = "0.0.0.0"
	DefaultBackendPort      = 8080
	DefaultBackendDataDir   = "data"
	DefaultCORSAllowOrigin  = "*"
	DefaultCacheTTLSeconds  = 30
	DefaultCacheTTLDuration = 30 * time.Second
)
