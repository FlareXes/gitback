// internal/mirror/worker.go

package mirror

import (
	"context"
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
