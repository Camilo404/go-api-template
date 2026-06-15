// Command api is the entry point of the service. All wiring lives
// here: configuration, logging, dependency injection, and the
// lifecycle of the HTTP server. The rest of the codebase is
// organised so this file is the only one that needs to change when
// you add a new resource or swap a backing store.
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Camilo404/go-api-template/internal/config"
	"github.com/Camilo404/go-api-template/internal/router"
	"github.com/Camilo404/go-api-template/internal/server"
	"github.com/Camilo404/go-api-template/internal/service"
	"github.com/Camilo404/go-api-template/internal/store"
)

// @title           Template API
// @version         1.0
// @description     REST API template written in Go (net/http) with structured logging, CORS, request IDs and graceful shutdown.
// @termsOfService  https://example.com/terms

// @contact.name    Platform Team
// @contact.url     https://example.com
// @contact.email   platform@example.com

// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host            localhost:8080
// @BasePath        /
// @schemes         http https

func main() {
	// Bootstrap logger. We rebuild it once the real log level is
	// known so the first emitted line can carry the right verbosity.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config_load_failed", slog.Any("error", err))
		os.Exit(1)
	}
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))
	logger.Info("config_loaded",
		slog.String("env", cfg.Env),
		slog.String("port", cfg.Port),
		slog.Any("cors_origins", cfg.CORSOrigins),
	)

	// --- Dependency wiring -----------------------------------------
	// Swap MemoryTaskStore for a real database implementation here.
	// The service layer does not care which one it gets, as long as
	// it satisfies store.TaskStorer.
	taskStore := store.NewMemoryTaskStore()
	taskSvc := service.NewTaskService(taskStore)
	// ---------------------------------------------------------------

	handler := router.New(router.Deps{
		Config: cfg,
		Logger: logger,
		Tasks:  taskSvc,
	})

	srv := server.New(cfg, logger, handler)
	if err := srv.Run(context.Background()); err != nil {
		logger.Error("server_failed", slog.Any("error", err))
		os.Exit(1)
	}
}
