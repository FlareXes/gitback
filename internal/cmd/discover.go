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

		rt, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer rt.Logger.Close()

		return executeDiscover(context.Background(), rt)
	},
}
