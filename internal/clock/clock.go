// internal/clock/clock.go
//
// Package clock declares gitback's two timestamp formats, so neither is
// ever hardcoded or re-decided ad hoc at a call site.
package clock

import "time"

// LocalRFC3339 formats t as local time in RFC3339, with a numeric UTC
// offset — e.g. "2026-09-28T11:26:16+05:30". Used for anything a
// person reads directly: structured log entries, gitback health's
// generated_at, mirrors.json's sync times.
func LocalRFC3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

// LocalRFC3339Now is LocalRFC3339(time.Now()).
func LocalRFC3339Now() string {
	return LocalRFC3339(time.Now())
}

// filenameLayout produces a UTC timestamp safe to embed in a filename
// and lexicographically sortable in chronological order, e.g.
// "2026-09-28T05-56-16Z".
const filenameLayout = "2006-01-02T15-04-05Z"

// FilenameUTC formats t using filenameLayout, always in UTC regardless
// of t's own location — e.g. time.Now() at 11:26:16+05:30 becomes
// "2026-09-28T05-56-16Z".
//
// Always UTC, deliberately: this format is used for machine-generated,
// embedded-in-a-filename timestamps (snapshot archive names, quarantine
// collision suffixes) where "lexical order == chronological order" is
// relied on elsewhere (ApplyRetention's snapshot pruning). Local time
// can move backward in wall-clock terms across a DST "fall back"
// transition, which would silently break that invariant; UTC has no
// DST, so it can't.
func FilenameUTC(t time.Time) string {
	return t.UTC().Format(filenameLayout)
}

// FilenameUTCNow is FilenameUTC(time.Now()).
func FilenameUTCNow() string {
	return FilenameUTC(time.Now())
}

// HumanDateLayout formats a date the way gitback shows it to a person —
// e.g. "Mon, Oct 12 2026". Used for every expiry-related date shown to
// the user.
const HumanDateLayout = "Mon, Jan 2 2006"

func HumanDate(t time.Time) string {
	return t.Format(HumanDateLayout)
}
