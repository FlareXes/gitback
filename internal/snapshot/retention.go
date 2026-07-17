// internal/snapshot/retention.go

package snapshot

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
)

// ApplyRetention removes old snapshots after a
// successful snapshot creation.
//
// Retention is disabled when configured <= 0.
func ApplyRetention(cfg *config.Config, logger *logging.Logger) error {

	snapshotDir := cfg.Snapshot.OutputDirectory
	snapshotRetentionCount := cfg.Snapshot.Retention

	if snapshotRetentionCount <= 0 {

		logger.Info(
			logging.Events.Snapshot.RetentionDisabled,
			"",
		)

		return nil
	}

	logger.Info(
		logging.Events.Snapshot.RetentionStarted,
		"",
	)

	entries, err := os.ReadDir(snapshotDir)

	if err != nil {
		return err
	}

	var snapshots []string

	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(name, ".tar.zst") {
			continue
		}

		snapshots = append(
			snapshots,
			name,
		)
	}

	//
	// Snapshot filenames use timestamps.
	// Lexical ordering matches chronological ordering.
	//
	sort.Strings(snapshots)

	if len(snapshots) <= snapshotRetentionCount {
		return nil
	}

	toDelete := snapshots[:len(snapshots)-snapshotRetentionCount]

	var deletedSnapshots []string
	var deletedChecksums []string
	var missingChecksums []string
	var failedDeletions []string

	for _, snapshot := range toDelete {

		// delete snapshot file
		snapshotPath := filepath.Join(
			snapshotDir,
			snapshot,
		)

		if err := os.Remove(snapshotPath); err != nil {

			failedDeletions = append(
				failedDeletions,
				snapshot,
			)

			logger.Error(
				logging.Events.Snapshot.RetentionFailed,
				"",
				err,
			)

			continue
		}

		deletedSnapshots = append(
			deletedSnapshots,
			snapshot,
		)

		// delete corresponding checksum file
		checksum := snapshot + ".sha256"

		checksumPath := filepath.Join(
			snapshotDir,
			checksum,
		)

		// if checksum file doesn't exist, skip it and keep filename for logs
		if _, err := os.Stat(checksumPath); os.IsNotExist(err) {

			missingChecksums = append(
				missingChecksums,
				checksum,
			)

			continue
		}

		if err := os.Remove(checksumPath); err != nil {

			failedDeletions = append(
				failedDeletions,
				checksum,
			)

			logger.Error(
				logging.Events.Snapshot.RetentionFailed,
				"",
				err,
			)

			continue
		}

		deletedChecksums = append(
			deletedChecksums,
			checksum,
		)
	}

	logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Snapshot.RetentionCompleted,

			Details: map[string]any{
				"retention": snapshotRetentionCount,

				"deleted_snapshots": deletedSnapshots,

				"deleted_checksums": deletedChecksums,

				"missing_checksums": missingChecksums,

				"failed_deletions": failedDeletions,
			},
		},
	)

	return nil
}
