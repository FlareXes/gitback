// internal/cmd/discover.go

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/github"
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

		client, err := github.New(cfg, logger)
		if err != nil {
			return err
		}

		if err := client.Discover(context.Background()); err != nil {

			logger.Error(
				logging.Events.GitHub.DiscoveryFailed,
				"",
				err,
			)

			fmt.Fprintf(
				os.Stderr,
				"Repository discovery failed: %v\n",
				err,
			)

			return err
		}

		fmt.Println("Repository discovery completed successfully")

		return nil
	},
}
