// internal/cmd/snapshot.go

package cmd

import (
	"context"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/lock"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/snapshot"
	"github.com/spf13/cobra"
)

var snapshotForce bool

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Create mirror snapshot",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := cfg.EnsureDirectories(); err != nil {
			return err
		}

		logger, err := logging.New(cfg.LogFile)
		if err != nil {
			return err
		}

		defer logger.Close()

		locker := lock.New(cfg.LockFile)

		unlock, err := locker.Acquire()
		if err != nil {
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

		engine := snapshot.New(cfg, logger)

		if err := engine.Create(context.Background(), snapshotForce); err != nil {

			logger.Error(
				logging.Events.Snapshot.Failed,
				"",
				err,
			)

			return err
		}

		return nil
	},
}

func init() {

	snapshotCmd.Flags().BoolVar(
		&snapshotForce,
		"force",
		false,
		"continue snapshot creation when repository/mirror health checks fail",
	)
}
