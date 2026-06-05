// internal/doctor/doctor.go
// Package doctor provides health check and diagnostic functionality.

package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/flarexes/gitback/internal/config"
	"github.com/google/go-github/v88/github"
)

type Check struct {
	Name           string
	Success        bool
	Recommendation string
}

type Report struct {
	Checks []Check
}

func Run(cfg *config.Config) (*Report, error) {

	report := &Report{}

	// Configuration
	report.add(
		checkFile(
			"config.yaml",
			cfg.ConfigDir+"/config.yaml",
			"Run: gitback init",
		),
	)

	report.add(
		checkFile(
			"github.token",
			cfg.TokenFile,
			"Run: gitback init",
		),
	)

	// Required binaries
	report.add(
		checkBinary(
			"git",
			"Install git",
		),
	)

	report.add(
		checkBinary(
			"tar",
			"Install tar",
		),
	)

	report.add(
		checkBinary(
			"zstd",
			"Install zstd",
		),
	)

	// Directories
	report.add(
		checkDirectory(
			"mirror directory",
			cfg.MirrorDir,
		),
	)

	report.add(
		checkDirectory(
			"snapshot directory",
			cfg.SnapshotDir,
		),
	)

	report.add(
		checkDirectory(
			"state directory",
			cfg.StateDir,
		),
	)

	// Log file writable
	report.add(
		checkLogFile(
			cfg.LogFile,
		),
	)

	// GitHub auth
	report.add(
		checkGitHub(
			cfg,
		),
	)

	return report, nil
}

func (r *Report) add(check Check) {

	r.Checks = append(
		r.Checks,
		check,
	)
}

func checkBinary(name string, recommendation string) Check {

	_, err := exec.LookPath(name)

	return Check{
		Name:           name,
		Success:        err == nil,
		Recommendation: recommendation,
	}
}

func checkDirectory(name string, path string) Check {

	info, err := os.Stat(path)

	return Check{
		Name:    name,
		Success: err == nil && info.IsDir(),
	}
}

func checkFile(name string, path string, recommendation string) Check {

	_, err := os.Stat(path)

	return Check{
		Name:           name,
		Success:        err == nil,
		Recommendation: recommendation,
	}
}

func checkLogFile(path string) Check {

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0600,
	)

	if err == nil {
		file.Close()
	}

	return Check{
		Name:    "log file writable",
		Success: err == nil,
	}
}

func checkGitHub(cfg *config.Config) Check {

	if cfg.GitHubToken == "" {

		return Check{
			Name:           "github authentication",
			Success:        false,
			Recommendation: "Run: gitback init",
		}
	}

	client, err := github.NewClient(
		github.WithAuthToken(
			cfg.GitHubToken,
		),
	)

	if err != nil {

		return Check{
			Name:    "github authentication",
			Success: false,
		}
	}

	_, _, err = client.Users.Get(
		context.Background(),
		"",
	)

	return Check{
		Name:           "github authentication",
		Success:        err == nil,
		Recommendation: "Verify token permissions",
	}
}

func (r *Report) Print() {

	for _, check := range r.Checks {

		if check.Success {

			fmt.Printf("[OK]   %s\n", check.Name)

			continue
		}

		fmt.Printf("[FAIL] %s\n", check.Name)

		if check.Recommendation != "" {

			fmt.Printf(
				"       Recommendation: %s\n",
				check.Recommendation,
			)
		}
	}
}
