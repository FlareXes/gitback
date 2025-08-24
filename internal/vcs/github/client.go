// Package github provides GitHub-specific VCS functionality.
package github

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v59/github"
	"golang.org/x/oauth2"

	"github.com/flarexes/gitback/internal/types"
)

const (
	// DefaultPerPage is the default number of items to request per page.
	DefaultPerPage = 100
	// DefaultTimeout is the default timeout for API requests.
	DefaultTimeout = 30 * time.Second
)

// GitHubVCS implements the VCS interface for GitHub.
type GitHubVCS struct {
	client *github.Client
	config *types.Config
}

// NewGitHubVCS creates a new GitHubVCS instance.
func NewGitHubVCS(cfg *types.Config) (*GitHubVCS, error) {
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	// If we have a token, use it for authentication
	if cfg.Token != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.Token},
		)
		httpClient.Transport = &oauth2.Transport{
			Source: ts,
		}
	}

	client := github.NewClient(httpClient)

	return &GitHubVCS{
		client: client,
		config: cfg,
	}, nil
}

// getUsername returns the username to use for API calls.
// If we're in no-auth mode, we use the provided username.
// Otherwise, we try to get the authenticated user's username.
func (g *GitHubVCS) getUsername(ctx context.Context) (string, error) {
	if g.config.NoAuth {
		if g.config.User == "" {
			return "", fmt.Errorf("username is required in no-auth mode")
		}
		return g.config.User, nil
	}

	// Get the authenticated user
	user, _, err := g.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}

	if user.Login == nil {
		return "", fmt.Errorf("no username found for authenticated user")
	}

	return *user.Login, nil
}

// Backup performs the backup of repositories and gists.
func (g *GitHubVCS) Backup(ctx context.Context) error {
	username, err := g.getUsername(ctx)
	if err != nil {
		return fmt.Errorf("failed to get username: %w", err)
	}

	// Backup repositories
	if err := g.backupRepositories(ctx, username); err != nil {
		return fmt.Errorf("failed to backup repositories: %w", err)
	}

	// Backup gists if configured to do so
	if g.config.IncludeGists {
		if err := g.backupGists(ctx, username); err != nil {
			return fmt.Errorf("failed to backup gists: %w", err)
		}
	}

	return nil
}
