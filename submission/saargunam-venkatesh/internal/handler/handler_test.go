package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/saargunam-venkatesh/config-service/internal/domain"
	"github.com/saargunam-venkatesh/config-service/internal/handler"
	"github.com/saargunam-venkatesh/config-service/internal/repository"
	"github.com/saargunam-venkatesh/config-service/internal/service"
)

// In-memory repo reused from service tests via the same interface.
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

func newTestHandler() http.Handler {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	svc := service.New(newMockRepo())
	h := handler.New(svc, log)
	return h.Router()
}

func TestPing(t *testing.T) {
	router := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "pong" {
		t.Errorf("expected pong, got %s", w.Body.String())
	}
}

func TestGetConfig_NotFound(t *testing.T) {
	router := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/configs/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpsertAndGetConfig(t *testing.T) {
	router := newTestHandler()

	body := `{"id":"cfg_1","host":"localhost","port":8080,"app_name":"config-service","log_level":"INFO"}`
	req := httptest.NewRequest(http.MethodPost, "/configs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on upsert, got %d — body: %s", w.Code, w.Body.String())
	}

	// Now read it back.
	req2 := httptest.NewRequest(http.MethodGet, "/configs/cfg_1", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 on get, got %d", w2.Code)
	}

	var cfg domain.Config
	if err := json.NewDecoder(w2.Body).Decode(&cfg); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if cfg.ID != "cfg_1" {
		t.Errorf("expected cfg_1, got %s", cfg.ID)
	}
}

func TestUpsertConfig_InvalidBody(t *testing.T) {
	router := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/configs", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpsertConfig_ValidationFailure(t *testing.T) {
	router := newTestHandler()
	// Missing host — should fail validation.
	body := `{"id":"cfg_1","host":"","port":8080,"app_name":"config-service","log_level":"INFO"}`
	req := httptest.NewRequest(http.MethodPost, "/configs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for validation failure, got %d", w.Code)
	}
}
