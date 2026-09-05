// internal/cmd/root.go

package cmd

import (
	"github.com/spf13/cobra"
)

const (
	groupSetup      = "setup"
	groupWorkflow   = "workflow"
	groupDiagnostic = "diagnostic"
)

var rootCmd = &cobra.Command{
	Use:           "gitback",
	Short:         "GitHub Backup Utility",
	Long:          `GitBack is an unattended backup tool for your GitHub repositories and gists.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {

	rootCmd.AddGroup(
		&cobra.Group{ID: groupSetup, Title: "Setup Commands:"},
		&cobra.Group{ID: groupWorkflow, Title: "Workflow Commands:"},
		&cobra.Group{ID: groupDiagnostic, Title: "Diagnostic Commands:"},
	)

	initCmd.GroupID = groupSetup
	configCmd.GroupID = groupSetup

	runCmd.GroupID = groupWorkflow
	discoverCmd.GroupID = groupWorkflow
	syncCmd.GroupID = groupWorkflow
	snapshotCmd.GroupID = groupWorkflow

	doctorCmd.GroupID = groupDiagnostic
	healthCmd.GroupID = groupDiagnostic
	versionCmd.GroupID = groupDiagnostic

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(snapshotCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(versionCmd)
}
