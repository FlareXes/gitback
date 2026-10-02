// internal/mirror/quarantine.go

package mirror

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/filesystem"
	"github.com/flarexes/gitback/internal/logging"
)

// quarantineTimeLayout is the timestamp quarantineMirror appends when a
// mirror's quarantine path is already taken, e.g.
// "name.git.20260920T163629Z". canonicalMirrorName (orphans.go) must
// recognize exactly this shape so every copy of one mirror can be found
// again; a test pins the two together.
const quarantineTimeLayout = "20060102T150405Z"

// quarantineMirror moves a corrupt mirror out of the active mirror tree while
// preserving its relative directory structure. The quarantined mirror is kept
// until a verified replacement has been created.
func (e *Engine) quarantineMirror(target string) (string, error) {

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Emit(logging.Events.Mirror.QuarantineStarted, logging.WithAsset(repoName))

	relative, err := filepath.Rel(
		e.cfg.Storage.MirrorRoot,
		target,
	)
	if err != nil {
		return "", fmt.Errorf("determine quarantine path: %w", err)
	}

	quarantinePath := filepath.Join(
		e.cfg.QuarantineDir(),
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

		quarantinePath += "." + time.Now().UTC().Format(quarantineTimeLayout)
	}

	if err := os.Rename(target, quarantinePath); err != nil {

		e.logger.Emit(
			logging.Events.Mirror.QuarantineFailed,
			logging.WithAsset(repoName),
			logging.WithError(err),
		)

		return "", err
	}

	e.logger.Emit(
		logging.Events.Mirror.QuarantineCompleted,
		logging.WithAsset(repoName),
		logging.WithDetails(map[string]any{
			"quarantine_path": quarantinePath,
		}),
	)

	return quarantinePath, nil
}

// cleanupQuarantine removes every quarantined copy of the mirror at
// target: the plain "name.git" and any "name.git.<timestamp>" copies left
// by repeated quarantines.
func (e *Engine) cleanupQuarantine(target string) error {

	relative, err := filepath.Rel(
		e.cfg.Storage.MirrorRoot,
		target,
	)
	if err != nil {
		return fmt.Errorf("compute quarantine path: %w", err)
	}

	// All copies of one mirror sit side by side in the same directory,
	// so a single directory listing finds every one of them.
	dir := filepath.Join(e.cfg.QuarantineDir(), filepath.Dir(relative))
	base := filepath.Base(relative)

	entries, err := os.ReadDir(dir)
	if err != nil {

		// The normal case: nothing was ever quarantined under this owner.
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	var removed []string
	var errs []error

	for _, entry := range entries {

		// canonicalMirrorName strips the timestamp suffix, so this
		// matches "x.git" and "x.git.<timestamp>" but not a sibling like
		// "x.git.bak" or a different mirror.
		if !entry.IsDir() || canonicalMirrorName(entry.Name()) != base {
			continue
		}

		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			errs = append(errs, err)
			continue
		}

		removed = append(removed, entry.Name())
	}

	if len(removed) > 0 {
		e.logger.Emit(
			logging.Events.Mirror.QuarantineCleanupCompleted,
			logging.WithAsset(assetName(base)),
			logging.WithDetails(map[string]any{"removed": removed}),
		)
	}

	return errors.Join(errs...)
}

// CountQuarantined reports how many distinct mirrors currently have a
// quarantined copy, split into repositories and gists.
//
// Distinct matters: a mirror quarantined more than once leaves several
// entries on disk, but it is one mirror needing attention. Counting
// lives here, next to the code that decides the quarantine layout, so
// the count can't drift from what quarantineMirror actually produces.
//
// If one category can't be read, the other is still counted and the
// error is returned alongside the partial result.
func CountQuarantined(cfg *config.Config) (repositories int, gists int, err error) {

	root := cfg.QuarantineDir()

	var errs []error

	found, scanErr := scanRepositoryMirrors(root)
	if scanErr != nil {
		errs = append(errs, fmt.Errorf("repositories: %w", scanErr))
	}
	repositories = countDistinct(found)

	found, scanErr = scanGistMirrors(root)
	if scanErr != nil {
		errs = append(errs, fmt.Errorf("gists: %w", scanErr))
	}
	gists = countDistinct(found)

	return repositories, gists, errors.Join(errs...)
}

// countDistinct counts unique mirrors by their canonical relative path.
func countDistinct(found []foundMirror) int {

	seen := make(map[string]struct{}, len(found))

	for _, m := range found {
		seen[m.rel] = struct{}{}
	}

	return len(seen)
}
