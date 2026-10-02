// internal/mirror/orphans.go
//
// Orphan detection: finds mirrors on disk that the current inventory (`gitback discover`)
// no longer lists. GitBack never deletes these. They are only reported,
// via `gitback health`, and the user decides what to do.

package mirror

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/flarexes/gitback/internal/state"
)

// Orphans lists mirrors present on disk but absent from the current
// inventory. Names are "owner/name" for repositories and the bare ID
// for gists, sorted and de-duplicated.
type Orphans struct {
	// Live mirrors: intact backups that `gitback sync` no longer updates.
	Repositories []string
	Gists        []string

	// Quarantined mirrors: corrupt, and their upstream is gone too, so
	// automatic recovery (a fresh clone) can never happen.
	QuarantinedRepositories []string
	QuarantinedGists        []string
}

// quarantineTimestampSuffix matches the suffix quarantineMirror adds
// when a path is already quarantined, e.g. "name.git.20260920T163629Z".
var quarantineTimestampSuffix = regexp.MustCompile(`\.\d{8}T\d{6}Z$`)

// canonicalMirrorName strips a quarantine timestamp suffix so a mirror
// quarantined more than once maps to one canonical name. It is a no-op
// for live mirrors, which never carry the suffix.
func canonicalMirrorName(name string) string {
	return quarantineTimestampSuffix.ReplaceAllString(name, "")
}

// repositoryRelPath returns where a repository's mirror lives relative
// to the mirror root, e.g. "repositories/owner/name.git". Both sync
// (repositoryMirrorPath) and orphan detection derive paths from this
// one function, so they can never disagree about a mirror's location.
func repositoryRelPath(repoURL string) string {

	repo := strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(repo, "/")

	// Unexpected URL shape: keep the long-standing fallback layout.
	if len(parts) < 2 {
		return filepath.Join("repositories", filepath.Base(repoURL))
	}

	owner := parts[len(parts)-2]
	name := parts[len(parts)-1]

	return filepath.Join("repositories", owner, name+".git")
}

// gistRelPath returns a gist mirror's path relative to the mirror
// root, e.g. "gists/<id>.git".
func gistRelPath(gistURL string) string {

	id := assetName(gistURL)

	return filepath.Join("gists", id+".git")
}

// foundMirror is one mirror directory discovered on disk. rel is its
// path relative to the scanned base (comparable with the inventory's
// relative paths); name is what to show a human.
type foundMirror struct {
	rel  string
	name string
}

// scanRepositoryMirrors lists every "<owner>/<name>.git" directory
// under base/repositories. A missing directory means none exist.
func scanRepositoryMirrors(base string) ([]foundMirror, error) {

	root := filepath.Join(base, "repositories")

	owners, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var found []foundMirror

	for _, owner := range owners {

		if !owner.IsDir() {
			continue
		}

		entries, err := os.ReadDir(filepath.Join(root, owner.Name()))
		if err != nil {
			return nil, err
		}

		for _, entry := range entries {

			name := canonicalMirrorName(entry.Name())

			// Only directories named *.git are mirrors; this also skips
			// leftovers like "<name>.git.tmp" from an interrupted recovery.
			if !entry.IsDir() || !strings.HasSuffix(name, ".git") {
				continue
			}

			found = append(found, foundMirror{
				rel:  filepath.Join("repositories", owner.Name(), name),
				name: owner.Name() + "/" + strings.TrimSuffix(name, ".git"),
			})
		}
	}

	return found, nil
}

// scanGistMirrors lists every "<id>.git" directory under base/gists.
func scanGistMirrors(base string) ([]foundMirror, error) {

	root := filepath.Join(base, "gists")

	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var found []foundMirror

	for _, entry := range entries {

		name := canonicalMirrorName(entry.Name())

		if !entry.IsDir() || !strings.HasSuffix(name, ".git") {
			continue
		}

		found = append(found, foundMirror{
			rel:  filepath.Join("gists", name),
			name: strings.TrimSuffix(name, ".git"),
		})
	}

	return found, nil
}

// orphansIn returns the sorted, de-duplicated names of mirrors under
// base that are not in keep.
func orphansIn(
	scan func(base string) ([]foundMirror, error),
	base string,
	keep map[string]bool,
) ([]string, error) {

	found, err := scan(base)
	if err != nil {
		return nil, err
	}

	// A set, because a mirror quarantined twice appears twice on disk
	// but is one logical orphan.
	seen := make(map[string]struct{})

	for _, m := range found {
		if !keep[m.rel] {
			seen[m.name] = struct{}{}
		}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)

	return names, nil
}

// findCategory finds orphans for one category (repositories or gists),
// both live and quarantined, against that category's inventory file.
//
// A missing inventory returns no orphans and no error: discovery hasn't
// run, so there is nothing to compare against. An inventory that exists
// but is empty is compared as-is, so every mirror is reported. That is
// accurate, and usually a useful signal (e.g. the token lost access).
func findCategory(
	inventoryFile string,
	relPath func(url string) string,
	scan func(base string) ([]foundMirror, error),
	mirrorRoot string,
	quarantineRoot string,
) (live []string, quarantined []string, err error) {

	urls, err := state.ReadInventory(inventoryFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	// Everything the inventory still lists, as mirror-root-relative paths.
	keep := make(map[string]bool, len(urls))
	for _, url := range urls {
		keep[relPath(url)] = true
	}

	live, err = orphansIn(scan, mirrorRoot, keep)
	if err != nil {
		return nil, nil, err
	}

	quarantined, err = orphansIn(scan, quarantineRoot, keep)
	if err != nil {
		return nil, nil, err
	}

	return live, quarantined, nil
}

// FindOrphans reports mirrors on disk that the current inventory no
// longer lists. It never modifies anything.
//
// If a category can't be checked (e.g. its inventory is unreadable),
// the other category's results are still returned alongside a non-nil
// error, so one bad file doesn't hide everything else.
func FindOrphans(cfg *config.Config, layout runtime.Layout) (Orphans, error) {

	var result Orphans
	var errs []error
	var err error

	result.Repositories, result.QuarantinedRepositories, err = findCategory(
		layout.RepositoryInventoryFile,
		repositoryRelPath,
		scanRepositoryMirrors,
		cfg.Storage.MirrorRoot,
		cfg.QuarantineDir(),
	)
	if err != nil {
		errs = append(errs, fmt.Errorf("repositories: %w", err))
	}

	// With gist backup disabled the gist inventory isn't refreshed, so
	// there is no reliable current truth to compare gists against.
	if cfg.GitHub.BackupGists {

		result.Gists, result.QuarantinedGists, err = findCategory(
			layout.GistInventoryFile,
			gistRelPath,
			scanGistMirrors,
			cfg.Storage.MirrorRoot,
			cfg.QuarantineDir(),
		)
		if err != nil {
			errs = append(errs, fmt.Errorf("gists: %w", err))
		}
	}

	return result, errors.Join(errs...)
}
