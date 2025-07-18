package main

import (
	"os"
)

func main() {
	os.Exit(runMain())
}

func runMain() int {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--lsp":
			// run in LSP mode for Language Server
			return runLSP()
		case "--api":
			// run in API mode for JS API
			return runAPI()
		}
	}
	// run in CLI mode for direct command line usage
	return int(runCMD())
}
