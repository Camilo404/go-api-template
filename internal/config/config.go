// Package config loads runtime configuration from environment variables.
//
// Each field has a sensible default so the binary can boot in development
// without any env file. In production, override values via the environment.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime settings for the service.
type Config struct {
	Port          string
	Env           string
	LogLevel      slog.Level
	CORSOrigins   []string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	ShutdownWait  time.Duration
	MaxBodyBytes  int64
	EnableSwagger bool
}

// Load reads configuration from the environment and validates it.
func Load() (*Config, error) {
	env := getString("APP_ENV", "development")
	cfg := &Config{
		Port:          getString("PORT", "8080"),
		Env:           env,
		LogLevel:      parseLogLevel(getString("LOG_LEVEL", "info")),
		CORSOrigins:   splitCSV(getString("CORS_ORIGINS", "*")),
		ReadTimeout:   getDuration("READ_TIMEOUT", 15*time.Second),
		WriteTimeout:  getDuration("WRITE_TIMEOUT", 15*time.Second),
		IdleTimeout:   getDuration("IDLE_TIMEOUT", 60*time.Second),
		ShutdownWait:  getDuration("SHUTDOWN_WAIT", 15*time.Second),
		MaxBodyBytes:  getInt64("MAX_BODY_BYTES", 1<<20), // 1 MiB
		EnableSwagger: getBool("ENABLE_SWAGGER", !strings.EqualFold(env, "production")),
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// IsProd reports whether the service is running in production mode.
func (c *Config) IsProd() bool { return strings.EqualFold(c.Env, "production") }

func (c *Config) validate() error {
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("invalid PORT %q: must be numeric", c.Port)
	}
	if c.MaxBodyBytes <= 0 {
		return fmt.Errorf("MAX_BODY_BYTES must be positive")
	}
	return nil
}

func getString(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

func getBool(key string, def bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func getInt64(key string, def int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
