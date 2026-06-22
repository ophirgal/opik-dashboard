package model

import "time"

// Metrics is the set of repository health metrics exposed by the dashboard API.
type Metrics struct {
	Repo        string    `json:"repo"`
	WindowDays  int       `json:"window_days"`
	GeneratedAt time.Time `json:"generated_at"`

	// PR metrics
	AvgTimeToAddressPRHours float64 `json:"avg_time_to_address_pr_hours"`
	AddressedPRCount        int     `json:"addressed_pr_count"`
	DisengagedPRCount       int     `json:"disengaged_pr_count"`
	DisengagedMultiplier    float64 `json:"disengaged_multiplier"`

	// Bug issue metrics (issues whose title is prefixed with "[BUG]")
	OpenBugIssueCount      int     `json:"open_bug_issue_count"`
	AvgTimeToCloseBugHours float64 `json:"avg_time_to_close_bug_hours"`
	ClosedBugCount         int     `json:"closed_bug_count"`
}
