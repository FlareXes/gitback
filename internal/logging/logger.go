// internal/logging/logger.go

package logging

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Logger struct {
	runID string

	encoder *json.Encoder

	file *os.File

	mu sync.Mutex
}

func New(logFile string) (*Logger, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf(
			"open log file: %w",
			err,
		)
	}

	return &Logger{
		runID:   generateRunID(),
		encoder: json.NewEncoder(file),
		file:    file,
	}, nil
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

	entry.Timestamp = time.Now().UTC().Format(time.RFC3339)

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
