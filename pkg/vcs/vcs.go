package vcs

import (
	"context"

	"github.com/etsubu/manticore-scanner/pkg/api"
)

// Context holds VCS-specific metadata.
type Context struct {
	Provider   string // "github", "gitlab", etc.
	Repository string // "owner/repo"
	PRNumber   int
	CommitSHA  string
	Token      string
}

// Provider publishes scan results to a VCS platform.
type Provider interface {
	// Name returns the provider identifier.
	Name() string
	// Detect checks if running in this provider's CI environment and extracts context.
	Detect() (*Context, error)
	// PostResults publishes results to the PR/MR.
	PostResults(ctx context.Context, vcsCtx *Context, results []api.BatchResultItem) error
}
