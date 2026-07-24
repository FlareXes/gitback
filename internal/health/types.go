// internal/health/types.go

package health

type HealthReport struct {
	GeneratedAt string `json:"generated_at"`

	Status string `json:"status"`

	Repositories AssetHealth `json:"repositories"`
	Gists        AssetHealth `json:"gists"`

	Quarantine QuarantineHealth `json:"quarantine"`

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
}

type QuarantineHealth struct {
	Repositories int `json:"repositories"`
	Gists        int `json:"gists"`
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
