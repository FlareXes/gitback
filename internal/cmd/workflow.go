// internal/cmd/workflow.go

package cmd

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/discovery"
	"github.com/flarexes/gitback/internal/lock"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/mirror"
	"github.com/flarexes/gitback/internal/snapshot"
)

// prepareRuntime loads configuration once and ensures GitBack's runtime
// directories exist before any command starts work.
func prepareRuntime() (*config.Config, *logging.Logger, error) {

	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	if err := cfg.EnsureRuntimeDirectories(); err != nil {
		return nil, nil, err
	}

	logger, err := logging.New(config.LogFile())
	if err != nil {
		return nil, nil, err
	}

	return cfg, logger, nil
}

// withLock runs fn while holding GitBack's global lock.
// Lock acquisition and release are logged here so individual commands do not
// need to duplicate the same boilerplate.
func withLock(logger *logging.Logger, fn func() error) error {

	locker := lock.New(config.LockFile())

	unlock, err := locker.Acquire()
	if err != nil {

		logger.Error(
			logging.Events.Lock.Busy,
			"",
			err,
		)

		return err
	}

	logger.Info(
		logging.Events.Lock.Acquired,
		"",
	)

	defer func() {
		unlock()

		logger.Info(
			logging.Events.Lock.Released,
			"",
		)
	}()

	return fn()
}

// executeDiscover runs repository and gist discovery using the shared runtime
// configuration and logger.
func executeDiscover(
	ctx context.Context,
	cfg *config.Config,
	logger *logging.Logger,
) error {

	logger.Info(
		logging.Events.GitHub.DiscoveryStarted,
		"",
	)

	client, err := discovery.New(cfg, logger)
	if err != nil {
		return err
	}

	if err := client.Discover(ctx); err != nil {

		logger.Error(
			logging.Events.GitHub.DiscoveryFailed,
			"",
			err,
		)

		return fmt.Errorf(
			"repository discovery failed: %w",
			err,
		)
	}

	return nil
}

// executeSync runs mirror synchronization using the shared runtime
// configuration and logger.
func executeSync(
	ctx context.Context,
	cfg *config.Config,
	logger *logging.Logger,
) error {

	logger.Info(
		logging.Events.Sync.Started,
		"",
	)

	engine := mirror.New(cfg, logger)

	if err := engine.Sync(ctx); err != nil {

		logger.Error(
			logging.Events.Sync.Failed,
			"",
			err,
		)

		return err
	}

	logger.Info(
		logging.Events.Sync.Completed,
		"",
	)

	return nil
}

// executeSnapshot creates a snapshot using the shared runtime configuration
// and logger.
func executeSnapshot(
	ctx context.Context,
	cfg *config.Config,
	logger *logging.Logger,
	force bool,
) error {

	logger.Info(
		logging.Events.Snapshot.Started,
		"",
	)

	engine := snapshot.New(cfg, logger)

	if err := engine.Create(ctx, force); err != nil {

		logger.Error(
			logging.Events.Snapshot.Failed,
			"",
			err,
		)

		return err
	}

	logger.Info(
		logging.Events.Snapshot.Completed,
		"",
	)

	return nil
}

// executeRun performs the unattended backup workflow under a single lock.
// It always uses best-effort snapshot mode so unattended runs continue even if
// a few mirrors failed to synchronize.
func executeRun(ctx context.Context) error {

	cfg, logger, err := prepareRuntime()
	if err != nil {
		return err
	}
	defer logger.Close()

	return withLock(logger, func() error {

		if err := executeDiscover(ctx, cfg, logger); err != nil {
			return err
		}

		if err := executeSync(ctx, cfg, logger); err != nil {
			return err
		}

		// Unattended runs should still produce a snapshot even when some
		// mirrors failed earlier in the pipeline.
		if err := executeSnapshot(ctx, cfg, logger, true); err != nil {
			return err
		}

		return nil
	})
}
