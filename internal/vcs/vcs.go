// Package vcs provides version control system operations for backing up repositories.
package vcs

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/types"
	"github.com/flarexes/gitback/internal/vcs/github"
)

// VCS defines the interface for version control system operations.
type VCS interface {
	// Backup performs the backup operation for repositories and gists.
	Backup(ctx context.Context) error
}

// New creates a new VCS instance based on the provided configuration.
func New(cfg *types.Config) (VCS, error) {
	// For now, we only support GitHub
	return github.NewGitHubVCS(cfg)
}

// Run executes the backup operation with the provided configuration.
func Run(ctx context.Context, cfg *types.Config) error {
	vcs, err := New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize VCS: %w", err)
	}

	return vcs.Backup(ctx)
}
