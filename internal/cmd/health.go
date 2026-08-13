// internal/cmd/health.go

package cmd

import (
	"encoding/json"
	"os"

	"github.com/flarexes/gitback/internal/health"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/spf13/cobra"
)

var healthJSON bool

// healthCmd reports the current status of an already-initialized GitBack
// installation: sync state, quarantined mirrors, snapshots, and disk
// space. Unlike doctorCmd, it requires config to load successfully —
// there's nothing meaningful to report on an uninitialized installation.
var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show GitBack health report",

	RunE: func(cmd *cobra.Command, args []string) error {

		// prepareRuntime resolves the layout, loads config, ensures
		// runtime directories exist, and opens the logger — the same
		// shared setup used by sync/snapshot/discover, so health
		// can't drift from how those commands behave.
		rt, err := prepareRuntime()
		if err != nil {
			return err
		}
		defer rt.Logger.Close()

		report, err := health.Generate(rt.Config, rt.Layout)
		if err != nil {
			return err
		}

		// Every health check is logged for later auditing, regardless
		// of output format below.
		rt.Logger.Emit(
			logging.Entry{
				Level: logging.Info,
				Event: logging.Events.Health.HealthReport,
				Details: map[string]any{
					"report": report,
				},
			},
		)

		if healthJSON {
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(report)
		}

		health.PrintReport(report)

		return nil
	},
}

func init() {

	healthCmd.Flags().BoolVar(
		&healthJSON,
		"json",
		false,
		"Output machine-readable JSON",
	)
}
