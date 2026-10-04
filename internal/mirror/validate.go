package mirror

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/flarexes/gitback/internal/logging"
)

var ErrMirrorCorrupt = errors.New("mirror is corrupt")

func (e *Engine) validateMirror(ctx context.Context, target string) error {

	repoName := assetName(target)

	e.logger.Emit(logging.Events.Mirror.FsckStarted, logging.WithAsset(repoName))

	fsck := exec.CommandContext(
		ctx,
		"git",
		"-C",
		target,
		"fsck",
		"--no-dangling",
	)

	output, err := fsck.CombinedOutput()
	if err != nil {

		// A canceled context kills fsck mid-check, which looks identical
		// to real corruption (non-zero exit). Without this check, an
		// interrupted but perfectly healthy mirror would be quarantined
		// for no reason.
		if isCancelled(err) {
			e.logger.Emit(
				logging.Events.Mirror.FsckInterrupted,
				logging.WithAsset(repoName),
				logging.WithCause(logging.CauseCancelled),
			)
			return err
		}

		detail := strings.TrimSpace(string(output))
		if detail == "" {
			// fsck can exit non-zero with no output at all — fall back
			// to the exit error itself
			detail = err.Error()
		}

		fsckErr := fmt.Errorf("%w: %s", ErrMirrorCorrupt, detail)

		e.logger.Emit(
			logging.Events.Mirror.FsckFailed,
			logging.WithAsset(repoName),
			logging.WithError(fsckErr),
			logging.WithCause(logging.CauseCorruption),
		)

		return fsckErr
	}

	e.logger.Emit(logging.Events.Mirror.FsckCompleted, logging.WithAsset(repoName))

	return nil
}
