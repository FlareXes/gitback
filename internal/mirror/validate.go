package mirror

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/logging"
)

func (e *Engine) validateMirror(ctx context.Context, target string) error {

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.FsckStarted,
		repoName,
	)

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
			"git fsck: %s",
			strings.TrimSpace(string(output)),
		)

		e.logger.Error(
			logging.Events.Mirror.FsckFailed,
			repoName,
			fsckErr,
		)

		return fsckErr
	}

	e.logger.Info(
		logging.Events.Mirror.FsckCompleted,
		repoName,
	)

	return nil
}
