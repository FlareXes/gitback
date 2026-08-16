// internal/mirror/git.go

package mirror

import (
	"context"
	"os/exec"
	"time"

	"github.com/flarexes/gitback/internal/logging"
)

// Execute a git command with retry support.
func (e *Engine) runGit(ctx context.Context, repo string, env []string, args ...string) ([]byte, error) {

	var lastErr error
	var lastOutput []byte

	retryAttempts := e.cfg.Sync.RetryAttempts

	for attempt := 1; attempt <= retryAttempts; attempt++ {

		cmd := exec.CommandContext(
			ctx,
			"git",
			args...,
		)
		cmd.Env = env

		output, err := cmd.CombinedOutput()
		if err == nil {
			return output, nil
		}

		lastErr = err
		lastOutput = output

		// If the context is already canceled (e.g. Ctrl+C/SIGTERM via
		// runCancelable), stop retrying immediately rather than
		// attempting another doomed subprocess spawn.
		if ctx.Err() != nil {
			return lastOutput, ctx.Err()
		}

		if attempt == retryAttempts {
			break
		}

		e.logger.Emit(
			logging.Entry{
				Level: logging.Warn,
				Event: logging.Events.Mirror.Retry,

				Repo: repo,

				Details: map[string]any{
					"attempt":      attempt,
					"max_attempts": retryAttempts,
				},
			},
		)

		// Linear backoff: attempt 1 -> 5s, attempt 2 -> 10s.
		//
		// time.Sleep is not context-aware — it always runs for its
		// full duration regardless of cancellation. Using a select on
		// ctx.Done() alongside a timer means a signal arriving mid-wait
		// interrupts the backoff immediately instead of forcing the
		// shutdown to wait out up to 10s of a pointless sleep.
		wait := time.Duration(attempt*5) * time.Second
		timer := time.NewTimer(wait)

		select {
		case <-ctx.Done():
			timer.Stop()
			return lastOutput, ctx.Err()
		case <-timer.C:
		}
	}

	return lastOutput, lastErr
}
