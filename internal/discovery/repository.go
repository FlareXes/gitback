// internal/discovery/discover.go

package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/flarexes/gitback/internal/clock"
	"github.com/flarexes/gitback/internal/ghauth"
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

			// Give a real explanation when the failure is GitHub
			// rejecting the token, rather than letting go-github's raw
			// error text (e.g. "Bad credentials").
			if msg, ok := ghauth.Diagnose(err, resp); ok {
				return DiscoverResult{}, fmt.Errorf("list repositories page=%d: %s", page, msg)
			}

			return DiscoverResult{}, fmt.Errorf("list repositories page=%d: %w", page, err)
		}

		lastResponse = resp

		for _, repo := range repos {

			all = append(
				all,
				repo.GetCloneURL(),
			)
		}

		c.logger.Emit(
			logging.Events.GitHub.PageFetched,
			logging.WithDetails(map[string]any{
				"resource":     "repositories",
				"page":         page,
				"items":        len(repos),
				"total_so_far": len(all),
			}),
		)

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	// Proactive expiry warning
	if status, expiresAt := ghauth.CheckExpiration(lastResponse); status == ghauth.ExpiringSoon || status == ghauth.Expired {

		days := int(time.Until(expiresAt).Hours() / 24)

		fmt.Printf(
			"[WARN] GitHub token expires on %s (in %d day(s)). Run `gitback init --force` before then.\n",
			clock.HumanDate(expiresAt), days,
		)

		c.logger.Emit(
			logging.Events.GitHub.TokenExpiringSoon,
			logging.WithDetails(map[string]any{
				"expires_at":     clock.LocalRFC3339(expiresAt),
				"days_remaining": days,
			}),
		)
	}

	return DiscoverResult{
		URLs:      all,
		RateLimit: lastResponse.Rate,
	}, nil
}
