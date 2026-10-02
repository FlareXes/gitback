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

// isCancelled reports whether err is (or wraps) a context cancellation,
// as opposed to a genuine operation failure.
func isCancelled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

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

// cloneMirror clones repo into target.
//
// The clone is written to a staging path (target + ".tmp") and only
// renamed into place once it completes successfully.
func (e *Engine) cloneMirror(ctx context.Context, repo string, target string) error {

	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(repo),
		".git",
	)

	e.logger.Emit(logging.Events.Mirror.CloneStarted, logging.WithAsset(repoName))

	askPass, err := e.createAskPassScript()
	if err != nil {
		return fmt.Errorf("create askpass script in %s: %w", e.layout.TempDir, err)
	}

	defer os.Remove(askPass)

	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {

		return fmt.Errorf(
			"create mirror directory %s: %w",
			filepath.Dir(target),
			err,
		)
	}

	staging := target + ".tmp"

	// Remove any stale staging directory (.tmp) left by a previous interrupted
	// clone attempt at this same path.
	if err := os.RemoveAll(staging); err != nil {
		return fmt.Errorf("remove stale staging directory %s: %w", staging, err)
	}

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"clone",
		"--mirror",
		repo,
		staging,
	)

	if err != nil {

		// Whatever partial state the clone left in staging is not
		// useful — a git clone that failed partway through cannot be
		// resumed by retrying against the same directory. Clean it up
		// so the next attempt starts from nothing rather than
		// accumulating .tmp directories with every failed attempt.
		_ = os.RemoveAll(staging)

		if isCancelled(err) {
			e.logger.Emit(
				logging.Events.Mirror.CloneInterrupted,
				logging.WithAsset(repoName),
				logging.WithCause(logging.CauseCancelled),
			)
			return err
		}

		e.logger.Emit(
			logging.Events.Mirror.CloneFailed,
			logging.WithAsset(repoName),
			logging.WithError(fmt.Errorf("%s", gitErrorMessage(output, err))),
		)
		return err
	}

	// runGit reporting success is not, by itself, proof staging exists on
	// disk — verify directly before the rename depends on it. Surfacing a
	// clear error here, with git's own output attached, is far more useful
	// for diagnosis than letting a missing staging directory show up later
	// as an opaque os.Rename "no such file or directory".
	if _, statErr := os.Stat(staging); statErr != nil {
		return fmt.Errorf(
			"git reported success but no mirror was produced at %s (git output: %s): %w",
			staging,
			strings.TrimSpace(string(output)),
			statErr,
		)
	}

	// The clone completed fully — activate it.
	if err := os.Rename(staging, target); err != nil {
		return fmt.Errorf("activate cloned mirror: %w", err)
	}

	e.logger.Emit(
		logging.Events.Mirror.CloneCompleted,
		logging.WithAsset(repoName),
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

	e.logger.Emit(logging.Events.Mirror.UpdateStarted, logging.WithAsset(repoName))

	askPass, err := e.createAskPassScript()
	if err != nil {
		return fmt.Errorf("create askpass script in %s: %w", e.layout.TempDir, err)
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

		if isCancelled(err) {
			e.logger.Emit(
				logging.Events.Mirror.UpdateInterrupted,
				logging.WithAsset(repoName),
				logging.WithCause(logging.CauseCancelled),
			)
			return err
		}

		e.logger.Emit(
			logging.Events.Mirror.UpdateFailed,
			logging.WithAsset(repoName),
			logging.WithError(fmt.Errorf("%s", gitErrorMessage(output, err))),
		)

		return err
	}

	e.logger.Emit(
		logging.Events.Mirror.UpdateCompleted,
		logging.WithAsset(repoName),
		logging.WithDuration(time.Since(start)),
	)

	return nil
}

// refreshMirror brings the live mirror at target up to date: it clones a
// missing mirror, recovers a corrupt one (quarantine, fresh clone, swap),
// or updates a healthy one. It does not touch leftover quarantined
// copies; syncMirror does that once this succeeds.
func (e *Engine) refreshMirror(ctx context.Context, url string, target string) error {

	// Clone if asset doesn't exist.
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return e.cloneMirror(ctx, url, target)
	}

	// Validate the existing mirror before attempting to update it.
	if err := e.validateMirror(ctx, target); err != nil {

		// Corrupt mirrors cannot be updated; return error for retry logic.
		if errors.Is(err, ErrMirrorCorrupt) {

			// Log the corruption event.
			repoName := strings.TrimSuffix(filepath.Base(target), ".git")

			e.logger.Emit(
				logging.Events.Mirror.CorruptionDetected,
				logging.WithAsset(repoName),
				logging.WithCause(logging.CauseCorruption),
			)

			// Quarantine the corrupt mirror.
			quarantinePath, qerr := e.quarantineMirror(target)
			if qerr != nil {
				return qerr
			}

			if ctx.Err() != nil {
				e.logger.Emit(
					logging.Events.Mirror.RecoveryDeferred,
					logging.WithAsset(repoName),
					logging.WithCause(logging.CauseCancelled),
				)
				return ctx.Err()
			}

			// Try to recover the corrupt mirror.
			if rerr := e.recoverCorruptMirror(ctx, url, target, quarantinePath); rerr != nil {

				if isCancelled(rerr) {
					e.logger.Emit(
						logging.Events.Mirror.RecoveryInterrupted,
						logging.WithAsset(repoName),
						logging.WithCause(logging.CauseCancelled),
					)
					return rerr
				}

				e.logger.Emit(
					logging.Events.Mirror.RecoveryFailed,
					logging.WithAsset(repoName),
					logging.WithError(rerr),
				)

				return rerr
			}

			e.logger.Emit(logging.Events.Mirror.RecoverySucceeded, logging.WithAsset(repoName))

			return nil
		}

		return err
	}

	// Update existing asset.
	return e.updateMirror(ctx, target)
}

// syncMirror brings the mirror at target up to date, then clears any
// quarantined copies of it.
//
// The rule is that quarantine holds only unresolved mirrors, and a
// mirror is resolved once its live copy is verified healthy.
func (e *Engine) syncMirror(ctx context.Context, url string, target string) error {

	if err := e.refreshMirror(ctx, url, target); err != nil {
		return err
	}

	// Best-effort: failing to tidy quarantine must not turn a mirror
	// that synced successfully into a failed one.
	if err := e.cleanupQuarantine(target); err != nil {
		e.logger.Emit(
			logging.Events.Mirror.QuarantineCleanupFailed,
			logging.WithAsset(strings.TrimSuffix(filepath.Base(target), ".git")),
			logging.WithError(err),
		)
	}

	return nil
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

	// cloneMirror already stages to target+".tmp" and only renames into
	// place on success, so recovery's "clone fresh, verify, activate"
	// need is met by cloneMirror + validateMirror directly.
	if err := e.cloneMirror(ctx, url, target); err != nil {
		return err
	}

	if err := e.validateMirror(ctx, target); err != nil {
		return err
	}

	// Remove the quarantined mirror after successful replacement.
	if err := os.RemoveAll(quarantine); err != nil {
		e.logger.Emit(
			logging.Events.Mirror.QuarantineCleanupFailed,
			logging.WithAsset(filepath.Base(target)),
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
