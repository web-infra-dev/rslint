package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type testPluginConfigPlan struct {
	owners  map[string]string
	configs map[string]*rslintconfig.MergedConfig
}

func (plan testPluginConfigPlan) OwnerForFile(filePath string) string {
	return plan.owners[filePath]
}

func (plan testPluginConfigPlan) ConfigForFile(filePath string) *rslintconfig.MergedConfig {
	return plan.configs[filePath]
}

// TestPluginConfigResolver_UsesGoOwnedCatalogKey proves the routing identity is
// the same normalized key Go published in its typed discovery catalog. Node
// treats that key as opaque when activating the matching plugin host.
func TestPluginConfigResolver_UsesGoOwnedCatalogKey(t *testing.T) {
	configDir := "C:/proj"
	merged := new(rslintconfig.MergedConfig)
	r := pluginConfigResolver{
		lintResolver: testPluginConfigPlan{
			owners:  map[string]string{configDir + "/src/a.ts": configDir},
			configs: map[string]*rslintconfig.MergedConfig{configDir + "/src/a.ts": merged},
		},
	}
	wireKey, merged := r.resolve(configDir + "/src/a.ts")
	if wireKey != configDir {
		t.Errorf("wire configKey = %q, want Go-owned catalog key %q", wireKey, configDir)
	}
	if merged == nil {
		t.Fatal("expected a merged config for the matched file")
	}

	// With no low-level API routing override, the owner key is used directly.
	posix := pluginConfigResolver{
		lintResolver: testPluginConfigPlan{
			owners:  map[string]string{"/posix/proj/a.ts": "/posix/proj"},
			configs: map[string]*rslintconfig.MergedConfig{"/posix/proj/a.ts": merged},
		},
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
	diags := <-dispatchPluginLintAsync(context.Background(), failing, pluginInput(), false, "off", nil)
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
	if diags := <-dispatchPluginLintAsync(context.Background(), canceled, pluginInput(), false, "off", nil); len(diags) != 0 {
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
	if diags := <-dispatchPluginLintAsync(context.Background(), dispatch, nil, false, "off", nil); len(diags) != 0 {
		t.Errorf("no inputs should yield 0 diagnostics, got %d", len(diags))
	}
}

// TestPluginConfigResolver_Branches covers resolve()'s non-match and
// single-config (configMap==nil) fallback branches.
func TestPluginConfigResolver_Branches(t *testing.T) {
	// Multi-config, file under no config -> ("", nil).
	r := pluginConfigResolver{
		lintResolver: testPluginConfigPlan{},
	}
	if wk, m := r.resolve("/elsewhere/a.ts"); wk != "" || m != nil {
		t.Errorf("no-match -> (\"\",nil), got (%q, nil=%v)", wk, m == nil)
	}

	// Single-config (configMap==nil): wireKey is currentDirectory; merged from rslintConfig.
	single := pluginConfigResolver{
		lintResolver: testPluginConfigPlan{
			owners:  map[string]string{"/proj/a.ts": "/proj"},
			configs: map[string]*rslintconfig.MergedConfig{"/proj/a.ts": new(rslintconfig.MergedConfig)},
		},
	}
	if wk, m := single.resolve("/proj/a.ts"); wk != "/proj" || m == nil {
		t.Errorf("single-config -> (currentDirectory, merged), got (%q, nil=%v)", wk, m == nil)
	}
}
