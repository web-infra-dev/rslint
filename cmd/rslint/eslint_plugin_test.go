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

func TestLintConfigResolver_BoundOwnersAndTypeInfoGate(t *testing.T) {
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
	resolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
		ConfigMap:     configMap,
		TypeInfoFiles: typeInfoFiles,
		OwnerConfigDirBySourcePath: map[string]string{
			"/repo/packages/app/src/gap.ts":   "/repo/packages/app",
			"/repo/packages/app/src/typed.ts": "/repo/packages/app",
			"/repo/root.ts":                   "/repo",
		},
		MergedConfigBySourcePath: map[string]*rslintconfig.MergedConfig{
			"/repo/packages/app/src/gap.ts":   configMap["/repo/packages/app"].GetConfigForFile("/repo/packages/app/src/gap.ts", "/repo/packages/app"),
			"/repo/packages/app/src/typed.ts": configMap["/repo/packages/app"].GetConfigForFile("/repo/packages/app/src/typed.ts", "/repo/packages/app"),
			"/repo/root.ts":                   configMap["/repo"].GetConfigForFile("/repo/root.ts", "/repo"),
		},
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
	resolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
		ConfigMap:                  configMap,
		OwnerConfigDirBySourcePath: map[string]string{sourcePath: "/repo"},
		MergedConfigBySourcePath: map[string]*rslintconfig.MergedConfig{
			sourcePath: configMap["/repo"].GetConfigForFile(sourcePath, "/repo"),
		},
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
	resolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
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

func TestPluginConfigResolverReusesBoundMergedConfig(t *testing.T) {
	const sourcePath = "/repo/src/selected.ts"
	selected := &rslintconfig.MergedConfig{
		Settings: rslintconfig.Settings{"selection": "predicate"},
		LanguageOptions: &rslintconfig.LanguageOptions{Raw: map[string]any{
			"sourceType": "module",
		}},
	}
	lintResolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
		ConfigMap: map[string]rslintconfig.RslintConfig{
			"/repo": {{Settings: rslintconfig.Settings{"selection": "legacy-rematch"}}},
		},
		OwnerConfigDirBySourcePath: map[string]string{sourcePath: "/repo"},
		MergedConfigBySourcePath:   map[string]*rslintconfig.MergedConfig{sourcePath: selected},
	})

	wireKey, merged := (pluginConfigResolver{
		lintResolvers: []*lintConfigResolver{lintResolver},
		pluginConfigDirByOwner: map[string]string{
			"/repo": "wire-owner",
		},
	}).resolveForView(0, sourcePath)
	if wireKey != "wire-owner" {
		t.Fatalf("wire key = %q, want wire-owner", wireKey)
	}
	if merged != selected {
		t.Fatalf("plugin resolver rematched config: got=%+v want exact bound=%+v", merged, selected)
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
	selected := &rslintconfig.MergedConfig{Rules: map[string]*rslintconfig.RuleConfig{
		"no-console": {Level: "error"},
	}}
	resolver := mustNewLintConfigResolver(t, lintConfigResolverOptions{
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
		MergedConfigBySourcePath: map[string]*rslintconfig.MergedConfig{
			"C:/REPO/SRC/A.ts": selected,
		},
		FS: fsys,
	})

	rules := resolver.ActiveRulesForFile(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("expected case-equivalent source mapping to retain config rules, got %v", configuredRuleNameSet(rules))
	}
}

func TestLintConfigResolverModernBindingRejectsMissingExactMergedConfig(t *testing.T) {
	_, err := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap: map[string]rslintconfig.RslintConfig{
			"/repo": {{Rules: rslintconfig.Rules{"no-console": "error"}}},
		},
		OwnerConfigDirBySourcePath: map[string]string{"/repo/a.ts": "/repo"},
		RequiredSourcePaths:        []string{"/repo/a.ts"},
	})
	if err == nil || !strings.Contains(err.Error(), "no exact merged config") {
		t.Fatalf("error = %v, want modern binding invariant failure", err)
	}
}

func TestLintConfigResolverModernBindingRejectsNilExactMergedConfig(t *testing.T) {
	_, err := newLintConfigResolver(lintConfigResolverOptions{
		ConfigMap: map[string]rslintconfig.RslintConfig{
			"/repo": {{Rules: rslintconfig.Rules{"no-console": "error"}}},
		},
		OwnerConfigDirBySourcePath: map[string]string{"/repo/a.ts": "/repo"},
		MergedConfigBySourcePath:   map[string]*rslintconfig.MergedConfig{"/repo/a.ts": nil},
		RequiredSourcePaths:        []string{"/repo/a.ts"},
	})
	if err == nil || !strings.Contains(err.Error(), "no exact merged config") {
		t.Fatalf("error = %v, want nil modern binding invariant failure", err)
	}
}

func mustNewLintConfigResolver(t *testing.T, options lintConfigResolverOptions) *lintConfigResolver {
	t.Helper()
	resolver, err := newLintConfigResolver(options)
	if err != nil {
		t.Fatalf("newLintConfigResolver: %v", err)
	}
	return resolver
}

func configuredRuleNameSet(rules []linter.ConfiguredRule) map[string]bool {
	names := make(map[string]bool, len(rules))
	for _, r := range rules {
		names[r.Name] = true
	}
	return names
}
