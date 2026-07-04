package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saargunam-venkatesh/config-service/internal/domain"
)

type postgresRepo struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a new Postgres-backed repository.
func NewPostgres(pool *pgxpool.Pool) Repository {
	return &postgresRepo{pool: pool}
}

func (r *postgresRepo) Get(ctx context.Context, id string) (*domain.Config, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, host, port, app_name, log_level, updated_at FROM configs WHERE id = $1`,
		id,
	)

	var cfg domain.Config
	err := row.Scan(&cfg.ID, &cfg.Host, &cfg.Port, &cfg.AppName, &cfg.LogLevel, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *postgresRepo) Upsert(ctx context.Context, cfg *domain.Config) (*domain.Config, error) {
	cfg.UpdatedAt = time.Now().UTC()

	_, err := r.pool.Exec(ctx,
		`INSERT INTO configs (id, host, port, app_name, log_level, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (id) DO UPDATE
		   SET host = EXCLUDED.host,
		       port = EXCLUDED.port,
		       app_name = EXCLUDED.app_name,
		       log_level = EXCLUDED.log_level,
		       updated_at = EXCLUDED.updated_at`,
		cfg.ID, cfg.Host, cfg.Port, cfg.AppName, cfg.LogLevel, cfg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
