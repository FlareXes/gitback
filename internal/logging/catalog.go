// internal/logging/catalog.go
//
// This file is the single source of truth for every event gitback can
// log. Logger.Emit only accepts an EventDef — there is no way to log a
// bare string — so adding a new log line always means adding an entry
// here first.
//
// Rules for adding a new event:
//  1. Pick the right Component* constant (types.go). Add a new one
//     only if this event genuinely belongs to a subsystem that
//     doesn't have one yet.
//  2. Give it a Code unique within that Component, snake_case,
//     verb-led ("clone_failed", not "failed_clone").
//  3. Set Level to the severity this event ALWAYS has. If the same
//     condition is sometimes worse than other times, that's two
//     events, not one event logged at two levels.
//  4. Write Message as a fixed, complete description — no variables.
//     Good: "Failed to clone repository mirror". Bad: "clone failed".
//  5. Write Remediation for anything Warn or above: a concrete next
//     action. Leave it empty only for Info, or when the action really
//     is "nothing — this will be retried automatically" (say that).
//  6. Check whether an existing EventDef already fits before adding a
//     new one — don't create "clone_error" next to "clone_failed".
package logging

type GitHubEvents struct {
	DiscoveryStarted   EventDef
	DiscoveryCompleted EventDef
	DiscoveryFailed    EventDef
	DiscoverySummary   EventDef
	PageFetched        EventDef
	InventoryLoaded    EventDef
	RateLimit          EventDef
}

type InventoryEvents struct {
	Missing    EventDef
	Empty      EventDef
	ReadFailed EventDef
}

type MirrorEvents struct {
	CloneStarted   EventDef
	CloneCompleted EventDef
	CloneFailed    EventDef

	UpdateStarted   EventDef
	UpdateCompleted EventDef
	UpdateFailed    EventDef

	Retry EventDef

	FsckStarted   EventDef
	FsckCompleted EventDef
	FsckFailed    EventDef

	QuarantineStarted          EventDef
	QuarantineCompleted        EventDef
	QuarantineFailed           EventDef
	QuarantineCleanupCompleted EventDef
	QuarantineCleanupFailed    EventDef

	CorruptionDetected EventDef
	RecoverySucceeded  EventDef
	RecoveryFailed     EventDef

	StateSaveFailed EventDef
}

type SyncEvents struct {
	Started   EventDef
	Completed EventDef
	Failed    EventDef
	Summary   EventDef
}

type SnapshotEvents struct {
	Started   EventDef
	Completed EventDef
	Failed    EventDef

	VerificationStarted EventDef
	VerificationPassed  EventDef
	VerificationFailed  EventDef

	ArchiveStarted   EventDef
	ArchiveCompleted EventDef

	CompressionStarted   EventDef
	CompressionCompleted EventDef

	ChecksumStarted   EventDef
	ChecksumCompleted EventDef

	Summary EventDef

	RetentionDisabled  EventDef
	RetentionStarted   EventDef
	RetentionCompleted EventDef
	RetentionFailed    EventDef

	CollisionDetected EventDef
}

type LockEvents struct {
	Acquired EventDef
	Released EventDef
	Busy     EventDef
}

type HealthEvents struct {
	HealthReport EventDef
}

type DoctorEvents struct {
	ReportGenerated EventDef
}

type LogRetentionEvents struct {
	Pruned           EventDef
	PruneFailed      EventDef
	UnrecognizedFile EventDef
}

type EventCatalog struct {
	GitHub       GitHubEvents
	Inventory    InventoryEvents
	Mirror       MirrorEvents
	Sync         SyncEvents
	Snapshot     SnapshotEvents
	Lock         LockEvents
	Health       HealthEvents
	Doctor       DoctorEvents
	LogRetention LogRetentionEvents
}

