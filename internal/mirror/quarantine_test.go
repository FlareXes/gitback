// internal/mirror/quarantine_test.go

package mirror

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/flarexes/gitback/internal/clock"
)

// quarantineNames lists the entries in one owner's quarantine directory.
func quarantineNames(t *testing.T, e *Engine, owner string) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join(e.cfg.QuarantineDir(), "repositories", owner))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read quarantine: %v", err)
	}

	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

// newSyncEngine builds an Engine over a disposable tree. A nil logger is
// fine: Emit is nil-safe.
func newSyncEngine(t *testing.T) *Engine {
	t.Helper()

	cfg, layout := orphanFixture(t, nil)

	// createAskPassScript writes into TempDir, which real runs get from
	// layout.EnsureDirs().
	mkdirAll(t, layout.TempDir)

	return New(cfg, layout, nil)
}

// makeRemote builds a small bare repository to sync from. Cloning a local
// path exercises the real git code paths with no network, so these tests
// need only the git binary (skipped if it is absent).
func makeRemote(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	dir := t.TempDir()
	work := filepath.Join(dir, "work")
	remote := filepath.Join(dir, "remote.git")

	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	git("init", work)

	if err := os.WriteFile(filepath.Join(work, "f.txt"), []byte("x"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	git("-C", work, "add", ".")

	// Identity and signing are set per command so the test neither
	// depends on nor modifies the developer's git config.
	git("-C", work,
		"-c", "user.name=test",
		"-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false",
		"commit", "-m", "init",
	)

	git("clone", "--bare", work, remote)

	return remote
}

// The cleanup must remove every copy of this mirror and nothing else.
func TestCleanupQuarantine_removesOnlyCopiesOfThisMirror(t *testing.T) {

	cfg, layout := orphanFixture(t, nil)
	e := New(cfg, layout, nil)

	q := filepath.Join(cfg.QuarantineDir(), "repositories")

	for _, name := range []string{
		"owner/x.git",                      // plain copy
		"owner/x.git.2026-09-20T16-36-29Z", // copies from repeated quarantines
		"owner/x.git.2026-09-21T01-01-01Z",
		"owner/y.git",     // a different mirror, same owner
		"owner/x.git.bak", // not a quarantine copy: the suffix isn't a timestamp
		"other/x.git",     // same name, different owner
	} {
		mkdirAll(t, filepath.Join(q, name))
	}

	target := filepath.Join(cfg.RepositoryMirrorRoot(), "owner", "x.git")

	if err := e.cleanupQuarantine(target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := quarantineNames(t, e, "owner"), []string{"x.git.bak", "y.git"}; !reflect.DeepEqual(got, want) {
		t.Errorf("owner entries = %v, want %v", got, want)
	}
	if got, want := quarantineNames(t, e, "other"), []string{"x.git"}; !reflect.DeepEqual(got, want) {
		t.Errorf("other entries = %v, want %v", got, want)
	}
}

// quarantineMirror produces the suffix and canonicalMirrorName matches
// it. If the two ever disagree, copies silently stop being found and
// cleaned up, so pin them together.
func TestCanonicalMirrorName_recognizesWhatQuarantineMirrorProduces(t *testing.T) {

	suffixed := "x.git." + clock.FilenameUTCNow()

	if got := canonicalMirrorName(suffixed); got != "x.git" {
		t.Errorf("canonicalMirrorName(%q) = %q, want %q", suffixed, got, "x.git")
	}
}

// Regression: a healthy mirror that merely updates never triggered
// cleanup, so copies left by earlier interrupted runs stayed forever.
func TestSyncMirror_healthyMirror_clearsStaleQuarantine(t *testing.T) {

	remote := makeRemote(t)
	e := newSyncEngine(t)
	ctx := context.Background()

	target := filepath.Join(e.cfg.RepositoryMirrorRoot(), "owner", "x.git")

	// The first sync clones the mirror.
	if err := e.syncMirror(ctx, remote, target); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	// Earlier interrupted runs left quarantined copies of it.
	q := filepath.Join(e.cfg.QuarantineDir(), "repositories", "owner")
	mkdirAll(t, filepath.Join(q, "x.git"))
	mkdirAll(t, filepath.Join(q, "x.git."+clock.FilenameUTCNow()))

	// The second sync takes the update path.
	if err := e.syncMirror(ctx, remote, target); err != nil {
		t.Fatalf("second sync: %v", err)
	}

	if got := quarantineNames(t, e, "owner"); len(got) != 0 {
		t.Errorf("quarantine still holds %v after a successful sync", got)
	}
}

// Regression for the incident: a corrupt mirror is quarantined and
// recovered while an older copy already sits in quarantine. Recovery used
// to remove only the newest copy, leaving the older one behind.
func TestSyncMirror_recovery_leavesNoQuarantineBehind(t *testing.T) {

	remote := makeRemote(t)
	e := newSyncEngine(t)
	ctx := context.Background()

	target := filepath.Join(e.cfg.RepositoryMirrorRoot(), "owner", "x.git")

	if err := e.syncMirror(ctx, remote, target); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	// Delete the mirror's object files, leaving refs pointing at nothing.
	// fsck fails on this for every git version, which is what sends the
	// next sync down the recovery path. Removing the mirror's own links
	// does not affect the source repository.
	err := filepath.WalkDir(filepath.Join(target, "objects"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return os.Remove(path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("corrupt mirror: %v", err)
	}

	// Guard the test's premise: if the mirror isn't actually broken, the
	// update path would also clean up and this test would pass without
	// exercising recovery at all.
	if err := exec.Command("git", "-C", target, "fsck").Run(); err == nil {
		t.Fatal("setup failed to corrupt the mirror")
	}

	// An earlier run already quarantined this mirror, so the new copy will
	// get a timestamp suffix and two copies will exist.
	mkdirAll(t, filepath.Join(e.cfg.QuarantineDir(), "repositories", "owner", "x.git"))

	if err := e.syncMirror(ctx, remote, target); err != nil {
		t.Fatalf("sync should recover the corrupt mirror: %v", err)
	}

	if got := quarantineNames(t, e, "owner"); len(got) != 0 {
		t.Errorf("quarantine still holds %v after recovery", got)
	}
}

// Regression: the old health count only saw names ending in ".git", so a
// mirror with only timestamped copies wasn't counted, and a mirror with
// several copies would have been counted several times.
func TestCountQuarantined_countsEachMirrorOnce(t *testing.T) {

	cfg, _ := orphanFixture(t, nil)
	q := cfg.QuarantineDir()

	ts := clock.FilenameUTCNow()

	mkdirAll(t, filepath.Join(q, "repositories", "owner", "a.git"))
	mkdirAll(t, filepath.Join(q, "repositories", "owner", "a.git."+ts)) // second copy of a
	mkdirAll(t, filepath.Join(q, "repositories", "owner", "b.git."+ts)) // only a timestamped copy
	mkdirAll(t, filepath.Join(q, "gists", "g1.git."+ts))

	repositories, gists, err := CountQuarantined(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repositories != 2 {
		t.Errorf("repositories = %d, want 2", repositories)
	}
	if gists != 1 {
		t.Errorf("gists = %d, want 1", gists)
	}
}
