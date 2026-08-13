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
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/flarexes/gitback/internal/state"
)

func Generate(cfg *config.Config, layout runtime.Layout) (*HealthReport, error) {

	report := &HealthReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),

		// Optimistic default. updateStatus function downgrades this based on
		// what the populate* functions below find.
		Status: "healthy",

		Retention: RetentionHealth{
			Enabled: cfg.Snapshot.Retention > 0,
			Keep:    cfg.Snapshot.Retention,
		},
	}

	// Each populate* function is independent and self-contained: it
	// only ever appends to warnings/recommendations on failure, never
	// returns an error. This keeps the report complete even when one
	// section can't be gathered.
	populateAssets(cfg, layout, report)
	populateQuarantine(cfg, report)
	populateSnapshots(cfg, report)
	populateDisk(cfg, report)

	populateWarnings(cfg, report)
	populateRecommendations(cfg, layout, report)
	updateStatus(cfg, report)

	return report, nil
}

func populateAssets(cfg *config.Config, layout runtime.Layout, report *HealthReport) {

	data, err := state.LoadMirrors(layout.MirrorsStateFile)
	if err != nil {

		if os.IsNotExist(err) {
			// Expected on a fresh install — nudge the user to sync.
			report.Warnings = append(report.Warnings, "mirror state unavailable")
			report.Recommendations = append(report.Recommendations, "run gitback sync")
		} else {
			// File exists but couldn't be read/parsed — likely
			// corruption, distinct from "hasn't synced yet".
			report.Warnings = append(
				report.Warnings,
				fmt.Sprintf("mirror state file is unreadable: %v", err),
			)
			report.Recommendations = append(
				report.Recommendations,
				fmt.Sprintf("inspect %s; it may be corrupted", layout.MirrorsStateFile),
			)
		}

		return
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

	// Gists are optional per config, so only count them if the user
	// has gist backup enabled — otherwise report.Gists stays zeroed
	// and PrintReport skips the section entirely.
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
}

// populateQuarantine counts mirrors that remain quarantined after
// automatic recovery attempts.
func populateQuarantine(cfg *config.Config, report *HealthReport) {

	repositories, err := countQuarantinedRepositories(cfg)
	if err != nil {
		report.Warnings = append(
			report.Warnings,
			fmt.Sprintf("could not inspect quarantined repositories: %v", err),
		)
	}

	gists, err := countQuarantinedGists(cfg)
	if err != nil {
		report.Warnings = append(
			report.Warnings,
			fmt.Sprintf("could not inspect quarantined gists: %v", err),
		)
	}

	report.Quarantine.Repositories = repositories
	report.Quarantine.Gists = gists
}

// populateSnapshots scans the snapshot output directory and records the
// count, total size, and most recent snapshot by filename (snapshot
// filenames are expected to sort lexicographically by creation time).
func populateSnapshots(cfg *config.Config, report *HealthReport) {

	entries, err := os.ReadDir(cfg.Snapshot.OutputDirectory)
	if err != nil {
		// Directory missing/unreadable is surfaced as a warning, not
		// a fatal error — the rest of the report is still useful.
		report.Warnings = append(
			report.Warnings,
			fmt.Sprintf("could not read snapshot directory: %v", err),
		)
		return
	}

	var snapshots []string
	var totalSize int64

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Only count GitBack's own snapshot archives; ignore any
		// other files a user might have placed in this directory.
		if !strings.HasSuffix(name, ".tar.zst") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			// Skip entries we can't stat (e.g. removed mid-scan)
			// rather than failing the whole scan over one file.
			continue
		}

		report.Snapshots.Count++
		totalSize += info.Size()
		snapshots = append(snapshots, name)
	}

	if len(snapshots) == 0 {
		return
	}

	report.Snapshots.Size = totalSize

	// Snapshot filenames are timestamp-prefixed, so a lexicographic
	// sort is equivalent to a chronological sort.
	sort.Strings(snapshots)
	report.Snapshots.Latest = snapshots[len(snapshots)-1]
}

// populateDisk reports free/total space for every distinct filesystem
// backing the mirror root and (if set) the snapshot output directory.
// Paths sharing a device are only reported once, since they share the
// same free-space figure.
func populateDisk(cfg *config.Config, report *HealthReport) {

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
			// A single unreadable path shouldn't prevent reporting
			// disk usage for the others.
			report.Warnings = append(
				report.Warnings,
				fmt.Sprintf("could not read disk usage for %s: %v", path, err),
			)
			continue
		}

		// Skip if we've already reported this device
		if _, ok := seen[disk.Device]; ok {
			continue
		}
		seen[disk.Device] = struct{}{}

		report.Disks = append(report.Disks, *disk)
	}
}

// populateWarnings appends human-readable warnings derived from counts
// already gathered by the populate* functions above. This is where
// thresholds (disk space, retention) are evaluated against config.
func populateWarnings(cfg *config.Config, report *HealthReport) {

	// Failed assets (repositories + gists)
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

	// Quarantine (repositories + gists)
	quarantined := report.Quarantine.Repositories + report.Quarantine.Gists
	if quarantined > 0 {

		report.Warnings = append(
			report.Warnings,
			fmt.Sprintf(
				"%d mirrors quarantined",
				quarantined,
			),
		)
	}

	// Disk space
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

	// Snapshot retention
	if cfg.Snapshot.Retention == 1 {

		report.Warnings = append(
			report.Warnings,
			"only one snapshot retained",
		)
	}
}

// populateRecommendations appends actionable next steps corresponding to
// the warnings above. Kept as a separate pass (rather than merged into
// populateWarnings) so warnings stay purely descriptive and
// recommendations stay purely prescriptive.
func populateRecommendations(cfg *config.Config, layout runtime.Layout, report *HealthReport) {

	// Failed assets
	if report.Repositories.Failed > 0 || report.Gists.Failed > 0 {
		report.Recommendations = append(
			report.Recommendations,
			fmt.Sprintf(
				"run `gitback sync` and inspect %s",
				layout.MirrorsStateFile,
			),
		)
	}

	// Quarantine
	if report.Quarantine.Repositories > 0 || report.Quarantine.Gists > 0 {

		report.Recommendations = append(
			report.Recommendations,
			fmt.Sprintf(
				"run `gitback sync` for automatic recovery; if mirrors remain quarantined afterwards, inspect them manually at %s",
				cfg.QuarantineDir(),
			),
		)
	}

	// Disk space
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

	// Snapshot retention
	if report.Snapshots.Count == 0 {
		report.Recommendations = append(
			report.Recommendations,
			"consider creating a snapshot",
		)
	}
}

// updateStatus derives the overall report.Status from everything gathered
// above. Severity only ever escalates (healthy -> warning -> critical) —
// disk space below threshold is treated as the most severe condition
// since it risks future sync/snapshot failures, not just past ones.
func updateStatus(cfg *config.Config, report *HealthReport) {

	report.Status = "healthy"

	if report.Repositories.Failed > 0 || report.Gists.Failed > 0 {
		report.Status = "warning"
	}

	if report.Quarantine.Repositories > 0 || report.Quarantine.Gists > 0 {
		report.Status = "warning"
	}

	for _, disk := range report.Disks {
		if disk.FreePercent < cfg.Health.MinimumFreeDiskPercent {
			report.Status = "critical"
			break
		}
	}
}

// diskUsage reports free/total space for the filesystem backing path,
// using statfs directly rather than a third-party dependency since only
// Linux is supported for now.
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
		// Fsid identifies the filesystem itself, used by populateDisk
		// to de-duplicate paths that share a device.
		Device: uint64(stat.Fsid.X__val[0]),
	}, nil
}
