// internal/mirror/mirror_test.go

package mirror

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestCloneMirror_interrupted_leavesTargetMissing is a regression test
// for the root cause behind repeated quarantine cycles during testing:
// an interrupted clone previously left a half-written, genuinely
// corrupt directory directly at target, which the next sync's fsck
// correctly (but expensively) detected as corruption. Staging the
// clone means an interrupted run leaves target simply absent instead.
func TestCloneMirror_interrupted_leavesTargetMissing(t *testing.T) {

	remote := makeRemote(t)
	e := newSyncEngine(t)

	target := filepath.Join(e.cfg.RepositoryMirrorRoot(), "owner", "x.git")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled before cloneMirror even starts

	err := e.cloneMirror(ctx, remote, target)
	if err == nil {
		t.Fatal("expected an error for an already-cancelled clone")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}

	if _, statErr := os.Stat(target); !os.IsNotExist(statErr) {
		t.Errorf("target %s should not exist after an interrupted clone, stat err = %v", target, statErr)
	}

	if _, statErr := os.Stat(target + ".tmp"); !os.IsNotExist(statErr) {
		t.Errorf("staging directory %s should be cleaned up, stat err = %v", target+".tmp", statErr)
	}
}

// TestCloneMirror_success_activatesAtTarget confirms the normal case
// still works: a successful clone ends up at target, not stuck in
// staging.
func TestCloneMirror_success_activatesAtTarget(t *testing.T) {

	remote := makeRemote(t)
	e := newSyncEngine(t)

	target := filepath.Join(e.cfg.RepositoryMirrorRoot(), "owner", "x.git")

	if err := e.cloneMirror(context.Background(), remote, target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "HEAD")); err != nil {
		t.Errorf("expected a valid mirror at %s, stat err = %v", target, err)
	}

	if _, err := os.Stat(target + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("staging directory should not remain after success, stat err = %v", err)
	}
}

// TestCloneMirror_staleStagingDirectory_isRemovedBeforeCloning guards
// the cleanup step explicitly: a .tmp directory left by some earlier
// interrupted attempt must not interfere with the next one.
func TestCloneMirror_staleStagingDirectory_isRemovedBeforeCloning(t *testing.T) {

	remote := makeRemote(t)
	e := newSyncEngine(t)

	target := filepath.Join(e.cfg.RepositoryMirrorRoot(), "owner", "x.git")
	staging := target + ".tmp"

	// Simulate leftover partial state from a previous interrupted clone.
	if err := os.MkdirAll(filepath.Join(staging, "garbage"), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := e.cloneMirror(context.Background(), remote, target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, "HEAD")); err != nil {
		t.Errorf("expected a valid mirror at %s despite stale staging dir, stat err = %v", target, err)
	}
}
