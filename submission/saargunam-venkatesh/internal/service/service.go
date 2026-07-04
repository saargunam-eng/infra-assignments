package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/saargunam-venkatesh/config-service/internal/domain"
	"github.com/saargunam-venkatesh/config-service/internal/repository"
)

var validLogLevels = map[string]bool{
	"TRACE": true, "DEBUG": true, "INFO": true,
	"WARN": true, "ERROR": true, "FATAL": true,
}

// Service handles business logic for configs.
type Service struct {
	repo repository.Repository
}

// New creates a new Service.
func New(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

// Get retrieves a config by ID.
func (s *Service) Get(ctx context.Context, id string) (*domain.Config, error) {
	return s.repo.Get(ctx, id)
}

// Upsert validates and saves a config.
func (s *Service) Upsert(ctx context.Context, cfg *domain.Config) (*domain.Config, error) {
	if err := validate(cfg); err != nil {
		return nil, err
	}
	cfg.LogLevel = strings.ToUpper(cfg.LogLevel)
	return s.repo.Upsert(ctx, cfg)
}

func validate(cfg *domain.Config) error {
	if strings.TrimSpace(cfg.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if strings.TrimSpace(cfg.Host) == "" {
		return fmt.Errorf("host is required")
	}
	if strings.TrimSpace(cfg.AppName) == "" {
		return fmt.Errorf("app_name is required")
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if !validLogLevels[strings.ToUpper(cfg.LogLevel)] {
		return fmt.Errorf("log_level must be one of TRACE, DEBUG, INFO, WARN, ERROR, FATAL")
	}
	return nil
}
