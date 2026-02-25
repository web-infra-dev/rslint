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
			// run in API mode for JS API
			return runAPI()
		}
	}
	// run in CLI mode for direct command line usage
	return runCMD()
}
