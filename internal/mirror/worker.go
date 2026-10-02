// internal/mirror/worker.go

package mirror

import (
	"context"
	"sync"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

type SyncFunc func(
	context.Context,
	string,
) error

// buildAsset builds the state.Asset outcome for one sync attempt,
// given its name and whatever error (if any) syncFn returned.
//
// Extracted as its own pure function specifically so this
// classification logic — success vs. genuine failure vs. cancelled —
// can be tested directly, with no channels, no WaitGroup, no
// goroutines at all.
func buildAsset(name string, err error) state.Asset {

	if err == nil {
		return state.Asset{
			Name:        name,
			LastSuccess: true,
		}
	}

	// isCancelled checks the specific error returned, not ctx.Err() at
	// some later point in time — see mirror.go's isCancelled doc
	// comment. A genuine failure on one repo must never be
	// misclassified just because some other repo's cancellation
	// happened to flip a shared context around the same moment.
	cause := logging.Cause("")
	if isCancelled(err) {
		cause = logging.CauseCancelled
	}

	return state.Asset{
		Name:        name,
		LastSuccess: false,
		Cause:       cause,
		Error:       err.Error(),
	}
}

func (e *Engine) worker(
	ctx context.Context,
	syncFn SyncFunc,
	jobs <-chan string,
	results chan<- state.Asset,
	wg *sync.WaitGroup,
) {

	defer wg.Done()

	for asset := range jobs {

		results <- buildAsset(assetName(asset), syncFn(ctx, asset))
	}
}

func (e *Engine) startWorkers(
	ctx context.Context,
	syncFn SyncFunc,
	jobs <-chan string,
	results chan<- state.Asset,
	wg *sync.WaitGroup,
) {

	for i := 0; i < e.cfg.Sync.Workers; i++ {

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
