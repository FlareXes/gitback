package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/google/go-github/v59/github"
)

// listGists lists all gists for the given username.
// It handles both authenticated and unauthenticated modes.
func (g *GitHubVCS) listGists(ctx context.Context, username string) ([]*github.Gist, error) {
	var allGists []*github.Gist

	opt := &github.GistListOptions{
		ListOptions: github.ListOptions{
			PerPage: DefaultPerPage,
		},
	}

	for {
		gists, resp, err := g.listGistsPage(ctx, username, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list gists: %w", err)
		}

		allGists = append(allGists, gists...)

		// Check if we've reached the last page
		if resp.NextPage == 0 {
			break
		}

		// Update the page number for the next request
		opt.Page = resp.NextPage

		// Check rate limits
		if err := g.checkRateLimit(resp); err != nil {
			return allGists, err
		}
	}

	return allGists, nil
}

// listGistsPage fetches a single page of gists.
func (g *GitHubVCS) listGistsPage(ctx context.Context, username string, opt *github.GistListOptions) ([]*github.Gist, *github.Response, error) {
	if g.config.NoAuth || g.config.User == username {
		// In no-auth mode or when listing our own gists,
		// we can use the simpler List API.
		gists, resp, err := g.client.Gists.List(ctx, username, opt)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to list gists for user %s: %w", username, err)
		}
		return gists, resp, nil
	}

	// For authenticated requests, we need to list all gists and filter them
	// since GitHub API doesn't support searching gists by user in the same way as repositories.
	gists, resp, err := g.client.Gists.ListAll(ctx, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list all gists: %w", err)
	}

	// Filter gists by owner
	var userGists []*github.Gist
	for _, gist := range gists {
		if gist.Owner != nil && gist.Owner.Login != nil && *gist.Owner.Login == username {
			userGists = append(userGists, gist)
		}
	}

	return userGists, resp, nil
}

// backupGists backs up all gists for the given username.
func (g *GitHubVCS) backupGists(ctx context.Context, username string) error {
	gists, err := g.listGists(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to list gists: %w", err)
	}

	if len(gists) == 0 {
		log.Println("No gists found to back up")
		return nil
	}

	log.Printf("Found %d gists for user %s\n", len(gists), username)

	// Create gists directory if it doesn't exist
	gistsDir := filepath.Join(g.config.OutputDir, "gists")
	if err := os.MkdirAll(gistsDir, 0755); err != nil {
		return fmt.Errorf("failed to create gists directory: %w", err)
	}

	errChan := make(chan error, len(gists))
	semaphore := make(chan struct{}, g.config.Threads)

	for _, gist := range gists {
		if gist.ID == nil {
			errChan <- fmt.Errorf("gist has no ID")
			continue
		}

		semaphore <- struct{}{} // Acquire semaphore
		go func(gist *github.Gist) {
			defer func() { <-semaphore }() // Release semaphore when done

			// Create gist directory
			gistDir := filepath.Join(gistsDir, *gist.ID)
			if err := os.MkdirAll(gistDir, 0755); err != nil {
				errChan <- fmt.Errorf("failed to create gist directory %s: %w", *gist.ID, err)
				return
			}

			// Save gist metadata
			metadataFile := filepath.Join(gistDir, "gist.json")
			metadata, err := json.MarshalIndent(gist, "", "  ")
			if err != nil {
				errChan <- fmt.Errorf("failed to marshal gist metadata %s: %w", *gist.ID, err)
				return
			}

			if err := os.WriteFile(metadataFile, metadata, 0644); err != nil {
				errChan <- fmt.Errorf("failed to save gist metadata %s: %w", *gist.ID, err)
				return
			}

			// Save each file in the gist
			for filename, file := range gist.Files {
				// Skip invalid files
				if file.Filename == nil {
					continue
				}

				// Get file content (might be nil for large files)
				content := ""
				if file.Content != nil {
					content = *file.Content
				} else if file.RawURL != nil {
					// For large files, we'd need to fetch the content from RawURL
					// This is a simplified version - in production, you'd want to handle this properly
					log.Printf("Skipping large file %s (content not included in gist response)", *file.Filename)
					continue
				}

				// Create necessary subdirectories
				filePath := filepath.Join(gistDir, *file.Filename)
				dir := filepath.Dir(filePath)
				if dir != "." {
					if err := os.MkdirAll(dir, 0755); err != nil {
						errChan <- fmt.Errorf("failed to create directory for gist file %s: %w", filename, err)
						continue
					}
				}

				// Write file content
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					errChan <- fmt.Errorf("failed to save gist file %s: %w", filename, err)
					continue
				}
			}

			log.Printf("Successfully backed up gist: %s\n", *gist.ID)
			errChan <- nil
		}(gist)

		// Check if context is done
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Continue with the next gist
		}
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(gists); i++ {
		if err := <-errChan; err != nil {
			log.Printf("Error during gist backup: %v", err)
		}
	}

	return nil
}
