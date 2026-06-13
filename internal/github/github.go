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

// func (c *Client) Discover(ctx context.Context) error {

// 	var all []string

// 	var lastResponse *github.Response

// 	opt := &github.RepositoryListByAuthenticatedUserOptions{
// 		Visibility: "all",

// 		ListOptions: github.ListOptions{
// 			PerPage: 100,
// 		},
// 	}

// 	// GitHub pagination: GitHub only returns a limited number of repos per request.
// 	for {

// 		fmt.Printf("Fetching repositories (page %d)\n", opt.Page+1)

// 		repos, resp, err := c.api.Repositories.ListByAuthenticatedUser(
// 			ctx,
// 			opt,
// 		)

// 		if err != nil {

// 			return fmt.Errorf("list repositories page=%d: %w",
// 				opt.Page,
// 				err,
// 			)
// 		}

// 		lastResponse = resp

// 		for _, repo := range repos {
// 			all = append(
// 				all,
// 				repo.GetCloneURL(),
// 			)
// 		}

// 		// No more pages
// 		if resp.NextPage == 0 {
// 			break
// 		}

// 		opt.Page = resp.NextPage
// 	}

// 	// Write repository inventory file.
// 	// This file is consumed later by the sync engine.
// 	f, err := os.OpenFile(
// 		c.cfg.RepositoryInventoryFile(),
// 		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
// 		0600,
// 	)
// 	if err != nil {

// 		return fmt.Errorf("open inventory file %s: %w",
// 			c.cfg.RepositoryInventoryFile(),
// 			err,
// 		)
// 	}

// 	defer f.Close()

// 	for _, repo := range all {

// 		_, err := fmt.Fprintln(
// 			f,
// 			repo,
// 		)

// 		if err != nil {

// 			return fmt.Errorf("write inventory file %s: %w",
// 				c.cfg.RepositoryInventoryFile(),
// 				err,
// 			)
// 		}
// 	}

// 	// Print total number of repositories discovered
// 	fmt.Printf("Discovered %d repositories\n", len(all))

// 	c.logger.Emit(
// 		logging.Entry{
// 			Level: logging.Info,
// 			Event: logging.Events.GitHub.DiscoveryCompleted,

// 			Details: map[string]any{
// 				"repo_count": len(all),
// 			},
// 		},
// 	)

// 	// Log remaining API quota.
// 	if lastResponse != nil {

// 		c.logger.Emit(
// 			logging.Entry{
// 				Level: logging.Info,
// 				Event: "github_rate_limit",

// 				Details: map[string]any{
// 					"limit":     lastResponse.Rate.Limit,
// 					"remaining": lastResponse.Rate.Remaining,
// 				},
// 			},
// 		)
// 	}

// 	return nil
// }

type DiscoverResult struct {
	Items     []string
	RateLimit github.Rate
}

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

		fmt.Printf("Fetching repositories (page %d)\n", opt.Page+1)

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

		if resp.NextPage == 0 {
			break
		}

		opt.Page = resp.NextPage
	}

	return DiscoverResult{
		Items:     all,
		RateLimit: lastResponse.Rate,
	}, nil
}

func (c *Client) discoverGists(ctx context.Context) (DiscoverResult, error) {

	var all []string
	var lastResponse *github.Response

	opt := &github.GistListOptions{

		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {

		fmt.Printf("Fetching gists (page %d)\n", opt.Page+1)

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
		Items:     all,
		RateLimit: lastResponse.Rate,
	}, nil
}

func writeInventory(path string, items []string) error {

	f, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0600,
	)

	if err != nil {
		return fmt.Errorf("open inventory file %s: %w",
			path,
			err,
		)
	}

	defer f.Close()

	for _, item := range items {

		if _, err := fmt.Fprintln(f, item); err != nil {

			return fmt.Errorf("write inventory file %s: %w",
				path,
				err,
			)
		}
	}

	return nil
}

func (c *Client) Discover(ctx context.Context) error {

	// Repository
	result, err := c.discoverRepositories(ctx)

	if err != nil {
		return err
	}

	if err := writeInventory(
		c.cfg.RepositoryInventoryFile(),
		result.Items,
	); err != nil {

		return err
	}

	// Log discovery completion
	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.DiscoveryCompleted,

			Details: map[string]any{
				"resource":   "repositories",
				"repo_count": len(result.Items),
			},
		},
	)

	// Log remaining API quota
	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.RateLimit,

			Details: map[string]any{
				"resource":  "repositories",
				"limit":     result.RateLimit.Limit,
				"remaining": result.RateLimit.Remaining,
			},
		},
	)

	// Gist
	gistCount := 0

	if c.cfg.BackupGists {

		gists, err := c.discoverGists(ctx)

		if err != nil {
			return err
		}

		if err := writeInventory(
			c.cfg.GistInventoryFile(),
			gists.Items,
		); err != nil {

			return err
		}

		gistCount = len(gists.Items)
	}

	fmt.Printf(
		"Discovered %d repositories and %d gists\n",
		len(result.Items),
		gistCount,
	)

	// Log gist completion
	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.DiscoveryCompleted,

			Details: map[string]any{
				"resource":   "gists",
				"gist_count": gistCount,
			},
		},
	)

	// Log remaining API quota
	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.RateLimit,

			Details: map[string]any{
				"resource":  "gists",
				"limit":     result.RateLimit.Limit,
				"remaining": result.RateLimit.Remaining,
			},
		},
	)

	return nil
}
