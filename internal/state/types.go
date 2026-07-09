// internal/state/types.go

package state

type Asset struct {
	Name        string `json:"name"`
	LastSuccess bool   `json:"last_success"`
	Error       string `json:"error,omitempty"`
}

type MirrorState struct {
	GeneratedAt     string `json:"generated_at"`
	SyncStartedAt   string `json:"sync_started_at,omitempty"`
	SyncCompletedAt string `json:"sync_completed_at,omitempty"`

	Repositories []Asset `json:"repositories"`
	Gists        []Asset `json:"gists"`
}
