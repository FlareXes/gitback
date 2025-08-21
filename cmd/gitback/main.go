/*
Gitback is a tool for backing up GitHub repositories and gists.

It provides functionality to backup GitHub repositories either using a GitHub Personal Access Token (PAT)
or without authentication (public repositories only).

Usage:
  gitback [flags]

Flags:
  --noauth        Disable GitHub Auth (limited to 60 requests/hour for public data only)
  --username      GitHub username (required when --noauth is set)
  --thread        Maximum number of concurrent connections (default: 5)
  --token         GitHub API token (optional, can also be set via GITHUB_TOKEN)
  --output-dir    Directory to store backups (default: ~/gitbackup)
  --timeout       Timeout in seconds for API requests (default: 30)
  --no-gists      Skip backing up gists
  --version       Show version information
  --help          Show help message

Environment Variables:
  GITBACK_NOAUTH         Set to "true" to disable authentication
  GITBACK_THREADS        Number of concurrent operations
  GITBACK_USER           GitHub username
  GITBACK_OUTPUT_DIR     Directory to store backups
  GITBACK_TIMEOUT        Timeout in seconds for API requests
  GITBACK_INCLUDE_GISTS  Set to "false" to skip gists
  GITHUB_TOKEN           GitHub Personal Access Token
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/flarexes/gitback/internal/types"
	"github.com/flarexes/gitback/internal/vcs"
	"github.com/flarexes/gitback/pkg/config"
)

// version is set during build (e.g., -ldflags="-X main.version=1.0.0")
var version = "dev"

func main() {
	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	setupSignalHandling(cancel)

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Print version and exit if requested
	if showVersion {
		printVersion()
		os.Exit(0)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Printf("Starting backup with configuration: %s\n", config.String(cfg))

	// Run the application
	if err := vcs.Run(ctx, cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// setupSignalHandling configures signal handling for graceful shutdown.
func setupSignalHandling(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal: %v. Shutting down...\n", sig)
		cancel()
	}()
}

// Command line flags
var (
	showVersion bool
	noAuth     bool
	threads    int
	token      string
	username   string
	outputDir  string
	timeout    int
	noGists    bool
)

// init initializes command line flags.
func init() {
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&noAuth, "noauth", false, "Disable GitHub Auth (limited to 60 requests/hour for public data only)")
	flag.IntVar(&threads, "thread", 5, "Maximum number of concurrent connections")
	flag.StringVar(&token, "token", "", "GitHub API token (can also be set via GITHUB_TOKEN)")
	flag.StringVar(&username, "username", "", "GitHub username (required when --noauth is set)")
	flag.StringVar(&outputDir, "output-dir", "", "Directory to store backups (default: ~/gitbackup)")
	flag.IntVar(&timeout, "timeout", 0, "Timeout in seconds for API requests (default: 30)")
	flag.BoolVar(&noGists, "no-gists", false, "Skip backing up gists")
}

// loadConfig loads configuration from command line flags, environment variables, and defaults.
func loadConfig() (*types.Config, error) {
	// Parse command line flags
	flag.Parse()

	// Load default configuration
	cfg := config.DefaultConfig()

	// Apply environment variable overrides
	if err := config.LoadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %w", err)
	}

	// Apply command line flag overrides (these take highest precedence)
	if noAuth {
		cfg.NoAuth = true
	}

	if threads > 0 {
		cfg.Threads = threads
	}

	if token != "" {
		cfg.Token = token
	}

	if username != "" {
		cfg.User = username
	}

	if outputDir != "" {
		cfg.OutputDir = outputDir
	}

	if timeout > 0 {
		cfg.Timeout = timeout
	}

	if noGists {
		cfg.IncludeGists = false
	}

	return cfg, nil
}

// printVersion prints the application version and exits.
func printVersion() {
	fmt.Printf("gitback version %s\n", version)
}
