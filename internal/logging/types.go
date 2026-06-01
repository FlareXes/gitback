// internal/logging/types.go

package logging

type Level string

const (
	Info  Level = "INFO"
	Warn  Level = "WARN"
	Error Level = "ERROR"
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
