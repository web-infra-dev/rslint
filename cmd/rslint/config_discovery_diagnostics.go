package main

import (
	"fmt"
	"io"
	"os"

	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

// writeConfigDiscoveryFailures renders non-fatal candidate failures for the
// surface adapter. Discovery itself stays transport-agnostic; CLI and API both
// use stderr so warnings never enter their framed stdout protocols.
func writeConfigDiscoveryFailures(w io.Writer, failures []discovery.ConfigFailure) {
	for _, failure := range failures {
		fmt.Fprintf(w, "Warning: skipping config %s: %s\n", failure.Path, failure.Message)
	}
}

func printConfigDiscoveryFailures(failures []discovery.ConfigFailure) {
	writeConfigDiscoveryFailures(os.Stderr, failures)
}
