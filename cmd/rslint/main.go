package main

import (
	"log"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	os.Exit(runMain())
}

const signalEnv = "RSLINT_STOP"

// runMain dispatches to the appropriate mode entry point.
//
// Three mode entries, each with its own orchestration but all sharing the
// `internal/linter` primitives at the bottom:
//
//	--lsp      → runLSP   (Language Server Protocol; long-running,
//	                       per-file incremental lint driven by VS Code /
//	                       editor)
//	--api      → runAPI   (request/response IPC for wasm / rslint-api JS
//	                       callers; not user-facing)
//	default    → runCLI   (every CLI invocation. The binary is an internal
//	                       npm artifact — users always reach it through
//	                       `rslint.cjs → cli.ts → engine.ts` which sends
//	                       an `init` IPC message carrying configs, plugin
//	                       entries, and runtime hints. The user's intent
//	                       (lint / --init / --help) is conveyed via the
//	                       forwarded argv, NOT via init payload fields,
//	                       so runCLI dispatches on the parsed flags
//	                       (`baseArgs.Init`, the `help` flag) rather
//	                       than on payload contents.)
//
// All CLI invocations flow through runCLI's IPC handshake — there is
// no direct-binary CLI path. The binary is shipped as an internal npm
// artifact; users always reach it through `rslint.cjs → cli.ts →
// engine.ts`, which handshakes via stdin/stdout IPC.
func runMain() int {
	waitForDebugSignal(10000 * time.Millisecond)
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--lsp":
			return runLSP(args[1:])
		case "--api":
			return runAPI()
		}
	}
	return runCLI(args)
}
