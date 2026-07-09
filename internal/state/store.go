// internal/state/store.go

package state

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/flarexes/gitback/internal/filesystem"
)

func Save(
	path string,
	syncStartedAt time.Time,
	syncCompletedAt time.Time,
	repositories []Asset,
	gists []Asset,
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
		Gists:        gists,
	}

	return filesystem.AtomicWriteFile(
		path,
		0600,
		func(w io.Writer) error {

			encoder := json.NewEncoder(w)
			encoder.SetIndent("", "  ")

			return encoder.Encode(data)
		},
	)
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
