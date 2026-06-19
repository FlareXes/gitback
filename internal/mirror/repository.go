// internal/mirror/repository.go

package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

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
			e.cfg.RepositoryMirrorDir(),
			filepath.Base(repoURL),
		)
	}

	owner := parts[len(parts)-2]
	name := parts[len(parts)-1]

	return filepath.Join(
		e.cfg.RepositoryMirrorDir(),
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

func (e *Engine) dispatchRepositoryJobs(jobs chan<- string) error {

	defer close(jobs)

	repositories, err := state.ReadInventory(e.cfg.RepositoryInventoryFile())

	if err != nil {

		e.logger.Warn(
			logging.Events.Inventory.Missing,
			e.cfg.RepositoryInventoryFile(),
			"repository inventory file not found",
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

		fmt.Printf("[REPO] %s\n", e.extractRepoName(repo))

		jobs <- repo
	}

	return nil
}
