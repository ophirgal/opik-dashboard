package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"fsa-boilerplate/backend/internal/model"
)

const bugPrefix = "[bug]"

// MetricsService computes repository health metrics from GitHub data. Results
// are cached in memory for the configured TTL to respect GitHub rate limits.
type MetricsService struct {
	fetcher    GitHubFetcher
	repo       string
	windowDays int
	multiplier float64
	ttl        time.Duration
	now        func() time.Time

	mu       sync.RWMutex
	cached   *model.Metrics
	cachedAt time.Time
}

func NewMetricsService(f GitHubFetcher, repo string, windowDays int, multiplier float64, ttl time.Duration) *MetricsService {
	return &MetricsService{
		fetcher:    f,
		repo:       repo,
		windowDays: windowDays,
		multiplier: multiplier,
		ttl:        ttl,
		now:        time.Now,
	}
}

// Compute returns the current metrics, serving a cached copy when it is still
// fresh, otherwise fetching from GitHub and recomputing.
func (s *MetricsService) Compute(ctx context.Context) (model.Metrics, error) {
	s.mu.RLock()
	if s.cached != nil && s.now().Sub(s.cachedAt) < s.ttl {
		m := *s.cached
		s.mu.RUnlock()
		return m, nil
	}
	s.mu.RUnlock()

	now := s.now()
	since := now.AddDate(0, 0, -s.windowDays)

	prs, err := s.fetcher.ListPullRequests(ctx, since)
	if err != nil {
		return model.Metrics{}, fmt.Errorf("fetch pull requests: %w", classifyFetchErr(err))
	}
	issues, err := s.fetcher.ListIssues(ctx, since)
	if err != nil {
		return model.Metrics{}, fmt.Errorf("fetch issues: %w", classifyFetchErr(err))
	}

	avgPR, addressed := avgAddressTimeHours(prs)
	avgBug, closedBugs := avgBugCloseTimeHours(issues)

	m := model.Metrics{
		Repo:                    s.repo,
		WindowDays:              s.windowDays,
		GeneratedAt:             now,
		AvgTimeToAddressPRHours: avgPR,
		AddressedPRCount:        addressed,
		DisengagedPRCount:       countDisengaged(prs, avgPR, s.multiplier, now),
		DisengagedMultiplier:    s.multiplier,
		OpenBugIssueCount:       countOpenBugs(issues),
		AvgTimeToCloseBugHours:  avgBug,
		ClosedBugCount:          closedBugs,
	}

	s.mu.Lock()
	s.cached = &m
	s.cachedAt = now
	s.mu.Unlock()

	return m, nil
}

// avgAddressTimeHours averages (closed_at - created_at) over closed PRs.
func avgAddressTimeHours(prs []model.PullRequest) (avg float64, count int) {
	var total float64
	for _, pr := range prs {
		if pr.ClosedAt == nil {
			continue
		}
		total += pr.ClosedAt.Sub(pr.CreatedAt).Hours()
		count++
	}
	if count == 0 {
		return 0, 0
	}
	return round2(total / float64(count)), count
}

// countDisengaged counts open PRs whose age exceeds avgHours*multiplier. When
// there is no average to compare against, nothing is considered disengaged.
func countDisengaged(prs []model.PullRequest, avgHours, multiplier float64, now time.Time) int {
	threshold := avgHours * multiplier
	if threshold <= 0 {
		return 0
	}
	count := 0
	for _, pr := range prs {
		if pr.ClosedAt != nil {
			continue // not open
		}
		if now.Sub(pr.CreatedAt).Hours() > threshold {
			count++
		}
	}
	return count
}

// countOpenBugs counts open issues whose title is prefixed with "[BUG]".
func countOpenBugs(issues []model.Issue) int {
	count := 0
	for _, is := range issues {
		if is.State == "open" && isBug(is.Title) {
			count++
		}
	}
	return count
}

// avgBugCloseTimeHours averages (closed_at - created_at) over closed "[BUG]" issues.
func avgBugCloseTimeHours(issues []model.Issue) (avg float64, count int) {
	var total float64
	for _, is := range issues {
		if is.ClosedAt == nil || !isBug(is.Title) {
			continue
		}
		total += is.ClosedAt.Sub(is.CreatedAt).Hours()
		count++
	}
	if count == 0 {
		return 0, 0
	}
	return round2(total / float64(count)), count
}

func isBug(title string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(title)), bugPrefix)
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
