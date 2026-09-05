package cmd

import (
	"fmt"

	examplecfg "github.com/flarexes/gitback/example"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect gitback configuration",
}

var configExampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Print a fully-documented example config.toml",
	Long: "Prints an example config.toml with every available setting, " +
		"its default, and notes on edge-case behavior. Redirect to a " +
		"file to use it as a starting point:\n\n" +
		"  gitback config example > config.toml",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(examplecfg.Config)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configExampleCmd)
}
