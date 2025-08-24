// Package github provides GitHub-specific VCS functionality.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	repos, err := g.listRepositories(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	// Backup gists if configured to do so
	if g.config.IncludeGists {
		gists, err := g.listGists(ctx, username)
		if err != nil {
			return fmt.Errorf("failed to list gists: %w", err)
		}
		log.Printf("Found %d gists for user %s\n", len(gists), username)

		// Create the gists directory if it doesn't exist
		gistsDir := filepath.Join(g.config.OutputDir, "gists")
		if err := os.MkdirAll(gistsDir, 0755); err != nil {
			return fmt.Errorf("failed to create gists directory: %w", err)
		}

		// Backup gists
		gistErrChan := make(chan error, len(gists))
		gistSemaphore := make(chan struct{}, g.config.Threads)

		for _, gist := range gists {
			if gist.ID == nil {
				gistErrChan <- fmt.Errorf("gist has no ID")
				continue
			}

			gistSemaphore <- struct{}{} // Acquire semaphore
			go func(gist *github.Gist) {
				defer func() { <-gistSemaphore }() // Release semaphore when done

				// Create gist directory
				gistDir := filepath.Join(gistsDir, *gist.ID)
				if err := os.MkdirAll(gistDir, 0755); err != nil {
					gistErrChan <- fmt.Errorf("failed to create gist directory %s: %w", *gist.ID, err)
					return
				}

				// Save gist metadata
				metadataFile := filepath.Join(gistDir, "gist.json")
				metadata, err := json.MarshalIndent(gist, "", "  ")
				if err != nil {
					gistErrChan <- fmt.Errorf("failed to marshal gist metadata %s: %w", *gist.ID, err)
					return
				}

				if err := os.WriteFile(metadataFile, metadata, 0644); err != nil {
					gistErrChan <- fmt.Errorf("failed to save gist metadata %s: %w", *gist.ID, err)
					return
				}

				// Save each file in the gist
				for _, file := range gist.Files {
					// Skip files without a filename
					if file.Filename == nil {
						log.Printf("Skipping gist file with no filename in gist %s", *gist.ID)
						continue
					}

					// Get the filename safely
					filenameStr := *file.Filename

					// Get file content (might be nil for large files)
					var content string
					if file.Content != nil {
						content = *file.Content
					} else if file.RawURL != nil {
						// Try to fetch the content from RawURL
						resp, err := http.Get(*file.RawURL)
						if err != nil {
							log.Printf("Failed to fetch content for large file %s: %v", filenameStr, err)
							continue
						}
						defer resp.Body.Close()

						if resp.StatusCode != http.StatusOK {
							log.Printf("Failed to fetch content for %s: %s", filenameStr, resp.Status)
							continue
						}

						contentBytes, err := io.ReadAll(resp.Body)
						if err != nil {
							log.Printf("Failed to read content for %s: %v", filenameStr, err)
							continue
						}
						content = string(contentBytes)
					}

					// Create necessary subdirectories
					filePath := filepath.Join(gistDir, filenameStr)
					dir := filepath.Dir(filePath)
					if dir != "." {
						if err := os.MkdirAll(dir, 0755); err != nil {
							gistErrChan <- fmt.Errorf("failed to create directory for gist file %s: %w", filenameStr, err)
							continue
						}
					}

					// Write file content
					if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
						gistErrChan <- fmt.Errorf("failed to save gist file %s: %w", filenameStr, err)
						continue
					}
				}

				log.Printf("Successfully backed up gist: %s\n", *gist.ID)
				gistErrChan <- nil
			}(gist)

			// Check if context is done
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Continue with the next gist
			}
		}

		// Wait for all gist backups to complete
		for i := 0; i < len(gists); i++ {
			if err := <-gistErrChan; err != nil {
				log.Printf("Error during gist backup: %v", err)
			}
		}
	}

	log.Printf("Found %d repositories for user %s\n", len(repos), username)

	// Create repositories directory
	reposDir := filepath.Join(g.config.OutputDir, "repos")
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return fmt.Errorf("failed to create repositories directory: %w", err)
	}

	// Backup repositories
	errChan := make(chan error, len(repos))
	semaphore := make(chan struct{}, g.config.Threads)

	for _, repo := range repos {
		semaphore <- struct{}{} // Acquire semaphore
		go func(r *github.Repository) {
			defer func() { <-semaphore }() // Release semaphore when done

			if r.Name == nil {
				errChan <- fmt.Errorf("repository has no name")
				return
			}

			if err := g.cloneRepository(ctx, r, reposDir); err != nil {
				errChan <- fmt.Errorf("failed to backup repository %s: %w", *r.Name, err)
				return
			}

			errChan <- nil
		}(repo)
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(repos); i++ {
		if err := <-errChan; err != nil {
			log.Printf("Error during backup: %v", err)
		}
	}

	return nil
}
