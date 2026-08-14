package main

import (
	"path/filepath"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/linter"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Program capability is the framework gate for RequiresTypeInfo rules. Config
// resolution can return the complete rule set without knowing how source
// services were assembled.
func TestGate_LinterFiltersTypeAwareRuleOnGapFile(t *testing.T) {
	tmpDir := tspath.NormalizePath(t.TempDir())
	gapFile := tspath.NormalizePath(filepath.Join(tmpDir, "gap.ts"))

	fs := bundled.WrapFS(cachedvfs.From(osvfs.FS()))
	fs = utils.NewOverlayVFS(fs, map[string]string{gapFile: "let a: any = 10;\na.b = 20;\n"})

	buildContext := utils.NewProgramBuildContext(fs)
	fallback, _ := createFallbackProgram([]string{gapFile}, false, tmpDir, buildContext)
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
	sourceOnly, err := lintprogram.NewFromBoundSources(fallback, fallback.SourceFiles())
	if err != nil {
		t.Fatalf("adapt fallback Program: %v", err)
	}
	result, err := linter.RunLinter(linter.RunLinterOptions{
		Programs: []*lintprogram.Program{sourceOnly},
		Scope:    linter.FileScope{Files: []string{gapFile}},
		GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
			return rules
		},
		Consumer: rule.DiagnosticConsumer{
			Report: func(rule.RuleDiagnostic) {},
		},
	})
	if err != nil {
		t.Fatalf("RunLinter: %v", err)
	}
	if _, ran := result.ExecutedRules["@typescript-eslint/no-unsafe-member-access"]; ran {
		t.Fatal("type-aware rule bypassed the linter's gap-file eligibility filter")
	}
}
