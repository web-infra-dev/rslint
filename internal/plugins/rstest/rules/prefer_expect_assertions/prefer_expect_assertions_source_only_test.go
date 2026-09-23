package prefer_expect_assertions_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_expect_assertions"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestPreferExpectAssertionsSourceOnly checks that renamed imports, TestContext
// expect, the options overload and named or TestContext hook callbacks still
// resolve when no TypeChecker is available, which is the path a source-only
// Program takes.
func TestPreferExpectAssertionsSourceOnly(t *testing.T) {
	code := `import { beforeEach, describe, expect as check, test as scenario } from '@rstest/core';

scenario('renamed', () => { check.hasAssertions(); });
scenario('context', (context) => { context.expect.assertions(1); });
scenario('options overload', { timeout: 100 }, ({ expect }) => { expect.hasAssertions(); });
describe('named hook', () => {
  beforeEach(requireAssertions);
  scenario('covered', () => {});
});
describe('context hook', () => {
  beforeEach(({ expect }) => expect.hasAssertions());
  scenario('covered', () => {});
});
function requireAssertions() { check.hasAssertions(); }

scenario('missing', () => { run(); });
scenario('options overload missing', { timeout: 100 }, () => { run(); });`
	root := fixtures.GetRootDir()
	fileName := tspath.ResolvePath(root.Dir, "prefer-expect-assertions-source-only.ts")
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
				Name:     prefer_expect_assertions.PreferExpectAssertionsRule.Name,
				Severity: rule.SeverityError,
				Run: func(ctx rule.RuleContext) rule.RuleListeners {
					return prefer_expect_assertions.PreferExpectAssertionsRule.Run(ctx, nil)
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

	wantLines := []int{16, 17}
	if len(diagnostics) != len(wantLines) {
		t.Fatalf("got %d diagnostics, want %d: %v", len(diagnostics), len(wantLines), diagnostics)
	}
	for index, diagnostic := range diagnostics {
		if diagnostic.Message.Id != "haveExpectAssertions" {
			t.Errorf("diagnostic %d = %q", index, diagnostic.Message.Id)
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
