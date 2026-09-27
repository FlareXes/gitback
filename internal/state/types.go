// internal/state/types.go

package state

import "github.com/flarexes/gitback/internal/logging"

// Asset records the outcome of syncing one repository or gist mirror.
//
// Cause and Error are only meaningful when LastSuccess is false; both
// are empty for a successful sync. Cause == logging.CauseCancelled
// specifically means the sync was deliberately interrupted (Ctrl+C,
// systemd stop) — the mirror isn't broken, the sync just didn't finish;
// see health.populateAssets, which relies on exactly this distinction
// to avoid reporting a false "warning" status after an ordinary
// interruption.
type Asset struct {
	Name        string `json:"name"`
	LastSuccess bool   `json:"last_success"`

	Cause logging.Cause `json:"cause,omitempty"`
	Error string        `json:"error,omitempty"`
}

type MirrorState struct {
	GeneratedAt     string `json:"generated_at"`
	SyncStartedAt   string `json:"sync_started_at,omitempty"`
	SyncCompletedAt string `json:"sync_completed_at,omitempty"`

	Repositories []Asset `json:"repositories"`
	Gists        []Asset `json:"gists"`
}
