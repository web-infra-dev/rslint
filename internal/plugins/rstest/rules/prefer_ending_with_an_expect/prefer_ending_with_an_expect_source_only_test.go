package prefer_ending_with_an_expect_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_ending_with_an_expect"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestPreferEndingWithAnExpectSourceOnly checks that renamed imports, the
// options overload and test-context assertions still resolve when no
// TypeChecker is available, which is the path a source-only Program takes.
func TestPreferEndingWithAnExpectSourceOnly(t *testing.T) {
	code := `import { test as scenario, expect as check } from '@rstest/core';

scenario('renamed assertion', () => { check(checkout()).toBe('ok'); });
scenario('context assertion', context => { context.expect(checkout()).toBe('ok'); });
scenario('options overload', { timeout: 100 }, () => { check(checkout()).toBe('ok'); });

scenario('renamed assertion missing', () => { checkout(); });
scenario('options overload missing', { timeout: 100 }, () => { checkout(); });
scenario('renamed assertion again', () => { check(checkout()).toBe('ok'); });`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "prefer-ending-with-an-expect-source-only.ts")
	fs := utils.NewOverlayVFS(root.FS, map[string]string{fileName: code})
	host := utils.CreateCompilerHost(root.Dir, fs)
	sourceProgram, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames:   []string{fileName},
		Host:            host,
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext},
		SingleThreaded:  true,
	})
	if err != nil {
		t.Fatalf("NewFromRoots: %v", err)
	}
	if sourceProgram.CanProvideTypeChecker(sourceProgram.SourceFiles()[0]) {
		t.Fatal("expected a source-only Program with no TypeChecker")
	}

	var diagnostics []rule.RuleDiagnostic
	lintPlan, err := linter.PrepareLintPlan(linter.PrepareLintPlanOptions{
		Programs:         []*lintprogram.Program{sourceProgram},
		TargetsByProgram: [][]string{{fileName}},
		SingleThreaded:   true,
		GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
			return []rule.ConfiguredRule{{
				Name:     prefer_ending_with_an_expect.PreferEndingWithAnExpectRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return prefer_ending_with_an_expect.PreferEndingWithAnExpectRule.Run(ctx, nil)
				},
			}}
		},
	})
	if err != nil {
		t.Fatalf("PrepareLintPlan: %v", err)
	}
	if _, err := linter.RunLinter(linter.RunLinterOptions{
		SingleThreaded: true,
		LintPlan:       lintPlan,
		Consumer: rule.DiagnosticConsumer{Report: func(diagnostic rule.RuleDiagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		}},
	}); err != nil {
		t.Fatalf("RunLinter: %v", err)
	}

	wantLines := []int{7, 8}
	if len(diagnostics) != len(wantLines) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(diagnostics), len(wantLines), diagnostics)
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Description != "Tests should end with an assertion" {
			t.Errorf("diagnostic %d = %q", index, diagnostic.Message.Description)
		}
		line, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(
			diagnostic.SourceFile,
			diagnostic.Range.Pos(),
		)
		if got := line + 1; got != wantLines[index] {
			t.Errorf("diagnostic %d starts on line %d, want %d", index, got, wantLines[index])
		}
	}
}
