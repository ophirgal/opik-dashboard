package main

import (
	"log"

	"fsa-boilerplate/backend/internal/api"
	"fsa-boilerplate/backend/internal/config"
	"fsa-boilerplate/backend/internal/dal"
	"fsa-boilerplate/backend/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := dal.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := dal.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	fetcher, err := service.NewGitHubFetcher(cfg.GitHubToken, cfg.GitHubRepo)
	if err != nil {
		log.Fatalf("failed to init github fetcher: %v", err)
	}
	metricsSvc := service.NewMetricsService(fetcher, cfg.GitHubRepo, cfg.MetricsWindowDays, cfg.DisengagedMultiplier, cfg.MetricsCacheTTL)

	r := api.New(db, metricsSvc)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("opik-dashboard server listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
