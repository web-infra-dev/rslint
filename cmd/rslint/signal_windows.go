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
