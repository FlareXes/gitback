package health

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/state"
)

func Generate(cfg *config.Config) (*HealthReport, error) {

	report := &HealthReport{
		GeneratedAt: time.Now().
			UTC().
			Format(time.RFC3339),

		Status: "healthy",

		Retention: RetentionHealth{
			Enabled: cfg.SnapshotRetention > 0,
			Keep:    cfg.SnapshotRetention,
		},
	}

	if err := populateAssets(cfg, report); err != nil {
		return nil, err
	}

	if err := populateSnapshots(cfg, report); err != nil {
		return nil, err
	}

	if err := populateDisk(cfg, report); err != nil {
		return nil, err
	}

	populateWarnings(
		cfg,
		report,
	)

	populateRecommendations(
		cfg,
		report,
	)

	updateStatus(cfg, report)

	return report, nil
}

func populateAssets(cfg *config.Config, report *HealthReport) error {

	data, err := state.LoadMirrors(
		cfg.MirrorsStateFile,
	)

	if err != nil {

		report.Warnings = append(
			report.Warnings,
			"mirror state unavailable",
		)

		report.Recommendations = append(
			report.Recommendations,
			"run gitback sync",
		)

		return nil
	}

	report.Sync.StartedAt = data.SyncStartedAt

	report.Sync.CompletedAt = data.SyncCompletedAt

	for _, repo := range data.Repositories {

		report.Repositories.Total++

		if repo.LastSuccess {

			report.Repositories.Healthy++

		} else {

			report.Repositories.Failed++
		}
	}

	for _, gist := range data.Gists {

		report.Gists.Total++

		if gist.LastSuccess {

			report.Gists.Healthy++

		} else {

			report.Gists.Failed++
		}
	}

	return nil
}

func populateSnapshots(cfg *config.Config, report *HealthReport) error {

	entries, err := os.ReadDir(
		cfg.SnapshotDir,
	)

	if err != nil {
		return err
	}

	var snapshots []string
	var totalSize int64

	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		if !strings.HasSuffix(
			name,
			".tar.zst",
		) {
			continue
		}

		info, err := entry.Info()

		if err != nil {
			continue
		}

		report.Snapshots.Count++

		totalSize += info.Size()

		snapshots = append(
			snapshots,
			name,
		)
	}

	if len(snapshots) == 0 {
		return nil
	}

	report.Snapshots.Size = humanSize(totalSize)

	sort.Strings(snapshots)

	report.Snapshots.Latest = snapshots[len(snapshots)-1]

	return nil
}

func populateDisk(cfg *config.Config, report *HealthReport) error {

	var stat syscall.Statfs_t

	if err := syscall.Statfs(
		cfg.DataDir,
		&stat,
	); err != nil {

		return err
	}

	free := stat.Bavail * uint64(stat.Bsize)
	total := stat.Blocks * uint64(stat.Bsize)

	if total == 0 {
		return nil
	}

	report.Disk.FreePercent =
		int(
			(free * 100) / total,
		)

	report.Disk.Free =
		humanSize(
			int64(free),
		)

	report.Disk.MinimumPercent =
		cfg.MinimumFreeDiskPercent

	return nil
}

func updateStatus(cfg *config.Config, report *HealthReport) {

	report.Status = "healthy"

	if report.Repositories.Failed > 0 || report.Gists.Failed > 0 {

		report.Status = "warning"
	}

	if report.Disk.FreePercent <
		cfg.MinimumFreeDiskPercent {

		report.Status = "critical"
	}
}

func populateWarnings(cfg *config.Config, report *HealthReport) {

	failedAssets := report.Repositories.Failed + report.Gists.Failed

	if failedAssets > 0 {

		report.Warnings = append(
			report.Warnings,
			fmt.Sprintf(
				"%d assets unhealthy",
				failedAssets,
			),
		)
	}

	if report.Disk.FreePercent <
		cfg.MinimumFreeDiskPercent {

		report.Warnings = append(
			report.Warnings,
			"disk space below configured threshold",
		)
	}

	if cfg.SnapshotRetention == 1 {

		report.Warnings = append(
			report.Warnings,
			"only one snapshot retained",
		)
	}
}

func populateRecommendations(cfg *config.Config, report *HealthReport) {

	if report.Repositories.Failed > 0 || report.Gists.Failed > 0 {

		report.Recommendations = append(
			report.Recommendations,
			fmt.Sprintf(
				"run gitback sync and inspect %s",
				cfg.MirrorsStateFile,
			),
		)
	}

	if report.Disk.FreePercent < cfg.MinimumFreeDiskPercent {

		report.Recommendations = append(
			report.Recommendations,
			"consider increasing available storage",
		)
	}

	if report.Snapshots.Count == 0 {

		report.Recommendations = append(
			report.Recommendations,
			"consider creating a snapshot",
		)
	}
}

func humanSize(b int64) string {

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
