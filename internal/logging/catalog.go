// internal/logging/catalog.go

package logging

type GitHubEvents struct {
	DiscoveryStarted   string
	DiscoveryCompleted string
	DiscoveryFailed    string
}

type MirrorEvents struct {
	CloneStarted   string
	CloneCompleted string
	CloneFailed    string

	UpdateStarted   string
	UpdateCompleted string
	UpdateFailed    string

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
}

type CooldownEvents struct {
	Started string
}

type EventCatalog struct {
	GitHub   GitHubEvents
	Mirror   MirrorEvents
	Snapshot SnapshotEvents
	Lock     LockEvents
	Health   HealthEvents
	Restore  RestoreEvents
	Sync     SyncEvents
	Cooldown CooldownEvents
}

var Events = EventCatalog{
	GitHub: GitHubEvents{
		DiscoveryStarted:   "repo_discovery_started",
		DiscoveryCompleted: "repo_discovery_completed",
		DiscoveryFailed:    "repo_discovery_failed",
	},

	Mirror: MirrorEvents{
		CloneStarted:   "mirror_clone_started",
		CloneCompleted: "mirror_clone_completed",
		CloneFailed:    "mirror_clone_failed",

		UpdateStarted:   "mirror_update_started",
		UpdateCompleted: "mirror_update_completed",
		UpdateFailed:    "mirror_update_failed",

		FsckStarted:   "mirror_fsck_started",
		FsckCompleted: "mirror_fsck_completed",
		FsckFailed:    "mirror_fsck_failed",

		StateSaveFailed: "mirror_state_save_failed",
	},

	Sync: SyncEvents{
		Started:   "sync_started",
		Completed: "sync_completed",
		Failed:    "sync_failed",
	},

	Snapshot: SnapshotEvents{
		Started:   "snapshot_started",
		Completed: "snapshot_completed",
		Failed:    "snapshot_failed",

		VerificationStarted: "snapshot_verification_started",
		VerificationPassed:  "snapshot_verification_passed",
		VerificationFailed:  "snapshot_verification_failed",
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

	Cooldown: CooldownEvents{
		Started: "cooldown_started",
	},

	Restore: RestoreEvents{
		Started:   "restore_started",
		Completed: "restore_completed",
		Failed:    "restore_failed",
	},
}
