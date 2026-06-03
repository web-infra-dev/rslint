//go:build js

package main

import (
	"context"
	"fmt"
	"os"
)

// runCLI on js/wasm is a native fallback. The IPC CLI mode (a Node parent
// over OS stdio, with SIGINT/SIGTERM/SIGHUP handling) isn't available here:
// syscall.SIGHUP is undefined on js/wasm and there is no Node IPC parent. The
// wasm build (packages/rslint-wasm) reaches this binary through the same
// cmd/rslint main entry, so we run the shared pipeline natively, exactly as
// the removed runCMD did.
func runCLI(args []string) int {
	parsed, help, fatal := parseLintFlags(args)
	if fatal != 0 {
		return fatal
	}
	if help {
		fmt.Fprint(os.Stderr, usage)
		return 0
	}
	// js/wasm has no Node IPC host, so no eslint-plugin dispatch (nil).
	return executeLintPipeline(parsed, context.Background(), nil)
}
