package doctor

// Check represents a single diagnostic performed by Doctor.
type Check struct {
	Name           string `json:"name"`
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
	Recommendation string `json:"recommendation,omitempty"`
}

// Report contains the complete result of a Doctor run.
type Report struct {
	Checks []Check `json:"checks"`
}

// AddCheck appends a single diagnostic result to the report.
func (r *Report) AddCheck(check Check) {
	r.Checks = append(r.Checks, check)
}

// AddChecks appends multiple diagnostic results to the report.
func (r *Report) AddChecks(checks []Check) {
	r.Checks = append(r.Checks, checks...)
}
