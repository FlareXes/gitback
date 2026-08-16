// internal/cmd/signal.go

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// runCancelable runs fn with a context that is canceled on SIGINT
// (Ctrl+C) or SIGTERM.
//
// Every execute* function in this package already threads its ctx
// parameter down into exec.CommandContext for git/tar/zstd subprocesses,
// so once fn receives a context that is actually wired to the OS
// signal, an in-flight subprocess is killed automatically the moment a
// signal arrives — no further plumbing is needed downstream. Before
// this, every call site passed context.Background(), which can never be
// canceled, so that existing ctx plumbing had nothing to respond to.
//
// fn still returns normally afterward (with a context.Canceled-derived
// error surfaced from whichever subprocess was interrupted), so callers
// such as withLock's deferred unlock still run cleanly — the lock is
// released deliberately rather than only by the kernel reclaiming the
// fd when the process exits.
func runCancelable(fn func(ctx context.Context) error) error {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// done lets the goroutine below distinguish "a signal arrived" from
	// "fn finished on its own" so the shutdown message is only printed
	// when a signal actually interrupted something — not on every
	// ordinary, successful exit.
	done := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			fmt.Fprintln(
				os.Stderr,
				"\n[INFO] shutdown signal received; finishing in-progress operation and releasing lock...",
			)
		case <-done:
		}
	}()

	err := fn(ctx)

	close(done)

	return err
}
