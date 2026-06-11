// internal/config/validate.go

package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func (c *Config) Validate() error {

	var issues []string

	requiredDirs := map[string]string{
		"config_dir": c.ConfigDir,
		"data_dir":   c.DataDir,
		"log_dir":    c.LogDir,
	}

	for name, path := range requiredDirs {

		if path == "" {

			issues = append(
				issues,
				fmt.Sprintf(
					"%s is not configured",
					name,
				),
			)

			continue
		}

		info, err := os.Stat(path)

		if err != nil {

			issues = append(
				issues,
				fmt.Sprintf(
					"%s does not exist (%s)",
					name,
					path,
				),
			)

			continue
		}

		if !info.IsDir() {

			issues = append(
				issues,
				fmt.Sprintf(
					"%s is not a directory (%s)",
					name,
					path,
				),
			)
		}
	}

	requiredPaths := map[string]string{
		"token_file":         c.TokenFile,
		"state_dir":          c.StateDir,
		"mirror_dir":         c.MirrorDir,
		"snapshot_dir":       c.SnapshotDir,
		"temp_dir":           c.TempDir,
		"repo_inventory":     c.RepoInventory,
		"lock_file":          c.LockFile,
		"log_file":           c.LogFile,
		"mirrors_state_file": c.MirrorsStateFile,
	}

	for name, path := range requiredPaths {

		if path == "" {

			issues = append(
				issues,
				fmt.Sprintf(
					"%s is not configured",
					name,
				),
			)
		}
	}

	if c.TokenFile == "" {

		issues = append(
			issues,
			"token_file is not configured",
		)
	}

	if c.GitHubToken == "" {

		issues = append(
			issues,
			fmt.Sprintf(
				"github token missing or empty (%s)",
				c.TokenFile,
			),
		)
	}

	if c.MinimumFreeDiskPercent < 0 ||
		c.MinimumFreeDiskPercent > 100 {

		issues = append(
			issues,
			"minimum_free_disk_percent must be between 0 and 100",
		)
	}

	if c.GitRetryAttempts < 1 {
		issues = append(
			issues,
			"git_retry_attempts must be >= 1",
		)
	}

	if len(issues) == 0 {
		return nil
	}

	var msg strings.Builder

	msg.WriteString("configuration validation failed:\n")

	for _, issue := range issues {

		fmt.Fprintf(&msg, " - %s\n", issue)
	}

	msg.WriteString("\nRun: gitback init")

	return errors.New(msg.String())
}
