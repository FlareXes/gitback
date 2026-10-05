// internal/ghauth/ghauth.go
//
// Package ghauth centralizes everything gitback knows about GitHub
// token health: reading the expiration GitHub attaches to
// authenticated API responses, and turning an authentication failure
// into a message that explains what's actually wrong.
package ghauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/flarexes/gitback/internal/clock"
	"github.com/google/go-github/v88/github"
)

// ExpirationHeader is the header GitHub sets on an authenticated API
// response when the token in use has an expiration configured. Only
// fine-grained personal access tokens carry this, and only when an
// expiration was actually set — a classic PAT, or a fine-grained token
// with "no expiration" enabled, never has it.
const ExpirationHeader = "github-authentication-token-expiration"

// WarnWindow is how far in advance to start warning about an upcoming
// expiration.
const WarnWindow = 14 * 24 * time.Hour // 14 days

// expirationLayout is the exact format GitHub uses for the header
// value, e.g. "2026-10-15 00:00:00 UTC".
const expirationLayout = "2006-01-02 15:04:05 MST"

// ParseExpiration reads and parses ExpirationHeader off resp. ok is
// false whenever there's nothing usable: a nil response (e.g. a
// network-level failure that never reached GitHub), a missing header,
// or a value in an unrecognized shape. Callers must treat ok=false as
// "no information available," never as "the token is fine" — most
// tokens simply never carry this header.
func ParseExpiration(resp *github.Response) (time.Time, bool) {

	if resp == nil || resp.Response == nil {
		return time.Time{}, false
	}

	raw := resp.Header.Get(ExpirationHeader)
	if raw == "" {
		return time.Time{}, false
	}

	expiresAt, err := time.Parse(expirationLayout, raw)
	if err != nil {
		return time.Time{}, false
	}

	return expiresAt, true
}

// ExpiryStatus classifies how urgent a token's expiration is.
type ExpiryStatus int

const (
	// NoExpiryInfo means ParseExpiration found nothing to report.
	NoExpiryInfo ExpiryStatus = iota
	ExpiryOK
	ExpiringSoon
	Expired
)

// CheckExpiration classifies resp's token expiration, if any is
// present at all.
func CheckExpiration(resp *github.Response) (status ExpiryStatus, expiresAt time.Time) {

	expiresAt, ok := ParseExpiration(resp)
	if !ok {
		return NoExpiryInfo, time.Time{}
	}

	switch {
	case time.Now().After(expiresAt):
		return Expired, expiresAt
	case time.Until(expiresAt) <= WarnWindow:
		return ExpiringSoon, expiresAt
	default:
		return ExpiryOK, expiresAt
	}
}

// IsUnauthorized reports whether err represents GitHub rejecting the
// request's credentials specifically (HTTP 401) — as opposed to a
// network failure, rate limiting, or any other error shape.
func IsUnauthorized(err error) bool {

	var ghErr *github.ErrorResponse

	if errors.As(err, &ghErr) {
		return ghErr.Response != nil && ghErr.Response.StatusCode == 401
	}

	return false
}

// Diagnose turns an authentication-related API error into a message
// that explains what's actually wrong, incorporating expiry
// information when GitHub provided it. ok is false when err isn't an
// authentication problem at all (network error, rate limiting, etc.) —
// callers should fall back to showing err directly in that case, since
// this has nothing more useful to add.
func Diagnose(err error, resp *github.Response) (message string, ok bool) {

	if err == nil || !IsUnauthorized(err) {
		return "", false
	}

	status, expiresAt := CheckExpiration(resp)

	if status == Expired {
		return fmt.Sprintf(
			"GitHub rejected the token: it expired on %s. Generate a new token and run `gitback init --force` (or `gitback init --use-env-token`).",
			clock.HumanDate(expiresAt),
		), true
	}

	return "GitHub rejected the token (invalid, revoked, or insufficient permissions). " +
		"Verify the token and its permissions, or generate a new one and run `gitback init --force`.", true
}
