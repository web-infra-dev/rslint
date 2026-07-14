package main

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

func TestWriteConfigDiscoveryFailures(t *testing.T) {
	var output strings.Builder
	writeConfigDiscoveryFailures(&output, []discovery.ConfigFailure{
		{Path: "/repo/broken-a/rslint.config.mjs", Message: "unexpected token"},
		{Path: "/repo/broken-b/rslint.config.ts", Message: "module not found"},
	})

	want := "Warning: skipping config /repo/broken-a/rslint.config.mjs: unexpected token\n" +
		"Warning: skipping config /repo/broken-b/rslint.config.ts: module not found\n"
	if got := output.String(); got != want {
		t.Fatalf("config discovery warnings = %q, want %q", got, want)
	}
}
