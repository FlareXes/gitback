// internal/health/types.go

package health

type HealthReport struct {
	GeneratedAt string `json:"generated_at"`

	Status string `json:"status"`

	Repositories RepositoryHealth `json:"repositories"`
	Snapshots    SnapshotHealth   `json:"snapshots"`
	Disk         DiskHealth       `json:"disk"`
	Sync         SyncHealth       `json:"sync"`
	Retention    RetentionHealth  `json:"retention"`
}

type RepositoryHealth struct {
	Total   int `json:"total"`
	Healthy int `json:"healthy"`
	Failed  int `json:"failed"`
}

type SnapshotHealth struct {
	Count     int    `json:"count"`
	SizeBytes int64  `json:"size_bytes"`
	Latest    string `json:"latest,omitempty"`
}

type DiskHealth struct {
	FreePercent    int `json:"free_percent"`
	MinimumPercent int `json:"minimum_percent"`
}

type SyncHealth struct {
	StartedAt   string `json:"started_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type RetentionHealth struct {
	Enabled bool `json:"enabled"`
	Keep    int  `json:"keep"`
}
