package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	os.Exit(runMain())
}

const signalEnv = "RSLINT_STOP"

func waitForDebugSignal(pollInterval time.Duration) {
	if os.Getenv(signalEnv) == "" {
		return
	}
	log.Printf("waiting for debug SIGUSR2 signal, send signal to pid(%d) to continue", os.Getpid())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGUSR2)
	sig := <-sigCh
	log.Println("SIGUSR2 signal:", sig)
}
func runMain() int {
	waitForDebugSignal(10000 * time.Millisecond)
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
