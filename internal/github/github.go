// internal/github/github.go

package github

import (
	"context"
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/google/go-github/v88/github"
)

type Client struct {
	cfg    *config.Config
	logger *logging.Logger
	api    *github.Client
}

func New(cfg *config.Config, logger *logging.Logger) (*Client, error) {

	if cfg.GitHubToken == "" {
		return nil, fmt.Errorf("github token not configured; run: gitback init")
	}

	api, err := github.NewClient(
		github.WithAuthToken(
			cfg.GitHubToken,
		),
	)

	if err != nil {
		return nil, err
	}

	return &Client{
		cfg:    cfg,
		logger: logger,
		api:    api,
	}, nil
}

func (c *Client) Discover(ctx context.Context) error {

	var all []string

	var lastResponse *github.Response

	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",

		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	// GitHub pagination: GitHub only returns a limited number of repos per request.
	for {

		fmt.Printf("Fetching repositories (page %d)\n", opt.Page+1)

		repos, resp, err := c.api.Repositories.ListByAuthenticatedUser(
			ctx,
			opt,
		)

		if err != nil {

			return fmt.Errorf("list repositories page=%d: %w",
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

		// No more pages
		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	// Write repository inventory file.
	// This file is consumed later by the sync engine.
	f, err := os.OpenFile(
		c.cfg.RepoInventory,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)
	if err != nil {

		return fmt.Errorf("open inventory file %s: %w",
			c.cfg.RepoInventory,
			err,
		)
	}

	defer f.Close()

	for _, repo := range all {

		_, err := fmt.Fprintln(
			f,
			repo,
		)

		if err != nil {

			return fmt.Errorf("write inventory file %s: %w",
				c.cfg.RepoInventory,
				err,
			)
		}
	}

	// Print total number of repositories discovered
	fmt.Printf("Discovered %d repositories\n", len(all))

	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.DiscoveryCompleted,

			Details: map[string]any{
				"repo_count": len(all),
			},
		},
	)

	// Log remaining API quota.
	if lastResponse != nil {

		c.logger.Emit(
			logging.Entry{
				Level: logging.Info,
				Event: "github_rate_limit",

				Details: map[string]any{
					"limit":     lastResponse.Rate.Limit,
					"remaining": lastResponse.Rate.Remaining,
				},
			},
		)
	}

	return nil
}
