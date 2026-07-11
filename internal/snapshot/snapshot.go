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
	"strings"
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

func (e *Engine) Create(ctx context.Context, force bool) error {

	start := time.Now()

	// Verify dependencies before doing anything.
	if err := e.verifyDependencies(); err != nil {

		return err
	}

	// Verify mirror state
	fmt.Println("[1/5] Verifying mirrors")
	if err := e.verifyMirrors(); err != nil {

		e.logger.Error(
			logging.Events.Snapshot.VerificationFailed,
			"",
			err,
		)

		if !force {
			return err
		}

		e.logger.Warn(
			logging.Events.Snapshot.VerificationFailed,
			"",
			"snapshot --force mode enabled, continuing snapshot",
		)
	}

	timestamp := time.Now().
		UTC().
		Format("2006-01-02T15-04-05Z")

	tarFile := filepath.Join(e.cfg.SnapshotDir, timestamp+".tar")
	archiveFile := tarFile + ".zst"
	checksumFile := archiveFile + ".sha256"

	// check if tarFile already exists to avoid collision
	if _, err := os.Stat(tarFile); err == nil {

		err_msg := fmt.Errorf("temporary archive already exists: %s", tarFile)

		e.logger.Error(
			logging.Events.Snapshot.CollisionDetected,
			"",
			err_msg,
		)

		return err_msg
	}

	// check if archiveFile already exists to avoid collision
	if _, err := os.Stat(archiveFile); err == nil {

		err_msg := fmt.Errorf("snapshot already exists: %s", archiveFile)

		e.logger.Error(
			logging.Events.Snapshot.CollisionDetected,
			"",
			err_msg,
		)

		return err_msg
	}

	// Create tar archive.
	e.logger.Info(
		logging.Events.Snapshot.ArchiveStarted,
		"",
	)

	fmt.Println("[2/5] Creating archive")

	if err := e.createTar(ctx, tarFile); err != nil {
		return err
	}

	e.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Snapshot.ArchiveCompleted,

			Details: map[string]any{
				"archive": tarFile,
			},
		},
	)

	// Always cleanup temporary tar file.
	defer os.Remove(tarFile)

	// Compress archive.
	e.logger.Info(
		logging.Events.Snapshot.CompressionStarted,
		"",
	)

	fmt.Println("[3/5] Compressing archive")

	if err := e.compress(ctx, tarFile); err != nil {
		return err
	}

	e.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Snapshot.CompressionCompleted,

			Details: map[string]any{
				"archive": archiveFile,
			},
		},
	)

	// Generate checksum.
	e.logger.Info(
		logging.Events.Snapshot.ChecksumStarted,
		"",
	)

	fmt.Println("[4/5] Generating checksum")

	if err := e.sha256(archiveFile, checksumFile); err != nil {
		return err
	}

	e.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Snapshot.ChecksumCompleted,

			Details: map[string]any{
				"checksum": checksumFile,
			},
		},
	)

	// Apply retention policy.
	fmt.Println("[5/5] Applying retention policy")
	if err := ApplyRetention(e.cfg, e.logger); err != nil {

		e.logger.Error(
			logging.Events.Snapshot.RetentionFailed,
			"",
			err,
		)
	}

	fmt.Println()
	fmt.Println("Snapshot saved at " + archiveFile)

	e.logger.Emit(
		logging.Entry{
			Level: logging.Info,
			Event: logging.Events.Snapshot.Summary,

			DurationMS: time.Since(start).Milliseconds(),

			Details: map[string]any{
				"archive":    archiveFile,
				"checksum":   checksumFile,
				"force_mode": force,
			},
		},
	)

	return nil
}

// verifyDependencies checks if required system dependencies are available.
func (e *Engine) verifyDependencies() error {

	required := []string{"tar", "zstd"}

	for _, binary := range required {

		if _, err := exec.LookPath(binary); err != nil {

			return fmt.Errorf("%s not found in PATH", binary)
		}
	}

	return nil
}

// verifyMirrors checks the health of all mirrored repositories.
func (e *Engine) verifyMirrors() error {

	e.logger.Info(
		logging.Events.Snapshot.VerificationStarted,
		"",
	)

	data, err := state.LoadMirrors(e.cfg.MirrorsStateFile)

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

		var builder strings.Builder

		builder.WriteString(
			"mirror verification failed\n\n",
		)

		builder.WriteString(
			"Failed repositories:\n",
		)

		for _, repo := range failed {

			fmt.Fprintf(
				&builder,
				" - %s\n",
				repo,
			)
		}

		builder.WriteString(
			"\nRun:\n  gitback sync\n",
		)

		builder.WriteString(
			"\nOr create a snapshot anyway:\n  gitback snapshot --force",
		)

		return fmt.Errorf("%s", builder.String())
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
	// tar -cf <output> -C <dataDir> mirrors state/mirrors.json
	cmd := exec.CommandContext(
		ctx,
		"tar",
		"-cf",
		output,

		"-C",
		filepath.Dir(e.cfg.MirrorDir),

		filepath.Base(e.cfg.MirrorDir),

		filepath.Join(
			filepath.Base(e.cfg.StateDir),
			filepath.Base(e.cfg.MirrorsStateFile),
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
