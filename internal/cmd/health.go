// internal/cmd/health.go

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/health"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/spf13/cobra"
)

var healthJSON bool

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show GitBack health report",

	RunE: func(cmd *cobra.Command, args []string) error {

		layout, err := runtime.New()
		if err != nil {
			return err
		}

		if err := layout.EnsureDirs(); err != nil {
			return err
		}

		cfg, err := config.Load(layout)
		if err != nil {
			return err
		}

		for _, dir := range []string{cfg.Storage.MirrorRoot, cfg.Snapshot.OutputDirectory} {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("mkdir %s: %w", dir, err)
			}
		}

		report, err := health.Generate(cfg, layout)
		if err != nil {
			return err
		}

		if err := logHealthReport(layout.LogFile, report); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"[WARN] Failed to write health report to log: %v\n",
				err,
			)
		}

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

func logHealthReport(logFile string, report *health.HealthReport) error {

	logger, err := logging.New(logFile)

	if err != nil {
		return err
	}
	defer logger.Close()

	logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Health.HealthReport,
			Details: map[string]any{
				"report": report,
			},
		},
	)

	return nil
}
