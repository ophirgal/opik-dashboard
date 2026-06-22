package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v66/github"
)

// Sentinel errors so the API layer can map GitHub failures to meaningful HTTP
// status codes and surface actionable messages to the user.
var (
	// ErrGitHubRateLimited indicates the GitHub API rate limit was hit. This is
	// the common case when running without a GITHUB_TOKEN (60 req/hr).
	ErrGitHubRateLimited = errors.New("GitHub API rate limit exceeded — set GITHUB_TOKEN in your .env to raise the limit from 60 to 5000 requests/hour, then restart the backend")

	// ErrGitHubUnauthorized indicates the provided token was rejected.
	ErrGitHubUnauthorized = errors.New("GitHub authentication failed — check that GITHUB_TOKEN is valid and has access to the repository")
)

// classifyFetchErr converts low-level go-github errors into the package's
// sentinel errors when they match a known, user-actionable condition.
func classifyFetchErr(err error) error {
	if err == nil {
		return nil
	}

	var rateErr *github.RateLimitError
	var abuseErr *github.AbuseRateLimitError
	if errors.As(err, &rateErr) || errors.As(err, &abuseErr) {
		return fmt.Errorf("%w (%v)", ErrGitHubRateLimited, err)
	}

	var respErr *github.ErrorResponse
	if errors.As(err, &respErr) && respErr.Response != nil {
		switch respErr.Response.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			// 403 without rate-limit headers is typically a permissions issue.
			return fmt.Errorf("%w (%v)", ErrGitHubUnauthorized, err)
		}
	}

	return err
}
