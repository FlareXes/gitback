// internal/config/config_test.go

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/runtime"
)

// discardLogger returns a *logging.Logger writing into a throwaway
// temp file, for tests that call logging-aware code but don't assert
// on log output.
func discardLogger(t *testing.T) *logging.Logger {
	t.Helper()

	layout := runtime.NewWithRoot(t.TempDir())

	logger, err := logging.New(layout.LogDir, 0, 0)
	if err != nil {
		t.Fatalf("create discard logger: %v", err)
	}
	t.Cleanup(func() { logger.Close() })

	return logger
}

// TestReadToken_envTakesPrecedenceOverFile confirms GITBACK_TOKEN wins
// even when a token file also exists. This precedence is relied on
// elsewhere — `gitback init --use-env-token` deliberately skips
// writing a file specifically because of it — so a regression here
// would silently break that guarantee.
func TestReadToken_envTakesPrecedenceOverFile(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())

	if err := os.MkdirAll(filepath.Dir(layout.TokenFile), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(layout.TokenFile, []byte("file-token\n"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Setenv("GITBACK_TOKEN", "env-token")

	token, err := ReadToken(layout, discardLogger(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "env-token" {
		t.Errorf("token = %q, want %q (env should win)", token, "env-token")
	}
}

// TestReadToken_fallsBackToFile_whenEnvUnset verifies the token file is
// used when GITBACK_TOKEN isn't set at all.
func TestReadToken_fallsBackToFile_whenEnvUnset(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())

	if err := os.MkdirAll(filepath.Dir(layout.TokenFile), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(layout.TokenFile, []byte("file-token\n"), 0600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Setenv("GITBACK_TOKEN", "")

	token, err := ReadToken(layout, discardLogger(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "file-token" {
		t.Errorf("token = %q, want %q", token, "file-token")
	}
}

// TestReadToken_neitherSourceAvailable_returnsActionableError checks
// the error message itself — this is exactly what a user sees when
// gitback fails to authenticate, so it needs to tell them what to do.
func TestReadToken_neitherSourceAvailable_returnsActionableError(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())
	t.Setenv("GITBACK_TOKEN", "")

	_, err := ReadToken(layout, discardLogger(t))
	if err == nil {
		t.Fatal("expected an error when no token is configured, got nil")
	}
	if !strings.Contains(err.Error(), "gitback init") {
		t.Errorf("error message should mention `gitback init`, got: %v", err)
	}
}

// TestReadToken_fileUnreadable_preservesUnderlyingError guards against
// the fix above regressing: a genuine read failure (not just "file
// doesn't exist") must keep its real cause visible, not be flattened
// into the generic "not configured" message.
func TestReadToken_fileUnreadable_preservesUnderlyingError(t *testing.T) {

	if os.Geteuid() == 0 {
		t.Skip("running as root ignores file permissions")
	}

	layout := runtime.NewWithRoot(t.TempDir())

	if err := os.MkdirAll(filepath.Dir(layout.TokenFile), 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(layout.TokenFile, []byte("token\n"), 0000); err != nil {
		t.Fatalf("setup: %v", err)
	}

	t.Setenv("GITBACK_TOKEN", "")

	_, err := ReadToken(layout, discardLogger(t))
	if err == nil {
		t.Fatal("expected an error for an unreadable token file, got nil")
	}
	if strings.Contains(err.Error(), "gitback init") {
		t.Errorf("expected the real read error, got the generic not-configured message: %v", err)
	}
}

// TestReadToken_nilLogger_doesNotPanic ensures ReadToken stays usable
// from contexts without a guaranteed logger — e.g. doctor's checks,
// which may run before its own logger is known to exist.
func TestReadToken_nilLogger_doesNotPanic(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())
	t.Setenv("GITBACK_TOKEN", "env-token")

	token, err := ReadToken(layout, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "env-token" {
		t.Errorf("token = %q, want %q", token, "env-token")
	}
}
