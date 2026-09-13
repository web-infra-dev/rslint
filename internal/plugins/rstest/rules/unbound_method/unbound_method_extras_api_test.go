package unbound_method_test

import (
	"encoding/json"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/api/server"
	"github.com/web-infra-dev/rslint/internal/testutil/txtarfs"
)

func TestUnboundMethodAPICapability(t *testing.T) {
	root := tspath.NormalizePath(txtarfs.MustParseFile(t, "testdata/capabilities.txtar").Materialize(t, ""))
	response, err := (&server.Handler{}).HandleLint(api.LintRequest{
		WorkingDirectory: root, ConfigDirectory: root, Files: []string{tspath.ResolvePath(root, "input.ts")},
		Config: json.RawMessage(`[{"plugins":["rstest"],"languageOptions":{"parserOptions":{"project":["./tsconfig.json"]}},"rules":{"rstest/unbound-method":"error"}}]`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.RuleCount != 1 || len(response.Diagnostics) != 8 {
		t.Fatalf("API capability: %+v", response)
	}
}
