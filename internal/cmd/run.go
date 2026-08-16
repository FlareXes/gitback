// internal/cmd/run.go

package cmd

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the full GitBack backup workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		// runCancelable supplies a context wired to SIGINT/SIGTERM in
		// place of context.Background(), so Ctrl+C or a systemd stop
		// now actually interrupts an in-flight discover/sync/snapshot
		// step instead of being ignored by the context layer.
		return runCancelable(executeRun)
	},
}
