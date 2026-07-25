// internal/cmd/discover.go

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover GitHub repositories",
	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, logger, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer logger.Close()

		return executeDiscover(context.Background(), cfg, logger)
	},
}
