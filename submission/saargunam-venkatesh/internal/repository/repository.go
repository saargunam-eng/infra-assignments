package repository

import (
	"context"
	"errors"

	"github.com/saargunam-venkatesh/config-service/internal/domain"
)

// ErrNotFound is returned when a config does not exist.
var ErrNotFound = errors.New("config not found")

// Repository is the persistence interface. The service layer depends only on this.
type Repository interface {
	Get(ctx context.Context, id string) (*domain.Config, error)
	Upsert(ctx context.Context, cfg *domain.Config) (*domain.Config, error)
}
