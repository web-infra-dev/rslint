package server

import (
	"fmt"
	"io"
	"os"

	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

func printConfigDiscoveryFailures(failures []discovery.ConfigFailure) {
	writeConfigDiscoveryFailures(os.Stderr, failures)
}

func writeConfigDiscoveryFailures(w io.Writer, failures []discovery.ConfigFailure) {
	for _, failure := range failures {
		fmt.Fprintf(w, "Warning: skipping config %s: %s\n", failure.Path, failure.Message)
	}
}

func reportShadowedPluginRules(shadowed []string) {
	writeShadowedPluginRules(os.Stderr, shadowed)
}

func writeShadowedPluginRules(w io.Writer, shadowed []string) {
	for _, ruleName := range shadowed {
		fmt.Fprintf(
			w,
			"rslint: plugin rule %q is shadowed by a built-in rule of the same name; using the built-in.\n",
			ruleName,
		)
	}
}
