//go:build js && wasm || wasip1 && wasm

package main

import (
	"time"
)

func waitForDebugSignal(pollInterval time.Duration) {
	// WASM doesn't support signals in the same way as Unix systems
	// This is a placeholder implementation for WASM builds
}
