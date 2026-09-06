// internal/mirror/repository.go

package mirror

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

func (e *Engine) repoMirrorRoot() string {
	return filepath.Join(
		e.cfg.Storage.MirrorRoot,
		"repositories",
	)
}

func (e *Engine) extractRepoName(repoURL string) string {

	repo := strings.TrimSuffix(
		repoURL,
		".git",
	)

	parts := strings.Split(repo, "/")

	if len(parts) < 2 {
		return repo
	}

	return fmt.Sprintf(
		"%s/%s",
		parts[len(parts)-2],
		parts[len(parts)-1],
	)
}

func (e *Engine) repositoryMirrorPath(repoURL string) string {

	repo := strings.TrimSuffix(repoURL, ".git")

	parts := strings.Split(repo, "/")

	if len(parts) < 2 {

		return filepath.Join(
			e.repoMirrorRoot(),
			filepath.Base(repoURL),
		)
	}

	owner := parts[len(parts)-2]
	name := parts[len(parts)-1]

	return filepath.Join(
		e.repoMirrorRoot(),
		owner,
		name+".git",
	)
}

func (e *Engine) syncRepository(ctx context.Context, repo string) error {

	return e.syncMirror(
		ctx,
		repo,
		e.repositoryMirrorPath(repo),
	)
}

func (e *Engine) syncRepositories(ctx context.Context) ([]state.Asset, error) {

	jobs := make(chan string)
	results := make(chan state.Asset)

	var wg sync.WaitGroup

	e.startWorkers(
		ctx,
		e.syncRepository,
		jobs,
		results,
		&wg,
	)

	dispatchErr := make(chan error, 1)

	go func() {
		dispatchErr <- e.dispatchRepositoryJobs(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var repositories []state.Asset

	for result := range results {

		repositories = append(
			repositories,
			result,
		)
	}

	if err := <-dispatchErr; err != nil {
		return nil, err
	}

	return repositories, nil
}

// dispatchRepositoryJobs feeds the repository inventory to the worker
// pool. A missing inventory file means discovery hasn't run yet and is
// not an error — sync should still succeed with zero repositories so a
// fresh install doesn't fail before init/discover have been run. Any
// other read error (permissions, corruption) is real and must
// propagate, or sync would silently process zero repositories and
// report success.
func (e *Engine) dispatchRepositoryJobs(jobs chan<- string) error {

	defer close(jobs)

	repositories, err := state.ReadInventory(e.layout.RepositoryInventoryFile)

	if err != nil {

		// If inventory file doesn't exist, return early
		if os.IsNotExist(err) {

			e.logger.Emit(
				logging.Events.Inventory.Missing,
				logging.WithDetails(map[string]any{"inventory_file": e.layout.RepositoryInventoryFile}),
			)

			fmt.Println(
				"[WARN] Repository inventory missing. Run: gitback discover",
			)

			return nil
		}

		// If there's a different error reading the inventory, log it and return
		// Such as: permission denied, file corrupted, etc.
		e.logger.Emit(
			logging.Events.Inventory.ReadFailed,
			logging.WithError(err),
			logging.WithDetails(map[string]any{"inventory_file": e.layout.RepositoryInventoryFile}),
		)

		return fmt.Errorf(
			"read repository inventory %s: %w",
			e.layout.RepositoryInventoryFile,
			err,
		)
	}

	if len(repositories) == 0 {

		e.logger.Emit(
			logging.Events.Inventory.Empty,
			logging.WithDetails(map[string]any{"inventory_file": e.layout.RepositoryInventoryFile}),
		)

		fmt.Println(
			"[WARN] Repository inventory empty. Run: gitback discover",
		)

		return nil
	}

	for _, repo := range repositories {

		jobs <- repo

		fmt.Printf("[REPO] %s\n", e.extractRepoName(repo))
	}

	return nil
}
