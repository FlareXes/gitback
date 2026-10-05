// internal/ghauth/ghauth_test.go

package ghauth

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v88/github"
)

func TestParseExpiration_validHeader_parsesCorrectly(t *testing.T) {

	resp := &github.Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set(ExpirationHeader, "2026-10-15 00:00:00 UTC")

	got, ok := ParseExpiration(resp)
	if !ok {
		t.Fatal("expected ok=true for a valid header")
	}

	want := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// Classic PATs, and fine-grained PATs with "no expiration" enabled,
// simply don't carry this header — that must not be treated as a
// failure.
func TestParseExpiration_missingHeader_returnsNotOK(t *testing.T) {

	resp := &github.Response{Response: &http.Response{Header: http.Header{}}}

	if _, ok := ParseExpiration(resp); ok {
		t.Error("expected ok=false when the header is absent")
	}
}

func TestParseExpiration_nilResponse_returnsNotOK(t *testing.T) {
	if _, ok := ParseExpiration(nil); ok {
		t.Error("expected ok=false for a nil response")
	}
}

func TestCheckExpiration_alreadyExpired(t *testing.T) {

	resp := responseWithExpiry(t, time.Now().Add(-24*time.Hour))

	status, _ := CheckExpiration(resp)
	if status != Expired {
		t.Errorf("status = %v, want Expired", status)
	}
}

func TestCheckExpiration_withinWarnWindow(t *testing.T) {

	resp := responseWithExpiry(t, time.Now().Add(3*24*time.Hour))

	status, _ := CheckExpiration(resp)
	if status != ExpiringSoon {
		t.Errorf("status = %v, want ExpiringSoon", status)
	}
}

func TestCheckExpiration_farInFuture(t *testing.T) {

	resp := responseWithExpiry(t, time.Now().Add(90*24*time.Hour))

	status, _ := CheckExpiration(resp)
	if status != ExpiryOK {
		t.Errorf("status = %v, want ExpiryOK", status)
	}
}

func responseWithExpiry(t *testing.T, at time.Time) *github.Response {
	t.Helper()
	resp := &github.Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set(ExpirationHeader, at.UTC().Format(expirationLayout))
	return resp
}

func TestIsUnauthorized_401_returnsTrue(t *testing.T) {

	err := &github.ErrorResponse{Response: &http.Response{StatusCode: 401}}

	if !IsUnauthorized(err) {
		t.Error("expected true for a 401 error")
	}
}

func TestIsUnauthorized_otherStatus_returnsFalse(t *testing.T) {

	err := &github.ErrorResponse{Response: &http.Response{StatusCode: 403}}

	if IsUnauthorized(err) {
		t.Error("expected false for a non-401 error")
	}
}

func TestIsUnauthorized_nonGitHubError_returnsFalse(t *testing.T) {
	if IsUnauthorized(fmt.Errorf("some other error")) {
		t.Error("expected false for an unrelated error type")
	}
}

func TestDiagnose_expiredToken_mentionsExpiry(t *testing.T) {

	err := &github.ErrorResponse{Response: &http.Response{StatusCode: 401}}
	resp := responseWithExpiry(t, time.Now().Add(-24*time.Hour))

	msg, ok := Diagnose(err, resp)
	if !ok {
		t.Fatal("expected ok=true for a 401 with expiry info")
	}
	if msg == "" {
		t.Error("expected a non-empty diagnosis")
	}
}

func TestDiagnose_nonAuthError_returnsNotOK(t *testing.T) {
	if _, ok := Diagnose(fmt.Errorf("network timeout"), nil); ok {
		t.Error("expected ok=false for a non-auth error")
	}
}
