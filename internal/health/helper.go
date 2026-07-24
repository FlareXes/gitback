// internal/health/helper.go

package health

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/config"
)

// countQuarantinedRepositories returns the number of quarantined repository mirrors.
func countQuarantinedRepositories(cfg *config.Config) (int, error) {

	root := filepath.Join(config.QuarantineDir(cfg), "repositories")

	// Missing directory simply means no repositories are quarantined.
	owners, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	count := 0

	for _, owner := range owners {

		if !owner.IsDir() {
			continue
		}

		repositories, err := os.ReadDir(
			filepath.Join(
				root,
				owner.Name(),
			),
		)

		if err != nil {
			return 0, err
		}

		for _, repository := range repositories {

			// Repository mirrors are stored as directories ending
			// with ".git".
			if repository.IsDir() && strings.HasSuffix(
				repository.Name(),
				".git",
			) {

				count++
			}
		}
	}

	return count, nil
}

// countQuarantinedGists returns the number of quarantined gist mirrors.
func countQuarantinedGists(cfg *config.Config) (int, error) {

	root := filepath.Join(
		config.QuarantineDir(cfg),
		"gists",
	)

	// Missing directory simply means no gists are quarantined.
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	count := 0

	for _, gist := range entries {

		// Gist mirrors are stored directly beneath the gists directory.
		if gist.IsDir() && strings.HasSuffix(
			gist.Name(),
			".git",
		) {

			count++
		}
	}

	return count, nil
}
