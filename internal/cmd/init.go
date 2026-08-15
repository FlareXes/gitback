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

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gitback environment",
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

		// Get GitHub username
		reader := bufio.NewReader(os.Stdin)

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

		configPath := layout.ConfigFile

		if err := config.Write(configPath, cfg); err != nil {
			return err
		}

		// Save token separately
		if err := os.WriteFile(layout.TokenFile, []byte(token+"\n"), 0600); err != nil {
			return err
		}

		if _, err := config.Load(layout); err != nil {
			return fmt.Errorf("post-init validation failed: %w", err)
		}

		fmt.Printf("Authenticated as: %s\n", user.GetLogin())
		fmt.Printf("Token file: %s\n", layout.TokenFile)
		fmt.Printf("Config file: %s\n", configPath)

		fmt.Println("\ngitback initialized successfully")

		return nil
	},
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
}
