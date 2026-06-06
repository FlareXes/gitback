// internal/health/types.go

package health

type Recommendation struct {
	Severity string
	Message  string
}

type HealthReport struct {
	RepositoryCount int
	HealthyCount    int
	FailedCount     int

	LastSync string

	SnapshotCount int
	LastSnapshot  string

	SnapshotBytes uint64

	DiskFreePercent int

	Recommendations []Recommendation
}
