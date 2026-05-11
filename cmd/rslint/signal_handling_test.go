//go:build !windows

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

// signal.NotifyContext registration contract: when we ask for SIGHUP
// (alongside the existing SIGINT/SIGTERM), a SIGHUP delivered to the
// current process MUST cancel the derived ctx. The previous setup only
// listed SIGINT/SIGTERM, so a controlling-terminal disconnect (SIGHUP)
// fell through to the OS default action — abrupt termination, IPC peer
// abandoned, profiling files truncated.
//
// We deliver the signal to ourselves via syscall.Kill(os.Getpid()) so
// it lands at this process's signal handler. Skipped on Windows —
// SIGHUP isn't delivered naturally there and the CLI/LSP code's SIGHUP
// registration is a documented no-op on that platform.
func TestSignalNotifyContext_AllThreeSignalsCancelCtx(t *testing.T) {
	for _, tc := range []struct {
		name string
		sig  syscall.Signal
	}{
		{"SIGINT", syscall.SIGINT},
		{"SIGTERM", syscall.SIGTERM},
		{"SIGHUP", syscall.SIGHUP},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, stop := signal.NotifyContext(
				context.Background(),
				syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP,
			)
			defer stop()

			if err := syscall.Kill(os.Getpid(), tc.sig); err != nil {
				t.Fatalf("syscall.Kill: %v", err)
			}

			select {
			case <-ctx.Done():
				if ctx.Err() == nil {
					t.Errorf("ctx.Done() fired but ctx.Err() is nil")
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("ctx did not fire within 2s after %s — signal handler missed?", tc.name)
			}
		})
	}
}

// signal.Notify channel registration: same three signals must reach
// the init-handshake channel pattern used by ipc_cli.go.
func TestSignalNotifyChannel_AllThreeSignalsDeliver(t *testing.T) {
	for _, tc := range []struct {
		name string
		sig  syscall.Signal
	}{
		{"SIGINT", syscall.SIGINT},
		{"SIGTERM", syscall.SIGTERM},
		{"SIGHUP", syscall.SIGHUP},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			defer signal.Stop(ch)

			if err := syscall.Kill(os.Getpid(), tc.sig); err != nil {
				t.Fatalf("syscall.Kill: %v", err)
			}

			select {
			case got := <-ch:
				if got != tc.sig {
					t.Errorf("got signal %v, want %v", got, tc.sig)
				}
			case <-time.After(2 * time.Second):
				t.Fatalf("channel did not receive %s within 2s", tc.name)
			}
		})
	}
}
