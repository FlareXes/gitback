// internal/logging/types.go

package logging

type Level string

const (
	Info  Level = "INFO"
	Warn  Level = "WARN"
	Error Level = "ERROR"
	Critical Level = "CRITICAL"
)

type Entry struct {
	Timestamp string `json:"ts"`

	Level Level `json:"level"`

	Event string `json:"event"`

	RunID string `json:"run_id,omitempty"`

	Repo string `json:"repo,omitempty"`

	DurationMS int64 `json:"duration_ms,omitempty"`

	Error string `json:"error,omitempty"`

	Details any `json:"details,omitempty"`
}

type MirrorState struct {
	Name        string `json:"name"`
	LastSuccess bool   `json:"last_success"`
	Error       string `json:"error,omitempty"`
}

type MirrorsState struct {
	GeneratedAt  string        `json:"generated_at"`
	Repositories []MirrorState `json:"repositories"`
}