var Events = EventCatalog{

	GitHub: GitHubEvents{
		DiscoveryStarted: EventDef{
			Component: ComponentGitHub,
			Code:      "discovery_started",
			Level:     Info,
			Message:   "GitHub discovery started",
		},
		DiscoveryCompleted: EventDef{
			Component: ComponentGitHub,
			Code:      "discovery_completed",
			Level:     Info,
			Message:   "GitHub discovery completed for a resource type",
		},
		DiscoveryFailed: EventDef{
			Component:   ComponentGitHub,
			Code:        "discovery_failed",
			Level:       Error,
			Message:     "GitHub discovery failed",
			Remediation: "Run `gitback doctor` to check network connectivity and token validity, then retry.",
		},
		DiscoverySummary: EventDef{
			Component: ComponentGitHub,
			Code:      "discovery_summary",
			Level:     Info,
			Message:   "GitHub discovery run summary",
		},
		PageFetched: EventDef{
			Component: ComponentGitHub,
			Code:      "page_fetched",
			Level:     Info,
			Message:   "Fetched a page of results from the GitHub API",
		},
		InventoryLoaded: EventDef{
			Component: ComponentGitHub,
			Code:      "inventory_loaded",
			Level:     Info,
			Message:   "Inventory file written",
		},
		RateLimit: EventDef{
			Component: ComponentGitHub,
			Code:      "rate_limit",
			Level:     Info,
			Message:   "GitHub API rate limit status",
		},
	},

	Inventory: InventoryEvents{
		Missing: EventDef{
			Component:   ComponentInventory,
			Code:        "missing",
			Level:       Warn,
			Message:     "Inventory file not found",
			Remediation: "Run `gitback discover` to generate it.",
		},
		Empty: EventDef{
			Component:   ComponentInventory,
			Code:        "empty",
			Level:       Warn,
			Message:     "Inventory file is empty",
			Remediation: "Run `gitback discover` to populate it.",
		},
		ReadFailed: EventDef{
			Component:   ComponentInventory,
			Code:        "read_failed",
			Level:       Error,
			Message:     "Inventory file could not be read",
			Remediation: "Check file permissions; if the file is corrupted, run `gitback discover` to regenerate it.",
		},
	},

	Mirror: MirrorEvents{
		CloneStarted: EventDef{
			Component: ComponentMirror,
			Code:      "clone_started",
			Level:     Info,
			Message:   "Mirror clone started",
		},
		CloneCompleted: EventDef{
			Component: ComponentMirror,
			Code:      "clone_completed",
			Level:     Info,
			Message:   "Mirror clone completed",
		},
		CloneFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "clone_failed",
			Level:       Error,
			Message:     "Mirror clone failed",
			Remediation: "Check network connectivity and GitHub token permissions; gitback retries automatically on the next sync.",
		},

		UpdateStarted: EventDef{
			Component: ComponentMirror,
			Code:      "update_started",
			Level:     Info,
			Message:   "Mirror update started",
		},
		UpdateCompleted: EventDef{
			Component: ComponentMirror,
			Code:      "update_completed",
			Level:     Info,
			Message:   "Mirror update completed",
		},
		UpdateFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "update_failed",
			Level:       Error,
			Message:     "Mirror update failed",
			Remediation: "Check network connectivity and GitHub token permissions; gitback retries automatically on the next sync.",
		},

		Retry: EventDef{
			Component:   ComponentMirror,
			Code:        "retry",
			Level:       Warn,
			Message:     "Retrying a failed git operation",
			Remediation: "No action needed yet; investigate if this repository keeps failing across multiple sync runs.",
		},

		FsckStarted: EventDef{
			Component: ComponentMirror,
			Code:      "fsck_started",
			Level:     Info,
			Message:   "Mirror integrity check started",
		},
		FsckCompleted: EventDef{
			Component: ComponentMirror,
			Code:      "fsck_completed",
			Level:     Info,
			Message:   "Mirror integrity check passed",
		},
		FsckFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "fsck_failed",
			Level:       Error,
			Message:     "Mirror integrity check failed",
			Remediation: "The mirror will be quarantined and gitback will attempt automatic recovery via a fresh clone.",
		},

		QuarantineStarted: EventDef{
			Component: ComponentMirror,
			Code:      "quarantine_started",
			Level:     Info,
			Message:   "Corrupt mirror quarantine started",
		},
		QuarantineCompleted: EventDef{
			Component: ComponentMirror,
			Code:      "quarantine_completed",
			Level:     Info,
			Message:   "Corrupt mirror quarantined",
		},
		QuarantineFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "quarantine_failed",
			Level:       Error,
			Message:     "Failed to quarantine a corrupt mirror",
			Remediation: "Check filesystem permissions on the quarantine directory; manual intervention may be required.",
		},
		QuarantineCleanupCompleted: EventDef{
			Component: ComponentMirror,
			Code:      "quarantine_cleanup_completed",
			Level:     Info,
			Message:   "Quarantined mirror removed after successful recovery",
		},
		QuarantineCleanupFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "quarantine_cleanup_failed",
			Level:       Warn,
			Message:     "Failed to remove a quarantined mirror after recovery",
			Remediation: "Manually inspect and remove the stale entry under the quarantine directory.",
		},

		CorruptionDetected: EventDef{
			Component:   ComponentMirror,
			Code:        "corruption_detected",
			Level:       Critical,
			Message:     "Mirror corruption detected",
			Remediation: "gitback will quarantine the mirror and attempt automatic recovery via a fresh clone.",
		},
		RecoverySucceeded: EventDef{
			Component: ComponentMirror,
			Code:      "recovery_succeeded",
			Level:     Info,
			Message:   "Corrupt mirror recovered successfully",
		},
		RecoveryFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "recovery_failed",
			Level:       Critical,
			Message:     "Automatic recovery of a corrupt mirror failed",
			Remediation: "Manual intervention required: inspect the quarantined mirror and consider a manual re-clone.",
		},

		StateSaveFailed: EventDef{
			Component:   ComponentMirror,
			Code:        "state_save_failed",
			Level:       Error,
			Message:     "Failed to save mirror sync state",
			Remediation: "Check disk space and permissions on the state directory; this run's results may be missing from `gitback health`.",
		},
	},

	Sync: SyncEvents{
		Started: EventDef{
			Component: ComponentSync,
			Code:      "started",
			Level:     Info,
			Message:   "Sync started",
		},
		Completed: EventDef{
			Component: ComponentSync,
			Code:      "completed",
			Level:     Info,
			Message:   "Sync completed",
		},
		Failed: EventDef{
			Component:   ComponentSync,
			Code:        "failed",
			Level:       Error,
			Message:     "Sync failed",
			Remediation: "Check the error details and run `gitback doctor` to diagnose the environment.",
		},
		Summary: EventDef{
			Component: ComponentSync,
			Code:      "summary",
			Level:     Info,
			Message:   "Sync run summary",
		},
	},

	Snapshot: SnapshotEvents{
		Started: EventDef{
			Component: ComponentSnapshot,
			Code:      "started",
			Level:     Info,
			Message:   "Snapshot creation started",
		},
		Completed: EventDef{
			Component: ComponentSnapshot,
			Code:      "completed",
			Level:     Info,
			Message:   "Snapshot creation completed",
		},
		Failed: EventDef{
			Component:   ComponentSnapshot,
			Code:        "failed",
			Level:       Error,
			Message:     "Snapshot creation failed",
			Remediation: "Ensure `tar` and `zstd` are installed and the snapshot output directory is writable, then retry.",
		},

		VerificationStarted: EventDef{
			Component: ComponentSnapshot,
			Code:      "verification_started",
			Level:     Info,
			Message:   "Pre-snapshot mirror verification started",
		},
		VerificationPassed: EventDef{
			Component: ComponentSnapshot,
			Code:      "verification_passed",
			Level:     Info,
			Message:   "Pre-snapshot mirror verification passed",
		},
		VerificationFailed: EventDef{
			Component:   ComponentSnapshot,
			Code:        "verification_failed",
			Level:       Error,
			Message:     "Pre-snapshot mirror verification failed",
			Remediation: "Run `gitback sync` to fix failing mirrors, or re-run with `--force` to snapshot anyway.",
		},

		ArchiveStarted: EventDef{
			Component: ComponentSnapshot,
			Code:      "archive_started",
			Level:     Info,
			Message:   "Snapshot archive creation started",
		},
		ArchiveCompleted: EventDef{
			Component: ComponentSnapshot,
			Code:      "archive_completed",
			Level:     Info,
			Message:   "Snapshot archive creation completed",
		},

		CompressionStarted: EventDef{
			Component: ComponentSnapshot,
			Code:      "compression_started",
			Level:     Info,
			Message:   "Snapshot compression started",
		},
		CompressionCompleted: EventDef{
			Component: ComponentSnapshot,
			Code:      "compression_completed",
			Level:     Info,
			Message:   "Snapshot compression completed",
		},

		ChecksumStarted: EventDef{
			Component: ComponentSnapshot,
			Code:      "checksum_started",
			Level:     Info,
			Message:   "Snapshot checksum generation started",
		},
		ChecksumCompleted: EventDef{
			Component: ComponentSnapshot,
			Code:      "checksum_completed",
			Level:     Info,
			Message:   "Snapshot checksum generation completed",
		},

		Summary: EventDef{
			Component: ComponentSnapshot,
			Code:      "summary",
			Level:     Info,
			Message:   "Snapshot run summary",
		},

		RetentionDisabled: EventDef{
			Component: ComponentSnapshot,
			Code:      "retention_disabled",
			Level:     Info,
			Message:   "Snapshot retention is disabled",
		},
		RetentionStarted: EventDef{
			Component: ComponentSnapshot,
			Code:      "retention_started",
			Level:     Info,
			Message:   "Snapshot retention cleanup started",
		},
		RetentionCompleted: EventDef{
			Component: ComponentSnapshot,
			Code:      "retention_completed",
			Level:     Info,
			Message:   "Snapshot retention cleanup completed",
		},
		RetentionFailed: EventDef{
			Component:   ComponentSnapshot,
			Code:        "retention_failed",
			Level:       Error,
			Message:     "Failed to delete an old snapshot during retention cleanup",
			Remediation: "Check filesystem permissions on the snapshot output directory.",
		},

		CollisionDetected: EventDef{
			Component:   ComponentSnapshot,
			Code:        "collision_detected",
			Level:       Error,
			Message:     "Snapshot output path already exists",
			Remediation: "Remove or move aside the conflicting file, or retry after the current minute has passed.",
		},
	},

	Lock: LockEvents{
		Acquired: EventDef{
			Component: ComponentLock,
			Code:      "acquired",
			Level:     Info,
			Message:   "Process lock acquired",
		},
		Released: EventDef{
			Component: ComponentLock,
			Code:      "released",
			Level:     Info,
			Message:   "Process lock released",
		},
		Busy: EventDef{
			Component:   ComponentLock,
			Code:        "busy",
			Level:       Warn,
			Message:     "Could not acquire the gitback process lock",
			Remediation: "Another gitback process is already running; wait for it to finish, or check for a stuck process holding the lock file.",
		},
	},

	Health: HealthEvents{
		HealthReport: EventDef{
			Component: ComponentHealth,
			Code:      "report_generated",
			Level:     Info,
			Message:   "Health report generated",
		},
	},

	Doctor: DoctorEvents{
		ReportGenerated: EventDef{
			Component: ComponentDoctor,
			Code:      "report_generated",
			Level:     Info,
			Message:   "Doctor diagnostic report generated",
		},
	},

	LogRetention: LogRetentionEvents{
		Pruned: EventDef{
			Component: ComponentLogRetention,
			Code:      "pruned",
			Level:     Info,
			Message:   "Old log files pruned",
		},
		PruneFailed: EventDef{
			Component:   ComponentLogRetention,
			Code:        "prune_failed",
			Level:       Warn,
			Message:     "Failed to prune one or more old log files",
			Remediation: "Check filesystem permissions on the log directory.",
		},
		UnrecognizedFile: EventDef{
			Component:   ComponentLogRetention,
			Code:        "unrecognized_file",
			Level:       Warn,
			Message:     "Found a file matching gitback's log naming pattern with an unparseable date",
			Remediation: "Manually inspect the file; remove it if it is not a genuine gitback log file.",
		},
	},
}
