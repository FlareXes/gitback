// internal/state/types.go

package state

type Repository struct {
	Name        string `json:"name"`
	LastSuccess bool   `json:"last_success"`
	Error       string `json:"error,omitempty"`
}

type Mirrors struct {
	GeneratedAt  string       `json:"generated_at"`
	Repositories []Repository `json:"repositories"`
}
