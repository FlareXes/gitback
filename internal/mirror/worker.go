// internal/mirror/worker.go

package mirror

import (
	"context"
	"fmt"
	"sync"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

func (e *Engine) worker(
	ctx context.Context,
	jobs <-chan string,
	results chan<- state.Asset,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for repo := range jobs {

		fmt.Printf("[REPO] %s\n", e.extractRepoName(repo))

		if err := e.syncRepository(ctx, repo); err != nil {

			e.logger.Error(
				logging.Events.Sync.Failed,
				repo,
				err,
			)

			results <- state.Asset{
				Name:        repo,
				LastSuccess: false,
				Error:       err.Error(),
			}

			continue
		}

		results <- state.Asset{
			Name:        repo,
			LastSuccess: true,
		}
	}
}

func (e *Engine) startWorkers(
	ctx context.Context,
	jobs <-chan string,
	results chan<- state.Asset,
	wg *sync.WaitGroup,
) {

	for i := 0; i < e.cfg.SyncWorkers; i++ {

		wg.Add(1)

		go e.worker(
			ctx,
			jobs,
			results,
			wg,
		)
	}
}

func (e *Engine) dispatchJobs(jobs chan<- string) error {

	defer close(jobs)

	repositories, err := state.ReadInventory(e.cfg.RepositoryInventoryFile())

	if err != nil {

		e.logger.Warn(
			logging.Events.Inventory.Missing,
			e.cfg.RepositoryInventoryFile(),
			"inventory file not found",
		)

		fmt.Println(
			"[WARN] Repository inventory missing. Run: gitback discover",
		)

		return nil
	}

	if len(repositories) == 0 {

		e.logger.Warn(
			logging.Events.Inventory.Empty,
			e.cfg.RepositoryInventoryFile(),
			"inventory file is empty",
		)

		fmt.Println(
			"[WARN] Repository inventory empty. Run: gitback discover",
		)

		return nil
	}

	for _, repo := range repositories {

		jobs <- repo
	}

	return nil
}
