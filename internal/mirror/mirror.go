// internal/mirror/mirror.go

package mirror

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

func printSyncSummary(label string, assets []state.Asset) {

	var failed []string
	var healthy int

	for _, asset := range assets {

		if asset.LastSuccess {
			healthy++
			continue
		}

		failed = append(failed, asset.Name)
	}

	fmt.Println()
	fmt.Println(label)

	fmt.Printf("  Total:   %d\n", len(assets))
	fmt.Printf("  Healthy: %d\n", healthy)
	fmt.Printf("  Failed:  %d\n", len(failed))

	if len(failed) > 0 {

		fmt.Println()
		fmt.Println("  Failed assets:")

		for _, asset := range failed {
			fmt.Printf("    - %s\n", asset)
		}
	}
}

func (e *Engine) cloneMirror(ctx context.Context, repo string, target string) error {

	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(repo),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.CloneStarted,
		repoName,
	)

	askPass, err := e.createAskPassScript()
	if err != nil {
		return err
	}

	defer os.Remove(askPass)

	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {

		return fmt.Errorf(
			"create mirror directory %s: %w",
			filepath.Dir(target),
			err,
		)
	}

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"clone",
		"--mirror",
		repo,
		target,
	)

	if err != nil {

		e.logger.Error(
			logging.Events.Mirror.CloneFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

		return err
	}

	e.logger.Duration(
		logging.Events.Mirror.CloneCompleted,
		repoName,
		time.Since(start),
	)

	return nil
}

func (e *Engine) updateMirror(ctx context.Context, target string) error {
	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.UpdateStarted,
		repoName,
	)

	askPass, err := e.createAskPassScript()
	if err != nil {
		return err
	}

	defer os.Remove(askPass)

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"-C",
		target,
		"remote",
		"update",
		"--prune",
	)

	if err != nil {

		e.logger.Error(
			logging.Events.Mirror.UpdateFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

		return err
	}

	e.logger.Duration(
		logging.Events.Mirror.UpdateCompleted,
		repoName,
		time.Since(start),
	)

	return nil
}

func (e *Engine) syncMirror(ctx context.Context, url string, target string) error {

	// clone if asset doesn't exist
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return e.cloneMirror(ctx, url, target)
	}

	// Validate the existing mirror before attempting to update it.
	if err := e.validateMirror(ctx, target); err != nil {

		// Corrupt mirrors cannot be updated; return error for retry logic.
		if errors.Is(err, ErrMirrorCorrupt) {

			// Log the corruption event
			repoName := filepath.Base(target)

			e.logger.Emit(
				logging.Entry{
					Level: logging.Critical,
					Event: "CorruptMirror",
					Repo:  repoName,
					Details: map[string]any{
						"action": "quarantined",
					},
				},
			)

			// TODO: Recovery is implemented in the next PR.
			if _, qerr := e.quarantineMirror(target); qerr != nil {
				return qerr
			}

		}

		return err
	}

	// update existing asset
	return e.updateMirror(ctx, target)
}
