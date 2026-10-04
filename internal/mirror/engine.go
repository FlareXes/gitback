// internal/mirror/engine.go

package mirror

import (
	"context"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/flarexes/gitback/internal/state"
)

type Engine struct {
	cfg    *config.Config
	layout runtime.Layout
	logger *logging.Logger
	token  string
}

func New(cfg *config.Config, layout runtime.Layout, logger *logging.Logger) *Engine {

	token, _ := config.ReadToken(layout, logger)

	return &Engine{
		cfg:    cfg,
		layout: layout,
		logger: logger,
		token:  token,
	}
}

func (e *Engine) Sync(ctx context.Context) error {

	syncStartedAt := time.Now()

	// Sync repositories
	repositories, err := e.syncRepositories(
		ctx,
	)

	if err != nil {
		return err
	}

	// Sync Gists
	var gists []state.Asset

	// Skip the gist phase entirely if repos were already interrupted —
	// starting a new phase after cancellation just repeats the burst.
	if e.cfg.GitHub.BackupGists && ctx.Err() == nil {
		gists, err = e.syncGists(ctx)
		if err != nil {
			return err
		}
	}

	printSyncSummary("Repositories", repositories)

	if e.cfg.GitHub.BackupGists {
		printSyncSummary("Gists", gists)
	}

	syncCompletedAt := time.Now()

	// Save assets metadata such URL with their failed/success status
	if err := state.SaveMirrors(
		e.layout.MirrorsStateFile,
		syncStartedAt,
		syncCompletedAt,
		repositories,
		gists,
	); err != nil {

		e.logger.Emit(logging.Events.Mirror.StateSaveFailed, logging.WithError(err))

		return err
	}

	e.logSyncSummary(syncStartedAt, repositories, gists)

	// Surface cancellation after state is persisted, so the caller can
	// log this as an interruption, not a failure, without losing progress.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return nil
}

func (e *Engine) logSyncSummary(
	syncStartedAt time.Time,
	repositories []state.Asset,
	gists []state.Asset,
) {
	repoHealthy, repoInterrupted, repoFailed := summarizeAssets(repositories)
	gistHealthy, gistInterrupted, gistFailed := summarizeAssets(gists)

	e.logger.Emit(
		logging.Events.Sync.Summary,
		logging.WithDuration(time.Since(syncStartedAt)),
		logging.WithDetails(map[string]any{
			"repositories_total":       len(repositories),
			"repositories_healthy":     repoHealthy,
			"repositories_interrupted": repoInterrupted,
			"repositories_failed":      repoFailed,

			"gists_enabled":     e.cfg.GitHub.BackupGists,
			"gists_total":       len(gists),
			"gists_healthy":     gistHealthy,
			"gists_interrupted": gistInterrupted,
			"gists_failed":      gistFailed,
		}),
	)
}

// summarizeAssets buckets assets into healthy/interrupted/failed using
// the same rule health.populateAssets already uses (Cause ==
// CauseCancelled, not just !LastSuccess).
func summarizeAssets(assets []state.Asset) (healthy, interrupted, failed int) {
	for _, a := range assets {
		switch {
		case a.LastSuccess:
			healthy++
		case a.Cause == logging.CauseCancelled:
			interrupted++
		default:
			failed++
		}
	}
	return
}
