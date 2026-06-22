package service

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v66/github"

	"fsa-boilerplate/backend/internal/model"
)

type fakeFetcher struct {
	prs    []model.PullRequest
	issues []model.Issue
	err    error
}

func (f *fakeFetcher) ListPullRequests(ctx context.Context, since time.Time) ([]model.PullRequest, error) {
	return f.prs, f.err
}

func (f *fakeFetcher) ListIssues(ctx context.Context, since time.Time) ([]model.Issue, error) {
	return f.issues, f.err
}

func ptr(t time.Time) *time.Time { return &t }

func TestComputeHappyPath(t *testing.T) {
	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	h := func(n int) time.Time { return now.Add(time.Duration(-n) * time.Hour) }

	fetcher := &fakeFetcher{
		prs: []model.PullRequest{
			// closed PRs -> address time 8h and 10h (avg 9h)
			{Number: 1, State: "closed", CreatedAt: h(10), ClosedAt: ptr(h(2))},
			{Number: 2, State: "closed", CreatedAt: h(20), ClosedAt: ptr(h(10))},
			// open PRs -> threshold = 9h * 2 = 18h
			{Number: 3, State: "open", CreatedAt: h(100)}, // age 100h > 18h -> disengaged
			{Number: 4, State: "open", CreatedAt: h(5)},   // age 5h < 18h -> engaged
		},
		issues: []model.Issue{
			{Number: 10, Title: "[BUG] crash on start", State: "open"},                                 // open bug
			{Number: 11, Title: "[BUG] memory leak", State: "closed", CreatedAt: h(30), ClosedAt: ptr(h(6))}, // close 24h
			{Number: 12, Title: "[bug] lowercase prefix", State: "closed", CreatedAt: h(10), ClosedAt: ptr(h(5))}, // close 5h
			{Number: 13, Title: "Feature request", State: "open"},                                      // not a bug
			{Number: 14, Title: "  [BUG] leading spaces", State: "open"},                                // open bug (trimmed)
		},
	}

	svc := NewMetricsService(fetcher, "comet-ml/opik", 90, 2.0, time.Minute)
	svc.now = func() time.Time { return now }

	m, err := svc.Compute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m.AvgTimeToAddressPRHours != 9 {
		t.Errorf("AvgTimeToAddressPRHours = %v, want 9", m.AvgTimeToAddressPRHours)
	}
	if m.AddressedPRCount != 2 {
		t.Errorf("AddressedPRCount = %v, want 2", m.AddressedPRCount)
	}
	if m.DisengagedPRCount != 1 {
		t.Errorf("DisengagedPRCount = %v, want 1", m.DisengagedPRCount)
	}
	if m.OpenBugIssueCount != 2 {
		t.Errorf("OpenBugIssueCount = %v, want 2", m.OpenBugIssueCount)
	}
	if m.AvgTimeToCloseBugHours != 14.5 {
		t.Errorf("AvgTimeToCloseBugHours = %v, want 14.5", m.AvgTimeToCloseBugHours)
	}
	if m.ClosedBugCount != 2 {
		t.Errorf("ClosedBugCount = %v, want 2", m.ClosedBugCount)
	}
	if m.Repo != "comet-ml/opik" || m.WindowDays != 90 || m.DisengagedMultiplier != 2.0 {
		t.Errorf("metadata mismatch: %+v", m)
	}
}

func TestComputeFetchError(t *testing.T) {
	fetcher := &fakeFetcher{err: errors.New("github unavailable")}
	svc := NewMetricsService(fetcher, "comet-ml/opik", 90, 2.0, time.Minute)

	if _, err := svc.Compute(context.Background()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestComputeRateLimitErrorIsClassified(t *testing.T) {
	rateErr := &github.RateLimitError{
		Rate:     github.Rate{Limit: 60, Remaining: 0},
		Response: &http.Response{StatusCode: http.StatusForbidden},
		Message:  "API rate limit exceeded",
	}
	fetcher := &fakeFetcher{err: rateErr}
	svc := NewMetricsService(fetcher, "comet-ml/opik", 90, 2.0, time.Minute)

	_, err := svc.Compute(context.Background())
	if !errors.Is(err, ErrGitHubRateLimited) {
		t.Fatalf("expected ErrGitHubRateLimited, got %v", err)
	}
}

func TestComputeCaches(t *testing.T) {
	now := time.Date(2026, 6, 22, 12, 0, 0, 0, time.UTC)
	fetcher := &fakeFetcher{}
	svc := NewMetricsService(fetcher, "comet-ml/opik", 90, 2.0, time.Minute)
	svc.now = func() time.Time { return now }

	if _, err := svc.Compute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Swap fetcher to one that errors; a cached result should still be returned.
	svc.fetcher = &fakeFetcher{err: errors.New("should not be called")}
	if _, err := svc.Compute(context.Background()); err != nil {
		t.Fatalf("expected cached result, got error: %v", err)
	}
}
