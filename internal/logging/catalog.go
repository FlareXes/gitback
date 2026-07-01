// internal/logging/catalog.go

package logging

type GitHubEvents struct {
	DiscoveryStarted   string
	DiscoveryCompleted string
	DiscoveryFailed    string
	DiscoverySummary   string

	PageFetched string

	InventoryLoaded string

	RateLimit string
}

type InventoryEvents struct {
	Missing string
	Empty   string
}

type MirrorEvents struct {
	CloneStarted   string
	CloneCompleted string
	CloneFailed    string

	UpdateStarted   string
	UpdateCompleted string
	UpdateFailed    string

	Retry string

	FsckStarted   string
	FsckCompleted string
	FsckFailed    string

	StateSaveFailed string
}

type SnapshotEvents struct {
	Started   string
	Completed string
	Failed    string

	VerificationStarted string
	VerificationPassed  string
	VerificationFailed  string

	ArchiveStarted   string
	ArchiveCompleted string

	CompressionStarted   string
	CompressionCompleted string

	ChecksumStarted   string
	ChecksumCompleted string

	Summary string

	RetentionDisabled  string
	RetentionStarted   string
	RetentionCompleted string
	RetentionFailed    string

	CollisionDetected string
}

type LockEvents struct {
	Acquired string
	Released string
	Busy     string
}

type HealthEvents struct {
	LowDisk      string
	RepoFailure  string
	StaleBackup  string
	HealthReport string
}

type RestoreEvents struct {
	Started   string
	Completed string
	Failed    string
}

type SyncEvents struct {
	Started   string
	Completed string
	Failed    string

	// Run-level summary.
	Summary string
}

type FilesystemEvents struct {
	DirectoryRecreated string
}

type DoctorEvents struct {
	ReportGenerated string
}

type EventCatalog struct {
	GitHub     GitHubEvents
	Inventory  InventoryEvents
	Mirror     MirrorEvents
	Snapshot   SnapshotEvents
	Lock       LockEvents
	Health     HealthEvents
	Restore    RestoreEvents
	Sync       SyncEvents
	Filesystem FilesystemEvents
	Doctor     DoctorEvents
}

var Events = EventCatalog{
	GitHub: GitHubEvents{
		DiscoveryStarted:   "discovery_started",
		DiscoveryCompleted: "discovery_completed",
		DiscoveryFailed:    "discovery_failed",

		DiscoverySummary: "discovery_summary",

		PageFetched: "github_page_fetched",

		InventoryLoaded: "inventory_loaded",

		RateLimit: "github_rate_limit",
	},

	Inventory: InventoryEvents{
		Missing: "inventory_missing",
		Empty:   "inventory_empty",
	},

	Mirror: MirrorEvents{
		CloneStarted:   "mirror_clone_started",
		CloneCompleted: "mirror_clone_completed",
		CloneFailed:    "mirror_clone_failed",

		UpdateStarted:   "mirror_update_started",
		UpdateCompleted: "mirror_update_completed",
		UpdateFailed:    "mirror_update_failed",

		Retry: "mirror_retry",

		FsckStarted:   "mirror_fsck_started",
		FsckCompleted: "mirror_fsck_completed",
		FsckFailed:    "mirror_fsck_failed",

		StateSaveFailed: "mirror_state_save_failed",
	},

	Sync: SyncEvents{
		Started:   "sync_started",
		Completed: "sync_completed",
		Failed:    "sync_failed",

		Summary: "sync_summary",
	},

	Snapshot: SnapshotEvents{
		Started:   "snapshot_started",
		Completed: "snapshot_completed",
		Failed:    "snapshot_failed",

		VerificationStarted: "snapshot_verification_started",
		VerificationPassed:  "snapshot_verification_passed",
		VerificationFailed:  "snapshot_verification_failed",

		ArchiveStarted:   "snapshot_archive_started",
		ArchiveCompleted: "snapshot_archive_completed",

		CompressionStarted:   "snapshot_compression_started",
		CompressionCompleted: "snapshot_compression_completed",

		ChecksumStarted:   "snapshot_checksum_started",
		ChecksumCompleted: "snapshot_checksum_completed",

		Summary: "snapshot_summary",

		RetentionDisabled:  "snapshot_retention_disabled",
		RetentionStarted:   "snapshot_retention_started",
		RetentionCompleted: "snapshot_retention_completed",
		RetentionFailed:    "snapshot_retention_failed",

		CollisionDetected: "snapshot_collision_detected",
	},

	Lock: LockEvents{
		Acquired: "lock_acquired",
		Released: "lock_released",
		Busy:     "lock_busy",
	},

	Health: HealthEvents{
		LowDisk:      "health_low_disk",
		RepoFailure:  "health_repo_failure",
		StaleBackup:  "health_stale_backup",
		HealthReport: "health_report_generated",
	},

	Restore: RestoreEvents{
		Started:   "restore_started",
		Completed: "restore_completed",
		Failed:    "restore_failed",
	},

	Filesystem: FilesystemEvents{
		DirectoryRecreated: "filesystem_directory_recreated",
	},

	Doctor: DoctorEvents{
		ReportGenerated: "doctor_report_generated",
	},
}
