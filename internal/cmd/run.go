// internal/cmd/run.go

package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the full GitBack backup workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		return executeRun(context.Background())
	},
}
