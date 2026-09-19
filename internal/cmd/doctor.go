// internal/cmd/doctor.go

package cmd

import (
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/doctor"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate gitback environment",

	RunE: func(cmd *cobra.Command, args []string) error {

		layout, err := runtime.New()
		if err != nil {
			return err
		}

		// Best-effort logger: doctor must still produce a full report
		// even if log writing itself is broken.
		// minKeep=0, retentionDays=0 disables pruning; doctor is read-only.
		logger, logErr := logging.New(layout.LogDir, 0, 0)
		if logErr != nil {
			fmt.Fprintf(os.Stderr, "[WARN] Could not open log file: %v\n", logErr)
		}
		defer logger.Close()

		report, err := doctor.Generate(layout, logger)
		if err != nil {
			return err
		}

		// doctor doesn't prune old logs; minKeep=0, retentionDays=0 disables pruning.
		if err := logDoctorReport(layout.LogDir, 0, 0, report); err != nil {

			fmt.Fprintf(
				os.Stderr,
				"[WARN] Failed to write doctor report to log: %v\n",
				err,
			)
		}

		printDoctorReport(report)

		return nil
	},
}

func logDoctorReport(logDir string, minKeep int, retentionDays int, report *doctor.Report) error {

	logger, err := logging.New(logDir, minKeep, retentionDays)
	if err != nil {
		return err
	}
	defer logger.Close()

	logger.Emit(
		logging.Events.Doctor.ReportGenerated,
		logging.WithDetails(map[string]any{
			"report": report,
		}),
	)

	return nil
}

func printDoctorReport(report *doctor.Report) {

	for _, check := range report.Checks {

		if check.Success {
			fmt.Printf("[OK]   %s\n", check.Name)
			continue
		}

		fmt.Printf("[FAIL] %s\n", check.Name)

		if check.Message != "" {
			fmt.Printf("       Reason: %s\n", check.Message)
		}

		if check.Recommendation != "" {
			fmt.Printf("       Recommendation: %s\n", check.Recommendation)
		}
	}
}
