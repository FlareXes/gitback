// internal/cmd/snapshot.go

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var snapshotForce bool

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Create mirror snapshot",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, logger, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer logger.Close()

		return withLock(logger, func() error {
			return executeSnapshot(context.Background(), cfg, logger, snapshotForce)
		})
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
