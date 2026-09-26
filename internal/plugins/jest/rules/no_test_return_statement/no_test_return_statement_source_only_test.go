package no_test_return_statement_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/no_test_return_statement"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestNoTestReturnStatementSourceOnly checks that renamed imports and named
// callbacks still resolve when no TypeChecker is available, which is the path
// a source-only Program takes. A shadowing parameter and a callback that is
// also called directly must stay unreported on that path too.
func TestNoTestReturnStatementSourceOnly(t *testing.T) {
	code := `import { test as scenario } from '@jest/globals';

scenario('inline', () => { return 1; });
scenario('named', named);
function named() { return 2; }
scenario('shared', shared);
function shared() { return 3; }
shared();
function register(named) { scenario('shadowed', named); }`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "no-test-return-statement-source-only.ts")
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
				Name:     no_test_return_statement.NoTestReturnStatementRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return no_test_return_statement.NoTestReturnStatementRule.Run(ctx, nil)
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

	wantLines := []int{3, 5}
	if len(diagnostics) != len(wantLines) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(diagnostics), len(wantLines), diagnostics)
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Description != "Jest tests should not return a value" {
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
