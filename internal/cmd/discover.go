// internal/cmd/discover.go

package cmd

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/discovery"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover GitHub repositories",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Load configuration from config.yaml
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		logger, err := logging.New(cfg.LogFile)
		if err != nil {
			return err
		}
		defer logger.Close()

		logger.Info(
			logging.Events.GitHub.DiscoveryStarted,
			"",
		)

		client, err := discovery.New(cfg, logger)
		if err != nil {
			return err
		}

		if err := client.Discover(context.Background()); err != nil {

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
	},
}
