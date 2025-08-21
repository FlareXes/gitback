// Package github provides GitHub-specific VCS functionality.
package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/go-github/v59/github"
)

// listRepositories lists all repositories for the given username.
func (g *GitHubVCS) listRepositories(ctx context.Context, username string) ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opts := &github.RepositoryListOptions{
		Type:        "all",
		Sort:        "pushed",
		Direction:   "desc",
		ListOptions: github.ListOptions{PerPage: DefaultPerPage},
	}

	for {
		repos, resp, err := g.client.Repositories.List(ctx, username, opts)
		if err != nil {
			return nil, fmt.Errorf("error listing repositories: %w", err)
		}

		allRepos = append(allRepos, repos...)

		// Check if we've reached the last page
		if resp.NextPage == 0 {
			break
		}

		// Check rate limits
		if err := g.checkRateLimit(resp); err != nil {
			return allRepos, err
		}

		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

// checkRateLimit checks the rate limit and sleeps if necessary.
func (g *GitHubVCS) checkRateLimit(resp *github.Response) error {
	if resp.Rate.Remaining <= 10 {
		sleepDuration := time.Until(resp.Rate.Reset.Time) + (10 * time.Second)
		log.Printf("Approaching rate limit, sleeping for %v until %v\n",
			sleepDuration,
			time.Now().Add(sleepDuration).Format(time.RFC3339))
		time.Sleep(sleepDuration)
	}
	return nil
}

// cloneRepository clones a single repository to the specified directory.
func (g *GitHubVCS) cloneRepository(ctx context.Context, repo *github.Repository, baseDir string) error {
	if repo == nil || repo.CloneURL == nil || repo.Name == nil {
		return fmt.Errorf("invalid repository")
	}

	targetDir := filepath.Join(baseDir, *repo.Name)

	// Check if directory already exists
	if _, err := os.Stat(targetDir); err == nil {
		log.Printf("Repository %s already exists, pulling latest changes\n", *repo.Name)
		return g.pullRepository(targetDir)
	}

	// Create the base directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", baseDir, err)
	}

	// Build the git clone command
	cloneURL := *repo.CloneURL
	if g.config.Token != "" {
		// Use SSH URL if we have a token (for private repos)
		if repo.SSHURL != nil {
			cloneURL = *repo.SSHURL
		}
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", cloneURL, *repo.Name)
	cmd.Dir = baseDir

	// Set up output capture
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository %s: %w\nOutput: %s", *repo.Name, err, string(output))
	}

	log.Printf("Successfully cloned %s to %s\n", *repo.Name, targetDir)
	return nil
}

// pullRepository updates an existing repository
func (g *GitHubVCS) pullRepository(repoDir string) error {
	// Change to the repository directory
	cmd := exec.Command("git", "remote", "update", "--prune")
	cmd.Dir = repoDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update repository: %w\nOutput: %s", err, string(output))
	}

	log.Printf("Successfully updated repository at %s\n", repoDir)
	return nil
}
