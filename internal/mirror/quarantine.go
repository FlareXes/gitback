// internal/mirror/quarantine.go

package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/config"
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

	if err := os.MkdirAll(
		filepath.Dir(quarantinePath),
		0700,
	); err != nil {
		return "", fmt.Errorf("create quarantine directory: %w", err)
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
