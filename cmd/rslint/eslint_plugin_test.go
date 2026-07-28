package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// TestPluginConfigResolver_UsesGoOwnedCatalogKey proves the routing identity is
// the same normalized key Go published in its typed discovery catalog. Node
// treats that key as opaque when activating the matching plugin host.
func TestPluginConfigResolver_UsesGoOwnedCatalogKey(t *testing.T) {
	configDir := tspath.NormalizePath(`C:\proj`)
	configMap := map[string]rslintconfig.RslintConfig{
		configDir: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
	}
	r := pluginConfigResolver{
		lintResolver: newLintConfigResolver(lintConfigResolverOptions{ConfigMap: configMap}),
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
		lintResolver: newLintConfigResolver(lintConfigResolverOptions{
			ConfigMap: map[string]rslintconfig.RslintConfig{"/posix/proj": configMap[configDir]},
		}),
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
	configMap := map[string]rslintconfig.RslintConfig{
		"/proj": {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
	}

	// Multi-config, file under no config -> ("", nil).
	r := pluginConfigResolver{
		lintResolver: newLintConfigResolver(lintConfigResolverOptions{ConfigMap: configMap}),
	}
	if wk, m := r.resolve("/elsewhere/a.ts"); wk != "" || m != nil {
		t.Errorf("no-match -> (\"\",nil), got (%q, nil=%v)", wk, m == nil)
	}

	// Single-config (configMap==nil): wireKey is currentDirectory; merged from rslintConfig.
	single := pluginConfigResolver{
		lintResolver: newLintConfigResolver(lintConfigResolverOptions{
			Config:           configMap["/proj"],
			CurrentDirectory: "/proj",
		}),
	}
	if wk, m := single.resolve("/proj/a.ts"); wk != "/proj" || m == nil {
		t.Errorf("single-config -> (currentDirectory, merged), got (%q, nil=%v)", wk, m == nil)
	}
}

func TestLintConfigResolver_NearestConfigAndTypeInfoGate(t *testing.T) {
	rslintconfig.RegisterAllRules()

	configMap := map[string]rslintconfig.RslintConfig{
		"/repo": {{
			Files: []string{"**/*.ts"},
			Rules: rslintconfig.Rules{"no-console": "error"},
		}},
		"/repo/packages/app": {{
			Files: []string{"**/*.ts"},
			Rules: rslintconfig.Rules{
				"@typescript-eslint/require-await": "error",
				"no-debugger":                      "error",
			},
		}},
	}
	typeInfoFiles := map[string]struct{}{
		"/repo/packages/app/src/typed.ts": {},
	}
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap:     configMap,
		TypeInfoFiles: typeInfoFiles,
	})

	gapRules := configuredRuleNameSet(resolver.ActiveRulesForFile("/repo/packages/app/src/gap.ts"))
	if gapRules["@typescript-eslint/require-await"] {
		t.Fatal("expected type-aware app rule to be filtered for file without type info")
	}
	if !gapRules["no-debugger"] {
		t.Fatal("expected nearest app config to enable no-debugger")
	}
	if gapRules["no-console"] {
		t.Fatal("did not expect parent config rule for file owned by nearest app config")
	}

	typedRules := configuredRuleNameSet(resolver.ActiveRulesForFile("/repo/packages/app/src/typed.ts"))
	if !typedRules["@typescript-eslint/require-await"] || !typedRules["no-debugger"] {
		t.Fatalf("expected typed app file to keep both app rules, got %v", typedRules)
	}

	rootRules := configuredRuleNameSet(resolver.ActiveRulesForFile("/repo/root.ts"))
	if !rootRules["no-console"] || rootRules["no-debugger"] {
		t.Fatalf("expected root file to use root config only, got %v", rootRules)
	}

	if rules := resolver.ActiveRulesForFile("/outside/a.ts"); len(rules) != 0 {
		t.Fatalf("expected file outside every config to have no rules, got %v", rules)
	}
}

func TestLintConfigResolver_UsesBoundOwnerForAliasedSource(t *testing.T) {
	rslintconfig.RegisterAllRules()

	configMap := map[string]rslintconfig.RslintConfig{
		"/repo": {{
			Files: []string{"packages/app/*.ts"},
			Rules: rslintconfig.Rules{"no-console": "error"},
		}},
		"/repo/packages/app": {{
			Rules: rslintconfig.Rules{"no-debugger": "error"},
		}},
	}
	sourcePath := "/repo/packages/app/a.ts"
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap:                  configMap,
		OwnerConfigDirBySourcePath: map[string]string{sourcePath: "/repo"},
	})

	rules := configuredRuleNameSet(resolver.ActiveRulesForFile(sourcePath))
	if !rules["no-console"] || rules["no-debugger"] {
		t.Fatalf("expected the binding's root owner to win over source-path inference, got %v", rules)
	}
}

func TestLintConfigResolver_UsesConfigPathAliasForRulesAndGlobals(t *testing.T) {
	rslintconfig.RegisterAllRules()

	cfg := rslintconfig.RslintConfig{{
		Files: []string{"src/**/*.ts"},
		LanguageOptions: &rslintconfig.LanguageOptions{Raw: map[string]any{
			"globals": map[string]any{
				"aliasedGlobal": "readonly",
			},
		}},
		Rules: rslintconfig.Rules{"no-console": "error"},
	}}
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		Config:                 cfg,
		CurrentDirectory:       "/repo",
		TypeInfoFiles:          map[string]struct{}{"/outside/real-a.ts": {}},
		ConfigPathBySourcePath: map[string]string{"/outside/real-a.ts": "/repo/src/a.ts"},
	})

	rules := resolver.ActiveRulesForFile("/outside/real-a.ts")
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("expected aliased source path to use config path rules, got %v", configuredRuleNameSet(rules))
	}
	if !rules[0].Globals["aliasedGlobal"] {
		t.Fatalf("expected aliased source path to carry globals from config path")
	}
	if resolver.ConfigForFile("/outside/real-a.ts") == nil {
		t.Fatalf("expected aliased source path to resolve merged config")
	}
}

type caseInsensitiveResolverFS struct {
	vfs.FS
}

func (f *caseInsensitiveResolverFS) UseCaseSensitiveFileNames() bool { return false }
func (f *caseInsensitiveResolverFS) Realpath(filePath string) string {
	return strings.ToLower(tspath.NormalizePath(filePath))
}

func TestLintConfigResolver_SourceMappingsUseCanonicalFilesystemIdentity(t *testing.T) {
	rslintconfig.RegisterAllRules()
	fsys := &caseInsensitiveResolverFS{FS: osvfs.FS()}
	sourcePath := "c:/repo/src/a.ts"
	resolver := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap: map[string]rslintconfig.RslintConfig{
			"C:/Repo": {{
				Files: []string{"src/**/*.ts"},
				Rules: rslintconfig.Rules{"no-console": "error"},
			}},
		},
		ConfigPathBySourcePath: map[string]string{"C:/REPO/SRC/A.ts": "c:/repo/src/a.ts"},
		OwnerConfigDirBySourcePath: map[string]string{
			"C:/REPO/SRC/A.ts": "c:/repo",
		},
		FS: fsys,
	})

	rules := resolver.ActiveRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("expected case-equivalent source mapping to retain config rules, got %v", configuredRuleNameSet(rules))
	}
}

func configuredRuleNameSet(rules []linter.ConfiguredRule) map[string]bool {
	names := make(map[string]bool, len(rules))
	for _, r := range rules {
		names[r.Name] = true
	}
	return names
}
