package main

import (
	"os"
	"time"
)

func main() {
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
			return runLSP()
		case "--api", "--ipc":
			// run in API/IPC mode for JS API
			return runAPI()
		}
	}
	// run in CLI mode for direct command line usage
	return int(runCMD())
}
