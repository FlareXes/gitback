// internal/state/store.go

package state

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/flarexes/gitback/internal/filesystem"
)

func SaveMirrors(
	path string,
	syncStartedAt time.Time,
	syncCompletedAt time.Time,
	repositories []Asset,
	gists []Asset,
) error {

	data := MirrorState{
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

func LoadMirrors(path string) (*MirrorState, error) {

	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf(
			"open mirror state %s: %w",
			path,
			err,
		)
	}

	defer file.Close()

	var data MirrorState

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, fmt.Errorf(
			"load mirror state %s: %w",
			path,
			err,
		)
	}

	return &data, nil
}
