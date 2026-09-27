// internal/mirror/worker_test.go

package mirror

import (
	"context"
	"errors"
	"testing"

	"github.com/flarexes/gitback/internal/logging"
)

// TestWorker_cancelledError_setsCancelledCause: a syncFn returning
// a context-cancellation error must produce an Asset with
// Cause == CauseCancelled, not just LastSuccess=false indistinguishable
// from a real failure.
func TestBuildAsset_cancelledError_setsCancelledCause(t *testing.T) {

	got := buildAsset("owner/repo", context.Canceled)

	if got.LastSuccess {
		t.Error("LastSuccess = true, want false for a cancelled sync")
	}
	if got.Cause != logging.CauseCancelled {
		t.Errorf("Cause = %q, want %q", got.Cause, logging.CauseCancelled)
	}
}

// TestBuildAsset_genuineError_leavesCauseEmpty confirms an ordinary,
// unclassified failure is never mislabeled as cancelled.
func TestBuildAsset_genuineError_leavesCauseEmpty(t *testing.T) {

	got := buildAsset("owner/repo", errors.New("authentication failed"))

	if got.LastSuccess {
		t.Error("LastSuccess = true, want false for a genuine failure")
	}
	if got.Cause != "" {
		t.Errorf("Cause = %q, want empty for an unclassified failure", got.Cause)
	}
}

// TestBuildAsset_success_marksHealthyWithNoCause confirms a
// successful sync carries no Cause or Error at all.
func TestBuildAsset_success_marksHealthyWithNoCause(t *testing.T) {

	got := buildAsset("owner/repo", nil)

	if !got.LastSuccess {
		t.Error("LastSuccess = false, want true for a successful sync")
	}
	if got.Cause != "" {
		t.Errorf("Cause = %q, want empty on success", got.Cause)
	}
	if got.Error != "" {
		t.Errorf("Error = %q, want empty on success", got.Error)
	}
}
