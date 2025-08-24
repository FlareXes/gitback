package github

import (
	"context"
	"fmt"

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
