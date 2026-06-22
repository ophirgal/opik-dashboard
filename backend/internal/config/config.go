package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL string
	Port        string

	// GitHub integration / metrics
	GitHubToken          string
	GitHubRepo           string // "owner/repo"
	MetricsWindowDays    int
	DisengagedMultiplier float64
	MetricsCacheTTL      time.Duration
}

func Load() *Config {
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Port:        os.Getenv("PORT"),

		GitHubToken:          os.Getenv("GITHUB_TOKEN"),
		GitHubRepo:           getString("GITHUB_REPO", "comet-ml/opik"),
		MetricsWindowDays:    getInt("METRICS_WINDOW_DAYS", 90),
		DisengagedMultiplier: getFloat("DISENGAGED_MULTIPLIER", 2.0),
		MetricsCacheTTL:      getDuration("METRICS_CACHE_TTL", 10*time.Minute),
	}
}

func getString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
