// internal/logging/types.go

package logging

// schemaVersion is bumped whenever Entry's field set or meaning
// changes in a way a log consumer (SIEM, dashboard, AI pipeline) needs
// to know about. Every entry carries it, so historical logs stay
// interpretable even after the schema evolves.
const schemaVersion = 1

type Level string

const (
	Info     Level = "INFO"
	Warn     Level = "WARN"
	Error    Level = "ERROR"
	Critical Level = "CRITICAL"
)

// Component names the internal subsystem an event belongs to.
// Constants, not free strings — a typo in catalog.go becomes a compile
// error instead of a silently-wrong value shipped to every consumer.
const (
	ComponentGitHub       = "github"
	ComponentInventory    = "inventory"
	ComponentMirror       = "mirror"
	ComponentSync         = "sync"
	ComponentSnapshot     = "snapshot"
	ComponentLock         = "lock"
	ComponentHealth       = "health"
	ComponentFilesystem   = "filesystem"
	ComponentDoctor       = "doctor"
	ComponentLogRetention = "log_retention"
)

// Cause is a small, fixed taxonomy of failure categories. Unlike Error
// (the raw, free-text message from whatever failed), Cause is
// machine-parseable and stable across versions and locales — this is
// what a SIEM rule or an AI triage pipeline should key off, not
// string-matching Error text.
//
// Only set Cause when a call site genuinely knows the category for
// certain. Leaving it empty is correct when the failure hasn't been
// classified — never guess.
type Cause string

const (
	CauseAuthFailure      Cause = "auth_failure"
	CauseNetworkError     Cause = "network_error"
	CausePermissionDenied Cause = "permission_denied"
	CauseDiskFull         Cause = "disk_full"
	CauseCorruption       Cause = "corruption"
	CauseNotFound         Cause = "not_found"
	CauseAlreadyExists    Cause = "already_exists"
	CauseTimeout          Cause = "timeout"
	CauseCancelled        Cause = "cancelled"
	CauseConfigInvalid    Cause = "config_invalid"
	CauseLockHeld         Cause = "lock_held"
	CauseChecksumMismatch Cause = "checksum_mismatch"
)

// EventDef is the single declaration point for one kind of log entry.
// See catalog.go's package doc for the rules governing how these are
// added.
type EventDef struct {
	// Component groups this event for filtering (e.g. a SIEM query
	// for "mirror.*"). Always one of the Component* constants above.
	Component string

	// Code identifies this event within Component, e.g. "clone_failed".
	// Combined as "<component>.<code>" for the entry's Event field.
	Code string

	// Level is this event's severity, fixed here rather than chosen
	// per call site — so the same conceptual event is never logged at
	// inconsistent severities in different places.
	Level Level

	// Message is a fixed, human-readable summary of what happened —
	// the "what". Never overridden per call site; variable context
	// (which repo, which path) belongs in Repo/Details, not Message.
	Message string

	// Remediation is the default "how to fix it" guidance, shown
	// directly in the log entry so a reader never has to cross-
	// reference a separate runbook. Required for Warn/Error/Critical
	// events with a knowable fix; empty is correct for Info events.
	// A call site may override this with WithRemediation when it has
	// more specific guidance (e.g. a concrete file path).
	Remediation string
}

// Entry is one structured log line. Everything through Message is
// always present; everything after is populated only when relevant.
// Field order here is also the emitted JSON field order — kept stable
// and deliberate so every gitback log line reads the same way, which
// matters for both human scanning and SIEM/AI field-position
// assumptions.
type Entry struct {
	// --- Envelope: present on every entry, always ---

	SchemaVersion int    `json:"schema_version"`
	Timestamp     string `json:"ts"`
	Level         Level  `json:"level"`
	RunID         string `json:"run_id"`
	Host          string `json:"host"`
	User          string `json:"user"`
	Component     string `json:"component"`
	Event         string `json:"event"`
	Message       string `json:"message"`

	// --- Context: populated when relevant ---

	// Asset names the specific repository or gist this entry concerns.
	// Empty for entries that aren't about one specific mirror (e.g. a
	// run-level summary).
	Asset string `json:"repo,omitempty"`

	DurationMS int64 `json:"duration_ms,omitempty"`

	// --- Diagnosis: the "why" and "how to fix" ---

	Cause Cause `json:"cause,omitempty"`

	// Error is the raw, free-text message from whatever underlying
	// operation failed — supplementary detail for deeper analysis.
	// Cause is what SIEM rules should match against, not this.
	Error string `json:"error,omitempty"`

	Remediation string `json:"remediation,omitempty"`

	// --- Escape hatch ---

	// Details carries structured, event-specific data that doesn't fit
	// the fields above — counts, name lists, rate-limit numbers. Not a
	// place for narrative text: a sentence that belongs here probably
	// belongs in this event's Message or Remediation instead.
	Details any `json:"details,omitempty"`
}
