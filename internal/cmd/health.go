// internal/cmd/health.go

package cmd

import (
	"encoding/json"
	"os"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/health"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show GitBack health report",

	RunE: func(cmd *cobra.Command, args []string) error {

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		if err := cfg.EnsureDirectories(); err != nil {
			return err
		}

		report, err := health.Generate(
			cfg,
		)

		if err != nil {
			return err
		}

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
	},
}
