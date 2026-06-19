package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flarexes/gitback/internal/version"
)

var showVersion bool

func init() {

	rootCmd.PersistentFlags().BoolVarP(
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
