// internal/health/types.go

package health

type HealthReport struct {
	GeneratedAt string `json:"generated_at"`

	Status string `json:"status"`

	Repositories AssetHealth `json:"repositories"`
	Gists        AssetHealth `json:"gists"`

	Quarantine QuarantineHealth `json:"quarantine"`

	// Orphaned lists mirrors on disk that discovery no longer lists.
	// Informational only: it never affects Status, because nothing is
	// broken. See mirror.FindOrphans.
	Orphaned OrphanHealth `json:"orphaned"`

	Sync      SyncHealth      `json:"sync"`
	Snapshots SnapshotHealth  `json:"snapshots"`
	Disks     []DiskHealth    `json:"disks"`
	Retention RetentionHealth `json:"retention"`

	Warnings        []string `json:"warnings,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

type AssetHealth struct {
	Total   int `json:"total"`
	Healthy int `json:"healthy"`
	Failed  int `json:"failed"`

	// Interrupted counts assets whose last sync attempt was cancelled
	// (state.Asset.Cause == logging.CauseCancelled) rather than
	// genuinely failed. Deliberately excluded from Failed and from
	// updateStatus's warning trigger — nothing is actually wrong; the
	// next sync will simply retry them.
	Interrupted int `json:"interrupted,omitempty"`
}

type QuarantineHealth struct {
	Repositories int `json:"repositories"`
	Gists        int `json:"gists"`
}

// OrphanList is a set of mirror names, split by kind.
type OrphanList struct {
	Repositories []string `json:"repositories,omitempty"`
	Gists        []string `json:"gists,omitempty"`
}

// OrphanHealth separates intact-but-no-longer-updated mirrors from
// quarantined ones, because the appropriate action differs.
type OrphanHealth struct {
	// Mirrors are intact backups that `gitback sync` no longer updates.
	Mirrors OrphanList `json:"mirrors"`

	// Quarantine holds corrupt mirrors whose upstream is also gone, so
	// automatic recovery can never happen.
	Quarantine OrphanList `json:"quarantine"`
}

type SnapshotHealth struct {
	Count  uint64 `json:"count"`
	Size   int64  `json:"size"`
	Latest string `json:"latest,omitempty"`
}

type DiskHealth struct {
	Path        string `json:"path"`
	Free        uint64 `json:"free"`
	Total       uint64 `json:"total"`
	FreePercent uint8  `json:"free_percent"`
	Device      uint64 `json:"-"` // Don't include this field when marshaling to JSON
}

type SyncHealth struct {
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type RetentionHealth struct {
	Enabled bool `json:"enabled"`
	Keep    int  `json:"keep"`
}

// Count returns the total number of names in the list.
func (l OrphanList) Count() int {
	return len(l.Repositories) + len(l.Gists)
}
