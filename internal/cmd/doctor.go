// internal/cmd/doctor.go

package cmd

import (
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/doctor"
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

		report.Print()

		return nil
	},
}
