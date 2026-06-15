package handlers

import (
	"net/http"
	"time"

	"github.com/Camilo404/go-api-template/internal/httpx"
)

// HealthResponse is the JSON shape returned by GET /health. Declared
// at package level so the OpenAPI generator can document the fields.
type HealthResponse struct {
	Status    string `json:"status" example:"ok"`
	StartedAt string `json:"started_at" example:"2026-06-15T12:00:00Z"`
}

// HealthHandler answers liveness/readiness probes. The started_at
// field is useful when debugging "how long has this pod been alive".
type HealthHandler struct {
	startedAt time.Time
}

// NewHealthHandler returns a handler that reports service uptime.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{startedAt: time.Now()}
}

// ServeHTTP godoc
// @Summary      Liveness probe
// @Description  Returns 200 OK with service uptime. Intended for liveness/readiness probes.
// @Tags         health
// @Produce      json
// @Success      200 {object} HealthResponse
// @Router       /health [get]
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, HealthResponse{
		Status:    "ok",
		StartedAt: h.startedAt.UTC().Format(time.RFC3339),
	})
}
