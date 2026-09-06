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
