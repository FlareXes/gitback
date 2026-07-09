// internal/mirror/worker.go

package mirror

import (
	"context"
	"sync"

	"github.com/flarexes/gitback/internal/state"
)

type SyncFunc func(
	context.Context,
	string,
) error

func (e *Engine) worker(
	ctx context.Context,
	syncFn SyncFunc,
	jobs <-chan string,
	results chan<- state.Asset,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for asset := range jobs {

		if err := syncFn(ctx, asset); err != nil {

			results <- state.Asset{
				Name:        asset,
				LastSuccess: false,
				Error:       err.Error(),
			}

			continue
		}

		results <- state.Asset{
			Name:        asset,
			LastSuccess: true,
		}
	}
}

func (e *Engine) startWorkers(
	ctx context.Context,
	syncFn SyncFunc,
	jobs <-chan string,
	results chan<- state.Asset,
	wg *sync.WaitGroup,
) {

	for i := 0; i < e.cfg.SyncWorkers; i++ {

		wg.Add(1)

		go e.worker(
			ctx,
			syncFn,
			jobs,
			results,
			wg,
		)
	}
}
