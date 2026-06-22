package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"fsa-boilerplate/backend/internal/service"
)

// MetricsHandler exposes repository health metrics over HTTP.
type MetricsHandler struct {
	svc *service.MetricsService
}

func NewMetricsHandler(svc *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{svc: svc}
}

// GetMetrics handles GET /api/v1/metrics.
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	metrics, err := h.svc.Compute(c.Request.Context())
	if err != nil {
		c.JSON(statusForErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// statusForErr maps known upstream failures to meaningful HTTP status codes,
// defaulting to 502 for any other GitHub problem.
func statusForErr(err error) int {
	switch {
	case errors.Is(err, service.ErrGitHubRateLimited):
		return http.StatusTooManyRequests // 429
	case errors.Is(err, service.ErrGitHubUnauthorized):
		return http.StatusUnauthorized // 401
	default:
		return http.StatusBadGateway // 502
	}
}
