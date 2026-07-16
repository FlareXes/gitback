// internal/doctor/doctor.go
// Package doctor provides health check and diagnostic functionality.

package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/flarexes/gitback/internal/config"
	"github.com/google/go-github/v88/github"
)

func Generate() (*Report, error) {

	report := &Report{}

	// ------------------------------------------------------------------
	// Environment
	//
	// Verify prerequisites that are independent of the current GitBack
	// installation.
	// ------------------------------------------------------------------

	report.AddCheck(
		checkOperatingSystem(),
	)

	report.AddChecks(
		checkEnvironment(),
	)

	// ------------------------------------------------------------------
	// Installation
	//
	// Installation-specific diagnostics require a valid GitBack
	// configuration. If the installation cannot be identified, return the
	// report after completing the environment checks.
	// ------------------------------------------------------------------

	cfg, err := config.Load()

	if err != nil {

		report.AddCheck(
			Check{
				Name:           "configuration",
				Success:        false,
				Message:        err.Error(),
				Recommendation: `Run "gitback init"`,
			},
		)

		return report, nil
	}

	report.AddCheck(
		checkFile(
			"github.token file",
			config.TokenFile(),
			`Run "gitback init"`,
		),
	)

	// ------------------------------------------------------------------
	// Filesystem
	// ------------------------------------------------------------------

	report.AddCheck(
		checkWritableFile(
			"log file",
			config.LogFile(),
		),
	)

	report.AddCheck(
		checkDirectory(
			"state directory",
			config.StateDir(),
		),
	)

	report.AddCheck(
		checkDirectory(
			"mirror directory",
			cfg.Storage.MirrorRoot,
		),
	)

	report.AddCheck(
		checkDirectory(
			"snapshot directory",
			cfg.Snapshot.OutputDirectory,
		),
	)

	// ------------------------------------------------------------------
	// Connectivity
	// ------------------------------------------------------------------

	token, _ := config.ReadToken()

	report.AddCheck(
		checkGitHub(token),
	)

	return report, nil
}

func checkOperatingSystem() Check {

	if runtime.GOOS != "linux" {

		return Check{
			Name:    "operating system",
			Success: false,
			Message: fmt.Sprintf(
				"%s is not currently supported",
				runtime.GOOS,
			),
			Recommendation: "Run GitBack on Linux.",
		}
	}

	return Check{
		Name:    "operating system",
		Success: true,
	}
}

// checkEnvironment verifies runtime dependencies that are independent
// of the current GitBack installation.
func checkEnvironment() []Check {

	return []Check{
		checkExecutable(
			"git",
			"Install Git.",
		),
		checkExecutable(
			"tar",
			"Install tar.",
		),
		checkExecutable(
			"zstd",
			"Install zstd.",
		),
	}
}

func checkExecutable(name string, recommendation string) Check {

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

func checkFile(name string, path string, recommendation string) Check {

	_, err := os.Stat(path)

	check := Check{
		Name:    name,
		Success: err == nil,
	}

	if err != nil {
		check.Message = err.Error()
		check.Recommendation = recommendation
	}

	return check
}

func checkWritableFile(name string, path string) Check {

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0600,
	)

	if err == nil {
		file.Close()
	}

	check := Check{
		Name:    name,
		Success: err == nil,
	}

	if err != nil {
		check.Message = err.Error()
		check.Recommendation = "Ensure the file path is writable and the parent directory exists."
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
				"Ensure the directory exists and is accessible: %s",
				path,
			),
		}
	}

	if !info.IsDir() {
		return Check{
			Name:    name,
			Success: false,
			Message: "path exists but is not a directory.",
			Recommendation: fmt.Sprintf(
				"Replace it with a directory: %s",
				path,
			),
		}
	}

	return Check{
		Name:    name,
		Success: true,
	}
}

func checkGitHub(token string) Check {

	if token == "" {

		return Check{
			Name:           "github authentication",
			Success:        false,
			Recommendation: `Run "gitback init"`,
			Message:        "GitHub token is not set.",
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
		Recommendation: "Verify the GitHub token and its permissions.",
	}
}
