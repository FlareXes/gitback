// internal/doctor/doctor_test.go

package doctor

import (
	"testing"

	"github.com/flarexes/gitback/internal/runtime"
)

// TestCheckTokenAvailable_envSet_succeedsWithoutFile guards against a
// false-positive doctor failure for an env-var-only setup — gitback
// init --use-env-token never creates a token file at all, and that's
// correct, not broken.
func TestCheckTokenAvailable_envSet_succeedsWithoutFile(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())
	t.Setenv("GITBACK_TOKEN", "env-token")

	check := checkTokenAvailable(layout.TokenFile)

	if !check.Success {
		t.Errorf("expected success when GITBACK_TOKEN is set, got failure: %s", check.Message)
	}
}

// TestCheckTokenAvailable_neitherPresent_fails confirms doctor still
// correctly flags a genuinely unconfigured installation.
func TestCheckTokenAvailable_neitherPresent_fails(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())
	t.Setenv("GITBACK_TOKEN", "")

	check := checkTokenAvailable(layout.TokenFile)

	if check.Success {
		t.Error("expected failure when neither GITBACK_TOKEN nor a token file is present")
	}
}
