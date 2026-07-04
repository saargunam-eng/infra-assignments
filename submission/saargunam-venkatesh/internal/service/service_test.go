package service_test

import (
	"context"
	"testing"

	"github.com/saargunam-venkatesh/config-service/internal/domain"
	"github.com/saargunam-venkatesh/config-service/internal/repository"
	"github.com/saargunam-venkatesh/config-service/internal/service"
)

// mockRepo is a simple in-memory stub — no external mock library needed.
type mockRepo struct {
	store map[string]*domain.Config
}

func newMockRepo() *mockRepo {
	return &mockRepo{store: make(map[string]*domain.Config)}
}

func (m *mockRepo) Get(_ context.Context, id string) (*domain.Config, error) {
	cfg, ok := m.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return cfg, nil
}

func (m *mockRepo) Upsert(_ context.Context, cfg *domain.Config) (*domain.Config, error) {
	m.store[cfg.ID] = cfg
	return cfg, nil
}

func validConfig() *domain.Config {
	return &domain.Config{
		ID:       "cfg_1",
		Host:     "localhost",
		Port:     8080,
		AppName:  "config-service",
		LogLevel: "INFO",
	}
}

func TestUpsert_Valid(t *testing.T) {
	svc := service.New(newMockRepo())
	cfg, err := svc.Upsert(context.Background(), validConfig())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.ID != "cfg_1" {
		t.Errorf("expected id cfg_1, got %s", cfg.ID)
	}
}

func TestUpsert_LogLevelNormalised(t *testing.T) {
	svc := service.New(newMockRepo())
	input := validConfig()
	input.LogLevel = "info" // lowercase — should be normalised to INFO
	cfg, err := svc.Upsert(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.LogLevel != "INFO" {
		t.Errorf("expected INFO, got %s", cfg.LogLevel)
	}
}

func TestUpsert_MissingID(t *testing.T) {
	svc := service.New(newMockRepo())
	cfg := validConfig()
	cfg.ID = ""
	_, err := svc.Upsert(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestUpsert_MissingHost(t *testing.T) {
	svc := service.New(newMockRepo())
	cfg := validConfig()
	cfg.Host = ""
	_, err := svc.Upsert(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestUpsert_InvalidPort(t *testing.T) {
	svc := service.New(newMockRepo())
	cases := []int{0, -1, 65536, 99999}
	for _, port := range cases {
		cfg := validConfig()
		cfg.Port = port
		_, err := svc.Upsert(context.Background(), cfg)
		if err == nil {
			t.Errorf("expected error for port %d", port)
		}
	}
}

func TestUpsert_InvalidLogLevel(t *testing.T) {
	svc := service.New(newMockRepo())
	cfg := validConfig()
	cfg.LogLevel = "VERBOSE"
	_, err := svc.Upsert(context.Background(), cfg)
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestGet_Found(t *testing.T) {
	repo := newMockRepo()
	svc := service.New(repo)
	svc.Upsert(context.Background(), validConfig())

	cfg, err := svc.Get(context.Background(), "cfg_1")
	if err != nil {
		t.Fatalf("expected config, got error: %v", err)
	}
	if cfg.ID != "cfg_1" {
		t.Errorf("expected cfg_1, got %s", cfg.ID)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := service.New(newMockRepo())
	_, err := svc.Get(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected ErrNotFound")
	}
}
