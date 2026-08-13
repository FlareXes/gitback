// internal/cmd/workflow.go

package cmd

import (
	"context"
	"fmt"

	"github.com/flarexes/gitback/internal/config"
	"github.com/flarexes/gitback/internal/discovery"
	"github.com/flarexes/gitback/internal/filesystem"
	"github.com/flarexes/gitback/internal/lock"
	"github.com/flarexes/gitback/internal/logging"
	"github.com/flarexes/gitback/internal/mirror"
	"github.com/flarexes/gitback/internal/runtime"
	"github.com/flarexes/gitback/internal/snapshot"
)

// Runtime contains everything required to execute a GitBack command.
type Runtime struct {
	Config *config.Config
	Layout runtime.Layout
	Logger *logging.Logger
}

// prepareRuntime resolves the layout, loads config, ensures directories
// exist, and opens the logger — once, shared by every command.
func prepareRuntime() (*Runtime, error) {
	layout, err := runtime.New()
	if err != nil {
		return nil, err
	}

	if err := layout.EnsureDirs(); err != nil {
		return nil, err
	}

	cfg, err := config.Load(layout)
	if err != nil {
		return nil, err
	}

	// Mirror/snapshot dirs are user-configured, so they're ensured here
	// rather than inside layout.EnsureDirs().
	for _, dir := range []string{cfg.Storage.MirrorRoot, cfg.Snapshot.OutputDirectory} {
		if _, err := filesystem.CreateDirectory(dir); err != nil {
			return nil, err
		}
	}

	logger, err := logging.New(layout.LogFile)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		Config: cfg,
		Layout: layout,
		Logger: logger,
	}, nil
}

func withLock(logger *logging.Logger, lockFile string, fn func() error) error {
	locker := lock.New(lockFile)

	unlock, err := locker.Acquire()
	if err != nil {
		logger.Error(logging.Events.Lock.Busy, "", err)
		return err
	}

	logger.Info(logging.Events.Lock.Acquired, "")
	defer func() {
		unlock()
		logger.Info(logging.Events.Lock.Released, "")
	}()

	return fn()
}

func executeRun(ctx context.Context) error {
	rt, err := prepareRuntime()
	if err != nil {
		return err
	}
	defer rt.Logger.Close()

	return withLock(rt.Logger, rt.Layout.LockFile, func() error {
		if err := executeDiscover(ctx, rt); err != nil {
			return err
		}
		if err := executeSync(ctx, rt); err != nil {
			return err
		}
		return executeSnapshot(ctx, rt, true)
	})
}

func executeDiscover(ctx context.Context, rt *Runtime) error {
	logger := rt.Logger
	logger.Info(logging.Events.GitHub.DiscoveryStarted, "")

	client, err := discovery.New(rt.Config, rt.Layout, logger)
	if err != nil {
		return err
	}

	if err := client.Discover(ctx); err != nil {
		logger.Error(logging.Events.GitHub.DiscoveryFailed, "", err)
		return fmt.Errorf("repository discovery failed: %w", err)
	}

	return nil
}

func executeSync(ctx context.Context, rt *Runtime) error {
	logger := rt.Logger
	logger.Info(logging.Events.Sync.Started, "")

	engine := mirror.New(rt.Config, rt.Layout, logger)
	if err := engine.Sync(ctx); err != nil {
		logger.Error(logging.Events.Sync.Failed, "", err)
		return err
	}

	logger.Info(logging.Events.Sync.Completed, "")
	return nil
}

func executeSnapshot(ctx context.Context, rt *Runtime, force bool) error {
	logger := rt.Logger
	logger.Info(logging.Events.Snapshot.Started, "")

	engine := snapshot.New(rt.Config, rt.Layout, logger)
	if err := engine.Create(ctx, force); err != nil {
		logger.Error(logging.Events.Snapshot.Failed, "", err)
		return err
	}

	logger.Info(logging.Events.Snapshot.Completed, "")
	return nil
}
