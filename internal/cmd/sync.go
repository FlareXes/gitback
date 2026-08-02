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

		return withLock(rt.Logger, rt.Paths.LockFile, func() error {
			return executeSync(context.Background(), rt)
		})
	},
}
