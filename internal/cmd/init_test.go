// internal/cmd/init_test.go

package cmd

import (
	"bufio"
	"strings"
	"testing"
)

// TestResolveInitToken_useEnvToken_readsFromEnvironment verifies that
// with --use-env-token, the token comes directly from GITBACK_TOKEN
// and stdin is never touched.
func TestResolveInitToken_useEnvToken_readsFromEnvironment(t *testing.T) {

	t.Setenv("GITBACK_TOKEN", "ghp_example123")

	// stdin is nil on purpose: the env-token path must never read from
	// it — a nil reader makes any accidental read attempt fail loudly
	// (a panic) rather than silently succeeding against test input
	// that happens to be present.
	token, err := resolveInitToken(true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ghp_example123" {
		t.Errorf("token = %q, want %q", token, "ghp_example123")
	}
}

// TestResolveInitToken_useEnvToken_missingVar_failsFast is the exact
// Dockerfile/CI failure mode --use-env-token exists to make loud and
// immediate, rather than silently falling back to an interactive
// prompt that would hang forever with no terminal attached.
func TestResolveInitToken_useEnvToken_missingVar_failsFast(t *testing.T) {

	t.Setenv("GITBACK_TOKEN", "")

	if _, err := resolveInitToken(true, nil); err == nil {
		t.Fatal("expected an error when GITBACK_TOKEN is unset, got nil")
	}
}

// TestResolveInitToken_interactive_readsFromStdin verifies the
// existing prompt-based flow still works unchanged.
func TestResolveInitToken_interactive_readsFromStdin(t *testing.T) {

	reader := bufio.NewReader(strings.NewReader("ghp_from_stdin\n"))

	token, err := resolveInitToken(false, reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "ghp_from_stdin" {
		t.Errorf("token = %q, want %q", token, "ghp_from_stdin")
	}
}

// TestResolveInitToken_interactive_emptyInput_returnsError guards
// against silently accepting a blank token when a user just presses
// enter.
func TestResolveInitToken_interactive_emptyInput_returnsError(t *testing.T) {

	reader := bufio.NewReader(strings.NewReader("\n"))

	if _, err := resolveInitToken(false, reader); err == nil {
		t.Fatal("expected an error for empty token input, got nil")
	}
}
