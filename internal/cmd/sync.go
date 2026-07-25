// internal/cmd/sync.go

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync repository mirrors",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, logger, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer logger.Close()

		return withLock(logger, func() error {
			return executeSync(context.Background(), cfg, logger)
		})
	},
}
