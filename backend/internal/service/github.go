package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"

	"fsa-boilerplate/backend/internal/model"
)

// GitHubFetcher retrieves the raw pull requests and issues needed to compute
// metrics. It is an interface so the calculation logic can be unit-tested with
// a fake implementation.
type GitHubFetcher interface {
	ListPullRequests(ctx context.Context, since time.Time) ([]model.PullRequest, error)
	ListIssues(ctx context.Context, since time.Time) ([]model.Issue, error)
}

// githubFetcher is the real GitHubFetcher backed by the GitHub REST API.
type githubFetcher struct {
	client *github.Client
	owner  string
	repo   string
}

// NewGitHubFetcher builds a fetcher for the given "owner/repo". An empty token
// uses unauthenticated access (subject to a low rate limit).
func NewGitHubFetcher(token, repo string) (GitHubFetcher, error) {
	owner, name, ok := strings.Cut(repo, "/")
	if !ok || owner == "" || name == "" {
		return nil, fmt.Errorf("invalid repo %q, expected \"owner/repo\"", repo)
	}

	client := github.NewClient(nil)
	if token != "" {
		client = client.WithAuthToken(token)
	}

	return &githubFetcher{client: client, owner: owner, repo: name}, nil
}

// ListPullRequests returns all currently-open PRs (regardless of age, so stale
// "disengaged" PRs are included) plus PRs closed within the window (for the
// average time-to-address calculation).
func (f *githubFetcher) ListPullRequests(ctx context.Context, since time.Time) ([]model.PullRequest, error) {
	open, err := f.listPRs(ctx, "open", time.Time{})
	if err != nil {
		return nil, err
	}
	closed, err := f.listPRs(ctx, "closed", since)
	if err != nil {
		return nil, err
	}
	return append(open, closed...), nil
}

// listPRs pages through PRs of a given state. When cutoff is non-zero, paging
// stops once a PR was last updated before it (the list is sorted updated-desc).
func (f *githubFetcher) listPRs(ctx context.Context, state string, cutoff time.Time) ([]model.PullRequest, error) {
	opt := &github.PullRequestListOptions{
		State:       state,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var out []model.PullRequest
	for {
		prs, resp, err := f.client.PullRequests.List(ctx, f.owner, f.repo, opt)
		if err != nil {
			return nil, err
		}
		for _, pr := range prs {
			if !cutoff.IsZero() && pr.GetUpdatedAt().Time.Before(cutoff) {
				return out, nil
			}
			out = append(out, model.PullRequest{
				Number:    pr.GetNumber(),
				State:     pr.GetState(),
				CreatedAt: pr.GetCreatedAt().Time,
				ClosedAt:  timePtr(pr.ClosedAt),
				MergedAt:  timePtr(pr.MergedAt),
			})
		}
		if resp.NextPage == 0 {
			return out, nil
		}
		opt.Page = resp.NextPage
	}
}

// ListIssues returns all currently-open issues (so stale open "[BUG]" issues
// are counted) plus issues closed within the window (for the average
// time-to-close calculation). Pull requests are excluded.
func (f *githubFetcher) ListIssues(ctx context.Context, since time.Time) ([]model.Issue, error) {
	open, err := f.listIssues(ctx, "open", time.Time{})
	if err != nil {
		return nil, err
	}
	closed, err := f.listIssues(ctx, "closed", since)
	if err != nil {
		return nil, err
	}
	return append(open, closed...), nil
}

// listIssues pages through issues of a given state. A non-zero since is applied
// server-side (filters by updated_at), bounding the closed set to the window.
func (f *githubFetcher) listIssues(ctx context.Context, state string, since time.Time) ([]model.Issue, error) {
	opt := &github.IssueListByRepoOptions{
		State:       state,
		Sort:        "updated",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if !since.IsZero() {
		opt.Since = since
	}

	var out []model.Issue
	for {
		issues, resp, err := f.client.Issues.ListByRepo(ctx, f.owner, f.repo, opt)
		if err != nil {
			return nil, err
		}
		for _, is := range issues {
			if is.IsPullRequest() {
				continue // the issues endpoint also returns PRs; skip them
			}
			out = append(out, model.Issue{
				Number:    is.GetNumber(),
				Title:     is.GetTitle(),
				State:     is.GetState(),
				CreatedAt: is.GetCreatedAt().Time,
				ClosedAt:  timePtr(is.ClosedAt),
			})
		}
		if resp.NextPage == 0 {
			return out, nil
		}
		opt.Page = resp.NextPage
	}
}

func timePtr(ts *github.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.Time
	return &t
}
