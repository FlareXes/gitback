// internal/mirror/mirror.go

package mirror

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

type Engine struct {
	cfg    *config.Config
	logger *logging.Logger
}

func New(cfg *config.Config, logger *logging.Logger) *Engine {
	return &Engine{
		cfg:    cfg,
		logger: logger,
	}
}

// Execute a git command with retry support.
func (e *Engine) runGit(
	ctx context.Context,
	repo string,
	env []string,
	args ...string,
) ([]byte, error) {

	var lastErr error

	for attempt := 1; attempt <= e.cfg.GitRetryAttempts; attempt++ {

		cmd := exec.CommandContext(
			ctx,
			"git",
			args...,
		)

		cmd.Env = env

		output, err := cmd.CombinedOutput()

		if err == nil {
			return output, nil
		}

		lastErr = err

		if attempt == e.cfg.GitRetryAttempts {
			break
		}

		e.logger.Emit(
			logging.Entry{
				Level: logging.Warn,
				Event: logging.Events.Mirror.Retry,

				Repo: repo,

				Details: map[string]any{
					"attempt":      attempt,
					"max_attempts": e.cfg.GitRetryAttempts,
				},
			},
		)

		// Linear backoff: attempt 1 -> 5s, attempt 2 -> 10s
		wait := time.Duration(attempt*5) * time.Second

		time.Sleep(wait)
	}

	return nil, lastErr
}

func (e *Engine) Sync(ctx context.Context) error {

	syncStartedAt := time.Now()

	repositories, err := e.syncRepositories(
		ctx,
	)

	if err != nil {
		return err
	}

	// Gists
	gists, err := e.syncGists(ctx)

	if err != nil {
		return err
	}

	printSyncSummary("Repositories", repositories)
	printSyncSummary("Gists", gists)

	syncCompletedAt := time.Now()

	if err := state.Save(
		e.cfg.MirrorsStateFile,
		syncStartedAt,
		syncCompletedAt,
		repositories,
		gists,
	); err != nil {

		e.logger.Error(
			logging.Events.Mirror.StateSaveFailed,
			"",
			err,
		)
	}

	return nil
}

func (e *Engine) syncRepositories(ctx context.Context) ([]state.Asset, error) {

	jobs := make(chan string)
	results := make(chan state.Asset)

	var wg sync.WaitGroup

	e.startWorkers(
		ctx,
		jobs,
		results,
		&wg,
	)

	dispatchErr := make(chan error, 1)

	go func() {
		dispatchErr <- e.dispatchJobs(jobs)
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

func (e *Engine) syncGists(ctx context.Context) ([]state.Asset, error) {

	file, err := os.Open(e.cfg.GistInventoryFile())

	if os.IsNotExist(err) {

		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var gists []state.Asset

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		gist := strings.TrimSpace(
			scanner.Text(),
		)

		if gist == "" {
			continue
		}

		fmt.Printf("[GIST] %s\n", e.extractGistName(gist))

		if err := e.syncGist(ctx, gist); err != nil {

			gists = append(
				gists,
				state.Asset{
					Name:        gist,
					LastSuccess: false,
					Error:       err.Error(),
				},
			)

			continue
		}

		gists = append(
			gists,
			state.Asset{
				Name:        gist,
				LastSuccess: true,
			},
		)
	}

	return gists, scanner.Err()
}

func (e *Engine) syncGist(ctx context.Context, gist string) error {

	gistID := strings.TrimSuffix(filepath.Base(gist), ".git")
	target := filepath.Join(e.cfg.GistMirrorDir(), gistID+".git")

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return e.cloneMirror(ctx, gist, target)
	}

	return e.updateMirror(ctx, target)
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

	file, err := os.Open(e.cfg.RepositoryInventoryFile())

	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		repo := strings.TrimSpace(scanner.Text())

		if repo == "" {
			continue
		}

		jobs <- repo
	}

	return scanner.Err()
}

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

func (e *Engine) syncRepository(ctx context.Context, repo string) error {
	repoName := strings.TrimSuffix(filepath.Base(repo), ".git")
	target := filepath.Join(e.cfg.RepositoryMirrorDir(), repoName+".git")

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return e.cloneMirror(ctx, repo, target)
	}

	return e.updateMirror(ctx, target)
}

func (e *Engine) cloneMirror(ctx context.Context, repo string, target string) error {

	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(repo),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.CloneStarted,
		repoName,
	)

	askPass, err := e.createAskPassScript()
	if err != nil {
		return err
	}

	defer os.Remove(askPass)

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"clone",
		"--mirror",
		repo,
		target,
	)

	if err != nil {

		e.logger.Error(
			logging.Events.Mirror.CloneFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

		return err
	}

	e.logger.Duration(
		logging.Events.Mirror.CloneCompleted,
		repoName,
		time.Since(start),
	)

	return nil
}

func (e *Engine) updateMirror(ctx context.Context, target string) error {
	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.UpdateStarted,
		repoName,
	)

	askPass, err := e.createAskPassScript()
	if err != nil {
		return err
	}

	defer os.Remove(askPass)

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"-C",
		target,
		"remote",
		"update",
		"--prune",
	)

	if err != nil {

		e.logger.Error(
			logging.Events.Mirror.UpdateFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

		return err
	}

	e.logger.Info(
		logging.Events.Mirror.FsckStarted,
		repoName,
	)

	fsck := exec.CommandContext(
		ctx,
		"git",
		"-C",
		target,
		"fsck",
		"--no-dangling",
	)

	if output, err := fsck.CombinedOutput(); err != nil {

		e.logger.Error(
			logging.Events.Mirror.FsckFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

	} else {

		e.logger.Info(
			logging.Events.Mirror.FsckCompleted,
			repoName,
		)
	}

	e.logger.Duration(
		logging.Events.Mirror.UpdateCompleted,
		repoName,
		time.Since(start),
	)

	return nil
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

func (e *Engine) extractGistName(gistURL string) string {

	return strings.TrimSuffix(
		filepath.Base(gistURL),
		".git",
	)
}

func printSyncSummary(label string, assets []state.Asset) {

	var failed []string
	var healthy int

	for _, asset := range assets {

		if asset.LastSuccess {
			healthy++
			continue
		}

		failed = append(failed, asset.Name)
	}

	fmt.Println()
	fmt.Println(label)

	fmt.Printf("  Total:   %d\n", len(assets))
	fmt.Printf("  Healthy: %d\n", healthy)
	fmt.Printf("  Failed:  %d\n", len(failed))

	if len(failed) > 0 {

		fmt.Println()
		fmt.Println("  Failed assets:")

		for _, asset := range failed {
			fmt.Printf("    - %s\n", asset)
		}
	}
}
