//go:build windows

package main

import (
	"log"
	"os"
	"time"
)

func waitForDebugSignal(pollInterval time.Duration) {
	if os.Getenv(signalEnv) == "" {
		return
	}
	log.Printf("Debug signal waiting is not supported on Windows. Continuing immediately.")
	// On Windows, we don't have SIGUSR2, so we just continue
	// In the future, we could implement Windows-specific debug mechanisms
	// like named pipes, events, or polling a file
}

// dumpSignals returns no signal on Windows: the standard syscall package has
// no SIGBREAK, and Go's runtime delivers CTRL_BREAK_EVENT as SIGINT, which the
// lint signal handler already owns. The on-demand signal dump is therefore
// unix-only; on Windows the hang watchdog (hangdiag.go) is the active dump
// mechanism and needs no signal.
func dumpSignals() []os.Signal {
	return nil
}
