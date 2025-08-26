//go:build linux || darwin

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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
