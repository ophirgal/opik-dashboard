package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"fsa-boilerplate/backend/internal/service"
)

func New(db *sql.DB, metrics *service.MetricsService) *gin.Engine {
	r := gin.Default()

	metricsHandler := NewMetricsHandler(metrics)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", Health)
		v1.GET("/metrics", metricsHandler.GetMetrics)
	}

	return r
}
