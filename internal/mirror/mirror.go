// internal/mirror/mirror.go

package mirror

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/state"
)

func printSyncSummary(label string, assets []state.Asset) {

	var failed []string
	var healthy int

	for _, asset := range assets {

		if asset.LastSuccess {
			healthy++
			continue
		}

		failed = append(failed, asset.Name)
	}

	fmt.Println()
	fmt.Println(label)

	fmt.Printf("  Total:   %d\n", len(assets))
	fmt.Printf("  Healthy: %d\n", healthy)
	fmt.Printf("  Failed:  %d\n", len(failed))

	if len(failed) > 0 {

		fmt.Println()
		fmt.Println("  Failed assets:")

		for _, asset := range failed {
			fmt.Printf("    - %s\n", asset)
		}
	}
}

func (e *Engine) cloneMirror(ctx context.Context, repo string, target string) error {

	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(repo),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.CloneStarted,
		repoName,
	)

	askPass, err := e.createAskPassScript()
	if err != nil {
		return err
	}

	defer os.Remove(askPass)

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"clone",
		"--mirror",
		repo,
		target,
	)

	if err != nil {

		e.logger.Error(
			logging.Events.Mirror.CloneFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

		return err
	}

	e.logger.Duration(
		logging.Events.Mirror.CloneCompleted,
		repoName,
		time.Since(start),
	)

	return nil
}

func (e *Engine) updateMirror(ctx context.Context, target string) error {
	start := time.Now()

	repoName := strings.TrimSuffix(
		filepath.Base(target),
		".git",
	)

	e.logger.Info(
		logging.Events.Mirror.UpdateStarted,
		repoName,
	)

	askPass, err := e.createAskPassScript()
	if err != nil {
		return err
	}

	defer os.Remove(askPass)

	output, err := e.runGit(
		ctx,
		repoName,
		e.gitEnv(askPass),

		"-C",
		target,
		"remote",
		"update",
		"--prune",
	)

	if err != nil {

		e.logger.Error(
			logging.Events.Mirror.UpdateFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

		return err
	}

	e.logger.Info(
		logging.Events.Mirror.FsckStarted,
		repoName,
	)

	fsck := exec.CommandContext(
		ctx,
		"git",
		"-C",
		target,
		"fsck",
		"--no-dangling",
	)

	if output, err := fsck.CombinedOutput(); err != nil {

		e.logger.Error(
			logging.Events.Mirror.FsckFailed,
			repoName,
			fmt.Errorf("%s", string(output)),
		)

	} else {

		e.logger.Info(
			logging.Events.Mirror.FsckCompleted,
			repoName,
		)
	}

	e.logger.Duration(
		logging.Events.Mirror.UpdateCompleted,
		repoName,
		time.Since(start),
	)

	return nil
}

func (e *Engine) syncMirror(ctx context.Context, url string, mirrorDir string) error {

	name := strings.TrimSuffix(filepath.Base(url), ".git")
	target := filepath.Join(mirrorDir, name+".git")

	// clone if asset doesn't exist
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return e.cloneMirror(ctx, url, target)
	}

	// update existing asset
	return e.updateMirror(ctx, target)
}
