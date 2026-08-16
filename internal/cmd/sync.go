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
		rt, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer rt.Logger.Close()

		// The signal-wired context is created here, outside withLock,
		// so a signal arriving while the lock itself is being acquired
		// is also respected, not just once inside executeSync.
		return runCancelable(func(ctx context.Context) error {
			return withLock(rt.Logger, rt.Layout.LockFile, func() error {
				return executeSync(ctx, rt)
			})
		})
	},
}
