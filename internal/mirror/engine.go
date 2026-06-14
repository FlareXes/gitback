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

	syncCompletedAt := time.Now()

	// Save assets metadata such URL with their failed/success status
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
