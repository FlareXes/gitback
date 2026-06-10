// internal/cmd/root.go

package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "gitback",
	Short:        "GitHub Backup Utility",
	SilenceUsage: true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(healthCmd)
}
