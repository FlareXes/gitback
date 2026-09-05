// internal/cmd/version.go

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flarexes/gitback/internal/version"
)

var showVersion bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print gitback version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("gitback %s\n", version.Get())
		return nil
	},
}

func init() {

	rootCmd.Flags().BoolVarP(
		&showVersion,
		"version",
		"v",
		false,
		"print version info",
	)

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !showVersion {
			return cmd.Help()
		}

		fmt.Printf("gitback %s\n", version.Get())

		return nil
	}
}
