// internal/logging/logger.go

package logging

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	runID string

	encoder *json.Encoder

	file *os.File

	mu sync.Mutex
}

// logFileNamePattern matches only gitback's own log filenames, e.g.
// "gitback-2026-09-05-saturday.log" — so pruning never touches
// unrelated files a user might keep in the same directory.
var logFileNamePattern = regexp.MustCompile(`^gitback-(\d{4}-\d{2}-\d{2})-[a-z]+\.log$`)

// dailyLogFileName builds today's log filename from t. Weekday is
// appended lowercase for readability without breaking shell globbing.
func dailyLogFileName(t time.Time) string {
	date := t.Format("2006-01-02")
	weekday := strings.ToLower(t.Format("Monday"))
	return fmt.Sprintf("gitback-%s-%s.log", date, weekday)
}

// CurrentLogFilePath returns the path New would open right now, given
// logDir. Used by config validation to check writability without opening a second
// file handle onto today's log.
func CurrentLogFilePath(logDir string) string {
	return filepath.Join(logDir, dailyLogFileName(time.Now()))
}

// New opens (creating if needed) today's log file under logDir.
//
// Local time decides the filename here, resolved once, at open time —
// not recomputed anywhere else in the Logger. Each gitback command is
// one short-lived process with exactly one Logger for its whole run,
// so a run spanning local midnight keeps writing to the file it
// opened at start; nothing re-checks the date mid-run.
//
// retentionDays prunes old log files in logDir; <= 0 disables pruning.
func New(logDir string, retentionDays int) (*Logger, error) {

	// Doctor may call this before EnsureDirs has ever run, so the
	// directory can't be assumed to already exist.
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	logFilePath := filepath.Join(logDir, dailyLogFileName(time.Now()))

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	logger := &Logger{
		runID:   generateRunID(),
		encoder: json.NewEncoder(file),
		file:    file,
	}

	// Best-effort: pruning must never stop gitback from running or
	// from writing today's log, so failures are logged, not returned.
	deleted, pruneErr := pruneOldLogs(logDir, retentionDays)

	if pruneErr != nil {

		logger.Warn(Events.LogRetention.PruneFailed, "", pruneErr.Error())

	} else if deleted > 0 {

		logger.Emit(Entry{
			Level: Info,
			Event: Events.LogRetention.Pruned,
			Details: map[string]any{
				"deleted": deleted,
			},
		})
	}

	return logger, nil
}

// pruneOldLogs removes gitback's own log files older than retentionDays
// retentionDays <= 0 disables pruning entirely (logs kept forever).
func pruneOldLogs(logDir string, retentionDays int) (int, error) {

	if retentionDays <= 0 {
		return 0, nil
	}

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return 0, fmt.Errorf("read log directory: %w", err)
	}

	// A file is kept as long as its date is on or after this day.
	cutoff := truncateToDay(time.Now().AddDate(0, 0, -retentionDays))

	deleted := 0
	var lastErr error

	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}

		match := logFileNamePattern.FindStringSubmatch(entry.Name())

		// Not one of gitback's dated log files — leave it alone.
		if match == nil {
			continue
		}

		fileDate, err := time.ParseInLocation("2006-01-02", match[1], time.Local)
		if err != nil {
			continue
		}

		if !fileDate.Before(cutoff) {
			continue
		}

		if err := os.Remove(filepath.Join(logDir, entry.Name())); err != nil {
			lastErr = err
			continue
		}

		deleted++
	}

	return deleted, lastErr
}

// truncateToDay strips the time-of-day component so cutoff comparisons
// are date-only, regardless of what time "now" happens to be.
func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func (l *Logger) Close() error {

	if l.file == nil {
		return nil
	}

	return l.file.Close()
}

func (l *Logger) Emit(entry Entry) {

	l.mu.Lock()
	defer l.mu.Unlock()

	// Local time, not UTC: readable at a glance for a human, and
	// RFC3339's numeric offset (e.g. "-07:00") still lets any log
	// consumer — a SIEM, jq, whatever — convert to UTC deterministically.
	entry.Timestamp = time.Now().Format(time.RFC3339)

	if entry.RunID == "" {
		entry.RunID = l.runID
	}

	_ = l.encoder.Encode(entry)
}

func (l *Logger) Info(event string, repo string) {
	l.Emit(Entry{
		Level: Info,
		Event: event,
		Repo:  repo,
	})
}

func (l *Logger) Warn(
	event string,
	repo string,
	message string,
) {
	l.Emit(Entry{
		Level: Warn,
		Event: event,
		Repo:  repo,
		Details: map[string]any{
			"message": message,
		},
	})
}

func (l *Logger) Error(
	event string,
	repo string,
	err error,
) {

	var errString string

	if err != nil {
		errString = err.Error()
	}

	l.Emit(Entry{
		Level: Error,
		Event: event,
		Repo:  repo,
		Error: errString,
	})
}

func (l *Logger) Duration(
	event string,
	repo string,
	duration time.Duration,
) {
	l.Emit(Entry{
		Level:      Info,
		Event:      event,
		Repo:       repo,
		DurationMS: duration.Milliseconds(),
	})
}

func generateRunID() string {

	buf := make([]byte, 4)

	if _, err := rand.Read(buf); err != nil {

		return time.Now().
			UTC().
			Format("20060102150405")
	}

	return hex.EncodeToString(buf)
}
