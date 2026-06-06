// internal/health/health.go

package health

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/state"
)

func Generate(cfg *config.Config) (*HealthReport, error) {

	report := &HealthReport{}

	mirrors, err := state.Load(cfg.MirrorsStateFile)

	if err != nil {
		return nil, err
	}

	report.RepositoryCount = len(mirrors.Repositories)

	report.LastSync = mirrors.SyncCompletedAt

	for _, repo := range mirrors.Repositories {

		if repo.LastSuccess {
			report.HealthyCount++
		} else {
			report.FailedCount++
		}
	}

	// Snapshot information
	entries, err := os.ReadDir(cfg.SnapshotDir)

	if err == nil {

		var newest string

		for _, entry := range entries {

			if entry.IsDir() {
				continue
			}

			if filepath.Ext(entry.Name()) != ".zst" {
				continue
			}

			report.SnapshotCount++

			// Add snapshot size to total.
			info, err := entry.Info()
			if err == nil {
				report.SnapshotBytes += uint64(info.Size())
			}

			// Snapshot filenames use:
			//
			// 2026-06-06T10-00-00Z.tar.zst
			//
			// Lexicographic sorting works because
			// YYYY-MM-DD is naturally sortable.
			if entry.Name() > newest {
				newest = entry.Name()
			}
		}

		report.LastSnapshot = newest
	}

	// Disk usage.
	var stat syscall.Statfs_t

	if err := syscall.Statfs(cfg.DataDir, &stat); err == nil {

		total := stat.Blocks
		free := stat.Bavail

		if total > 0 {

			report.DiskFreePercent = int(
				(free * 100) / total,
			)
		}
	}

	// Recommendations
	report.buildRecommendations(cfg)

	return report, nil
}

func (r *HealthReport) buildRecommendations(cfg *config.Config) {

	if r.FailedCount > 0 {

		r.Recommendations = append(
			r.Recommendations,
			Recommendation{
				Severity: "WARN",
				Message: fmt.Sprintf(
					"%d repositories failed during last sync",
					r.FailedCount,
				),
			},
		)
	}

	if r.SnapshotCount == 0 {

		r.Recommendations = append(
			r.Recommendations,
			Recommendation{
				Severity: "WARN",
				Message:  "no snapshots found",
			},
		)
	}

	if r.DiskFreePercent < cfg.MinimumFreeDiskPercent {

		r.Recommendations = append(
			r.Recommendations,
			Recommendation{
				Severity: "CRITICAL",
				Message: fmt.Sprintf(
					"disk free space below %d%%",
					cfg.MinimumFreeDiskPercent,
				),
			},
		)
	}
}

func HumanSize(b uint64) string {

	// Unit names in order.
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	// Work with float64 so we can divide repeatedly.
	size := float64(b)

	// Current unit index.
	// 0 = B
	// 1 = KB
	// 2 = MB
	// ...
	unit := 0

	// Keep dividing by 1024 until the number
	// becomes smaller than 1024.
	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}

	// For bytes, don't show decimal places.
	if unit == 0 {
		return fmt.Sprintf("%d %s", b, units[unit])
	}

	return fmt.Sprintf("%.1f %s", size, units[unit])
}
