package model

import "time"

// PullRequest is a normalized view of a GitHub pull request, holding only the
// fields the metric calculations need.
type PullRequest struct {
	Number    int        `json:"number"`
	State     string     `json:"state"` // "open" or "closed"
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
	MergedAt  *time.Time `json:"merged_at,omitempty"`
}

// Issue is a normalized view of a GitHub issue (pull requests are excluded
// upstream, so these are issues only).
type Issue struct {
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	State     string     `json:"state"` // "open" or "closed"
	CreatedAt time.Time  `json:"created_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty"`
}
