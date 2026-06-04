// internal/snapshot/snapshot.go

package snapshot

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

type Engine struct {
	cfg    *config.Config
	logger *logging.Logger
}

func New(cfg *config.Config, logger *logging.Logger) *Engine {
	return &Engine{
		cfg:    cfg,
		logger: logger,
	}
}

func (e *Engine) Create(ctx context.Context, headless bool) error {

	start := time.Now()

	// Verify dependencies before doing anything.
	if err := e.verifyDependencies(); err != nil {

		return err
	}

	// Verify mirror state.
	if err := e.verifyMirrors(); err != nil {

		e.logger.Error(
			logging.Events.Snapshot.VerificationFailed,
			"",
			err,
		)

		if !headless {
			return err
		}

		e.logger.Warn(
			logging.Events.Snapshot.VerificationFailed,
			"",
			"headless mode enabled, continuing snapshot",
		)
	}

	timestamp := time.Now().
		UTC().
		Format("2006-01-02T15-04-05Z")

	tarFile := filepath.Join(e.cfg.SnapshotDir, timestamp+".tar")
	archiveFile := tarFile + ".zst"
	checksumFile := archiveFile + ".sha256"

	// Create tar archive.
	if err := e.createTar(ctx, tarFile); err != nil {

		return err
	}

	// Compress archive.
	if err := e.compress(ctx, tarFile); err != nil {

		return err
	}

	// Generate checksum.
	if err := e.sha256(archiveFile, checksumFile); err != nil {

		return err
	}

	// Remove uncompressed tar.
	_ = os.Remove(tarFile)

	e.logger.Duration(
		logging.Events.Snapshot.Completed,
		"",
		time.Since(start),
	)

	return nil
}

func (e *Engine) verifyDependencies() error {

	required := []string{"tar", "zstd"}

	for _, binary := range required {

		if _, err := exec.LookPath(binary); err != nil {

			return fmt.Errorf("%s not found in PATH", binary)
		}
	}

	return nil
}

func (e *Engine) verifyMirrors() error {

	e.logger.Info(
		logging.Events.Snapshot.VerificationStarted,
		"",
	)

	data, err := state.Load(e.cfg.MirrorsStateFile)

	if err != nil {

		if os.IsNotExist(err) {

			return fmt.Errorf(
				"mirror state file not found: run `gitback sync` first",
			)
		}

		return err
	}

	var failed []string

	for _, repo := range data.Repositories {

		if !repo.LastSuccess {

			failed = append(
				failed,
				repo.Name,
			)
		}
	}

	if len(failed) > 0 {

		return fmt.Errorf(
			"mirror verification failed: %d repositories unhealthy",
			len(failed),
		)
	}

	e.logger.Info(
		logging.Events.Snapshot.VerificationPassed,
		"",
	)

	return nil
}

// Create a tar archive containing:
//
//	mirrors/
//	state/mirrors.json
func (e *Engine) createTar(ctx context.Context, output string) error {

	// Equivalent shell command:
	// tar -cf <output> -C <baseDir> mirrors state/mirrors.json
	cmd := exec.CommandContext(
		ctx,
		"tar",
		"-cf",
		output,

		"-C",
		e.cfg.BaseDir,

		"mirrors",

		filepath.Join(
			"state",
			"mirrors.json",
		),
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("tar failed: %s", string(out))
	}

	return nil
}

// Compress a tar archive using zstd.
func (e *Engine) compress(ctx context.Context, tarFile string) error {

	cmd := exec.CommandContext(
		ctx,
		"zstd",
		"-T2",
		tarFile,
	)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("zstd failed: %s", string(out))
	}

	return nil
}

// Generate a SHA-256 checksum for a file.
func (e *Engine) sha256(input string, output string) error {

	file, err := os.Open(input)
	if err != nil {
		return err
	}

	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	sum := fmt.Sprintf(
		"%x  %s\n",
		hash.Sum(nil),
		filepath.Base(input),
	)

	return os.WriteFile(
		output,
		[]byte(sum),
		0600,
	)
}
