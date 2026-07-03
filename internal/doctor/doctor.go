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
	Name           string `json:"name"`
	Success        bool   `json:"success"`
	Recommendation string `json:"recommendation,omitempty"`
	Message        string `json:"message,omitempty"`
}

type Report struct {
	Checks []Check `json:"checks"`
}

func Generate() (*Report, error) {

	report := &Report{}

	cfg := config.Default()

	// Load config
	if err := config.ReadConfig(&cfg); err != nil {
		return nil, err
	}

	// Load token
	_ = config.ReadToken(&cfg)

	// Validate config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

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
			cfg.GitHubToken,
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

	check := Check{
		Name:           name,
		Success:        err == nil,
		Recommendation: recommendation,
	}

	if err != nil {
		check.Message = err.Error()
	}

	return check
}

func checkDirectory(name, path string) Check {

	info, err := os.Stat(path)

	if err != nil {
		return Check{
			Name:    name,
			Success: false,
			Message: err.Error(),
			Recommendation: fmt.Sprintf(
				"Run \"gitback init\" or ensure the directory exists and is accessible: %s",
				path,
			),
		}
	}

	if !info.IsDir() {
		return Check{
			Name:    name,
			Success: false,
			Message: "path exists but is not a directory",
			Recommendation: fmt.Sprintf(
				"Run \"gitback init\" or replace it with a directory: %s",
				path,
			),
		}
	}

	return Check{
		Name:    name,
		Success: true,
	}
}

func checkFile(name string, path string, recommendation string) Check {

	_, err := os.Stat(path)

	check := Check{
		Name:    name,
		Success: err == nil,
	}

	if err != nil {
		check.Message = err.Error()
		check.Recommendation = fmt.Sprintf(recommendation, path)
	}

	return check
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

	check := Check{
		Name:    "log file writable",
		Success: err == nil,
	}

	if err != nil {
		check.Message = err.Error()
		check.Recommendation = "Ensure the log file path is writable and the parent directory exists"
	}

	return check
}

func checkGitHub(token string) Check {

	if token == "" {

		return Check{
			Name:           "github authentication",
			Success:        false,
			Recommendation: "Run: gitback init",
			Message:        "GitHub token is not set",
		}
	}

	client, err := github.NewClient(
		github.WithAuthToken(
			token,
		),
	)

	if err != nil {

		return Check{
			Name:           "github authentication",
			Success:        false,
			Message:        err.Error(),
			Recommendation: "Verify the GitHub token and its permissions.",
		}
	}

	_, _, err = client.Users.Get(
		context.Background(),
		"",
	)

	return Check{
		Name:           "github authentication",
		Success:        err == nil,
		Recommendation: "Verify the GitHub token and its permissions",
	}
}
