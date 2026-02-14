package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// shared
	DBDSN string

	// api
	Port string

	// worker
	WorkerID     string
	Workers      int
	LeaseSeconds int
	PollInterval time.Duration
}

func Load() Config {
	return Config{
		DBDSN:        envOr("DB_DSN", ""),
		Port:         envOr("PORT", "8080"),
		WorkerID:     envOr("WORKER_ID", "worker-1"),
		Workers:      envInt("WORKERS", 8),
		LeaseSeconds: envInt("LEASE_SECONDS", 30),
		PollInterval: time.Duration(envInt("POLL_INTERVAL_MS", 500)) * time.Millisecond,
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
