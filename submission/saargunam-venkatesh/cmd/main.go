package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saargunam-venkatesh/config-service/internal/handler"
	"github.com/saargunam-venkatesh/config-service/internal/repository"
	"github.com/saargunam-venkatesh/config-service/internal/service"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := getEnv("APP_PORT", "8080")
	dbURL := mustEnv("DATABASE_URL", log)

	// Connect to Postgres with a retry loop — the pod may start before Postgres is ready.
	pool, err := connectWithRetry(context.Background(), dbURL, log)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Run schema migration.
	if err := runMigration(context.Background(), pool, log); err != nil {
		log.Error("migration failed", "err", err)
		os.Exit(1)
	}

	repo := repository.NewPostgres(pool)
	svc := service.New(repo)
	h := handler.New(svc, log)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h.Router(),
	}

	// Graceful shutdown on SIGINT/SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	log.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

// connectWithRetry retries Postgres connection up to 10 times before giving up.
func connectWithRetry(ctx context.Context, dsn string, log *slog.Logger) (*pgxpool.Pool, error) {
	var pool *pgxpool.Pool
	var err error
	for i := range 10 {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				log.Info("connected to database")
				return pool, nil
			}
			pool.Close()
		}
		log.Warn("database not ready, retrying", "attempt", i+1)
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("could not connect to database after retries: %w", err)
}

// runMigration creates the configs table if it doesn't exist.
func runMigration(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	sql := `CREATE TABLE IF NOT EXISTS configs (
		id         TEXT        PRIMARY KEY,
		host       TEXT        NOT NULL,
		port       INTEGER     NOT NULL CHECK (port > 0 AND port <= 65535),
		app_name   TEXT        NOT NULL,
		log_level  TEXT        NOT NULL DEFAULT 'INFO',
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`
	_, err := pool.Exec(ctx, sql)
	if err != nil {
		return err
	}
	log.Info("migration applied")
	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string, log *slog.Logger) string {
	v := os.Getenv(key)
	if v == "" {
		log.Error("required env var not set", "key", key)
		os.Exit(1)
	}
	return v
}
