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

func runMain() int {
	waitForDebugSignal(10000 * time.Millisecond)
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--lsp":
			// run in LSP mode for Language Server
			return runLSP(args[1:])
		case "--api":
			// run in API mode for JavaScript API
			return runAPI()
		}
	}
	// Default: unified IPC CLI mode — the Node parent (cli.ts → engine.ts)
	// drives stdin/stdout as a length-prefixed-JSON IPC frame stream; the Go
	// child runs the init handshake + lint pipeline and forwards output /
	// acks shutdown. On js/wasm this resolves to the native fallback
	// (ipc_cli_js.go).
	return runCLI(args)
}
