// internal/state/types.go

package state

type Repository struct {
	Name        string `json:"name"`
	LastSuccess bool   `json:"last_success"`
	Error       string `json:"error,omitempty"`
}

type Mirrors struct {
	GeneratedAt     string       `json:"generated_at"`
	SyncStartedAt   string       `json:"sync_started_at,omitempty"`
	SyncCompletedAt string       `json:"sync_completed_at,omitempty"`
	Repositories    []Repository `json:"repositories"`
}
