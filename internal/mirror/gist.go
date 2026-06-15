// internal/mirror/gist.go

package mirror

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/state"
)

func (e *Engine) extractGistName(gistURL string) string {

	return strings.TrimSuffix(
		filepath.Base(gistURL),
		".git",
	)
}

func (e *Engine) gistMirrorPath(gistURL string) string {

	id := strings.TrimSuffix(
		filepath.Base(gistURL),
		".git",
	)

	return filepath.Join(
		e.cfg.GistMirrorDir(),
		id+".git",
	)
}

func (e *Engine) syncGist(ctx context.Context, gistURL string) error {

	return e.syncMirror(
		ctx,
		gistURL,
		e.gistMirrorPath(gistURL),
	)
}

func (e *Engine) syncGists(ctx context.Context) ([]state.Asset, error) {

	gistURLs, err := state.ReadInventory(
		e.cfg.GistInventoryFile(),
	)

	if err != nil {

		return nil, fmt.Errorf(
			"read gist inventory: %w",
			err,
		)
	}

	var gists []state.Asset

	for _, gistURL := range gistURLs {

		fmt.Printf("[GIST] %s\n", e.extractGistName(gistURL))

		if err := e.syncGist(ctx, gistURL); err != nil {

			gists = append(
				gists,
				state.Asset{
					Name:        gistURL,
					LastSuccess: false,
					Error:       err.Error(),
				},
			)

			continue
		}

		gists = append(
			gists,
			state.Asset{
				Name:        gistURL,
				LastSuccess: true,
			},
		)
	}

	return gists, nil
}
