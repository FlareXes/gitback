// internal/mirror/gist.go

package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

func (e *Engine) gistMirrorRoot() string {
	return filepath.Join(
		e.cfg.Storage.MirrorRoot,
		"gists",
	)
}

func (e *Engine) extractGistName(gistURL string) string {

	return strings.TrimSuffix(
		filepath.Base(gistURL),
		".git",
	)
}

func (e *Engine) gistMirrorPath(gistURL string) string {

	id := strings.TrimSuffix(
		filepath.Base(gistURL),
		".git",
	)

	return filepath.Join(
		e.gistMirrorRoot(),
		id+".git",
	)
}

func (e *Engine) syncGist(ctx context.Context, gistURL string) error {

	return e.syncMirror(
		ctx,
		gistURL,
		e.gistMirrorPath(gistURL),
	)
}

func (e *Engine) syncGists(ctx context.Context) ([]state.Asset, error) {

	jobs := make(chan string)
	results := make(chan state.Asset)

	var wg sync.WaitGroup

	e.startWorkers(
		ctx,
		e.syncGist,
		jobs,
		results,
		&wg,
	)

	dispatchErr := make(chan error, 1)

	go func() {
		dispatchErr <- e.dispatchGistJobs(
			jobs,
		)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var gists []state.Asset

	for result := range results {

		gists = append(
			gists,
			result,
		)
	}

	if err := <-dispatchErr; err != nil {
		return nil, err
	}

	return gists, nil
}

func (e *Engine) dispatchGistJobs(jobs chan<- string) error {

	defer close(jobs)

	gists, err := state.ReadInventory(config.GistInventoryFile())

	if err != nil {

		e.logger.Warn(
			logging.Events.Inventory.Missing,
			config.GistInventoryFile(),
			"gist inventory file not found",
		)

		fmt.Println(
			"[WARN] Gist inventory missing. Run: gitback discover",
		)

		return nil
	}

	if len(gists) == 0 {

		e.logger.Warn(
			logging.Events.Inventory.Empty,
			config.GistInventoryFile(),
			"gist inventory file is empty",
		)

		fmt.Println(
			"[WARN] Gist inventory empty. Run: gitback discover",
		)

		return nil
	}

	for _, gist := range gists {

		fmt.Printf("[GIST] %s\n", e.extractGistName(gist))

		jobs <- gist
	}

	return nil
}
