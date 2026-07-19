// internal/config/validate.go

package config

import (
	"errors"
	"fmt"
	"strings"
)

func (c *Config) Validate() error {

	var issues []string

	if c.Storage.MirrorRoot == "" {
		issues = append(
			issues,
			"storage.mirror_root is required",
		)
	}

	if c.Snapshot.OutputDirectory == "" {
		issues = append(
			issues,
			"snapshot.output_directory is required",
		)
	}

	if c.Sync.Workers < 1 {
		issues = append(
			issues,
			"sync.workers must be >= 1",
		)
	}

	if c.Sync.RetryAttempts < 1 {
		issues = append(
			issues,
			"sync.retry_attempts must be >= 1",
		)
	}

	if c.Health.MinimumFreeDiskPercent > 100 {

		issues = append(
			issues,
			"health.minimum_free_disk_percent must be between 0 and 100",
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

	return errors.New(msg.String())
}
