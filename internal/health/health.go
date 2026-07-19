package health

import (
	"fmt"
	"os"
	"path/filepath"
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
			Enabled: cfg.Snapshot.Retention > 0,
			Keep:    cfg.Snapshot.Retention,
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
		config.MirrorsStateFile(),
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

	if cfg.GitHub.BackupGists {
		for _, gist := range data.Gists {

			report.Gists.Total++

			if gist.LastSuccess {

				report.Gists.Healthy++

			} else {

				report.Gists.Failed++
			}
		}
	}

	return nil
}

func populateSnapshots(cfg *config.Config, report *HealthReport) error {

	entries, err := os.ReadDir(
		cfg.Snapshot.OutputDirectory,
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

	report.Snapshots.Size = totalSize

	sort.Strings(snapshots)

	report.Snapshots.Latest = snapshots[len(snapshots)-1]

	return nil
}

func populateDisk(cfg *config.Config, report *HealthReport) error {

	// More locations can be added here in the future without changing
	// the rest of the implementation.
	paths := []string{
		cfg.Storage.MirrorRoot,
	}

	// Snapshots are optional.
	if cfg.Snapshot.OutputDirectory != "" {
		paths = append(paths, cfg.Snapshot.OutputDirectory)
	}

	// Multiple configured paths may live on the same filesystem.
	// Track device IDs so we only report each filesystem once.
	seen := make(map[uint64]struct{})

	for _, path := range paths {

		disk, err := diskUsage(path)

		if err != nil {
			return err
		}

		if _, ok := seen[disk.Device]; ok {
			continue
		}

		seen[disk.Device] = struct{}{}

		report.Disks = append(report.Disks, *disk)
	}

	return nil
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

	for _, disk := range report.Disks {
		if disk.FreePercent < cfg.Health.MinimumFreeDiskPercent {
			report.Warnings = append(
				report.Warnings,
				fmt.Sprintf(
					"disk space below configured threshold on %s",
					disk.Path,
				),
			)
		}
	}

	if cfg.Snapshot.Retention == 1 {

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
				config.MirrorsStateFile(),
			),
		)
	}

	for _, disk := range report.Disks {
		if disk.FreePercent < cfg.Health.MinimumFreeDiskPercent {
			report.Recommendations = append(
				report.Recommendations,
				fmt.Sprintf(
					"consider increasing available storage on %s",
					disk.Path,
				),
			)
		}
	}

	if report.Snapshots.Count == 0 {

		report.Recommendations = append(
			report.Recommendations,
			"consider creating a snapshot",
		)
	}
}

func updateStatus(cfg *config.Config, report *HealthReport) {

	report.Status = "healthy"

	if report.Repositories.Failed > 0 || report.Gists.Failed > 0 {

		report.Status = "warning"
	}

	for _, disk := range report.Disks {
		if disk.FreePercent < cfg.Health.MinimumFreeDiskPercent {
			report.Status = "critical"
			break
		}
	}
}

func diskUsage(path string) (*DiskHealth, error) {

	path = filepath.Clean(path)

	var stat syscall.Statfs_t

	if err := syscall.Statfs(path, &stat); err != nil {
		return nil, err
	}

	total := stat.Blocks * uint64(stat.Bsize)

	if total == 0 {
		return nil, fmt.Errorf(
			"filesystem %q reports zero capacity",
			path,
		)
	}

	free := stat.Bavail * uint64(stat.Bsize)

	return &DiskHealth{
		Path:        path,
		Free:        free,
		Total:       total,
		FreePercent: uint8((free * 100) / total),
		Device:      uint64(stat.Fsid.X__val[0]),
	}, nil
}
