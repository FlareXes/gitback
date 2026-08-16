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
		rt, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer rt.Logger.Close()

		// The signal-wired context is created here, outside withLock,
		// so a signal arriving while the lock itself is being acquired
		// is also respected, not just once inside executeSnapshot.
		return runCancelable(func(ctx context.Context) error {
			return withLock(rt.Logger, rt.Layout.LockFile, func() error {
				return executeSnapshot(ctx, rt, snapshotForce)
			})
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
