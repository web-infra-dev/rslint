package server

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

func TestConfigDiagnosticsPreserveAPIMessages(t *testing.T) {
	var output strings.Builder
	writeConfigDiscoveryFailures(&output, []discovery.ConfigFailure{{
		Path:    "/repo/rslint.config.js",
		Message: "module failed",
	}})
	if want := "Warning: skipping config /repo/rslint.config.js: module failed\n"; output.String() != want {
		t.Fatalf("config failure = %q, want %q", output.String(), want)
	}

	output.Reset()
	writeShadowedPluginRules(&output, []string{"plugin/rule"})
	if want := "rslint: plugin rule \"plugin/rule\" is shadowed by a built-in rule of the same name; using the built-in.\n"; output.String() != want {
		t.Fatalf("shadow warning = %q, want %q", output.String(), want)
	}
}
