package require_top_level_describe_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/require_top_level_describe"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestRequireTopLevelDescribeSourceOnly checks that renamed imports and suite
// membership through named callbacks still resolve when no TypeChecker is
// available, which is the path a source-only Program takes.
func TestRequireTopLevelDescribeSourceOnly(t *testing.T) {
	code := `import { describe as suite, it as scenario, beforeEach as setup } from '@rstest/core';

function inlineBody() { scenario('inline body case', () => {}); }
suite('owned by a named callback', inlineBody);

{
  const blockBody = () => { setup(() => {}); };
  suite('owned from a block scope', blockBody);
}

let reassigned = () => { scenario('stale body case', () => {}); };
reassigned = () => {};
suite('owned by a reassigned binding', reassigned);

scenario('top level case', () => {});
setup(() => {});`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "require-top-level-describe-source-only.ts")
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
				Name:     require_top_level_describe.RequireTopLevelDescribeRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return require_top_level_describe.RequireTopLevelDescribeRule.Run(ctx, nil)
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

	wantLines := []int{11, 15, 16}
	wantMessages := []string{
		"All test cases must be wrapped in a describe block",
		"All test cases must be wrapped in a describe block",
		"All hooks must be wrapped in a describe block",
	}
	if len(diagnostics) != len(wantLines) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(diagnostics), len(wantLines), diagnostics)
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Description != wantMessages[index] {
			t.Errorf("diagnostic %d = %q, want %q", index, diagnostic.Message.Description, wantMessages[index])
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
