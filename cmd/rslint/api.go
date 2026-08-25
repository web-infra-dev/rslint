package main

import (
	"fmt"
	"os"

	"github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/api/server"
)

// runAPI binds the API service to the process stdio streams.
func runAPI() int {
	service := api.NewService(os.Stdin, os.Stdout, &server.Handler{})
	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
