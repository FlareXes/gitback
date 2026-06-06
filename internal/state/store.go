// internal/state/store.go

package state

import (
	"encoding/json"
	"os"
	"time"
)

func Save(
	path string,
	syncStartedAt time.Time,
	syncCompletedAt time.Time,
	repositories []Repository,
) error {

	data := Mirrors{
		GeneratedAt: time.Now().
			UTC().
			Format(time.RFC3339),

		SyncStartedAt: syncStartedAt.
			UTC().
			Format(time.RFC3339),

		SyncCompletedAt: syncCompletedAt.
			UTC().
			Format(time.RFC3339),

		Repositories: repositories,
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

func Load(path string) (*Mirrors, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var data Mirrors

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
