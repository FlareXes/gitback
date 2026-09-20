package mirror

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/logging"
)

var ErrMirrorCorrupt = errors.New("mirror is corrupt")

func (e *Engine) validateMirror(ctx context.Context, target string) error {

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

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

		fsckErr := fmt.Errorf(
			"%w: %s",
			ErrMirrorCorrupt,
			strings.TrimSpace(string(output)),
		)

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
