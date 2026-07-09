// internal/mirror/engine.go

package mirror

import (
	"context"
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
	gists, err := e.syncGists(ctx)

	if err != nil {
		return err
	}

	printSyncSummary("Repositories", repositories)
	printSyncSummary("Gists", gists)

	// Log sync summary
	e.logSyncSummary(syncStartedAt, repositories, gists)

	syncCompletedAt := time.Now()

	// Save assets metadata such URL with their failed/success status
	if err := state.SaveMirrors(
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

func (e *Engine) logSyncSummary(
	syncStartedAt time.Time,
	repositories []state.Asset,
	gists []state.Asset,
) {
	var repositoryHealthy int
	var repositoryFailed int

	for _, repo := range repositories {

		if repo.LastSuccess {
			repositoryHealthy++
		} else {
			repositoryFailed++
		}
	}

	var gistHealthy int
	var gistFailed int

	for _, gist := range gists {

		if gist.LastSuccess {
			gistHealthy++
		} else {
			gistFailed++
		}
	}

	// Run-level summary event.
	e.logger.Emit(
		logging.Entry{
			Level:      logging.Info,
			Event:      logging.Events.Sync.Summary,
			DurationMS: time.Since(syncStartedAt).Milliseconds(),

			Details: map[string]any{
				"repositories_total":   len(repositories),
				"repositories_healthy": repositoryHealthy,
				"repositories_failed":  repositoryFailed,

				"gists_total":   len(gists),
				"gists_healthy": gistHealthy,
				"gists_failed":  gistFailed,
			},
		},
	)
}
