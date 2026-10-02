// internal/mirror/display.go
package mirror

import (
	"path/filepath"
	"strings"

	"github.com/flarexes/gitback/internal/config"
)

// mirrorDisplayName derives the name used to identify a mirror in log
// entries and in mirrors.json, from either its clone URL or its
// filesystem path: the final path segment, with a trailing ".git"
// removed if present.
//
// This is the single place this derivation happens. Several call sites
// previously repeated this inline, and worker.go's buildAsset call
// didn't perform it at all — it used the raw inventory URL verbatim as
// state.Asset.Name, so a failed sync's summary could show
// "https://github.com/owner/repo.git" while every log line for the
// same failure showed "repo".
//
// This returns a single path segment, not "owner/name" — see
// repository.go's extractRepoName for the two-segment form used only
// for the dispatch-time [REPO] print line. A single-segment name is
// ambiguous if two different owners have a same-named repository; that
// is a known, accepted limitation here, consistent with the earlier
// decision not to thread repo/gist-kind awareness through these shared
// functions (see Entry.Asset's own doc comment for the same trade-off
// made deliberately elsewhere).
func assetName(pathOrURL string) string {
	return strings.TrimSuffix(filepath.Base(pathOrURL), ".git")
}

// CountGistMirrors returns how many gist mirrors currently exist on
// disk, regardless of the inventory. Used by gitback health to report
// leftover gist mirrors when gist backup has been disabled — unlike
// FindOrphans, this deliberately does not compare against the
// inventory, since a disabled category's inventory isn't being kept
// current at all.
func CountLiveGistMirrors(cfg *config.Config) (int, error) {
	found, err := scanGistMirrors(cfg.Storage.MirrorRoot)
	if err != nil {
		return 0, err
	}
	return len(found), nil
}
