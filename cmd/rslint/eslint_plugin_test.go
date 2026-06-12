package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// TestPluginConfigResolver_NormalizedMatchRawWireKey proves the routing split on
// the Go side (no Programs / fs needed): the resolver matches files against Go's
// NORMALIZED config key but emits the RAW configDirectory the host sent as the
// wire configKey (what the worker keys its plugin map on). This is the exact
// assertion that fails under a normalize-the-wire-key design and passes here.
func TestPluginConfigResolver_NormalizedMatchRawWireKey(t *testing.T) {
	data := []byte(`{"configs":[{"configDirectory":"C:\\proj","entries":[{"rules":{"no-debugger":"error"}}]}]}`)
	p, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	raw := `C:\proj`
	norm := tspath.NormalizePath(raw) // "C:/proj"

	// Windows: file matches on the normalized key, wire key is the RAW string.
	r := pluginConfigResolver{configMap: p.ConfigMap, originalConfigDir: p.OriginalConfigDir}
	wireKey, merged := r.resolve(norm + "/src/a.ts")
	if wireKey != raw {
		t.Errorf("wire configKey = %q, want RAW %q (not the normalized %q)", wireKey, raw, norm)
	}
	if merged == nil {
		t.Fatal("expected a merged config for the matched file")
	}

	// POSIX / no raw mapping: wireKey falls back to the (already-slash) key.
	posix := pluginConfigResolver{
		configMap: map[string]rslintconfig.RslintConfig{"/posix/proj": p.ConfigMap[norm]},
	}
	if wk, m := posix.resolve("/posix/proj/a.ts"); wk != "/posix/proj" || m == nil {
		t.Errorf("POSIX fallback: wireKey=%q merged-nil=%v, want /posix/proj + non-nil", wk, m == nil)
	}
}

func pluginInput() []linter.EslintPluginFileInput {
	return []linter.EslintPluginFileInput{
		{Path: "/proj/a.ts", ConfigKey: "/proj", Rules: []linter.ConfiguredRule{
			{Name: "uc/x", Severity: rule.SeverityError, IsEslintPluginRule: true},
		}},
	}
}

// TestDispatchPluginLintAsync_DispatchErrorSurfacesDiagnostic pins U1: a total
// dispatch failure (the whole plugin-lint phase never ran) must surface one
// error diagnostic so the CLI exit code reflects it, not a stderr-only false
// green.
func TestDispatchPluginLintAsync_DispatchErrorSurfacesDiagnostic(t *testing.T) {
	failing := func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		return nil, errors.New("WorkerPool: closed")
	}
	diags := <-dispatchPluginLintAsync(context.Background(), failing, pluginInput(), false, "off")
	if len(diags) != 1 {
		t.Fatalf("dispatch failure should surface 1 diagnostic, got %d", len(diags))
	}
	if diags[0].RuleName != "rslint/plugin-lint-error" || diags[0].Severity != rule.SeverityError {
		t.Errorf("want rslint/plugin-lint-error/SeverityError, got %q/%v", diags[0].RuleName, diags[0].Severity)
	}
	if diags[0].FilePath != "/proj/a.ts" {
		t.Errorf("diagnostic should anchor to the first input file, got %q", diags[0].FilePath)
	}
	if !strings.Contains(diags[0].Message.Description, "WorkerPool: closed") {
		t.Errorf("message should include the dispatch error, got %q", diags[0].Message.Description)
	}
}

// TestDispatchPluginLintAsync_CanceledNoDiagnostic verifies context.Canceled is
// a cooperative drop (editor/CLI aborted), NOT a false green, so it must not add
// an error diagnostic.
func TestDispatchPluginLintAsync_CanceledNoDiagnostic(t *testing.T) {
	canceled := func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		return nil, context.Canceled
	}
	if diags := <-dispatchPluginLintAsync(context.Background(), canceled, pluginInput(), false, "off"); len(diags) != 0 {
		t.Errorf("context.Canceled should yield 0 diagnostics, got %d", len(diags))
	}
}

// TestDispatchPluginLintAsync_NoInputsNoDiagnostic verifies the empty/no-op
// paths contribute nothing.
func TestDispatchPluginLintAsync_NoInputsNoDiagnostic(t *testing.T) {
	dispatch := func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		t.Fatal("dispatch must not be called with no inputs")
		return nil, errors.New("unreachable")
	}
	if diags := <-dispatchPluginLintAsync(context.Background(), dispatch, nil, false, "off"); len(diags) != 0 {
		t.Errorf("no inputs should yield 0 diagnostics, got %d", len(diags))
	}
}

// TestPluginConfigResolver_Branches covers resolve()'s non-match and
// single-config (configMap==nil) fallback branches.
func TestPluginConfigResolver_Branches(t *testing.T) {
	data := []byte(`{"configs":[{"configDirectory":"/proj","entries":[{"rules":{"no-debugger":"error"}}]}]}`)
	p, err := parseConfigPayload(data)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Multi-config, file under no config -> ("", nil).
	r := pluginConfigResolver{configMap: p.ConfigMap, originalConfigDir: p.OriginalConfigDir}
	if wk, m := r.resolve("/elsewhere/a.ts"); wk != "" || m != nil {
		t.Errorf("no-match -> (\"\",nil), got (%q, nil=%v)", wk, m == nil)
	}

	// Single-config (configMap==nil): wireKey is currentDirectory; merged from rslintConfig.
	single := pluginConfigResolver{rslintConfig: p.ConfigMap["/proj"], currentDirectory: "/proj"}
	if wk, m := single.resolve("/proj/a.ts"); wk != "/proj" || m == nil {
		t.Errorf("single-config -> (currentDirectory, merged), got (%q, nil=%v)", wk, m == nil)
	}
}
