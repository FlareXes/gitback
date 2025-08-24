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

// backupGists orchestrates the backup of all gists for a user.
func (g *GitHubVCS) backupGists(ctx context.Context, username string) error {
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

	// Backup gists concurrently
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

			if err := g.backupSingleGist(gist, gistsDir); err != nil {
				gistErrChan <- fmt.Errorf("failed to backup gist %s: %w", *gist.ID, err)
				return
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

	return nil
}

// backupSingleGist backs up a single gist to the specified directory.
func (g *GitHubVCS) backupSingleGist(gist *github.Gist, baseDir string) error {
	if gist.ID == nil {
		return fmt.Errorf("gist has no ID")
	}

	// Create gist directory
	gistDir := filepath.Join(baseDir, *gist.ID)
	if err := os.MkdirAll(gistDir, 0755); err != nil {
		return fmt.Errorf("failed to create gist directory %s: %w", *gist.ID, err)
	}

	// Save gist metadata
	if err := g.saveGistMetadata(gist, gistDir); err != nil {
		return fmt.Errorf("failed to save gist metadata: %w", err)
	}

	// Save each file in the gist
	if err := g.saveGistFiles(gist, gistDir); err != nil {
		return fmt.Errorf("failed to save gist files: %w", err)
	}

	return nil
}

// saveGistMetadata saves the gist metadata as a JSON file.
func (g *GitHubVCS) saveGistMetadata(gist *github.Gist, gistDir string) error {
	metadataFile := filepath.Join(gistDir, "gist.json")
	metadata, err := json.MarshalIndent(gist, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal gist metadata: %w", err)
	}

	if err := os.WriteFile(metadataFile, metadata, 0644); err != nil {
		return fmt.Errorf("failed to save gist metadata: %w", err)
	}

	return nil
}

// saveGistFiles saves all files in a gist to the specified directory.
func (g *GitHubVCS) saveGistFiles(gist *github.Gist, gistDir string) error {
	for _, file := range gist.Files {
		// Skip files without a filename
		if file.Filename == nil {
			log.Printf("Skipping gist file with no filename in gist %s", *gist.ID)
			continue
		}

		if err := g.saveGistFile(file, gistDir); err != nil {
			log.Printf("Failed to save gist file %s: %v", *file.Filename, err)
			continue
		}
	}

	return nil
}

// saveGistFile saves a single gist file to the specified directory.
func (g *GitHubVCS) saveGistFile(file github.GistFile, gistDir string) error {
	// Get the filename safely
	filenameStr := *file.Filename

	// Get file content (might be nil for large files)
	content, err := g.getGistFileContent(file)
	if err != nil {
		return fmt.Errorf("failed to get content for file %s: %w", filenameStr, err)
	}

	// Create necessary subdirectories
	filePath := filepath.Join(gistDir, filenameStr)
	dir := filepath.Dir(filePath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for gist file %s: %w", filenameStr, err)
		}
	}

	// Write file content
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save gist file %s: %w", filenameStr, err)
	}

	return nil
}

// getGistFileContent retrieves the content of a gist file.
func (g *GitHubVCS) getGistFileContent(file github.GistFile) (string, error) {
	// Get file content (might be nil for large files)
	if file.Content != nil {
		return *file.Content, nil
	}

	if file.RawURL != nil {
		// Try to fetch the content from RawURL
		resp, err := http.Get(*file.RawURL)
		if err != nil {
			return "", fmt.Errorf("failed to fetch content for large file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to fetch content: %s", resp.Status)
		}

		contentBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read content: %w", err)
		}

		return string(contentBytes), nil
	}

	return "", fmt.Errorf("no content available for file")
}
