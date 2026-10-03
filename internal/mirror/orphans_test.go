// internal/mirror/orphans_test.go

package mirror

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/flarexes/gitback/internal/state"
)

// orphanFixture returns a Config/Layout over a disposable tree. A nil
// repoInventory means the inventory file is never written.
func orphanFixture(t *testing.T, repoInventory []string) (*config.Config, runtime.Layout) {
	t.Helper()

	layout := runtime.NewWithRoot(t.TempDir())
	cfg := config.Default(layout)

	// WriteInventory doesn't create its parent directory; real callers
	// always go through layout.EnsureDirs() first.
	if err := os.MkdirAll(layout.StateDir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if repoInventory != nil {
		if err := state.WriteInventory(layout.RepositoryInventoryFile, repoInventory); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	return &cfg, layout
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

// A mirror on disk that the inventory doesn't list is exactly what this
// feature exists to surface.
func TestFindOrphans_liveMirrorNotInInventory_isReported(t *testing.T) {

	cfg, layout := orphanFixture(t, []string{"https://github.com/owner/listed.git"})
	mkdirAll(t, filepath.Join(cfg.RepositoryMirrorRoot(), "owner", "listed.git"))
	mkdirAll(t, filepath.Join(cfg.RepositoryMirrorRoot(), "owner", "unlisted.git"))

	got, err := FindOrphans(cfg, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := []string{"owner/unlisted"}; !reflect.DeepEqual(got.Repositories, want) {
		t.Errorf("Repositories = %v, want %v", got.Repositories, want)
	}
}

// Guards against sync and orphan detection drifting apart. A mirror
// placed exactly where sync would put it must never be called an
// orphan, or every healthy mirror would be reported as one.
func TestFindOrphans_mirrorAtSyncPath_isNotOrphan(t *testing.T) {

	url := "https://github.com/owner/listed.git"
	cfg, layout := orphanFixture(t, []string{url})

	mkdirAll(t, New(cfg, layout, nil).repositoryMirrorPath(url))

	got, err := FindOrphans(cfg, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Repositories) != 0 {
		t.Errorf("Repositories = %v, want none", got.Repositories)
	}
}

// A quarantined copy that collided earlier carries a timestamp suffix.
// It must still match its still-listed repo, and an unlisted one must
// still be reported under its canonical name.
func TestFindOrphans_quarantineTimestampSuffix_matchesCanonicalName(t *testing.T) {

	cfg, layout := orphanFixture(t, []string{"https://github.com/owner/still-exists.git"})
	q := filepath.Join(cfg.QuarantineDir(), "repositories", "owner")
	mkdirAll(t, filepath.Join(q, "still-exists.git.2026-09-20T16-36-29Z"))
	mkdirAll(t, filepath.Join(q, "gone.git.2026-09-20T16-36-29Z"))

	got, err := FindOrphans(cfg, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if want := []string{"owner/gone"}; !reflect.DeepEqual(got.QuarantinedRepositories, want) {
		t.Errorf("QuarantinedRepositories = %v, want %v", got.QuarantinedRepositories, want)
	}
}

// Without an inventory (discover hasn't run) nothing can be compared,
// so nothing may be reported. Otherwise every mirror looks orphaned.
func TestFindOrphans_missingInventory_reportsNothing(t *testing.T) {

	cfg, layout := orphanFixture(t, nil)
	mkdirAll(t, filepath.Join(cfg.RepositoryMirrorRoot(), "owner", "anything.git"))

	got, err := FindOrphans(cfg, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Repositories) != 0 {
		t.Errorf("Repositories = %v, want none when inventory is missing", got.Repositories)
	}
}

// With gist backup disabled the gist inventory isn't refreshed, so gists
// must not be reported even if a stale inventory says otherwise.
func TestFindOrphans_backupGistsDisabled_ignoresGists(t *testing.T) {

	cfg, layout := orphanFixture(t, nil)
	cfg.GitHub.BackupGists = false

	if err := state.WriteInventory(layout.GistInventoryFile, []string{"https://gist.github.com/listed.git"}); err != nil {
		t.Fatalf("setup: %v", err)
	}
	mkdirAll(t, filepath.Join(cfg.GistMirrorRoot(), "unlisted.git"))

	got, err := FindOrphans(cfg, layout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Gists) != 0 {
		t.Errorf("Gists = %v, want none with BackupGists disabled", got.Gists)
	}
}
