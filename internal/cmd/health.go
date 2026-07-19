// internal/cmd/health.go

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/health"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/spf13/cobra"
)

var healthJSON bool

func init() {

	rootCmd.AddCommand(healthCmd)

	healthCmd.Flags().BoolVar(
		&healthJSON,
		"json",
		false,
		"Output machine-readable JSON",
	)
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show GitBack health report",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := cfg.EnsureRuntimeDirectories(); err != nil {
			return err
		}

		report, err := health.Generate(
			cfg,
		)

		if err != nil {
			return err
		}

		if err := logHealthReport(config.LogFile(), report); err != nil {
			fmt.Fprintf(
				os.Stderr,
				"[WARN] Failed to write health report to log: %v\n",
				err,
			)
		}

		if healthJSON {

			encoder := json.NewEncoder(
				os.Stdout,
			)

			encoder.SetIndent(
				"",
				"  ",
			)

			return encoder.Encode(
				report,
			)
		}

		health.PrintReport(report)

		return nil
	},
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
