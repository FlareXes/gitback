// internal/discovery/gist.go

package discovery

import (
	"context"
	"fmt"

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

		fmt.Printf("Fetching gists         (page %d)\n", opt.Page+1)

		gists, resp, err := c.api.Gists.List(
			ctx,
			"",
			opt,
		)

		if err != nil {
			return DiscoverResult{}, fmt.Errorf("list gists page=%d: %w",
				opt.Page,
				err,
			)
		}

		lastResponse = resp

		for _, gist := range gists {

			all = append(
				all,
				gist.GetGitPullURL(),
			)
		}

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
