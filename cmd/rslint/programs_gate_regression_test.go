package main

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The config resolver normally filters RequiresTypeInfo rules for gap files.
// RunLinter repeats that filter so a caller cannot accidentally execute one
// against a fallback Program. Non-type-aware rules continue in the established
// gap-file mode with no TypeChecker.
func TestGate_LinterFiltersTypeAwareRuleOnGapFile(t *testing.T) {
	tmpDir := tspath.NormalizePath(t.TempDir())
	gapFile := tspath.NormalizePath(filepath.Join(tmpDir, "gap.ts"))

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	fs = utils.NewOverlayVFS(fs, map[string]string{gapFile: "let a: any = 10;\na.b = 20;\n"})

	parseCache := utils.NewParseCache()
	fallbackResult, _ := createFallbackProgram([]string{gapFile}, false, tmpDir, fs, parseCache)
	fallback := fallbackResult.program
	if fallback == nil {
		t.Fatal("could not create fallback Program")
	}

	rslintconfig.RegisterAllRules()
	cfg := rslintconfig.RslintConfig{
		rslintconfig.ConfigEntry{
			Files:   []string{"**/*.ts"},
			Rules:   rslintconfig.Rules{"@typescript-eslint/no-unsafe-member-access": "error"},
			Plugins: []string{"@typescript-eslint"},
		},
	}
	// Deliberately bypass the config resolver's type-info gate.
	rules, _ := rslintconfig.GlobalRuleRegistry.GetEnabledRules(cfg, gapFile, tmpDir, false)
	if len(rules) != 1 || rules[0].Name != "@typescript-eslint/no-unsafe-member-access" || !rules[0].RequiresTypeInfo {
		t.Fatalf("fixture did not resolve the expected type-aware rule: %+v", rules)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		Programs:      []*compiler.Program{fallback},
		Scope:         linter.FileScope{Files: []string{gapFile}},
		TypeInfoFiles: map[string]struct{}{},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return rules
		},
		OnDiagnostic: func(rule.RuleDiagnostic) {},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if _, ran := result.ExecutedRules["@typescript-eslint/no-unsafe-member-access"]; ran {
		t.Fatal("type-aware rule bypassed the linter's gap-file eligibility filter")
	}
}
