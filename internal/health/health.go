package health

import (
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

	if err := populateRepositories(cfg, report); err != nil {
		return nil, err
	}

	if err := populateSnapshots(cfg, report); err != nil {
		return nil, err
	}

	if err := populateDisk(cfg, report); err != nil {
		return nil, err
	}

	updateStatus(cfg, report)

	return report, nil
}

func populateRepositories(cfg *config.Config, report *HealthReport) error {

	data, err := state.Load(
		cfg.MirrorsStateFile,
	)

	if err != nil {
		return err
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

		report.Snapshots.SizeBytes +=
			info.Size()

		snapshots = append(
			snapshots,
			name,
		)
	}

	if len(snapshots) == 0 {
		return nil
	}

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

	report.Disk.MinimumPercent =
		cfg.MinimumFreeDiskPercent

	return nil
}

func updateStatus(cfg *config.Config, report *HealthReport) {

	if report.Repositories.Failed > 0 {
		report.Status = "degraded"
	}

	if report.Disk.FreePercent <
		cfg.MinimumFreeDiskPercent {

		report.Status = "warning"
	}
}
