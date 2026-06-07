// internal/mirror/mirror.go

package mirror

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	// Consumed for integrity check during snapshot
	var repositories []state.Repository

	// Time Sync
	var syncStartedAt = time.Now()

	file, err := os.Open(e.cfg.RepoInventory)
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

		if err := e.syncRepository(ctx, repo); err != nil {

			e.logger.Error(
				logging.Events.Sync.Failed,
				repo,
				err,
			)

			repositories = append(repositories, state.Repository{
				Name:        repo,
				LastSuccess: false,
				Error:       err.Error(),
			})

			continue
		}

		repositories = append(repositories, state.Repository{
			Name:        repo,
			LastSuccess: true,
		})

		e.cooldown()
	}

	// Time Sync
	var syncCompletedAt = time.Now()

	// Save mirrors metabase
	if err := state.Save(
		e.cfg.MirrorsStateFile,
		syncStartedAt,
		syncCompletedAt,
		repositories,
	); err != nil {

		e.logger.Error(
			logging.Events.Mirror.StateSaveFailed,
			"",
			err,
		)
	}

	return scanner.Err()
}

func (e *Engine) syncRepository(ctx context.Context, repo string) error {
	repoName := strings.TrimSuffix(filepath.Base(repo), ".git")
	target := filepath.Join(e.cfg.MirrorDir, repoName+".git")

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

	// output, err := cmd.CombinedOutput()
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

func (e *Engine) cooldown() {

	min := e.cfg.CooldownMinSeconds
	max := e.cfg.CooldownMaxSeconds

	seconds := rand.Intn(max-min+1) + min

	e.logger.Info(
		logging.Events.Cooldown.Started,
		"",
	)

	time.Sleep(
		time.Duration(seconds) * time.Second,
	)
}
