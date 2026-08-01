// internal/mirror/quarantine.go

package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/filesystem"
	"github.com/flarexes/gitback/internal/logging"
)

// quarantineMirror moves a corrupt mirror out of the active mirror tree while
// preserving its relative directory structure. The quarantined mirror is kept
// until a verified replacement has been created.
func (e *Engine) quarantineMirror(target string) (string, error) {

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.QuarantineStarted,
		repoName,
	)

	relative, err := filepath.Rel(
		e.cfg.Storage.MirrorRoot,
		target,
	)
	if err != nil {
		return "", fmt.Errorf("determine quarantine path: %w", err)
	}

	quarantinePath := filepath.Join(
		config.QuarantineDir(e.cfg),
		relative,
	)

	// Create quarantine directory structure for appropriate resource.
	if _, err := filesystem.CreateDirectory(
		filepath.Dir(quarantinePath),
	); err != nil {

		return "", fmt.Errorf(
			"create quarantine directory: %w",
			err,
		)
	}

	// If a previous quarantined mirror already exists, preserve it by adding
	// a timestamp to the new quarantine path.
	if _, err := os.Stat(quarantinePath); err == nil {

		quarantinePath += "." + time.Now().UTC().Format("20060102T150405Z")
	}

	if err := os.Rename(target, quarantinePath); err != nil {

		e.logger.Error(
			logging.Events.Mirror.QuarantineFailed,
			repoName,
			err,
		)

		return "", err
	}

	e.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Mirror.QuarantineCompleted,
			Repo:  repoName,
			Details: map[string]any{
				"quarantine_path": quarantinePath,
			},
		},
	)

	return quarantinePath, nil
}

func (e *Engine) cleanupQuarantine(target string) error {

	relative, err := filepath.Rel(
		e.cfg.Storage.MirrorRoot,
		target,
	)

	if err != nil {
		return fmt.Errorf("compute quarantine path: %w", err)
	}

	quarantine := filepath.Join(
		config.QuarantineDir(e.cfg),
		relative,
	)

	if _, err := os.Stat(quarantine); err != nil {

		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	repoName := filepath.Base(target)

	if err := os.RemoveAll(quarantine); err != nil {
		return err
	}

	e.logger.Info(
		logging.Events.Mirror.QuarantineCleanupCompleted,
		repoName,
	)

	return nil
}
