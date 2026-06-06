// internal/cmd/health.go

package cmd

import (
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/health"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show backup system health",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		report, err := health.Generate(cfg)

		if err != nil {
			return err
		}

		fmt.Printf(
			"Repositories: %d\n",
			report.RepositoryCount,
		)

		fmt.Printf(
			"Healthy: %d\n",
			report.HealthyCount,
		)

		fmt.Printf(
			"Failed: %d\n",
			report.FailedCount,
		)

		fmt.Println()

		fmt.Printf(
			"Last Sync:\n%s\n\n",
			report.LastSync,
		)

		fmt.Printf(
			"Last Snapshot:\n%s\n\n",
			report.LastSnapshot,
		)

		fmt.Printf(
			"Snapshots:\n%d\n\n",
			report.SnapshotCount,
		)

		fmt.Printf(
			"Snapshots Size:\n%s\n\n",
			health.HumanSize(
				report.SnapshotBytes,
			),
		)

		fmt.Printf(
			"Disk Free:\n%d%%\n\n",
			report.DiskFreePercent,
		)

		if len(report.Recommendations) > 0 {

			fmt.Println(
				"Recommendations:",
			)

			fmt.Println()

			for _, rec := range report.Recommendations {

				fmt.Printf(
					"[%s] %s\n",
					rec.Severity,
					rec.Message,
				)
			}
		}

		return nil
	},
}
