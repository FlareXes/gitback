// internal/cmd/sync.go

package cmd

import (
	"context"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/lock"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/mirror"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync repository mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Load configuration from config.yaml
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := cfg.EnsureRuntimeDirectories(); err != nil {
			return err
		}

		logger, err := logging.New(config.LogFile())
		if err != nil {
			return err
		}
		defer logger.Close()

		locker := lock.New(config.LockFile())

		// Prevent multiple sync processes running simultaneously
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

		// Always release lock when sync is done
		defer func() {
			unlock()

			logger.Info(
				logging.Events.Lock.Released,
				"",
			)
		}()

		logger.Info(
			logging.Events.Sync.Started,
			"",
		)

		engine := mirror.New(cfg, logger)

		// Run mirror synchronization
		if err := engine.Sync(context.Background()); err != nil {

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
	},
}
