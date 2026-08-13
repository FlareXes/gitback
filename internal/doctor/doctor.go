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
	rt "github.com/flarexes/gitback/internal/runtime"
	"github.com/google/go-github/v88/github"
)

// Generate runs every diagnostic doctor is able to run given whatever state
// is currently on disk. Nothing here requires a valid config — a missing
// or invalid config is reported as a failed check, not a hard error, so
// the rest of the checks still run and the user gets a full picture of
// what's wrong in one pass.
func Generate(layout rt.Layout) (*Report, error) {

	report := &Report{}

	// ------------------------------------------------------------------
	// Environment
	//
	// Prerequisites independent of any GitBack installation.
	// ------------------------------------------------------------------

	report.AddCheck(
		checkOperatingSystem(),
	)

	report.AddChecks(
		checkEnvironment(),
	)

	// ------------------------------------------------------------------
	// Configuration
	//
	// A missing/invalid config is reported as a failed check, not a
	// reason to stop — everything else below is layout-based and can
	// still be checked without it.
	// ------------------------------------------------------------------

	report.AddCheck(
		checkFile(
			"config file",
			layout.ConfigFile,
			`Run "gitback init"`,
		),
	)

	cfg, err := config.Load(layout)

	if err != nil {

		report.AddCheck(
			Check{
				Name:           "config validity",
				Success:        false,
				Message:        err.Error(),
				Recommendation: `Run "gitback init"`,
			},
		)

	} else {

		report.AddCheck(
			Check{
				Name:    "config validity",
				Success: true,
			},
		)
	}

	// ------------------------------------------------------------------
	// Filesystem
	//
	// Token, log, and state dir come from Layout, so they're checked
	// regardless of whether config loaded successfully.
	// ------------------------------------------------------------------

	report.AddCheck(
		checkFile(
			"github.token file",
			layout.TokenFile,
			`Run "gitback init"`,
		),
	)

	report.AddCheck(
		checkWritableFile(
			"log file",
			layout.LogFile,
		),
	)

	report.AddCheck(
		checkDirectory(
			"state directory",
			layout.StateDir,
		),
	)

	// Mirror/snapshot directories are user-configured, so these can only
	// run once config has actually loaded.
	if cfg != nil {

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
	}

	// ------------------------------------------------------------------
	// Connectivity
	//
	// ReadToken checks GITBACK_TOKEN and the token file independently of
	// config, so this still runs even if config failed to load.
	// ------------------------------------------------------------------

	token, _ := config.ReadToken(layout)

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
