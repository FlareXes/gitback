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

	e.logger.Emit(logging.Events.Mirror.CloneStarted, logging.WithRepo(repoName))

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
		e.logger.Emit(
			logging.Events.Mirror.CloneFailed,
			logging.WithRepo(repoName),
			logging.WithError(fmt.Errorf("%s", gitErrorMessage(output, err))),
		)
		return err
	}

	e.logger.Emit(
		logging.Events.Mirror.CloneCompleted,
		logging.WithRepo(repoName),
		logging.WithDuration(time.Since(start)),
	)

	return nil
}

func (e *Engine) updateMirror(ctx context.Context, target string) error {
	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Emit(logging.Events.Mirror.UpdateStarted, logging.WithRepo(repoName))

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

		e.logger.Emit(
			logging.Events.Mirror.UpdateFailed,
			logging.WithRepo(repoName),
			logging.WithError(fmt.Errorf("%s", gitErrorMessage(output, err))),
		)

		return err
	}

	e.logger.Emit(
		logging.Events.Mirror.UpdateCompleted,
		logging.WithRepo(repoName),
		logging.WithDuration(time.Since(start)),
	)

	return nil
}

func (e *Engine) syncMirror(ctx context.Context, url string, target string) error {

	// Clone if asset doesn't exist.
	if _, err := os.Stat(target); os.IsNotExist(err) {

		if err := e.cloneMirror(ctx, url, target); err != nil {
			return err
		}

		// Remove any stale quarantined copy so the
		// quarantine directory only contains unresolved mirrors.
		repoName := filepath.Base(target)

		if err := e.cleanupQuarantine(target); err != nil {

			e.logger.Emit(
				logging.Events.Mirror.QuarantineCleanupFailed,
				logging.WithRepo(repoName),
				logging.WithError(err),
			)
		}

		return nil
	}

	// Validate the existing mirror before attempting to update it.
	if err := e.validateMirror(ctx, target); err != nil {

		// Corrupt mirrors cannot be updated; return error for retry logic.
		if errors.Is(err, ErrMirrorCorrupt) {

			// Log the corruption event.
			repoName := filepath.Base(target)

			e.logger.Emit(
				logging.Events.Mirror.CorruptionDetected,
				logging.WithRepo(repoName),
				logging.WithCause(logging.CauseCorruption),
			)

			// Quarantine the corrupt mirror.
			quarantinePath, qerr := e.quarantineMirror(target)
			if qerr != nil {
				return qerr
			}

			// Try to recover the corrupt mirror.
			if rerr := e.recoverCorruptMirror(ctx, url, target, quarantinePath); rerr != nil {

				e.logger.Emit(
					logging.Events.Mirror.RecoveryFailed,
					logging.WithRepo(repoName),
					logging.WithError(rerr),
				)

				return rerr
			}

			e.logger.Emit(logging.Events.Mirror.RecoverySucceeded, logging.WithRepo(repoName))

			return nil
		}

		return err
	}

	// Update existing asset.
	return e.updateMirror(ctx, target)
}

// recoverCorruptMirror clones a fresh mirror, validates it, and atomically replaces
// the active mirror. The quarantined mirror is removed only after the
// replacement has been verified.
func (e *Engine) recoverCorruptMirror(
	ctx context.Context,
	url string,
	target string,
	quarantine string,
) error {

	tmp := target + ".tmp"

	// Remove any stale temporary mirror from a previous failed run.
	if err := os.RemoveAll(tmp); err != nil {
		return fmt.Errorf("remove temporary mirror: %w", err)
	}

	// Ensure temporary files are cleaned up if replacement fails.
	defer os.RemoveAll(tmp)

	// Clone a fresh mirror to the temporary path.
	if err := e.cloneMirror(ctx, url, tmp); err != nil {
		return err
	}

	// Validate the fresh mirror before replacing the active one.
	if err := e.validateMirror(ctx, tmp); err != nil {
		return err
	}

	// Atomically replace the active mirror with the fresh one.
	if err := os.Rename(tmp, target); err != nil {
		return fmt.Errorf(
			"activate replacement mirror: %w",
			err,
		)
	}

	// Remove the quarantined mirror after successful replacement.
	if err := os.RemoveAll(quarantine); err != nil {
		e.logger.Emit(
			logging.Events.Mirror.QuarantineCleanupFailed,
			logging.WithRepo(filepath.Base(target)),
			logging.WithError(err),
		)
	}

	return nil
}

// gitErrorMessage picks the most useful diagnostic text available after
// a failed runGit call. When git ran and exited non-zero, output holds
// git's actual stderr/stdout text and err is just an *exec.ExitError
// carrying an exit code with no message — so output is preferred. When
// the process never started at all (e.g. permission denied on the git
// binary), output is empty and err itself carries the real message, so
// that's used as the fallback.
func gitErrorMessage(output []byte, err error) string {

	msg := strings.TrimSpace(string(output))

	if msg != "" {
		return msg
	}

	if err != nil {
		return err.Error()
	}

	return "unknown error"
}
