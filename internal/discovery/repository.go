// internal/discovery/discover.go

package discovery

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/google/go-github/v88/github"
)

func (c *Client) discoverRepositories(ctx context.Context) (DiscoverResult, error) {

	var all []string
	var lastResponse *github.Response

	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {

		page := opt.Page + 1

		fmt.Printf("Fetching repositories  (page %d)\n", page)

		repos, resp, err := c.api.Repositories.ListByAuthenticatedUser(
			ctx,
			opt,
		)

		if err != nil {

			return DiscoverResult{}, fmt.Errorf("list repositories page=%d: %w",
				opt.Page,
				err,
			)
		}

		lastResponse = resp

		for _, repo := range repos {

			all = append(
				all,
				repo.GetCloneURL(),
			)
		}

		c.logger.Emit(
			logging.Entry{
				Level: logging.Info,
				Event: logging.Events.GitHub.PageFetched,

				Details: map[string]any{
					"resource":     "repositories",
					"page":         page,
					"items":        len(repos),
					"total_so_far": len(all),
				},
			},
		)

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
