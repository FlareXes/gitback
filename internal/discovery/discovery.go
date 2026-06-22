// internal/discover/discovery.go

package discovery

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
	"github.com/google/go-github/v88/github"
)

type DiscoverResult struct {
	URLs      []string
	RateLimit github.Rate
}

func (c *Client) Discover(ctx context.Context) error {

	// Repository
	result, err := c.discoverRepositories(ctx)

	if err != nil {
		return err
	}

	repoCount := len(result.URLs)

	// Save repository URLs to inventory file
	if err := state.WriteInventory(
		c.cfg.RepositoryInventoryFile(),
		result.URLs,
	); err != nil {

		return err
	}

	// Log discovery completion
	c.logDiscovery("repositories", repoCount, c.cfg.RepositoryInventoryFile(), result.RateLimit)

	// Gist
	gistCount := 0

	if c.cfg.BackupGists {

		result, err := c.discoverGists(ctx)

		if err != nil {
			return err
		}

		gistCount = len(result.URLs)

		// Save gist URLs to inventory file
		if err := state.WriteInventory(
			c.cfg.GistInventoryFile(),
			result.URLs,
		); err != nil {

			return err
		}

		// Log gist completion
		c.logDiscovery("gists", gistCount, c.cfg.GistInventoryFile(), result.RateLimit)
	}

	fmt.Println()
	fmt.Println("Repository: ", repoCount)

	if c.cfg.BackupGists {
		fmt.Println("Gist:       ", gistCount)
	}

	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.DiscoverySummary,

			Details: map[string]any{
				"repositories": repoCount,
				"gists":        gistCount,
				"total":        repoCount + gistCount,
			},
		},
	)

	return nil
}

func (c *Client) logDiscovery(
	resource string,
	count int,
	inventoryPath string,
	rate github.Rate,
) {

	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.InventoryLoaded,

			Details: map[string]any{
				"resource": resource,
				"count":    count,
				"path":     inventoryPath,
			},
		},
	)

	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.DiscoveryCompleted,

			Details: map[string]any{
				"resource": resource,
				"count":    count,
			},
		},
	)

	c.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.GitHub.RateLimit,

			Details: map[string]any{
				"resource":  resource,
				"limit":     rate.Limit,
				"remaining": rate.Remaining,
			},
		},
	)
}
