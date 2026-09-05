package lint

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rules"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func newBaseResolver(options ResolverOptions) *Resolver {
	options.Catalog = rules.All()
	configs := options.ConfigsByOwner
	if configs == nil {
		configs = map[string]config.RslintConfig{
			options.ConfigDirectory: options.Config,
		}
	}
	options.PathSpaces = config.NewPathSpaceSnapshot(configs, options.FS)
	return NewResolver(options)
}

func TestResolverRequiresCatalog(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil rule catalog to panic")
		}
	}()
	NewResolver(ResolverOptions{})
}

func TestResolverRequiresPathSpaceSnapshot(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil path-space snapshot to panic")
		}
	}()
	NewResolver(ResolverOptions{Catalog: rules.All()})
}

func TestResolverUsesOnlyBoundTargetOwnership(t *testing.T) {
	configs := map[string]config.RslintConfig{
		"/repo": {{
			Files: []string{"**/*.ts"},
			Rules: config.Rules{"no-console": "error"},
		}},
		"/repo/packages/app": {{
			Files:   []string{"**/*.ts"},
			Plugins: []string{"@typescript-eslint"},
			Rules: config.Rules{
				"@typescript-eslint/require-await": "error",
				"no-debugger":                      "error",
			},
		}},
	}
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: configs,
		TargetsBySourcePath: map[string]target.File{
			"/repo/packages/app/src/gap.ts":   targetForTest("/repo/packages/app/src/gap.ts", "/repo/packages/app"),
			"/repo/packages/app/src/typed.ts": targetForTest("/repo/packages/app/src/typed.ts", "/repo/packages/app"),
			"/repo/root.ts":                   targetForTest("/repo/root.ts", "/repo"),
		},
	})

	gapRules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath("/repo/packages/app/src/gap.ts"))
	if !gapRules["@typescript-eslint/require-await"] || !gapRules["no-debugger"] || gapRules["no-console"] {
		t.Fatalf("app target resolved against the wrong owner: %v", gapRules)
	}
	typedRules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath("/repo/packages/app/src/typed.ts"))
	if !typedRules["@typescript-eslint/require-await"] || !typedRules["no-debugger"] {
		t.Fatalf("typed app target lost its owner rules: %v", typedRules)
	}
	rootRules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath("/repo/root.ts"))
	if !rootRules["no-console"] || rootRules["no-debugger"] {
		t.Fatalf("root target resolved against the wrong owner: %v", rootRules)
	}
	if rules := resolver.EnabledRulesForSourcePath("/outside/a.ts"); len(rules) != 0 {
		t.Fatalf("unbound multi-config source received rules: %v", rules)
	}
}

func TestResolverUsesBoundOwnerForAliasedSource(t *testing.T) {
	configs := map[string]config.RslintConfig{
		"/repo": {{
			Files: []string{"packages/app/*.ts"},
			Rules: config.Rules{"no-console": "error"},
		}},
		"/repo/packages/app": {{Rules: config.Rules{"no-debugger": "error"}}},
	}
	sourcePath := "/repo/packages/app/a.ts"
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: configs,
		TargetsBySourcePath: map[string]target.File{
			sourcePath: targetForTest(sourcePath, "/repo"),
		},
	})

	rules := configuredRuleNameSet(resolver.EnabledRulesForSourcePath(sourcePath))
	if !rules["no-console"] || rules["no-debugger"] {
		t.Fatalf("source path overrode the binding's owner: %v", rules)
	}
}

func TestResolverUsesBoundTargetForRulesAndGlobals(t *testing.T) {
	cfg := config.RslintConfig{{
		Files: []string{"src/**/*.ts"},
		LanguageOptions: &config.LanguageOptions{Raw: map[string]any{
			"globals": map[string]any{"aliasedGlobal": "readonly"},
		}},
		Rules: config.Rules{"no-console": "error"},
	}}
	resolver := newBaseResolver(ResolverOptions{
		Config:          cfg,
		ConfigDirectory: "/repo",
		TargetsBySourcePath: map[string]target.File{
			"/outside/real-a.ts": {
				PathIdentity: config.PathIdentity{
					Path:          "/repo/src/a.ts",
					CanonicalPath: "/outside/real-a.ts",
				},
				ConfigDirectory: "/repo",
			},
		},
	})

	rules := resolver.EnabledRulesForSourcePath("/outside/real-a.ts")
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("aliased source did not use its target config: %v", configuredRuleNameSet(rules))
	}
	if access := rules[0].Environment.Globals["aliasedGlobal"]; access != utils.GlobalAccessReadonly {
		t.Fatalf("aliased source lost target globals: %v", access)
	}
	_, resolved, ok := resolver.ResolveSourcePath("/outside/real-a.ts")
	if !ok || resolved.MergedConfig == nil {
		t.Fatal("aliased source did not resolve its merged config")
	}
}

type caseInsensitiveResolverFS struct {
	vfs.FS
}

func (fs *caseInsensitiveResolverFS) UseCaseSensitiveFileNames() bool { return false }
func (fs *caseInsensitiveResolverFS) Realpath(filePath string) string {
	return strings.ToLower(tspath.NormalizePath(filePath))
}

func TestResolverSourceMappingsUseCanonicalFilesystemIdentity(t *testing.T) {
	fsys := &caseInsensitiveResolverFS{FS: osvfs.FS()}
	sourcePath := "c:/repo/src/a.ts"
	resolver := newBaseResolver(ResolverOptions{
		ConfigsByOwner: map[string]config.RslintConfig{
			"C:/Repo": {{
				Files: []string{"src/**/*.ts"},
				Rules: config.Rules{"no-console": "error"},
			}},
		},
		TargetsBySourcePath: map[string]target.File{
			"C:/REPO/SRC/A.ts": targetForTest("c:/repo/src/a.ts", "c:/repo"),
		},
		FS: fsys,
	})

	rules := resolver.EnabledRulesForSourcePath(sourcePath)
	if len(rules) != 1 || rules[0].Name != "no-console" {
		t.Fatalf("case-equivalent source mapping lost config rules: %v", configuredRuleNameSet(rules))
	}
}

func TestResolverSingleConfigAcceptsUnboundSource(t *testing.T) {
	resolver := newBaseResolver(ResolverOptions{
		ConfigDirectory: "/repo",
		Config: config.RslintConfig{{
			Rules: config.Rules{"no-debugger": "error"},
		}},
	})
	owner, resolved, ok := resolver.ResolveSourcePath("/repo/a.ts")
	if !ok || owner != "/repo" || len(resolved.EnabledRules) != 1 {
		t.Fatalf("single config did not accept an unbound source: owner=%q resolved=%+v ok=%v", owner, resolved, ok)
	}
}

func targetForTest(path string, owner string) target.File {
	return target.File{
		PathIdentity:    config.PathIdentity{Path: path, CanonicalPath: path},
		ConfigDirectory: owner,
	}
}

func configuredRuleNameSet(configuredRules []rule.ConfiguredRule) map[string]bool {
	names := make(map[string]bool, len(configuredRules))
	for _, configuredRule := range configuredRules {
		names[configuredRule.Name] = true
	}
	return names
}
