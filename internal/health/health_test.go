// internal/health/health_test.go

package health

import (
	"os"
	"testing"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/flarexes/gitback/internal/state"
)

// TestPopulateAssets_cancelledCause_notCountedAsFailed is a regression
// test for a real incident: Ctrl+C during sync left several repos with
// LastSuccess=false, and health.Generate reported an alarming
// "warning" status even though nothing was actually broken — those
// repos just needed a normal re-sync.
func TestPopulateAssets_cancelledCause_notCountedAsFailed(t *testing.T) {

	layout := runtime.NewWithRoot(t.TempDir())
	cfg := config.Default(layout)

	if err := os.MkdirAll(layout.StateDir, 0700); err != nil {
		t.Fatalf("setup: %v", err)
	}

	now := time.Now()

	if err := state.SaveMirrors(
		layout.MirrorsStateFile,
		now, now,
		[]state.Asset{
			{Name: "owner/genuinely-broken", LastSuccess: false, Error: "authentication failed"},
			{Name: "owner/interrupted-one", LastSuccess: false, Cause: logging.CauseCancelled, Error: "context canceled"},
			{Name: "owner/interrupted-two", LastSuccess: false, Cause: logging.CauseCancelled, Error: "context canceled"},
			{Name: "owner/healthy", LastSuccess: true},
		},
		nil,
	); err != nil {
		t.Fatalf("setup: %v", err)
	}

	report := &HealthReport{}
	populateAssets(&cfg, layout, report)

	if report.Repositories.Failed != 1 {
		t.Errorf("Failed = %d, want 1 (only the genuine failure)", report.Repositories.Failed)
	}
	if report.Repositories.Interrupted != 2 {
		t.Errorf("Interrupted = %d, want 2", report.Repositories.Interrupted)
	}
	if report.Repositories.Healthy != 1 {
		t.Errorf("Healthy = %d, want 1", report.Repositories.Healthy)
	}
}

// TestUpdateStatus_onlyInterrupted_staysHealthy confirms interrupted-only
// results never trigger the "warning" status — only genuine failures
// should.
func TestUpdateStatus_onlyInterrupted_staysHealthy(t *testing.T) {

	cfg := config.Default(runtime.NewWithRoot(t.TempDir()))

	report := &HealthReport{
		Repositories: AssetHealth{Total: 2, Interrupted: 2},
	}

	updateStatus(&cfg, report)

	if report.Status != "healthy" {
		t.Errorf("Status = %q, want %q", report.Status, "healthy")
	}
}
