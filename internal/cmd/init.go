// internal/cmd/init.go

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/google/go-github/v88/github"
	"github.com/spf13/cobra"
)

var (
	initForce       bool
	initUseEnvToken bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gitback environment",
	Long: `Init walks you through GitHub token setup and creates gitback's
configuration file with default settings.

If gitback is already initialized, init refuses to run unless --force
is passed, to avoid silently overwriting your saved token and config.

By default, init prompts for a GitHub token and saves it to disk.
Pass --use-env-token to instead read the token from GITBACK_TOKEN —
init won't prompt, and nothing is written to the token file.`,

	Example: `  # First-time setup
  gitback init

  # Re-run setup, overwriting the existing config and token
  gitback init --force

  # Non-interactive setup using an environment variable
  GITBACK_TOKEN=ghp_xxx gitback init --use-env-token`,

	RunE: func(cmd *cobra.Command, args []string) error {

		layout, err := runtime.New()
		if err != nil {
			return err
		}

		found, err := existingInstallation(layout)
		if err != nil {
			return err
		}

		if len(found) > 0 {
			if !initForce {
				return fmt.Errorf(
					"gitback appears to be already initialized (found: %s); use --force to reinitialize",
					strings.Join(found, ", "),
				)
			}

			fmt.Printf(
				"[WARN] Reinitializing existing installation (found: %s)\n",
				strings.Join(found, ", "),
			)
		}

		if err := layout.EnsureDirs(); err != nil {
			return err
		}

		cfg := config.Default(layout)

		for _, dir := range []string{cfg.Storage.MirrorRoot, cfg.Snapshot.OutputDirectory} {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("mkdir %s: %w", dir, err)
			}
		}

		token, err := resolveInitToken(initUseEnvToken, bufio.NewReader(os.Stdin))
		if err != nil {
			return err
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

		configPath := layout.ConfigFile

		if err := config.Write(configPath, cfg); err != nil {
			return err
		}

		// --use-env-token's entire point is that the token is never
		// persisted — the promise it makes is "gitback will read this
		// from your environment every time", so writing it to disk
		// here would silently break that promise.
		if !initUseEnvToken {
			if err := os.WriteFile(layout.TokenFile, []byte(token+"\n"), 0600); err != nil {
				return err
			}
		}

		if _, err := config.Load(layout); err != nil {
			return fmt.Errorf("post-init validation failed: %w", err)
		}

		fmt.Printf("Authenticated as: %s\n", user.GetLogin())

		if initUseEnvToken {
			fmt.Println("Token source:     GITBACK_TOKEN environment variable (not saved to disk)")
		} else {
			fmt.Printf("Token file:       %s\n", layout.TokenFile)
		}

		fmt.Printf("Config file:      %s\n", configPath)

		fmt.Println("\ngitback initialized successfully")

		return nil
	},
}

// resolveInitToken determines the GitHub token to use during init,
// either by reading GITBACK_TOKEN directly (useEnvToken) or by
// prompting interactively via stdin. It deliberately touches neither
// the network nor the filesystem — callers validate the returned token
// against GitHub and decide whether to persist it — which keeps this
// function cheaply testable without mocking either.
//
// stdin is only read when useEnvToken is false; pass nil safely when
// useEnvToken is true.
func resolveInitToken(useEnvToken bool, stdin *bufio.Reader) (string, error) {

	if useEnvToken {

		token := strings.TrimSpace(os.Getenv("GITBACK_TOKEN"))

		if token == "" {
			return "", fmt.Errorf(
				"--use-env-token was set but GITBACK_TOKEN is empty or unset; export it first",
			)
		}

		return token, nil
	}

	fmt.Println("Create a GitHub Personal Access Token.")
	fmt.Println("")
	fmt.Println("Option 1: Classic PAT:")
	fmt.Println("  Scope:")
	fmt.Println("    repo")
	fmt.Println("")
	fmt.Println("Option 2: Fine-grained PAT:")
	fmt.Println("  Repository access:")
	fmt.Println("    All repositories")
	fmt.Println("")
	fmt.Println("  Permissions:")
	fmt.Println("    Contents: Read-only")
	fmt.Println("    Metadata: Read-only")
	fmt.Println("")
	fmt.Print("GitHub token: ")

	token, err := stdin.ReadString('\n')
	if err != nil {
		return "", err
	}

	token = strings.TrimSpace(token)

	if token == "" {
		return "", fmt.Errorf("github token cannot be empty")
	}

	return token, nil
}

// existingInstallation reports which markers of a previous `gitback init`
// are present on disk. To avoid overwriting existing data & token
func existingInstallation(layout runtime.Layout) ([]string, error) {

	candidates := []struct {
		label string
		path  string
	}{
		{"config file", layout.ConfigFile},
		{"github token", layout.TokenFile},
		{"mirror state", layout.MirrorsStateFile},
		{"repository inventory", layout.RepositoryInventoryFile},
		{"gist inventory", layout.GistInventoryFile},
	}

	var found []string

	for _, c := range candidates {

		_, err := os.Stat(c.path)

		if err == nil {
			found = append(found, c.label)
			continue
		}

		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("check %s: %w", c.label, err)
		}
	}

	return found, nil
}

func init() {

	initCmd.Flags().BoolVar(
		&initForce,
		"force",
		false,
		"reinitialize even if gitback is already initialized (overwrites config.toml and github.token)",
	)

	initCmd.Flags().BoolVar(
		&initUseEnvToken,
		"use-env-token",
		false,
		"read the GitHub token from GITBACK_TOKEN instead of prompting; the token is never written to disk",
	)
}
