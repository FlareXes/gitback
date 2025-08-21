// Package config provides configuration management for the application.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flarexes/gitback/internal/types"
)

// DefaultConfig returns a new Config instance with default values.
func DefaultConfig() *types.Config {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return &types.Config{
		NoAuth:      false,
		Threads:     5,
		Token:       "",
		User:        "",
		OutputDir:   filepath.Join(homeDir, "gitbackup"),
		Timeout:     30, // seconds
		IncludeGists: true,
	}
}

// LoadFromEnv loads configuration from environment variables.
// Environment variables take precedence over default values.
func LoadFromEnv(cfg *types.Config) error {
	if val, ok := os.LookupEnv("GITBACK_NOAUTH"); ok {
		cfg.NoAuth = strings.ToLower(val) == "true"
	}

	if val, ok := os.LookupEnv("GITBACK_THREADS"); ok {
		threads, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid GITBACK_THREADS value: %v", err)
		}
		cfg.Threads = threads
	}

	if val, ok := os.LookupEnv("GITHUB_TOKEN"); ok {
		cfg.Token = val
	}

	if val, ok := os.LookupEnv("GITBACK_USER"); ok {
		cfg.User = val
	}

	if val, ok := os.LookupEnv("GITBACK_OUTPUT_DIR"); ok {
		cfg.OutputDir = val
	}

	if val, ok := os.LookupEnv("GITBACK_TIMEOUT"); ok {
		timeout, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("invalid GITBACK_TIMEOUT value: %v", err)
		}
		cfg.Timeout = timeout
	}

	if val, ok := os.LookupEnv("GITBACK_INCLUDE_GISTS"); ok {
		cfg.IncludeGists = strings.ToLower(val) == "true"
	}

	return nil
}

// Validate checks if the configuration is valid.
func Validate(cfg *types.Config) error {
	if cfg.NoAuth && cfg.User == "" {
		return errors.New("username is required when running in no-auth mode")
	}

	if cfg.Threads < 1 {
		return errors.New("thread count must be greater than 0")
	}

	if cfg.Timeout < 1 {
		return errors.New("timeout must be greater than 0")
	}

	if cfg.OutputDir == "" {
		return errors.New("output directory is required")
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}

// String returns a string representation of the configuration.
// Sensitive information (like tokens) is redacted.
func String(cfg *types.Config) string {
	if cfg == nil {
		return "<nil>"
	}

	// Always redact token in string representation
	tokenValue := "[REDACTED]"

	return fmt.Sprintf(
		"{NoAuth:%v Threads:%d Token:%s User:%s OutputDir:%s Timeout:%d IncludeGists:%v}",
		cfg.NoAuth,
		cfg.Threads,
		tokenValue, // Always redact token
		cfg.User,
		cfg.OutputDir,
		cfg.Timeout,
		cfg.IncludeGists,
	)
}

// Sanitize returns a copy of the configuration with sensitive data redacted.
// This should be used whenever logging or displaying the configuration.
func Sanitize(cfg *types.Config) *types.Config {
	if cfg == nil {
		return nil
	}

	// Create a deep copy of the config
	sanitized := &types.Config{
		NoAuth:      cfg.NoAuth,
		Threads:     cfg.Threads,
		Token:       "", // Always clear the token
		User:        cfg.User,
		OutputDir:   cfg.OutputDir,
		Timeout:     cfg.Timeout,
		IncludeGists: cfg.IncludeGists,
	}

	return sanitized
}
