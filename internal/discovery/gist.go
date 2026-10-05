// internal/discovery/gist.go

package discovery

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/ghauth"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/google/go-github/v88/github"
)

func (c *Client) discoverGists(ctx context.Context) (DiscoverResult, error) {

	var all []string
	var lastResponse *github.Response

	opt := &github.GistListOptions{

		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {

		page := opt.Page + 1

		fmt.Printf("Fetching gists         (page %d)\n", page)

		gists, resp, err := c.api.Gists.List(ctx, "", opt)

		if err != nil {

			// Give a real explanation when the failure is GitHub
			// rejecting the token, rather than letting go-github's raw
			// error text (e.g. "Bad credentials").
			if msg, ok := ghauth.Diagnose(err, resp); ok {
				return DiscoverResult{}, fmt.Errorf("list gists page=%d: %s", page, msg)
			}

			return DiscoverResult{}, fmt.Errorf("list gists page=%d: %w", page, err)
		}

		lastResponse = resp

		for _, gist := range gists {

			all = append(
				all,
				gist.GetGitPullURL(),
			)
		}

		c.logger.Emit(
			logging.Events.GitHub.PageFetched,
			logging.WithDetails(map[string]any{
				"resource":     "gists",
				"page":         page,
				"items":        len(gists),
				"total_so_far": len(all),
			}),
		)

		// No more pages
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return DiscoverResult{
		URLs:      all,
		RateLimit: lastResponse.Rate,
	}, nil
}
