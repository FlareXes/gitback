// internal/cmd/init.go

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/google/go-github/v88/github"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gitback environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Default()

		dirs := []string{
			cfg.BaseDir,
			cfg.ConfigDir,

			cfg.MirrorDir,
			cfg.SnapshotDir,
			cfg.StateDir,

			cfg.LogDir,
			cfg.HealthDir,

			cfg.TempDir,
		}

		// Create all required directories
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("mkdir %s: %w", dir, err)
			}

		}

		// Get GitHub username
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("GitHub username: ")

		owner, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		owner = strings.TrimSpace(owner)

		if owner == "" {
			return fmt.Errorf("github username cannot be empty")
		}

		fmt.Print("GitHub token: ")

		// Get GitHub token (PAT - Personal Access Token)
		token, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		token = strings.TrimSpace(token)

		if token == "" {
			return fmt.Errorf("github token cannot be empty")
		}

		// Validate token before saving anything.
		ctx, cancel := context.WithTimeout(
			context.Background(),
			30*time.Second,
		)
		defer cancel()

		client, err := github.NewClient(
			github.WithAuthToken(
				token,
			),
		)

		if err != nil {
			return err
		}

		user, _, err := client.Users.Get(ctx, "")

		if err != nil {
			return fmt.Errorf(
				"github authentication failed: %w",
				err,
			)
		}

		cfg.GitHubOwner = owner

		configPath := filepath.Join(cfg.ConfigDir, "config.yaml")

		if err := config.Write(configPath, cfg); err != nil {
			return err
		}

		// Save token separately
		if err := os.WriteFile(cfg.TokenFile, []byte(token+"\n"), 0600); err != nil {
			return err
		}

		if _, err := config.Load(); err != nil {

			return fmt.Errorf(
				"post-init validation failed: %w",
				err,
			)
		}

		fmt.Printf("Authenticated as: %s\n", user.GetLogin())
		fmt.Printf("Token file: %s\n", cfg.TokenFile)
		fmt.Printf("Config file: %s\n", configPath)

		fmt.Println("\ngitback initialized successfully")

		return nil
	},
}
