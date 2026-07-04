package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/saargunam-venkatesh/config-service/internal/domain"
	"github.com/saargunam-venkatesh/config-service/internal/repository"
	"github.com/saargunam-venkatesh/config-service/internal/service"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	svc *service.Service
	log *slog.Logger
}

// New creates a new Handler.
func New(svc *service.Service, log *slog.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Router wires all routes and returns the mux.
func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()
	r.Get("/ping", h.Ping)
	r.Get("/configs/{id}", h.GetConfig)
	r.Post("/configs", h.UpsertConfig)
	return r
}

// Ping handles GET /ping — used as liveness probe.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// GetConfig handles GET /configs/{id}.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cfg, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}
		h.log.Error("get config", "id", id, "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

// UpsertConfig handles POST /configs.
func (h *Handler) UpsertConfig(w http.ResponseWriter, r *http.Request) {
	var cfg domain.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.svc.Upsert(r.Context(), &cfg)
	if err != nil {
		// Validation errors come back as plain errors from the service.
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
