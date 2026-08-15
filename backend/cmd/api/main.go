// Package main is the composition root of the KinSpace API. It wires the
// configuration, storage, application services and HTTP server, then
// manages the process lifecycle (startup, signals, graceful shutdown).
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fabien-Halaby/kinspace/backend/internal/application"
	"github.com/Fabien-Halaby/kinspace/backend/internal/config"
	"github.com/Fabien-Halaby/kinspace/backend/internal/httpapi"
	"github.com/Fabien-Halaby/kinspace/backend/internal/logger"
	"github.com/Fabien-Halaby/kinspace/backend/internal/storage/postgres"
	"github.com/Fabien-Halaby/kinspace/backend/internal/token"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		os.Stderr.WriteString("config: " + err.Error() + "\n")
		os.Exit(1)
	}
	log := logger.New(cfg.Environment)

	ctx := context.Background()
	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		log.Error("apply migrations", "error", err)
		os.Exit(1)
	}

	tokens := token.NewHS256Manager(cfg.JWTSecret, cfg.JWTTTL)
	userRepo := postgres.NewUserRepository(pool)
	familyRepo := postgres.NewFamilyRepository(pool)
	relationRepo := postgres.NewRelationRepository(pool)

	deps := httpapi.Dependencies{
		Environment: cfg.Environment,
		Logger:      log,
		Tokens:      tokens,
		Auth:        application.NewAuthService(userRepo, tokens),
		Families:    application.NewFamilyService(familyRepo),
		Relations:   application.NewRelationService(relationRepo),
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.NewRouter(deps),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("kinspace api listening", "addr", server.Addr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	log.Info("server stopped")
}
