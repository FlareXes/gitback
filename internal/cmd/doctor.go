// internal/cmd/doctor.go

package cmd

import (
	"fmt"
	"os"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/doctor"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate gitback environment",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()

		if err != nil {

			fmt.Println("[FAIL] configuration")
			fmt.Println(err)

			fmt.Println("")
			fmt.Println("Run: gitback init")

			return nil
		}

		report, err := doctor.Run(cfg)

		if err != nil {
			return err
		}

		if err := logDoctorReport(cfg.LogFile, report); err != nil {

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

func logDoctorReport(logFile string, report *doctor.Report) error {

	logger, err := logging.New(logFile)

	if err != nil {
		return err
	}

	defer logger.Close()

	logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Doctor.ReportGenerated,

			Details: map[string]any{
				"report": report,
			},
		},
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

		if check.Recommendation != "" {

			fmt.Printf(
				"       Recommendation: %s\n",
				check.Recommendation,
			)
		}
	}
}
