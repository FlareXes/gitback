// internal/mirror/git.go

package mirror

import (
	"context"
	"os/exec"
	"time"

	"github.com/flarexes/gitback/internal/logging"
)

// Execute a git command with retry support.
func (e *Engine) runGit(
	ctx context.Context,
	repo string,
	env []string,
	args ...string,
) ([]byte, error) {

	var lastErr error

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

		// Linear backoff: attempt 1 -> 5s, attempt 2 -> 10s
		wait := time.Duration(attempt*5) * time.Second

		time.Sleep(wait)
	}

	return nil, lastErr
}
