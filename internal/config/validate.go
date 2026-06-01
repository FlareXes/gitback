package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func (c *Config) Validate() error {

	var issues []string

	//
	// GitHub settings
	//

	if c.GitHubOwner == "" ||
		c.GitHubOwner == "CHANGE_ME" {

		issues = append(
			issues,
			"github_owner is not configured",
		)
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

	//
	// Required directories
	//

	requiredDirs := map[string]string{
		"base_dir":     c.BaseDir,
		"config_dir":   c.ConfigDir,
		"mirror_dir":   c.MirrorDir,
		"snapshot_dir": c.SnapshotDir,
		"state_dir":    c.StateDir,
		"log_dir":      c.LogDir,
		"health_dir":   c.HealthDir,
		"temp_dir":     c.TempDir,
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

	//
	// Required file paths
	//
	// These do not necessarily need to exist yet,
	// but they must be configured.
	//

	requiredFiles := map[string]string{
		"repo_inventory": c.RepoInventory,
		"log_file":       c.LogFile,
		"lock_file":      c.LockFile,
		"token_file":     c.TokenFile,
	}

	for name, path := range requiredFiles {

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

	//
	// Cooldown validation
	//

	if c.CooldownMinSeconds < 0 {

		issues = append(
			issues,
			"cooldown_min_seconds must be >= 0",
		)
	}

	if c.CooldownMaxSeconds < 0 {

		issues = append(
			issues,
			"cooldown_max_seconds must be >= 0",
		)
	}

	if c.CooldownMaxSeconds <
		c.CooldownMinSeconds {

		issues = append(
			issues,
			"cooldown_max_seconds must be >= cooldown_min_seconds",
		)
	}

	//
	// Disk threshold validation
	//

	if c.MinimumFreeDiskPercent < 0 ||
		c.MinimumFreeDiskPercent > 100 {

		issues = append(
			issues,
			"minimum_free_disk_percent must be between 0 and 100",
		)
	}

	//
	// Final result
	//

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
