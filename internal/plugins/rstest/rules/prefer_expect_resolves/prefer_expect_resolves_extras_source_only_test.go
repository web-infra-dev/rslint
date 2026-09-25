package prefer_expect_resolves_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	impl "github.com/web-infra-dev/rslint/internal/plugins/rstest/rules/prefer_expect_resolves"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestRstestResolvesSourceOnlyAndEditDemand(t *testing.T) {
	code := "import { expect as check, test } from '@rstest/core'; check(await pending).toBe(1); test.for([1])('row', async (row, ctx) => { ctx.expect(await pending).toBe(1); }); test('context', async ({ expect: local }) => { local(await pending).toBe(1); });"
	root := fixtures.GetRootDir()
	filename := tspath.ResolvePath(root.Dir, "resolves-source.ts")
	program, err := lintprogram.NewFromRoots(lintprogram.RootOptions{
		RootFileNames: []string{filename}, Host: utils.CreateCompilerHost(root.Dir, utils.NewOverlayVFS(root.FS, map[string]string{filename: code})),
		CompilerOptions: &core.CompilerOptions{Module: core.ModuleKindESNext}, SingleThreaded: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if program.CanProvideTypeChecker(program.SourceFiles()[0]) {
		t.Fatal("expected source-only program")
	}
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: program, File: filename,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: impl.PreferExpectResolvesRule.Name, Run: func(ctx rule.RuleContext) rule.RuleListeners { return impl.PreferExpectResolvesRule.Run(ctx, nil) }}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { diagnostics = append(diagnostics, d) }},
		})
		if len(diagnostics) != 3 {
			t.Fatalf("demand %d: got %d diagnostics, want 3", demand, len(diagnostics))
		}
		for _, d := range diagnostics {
			if d.FixesPtr != nil {
				t.Fatal("source-only assertion received automatic fixes")
			}
			if (d.Suggestions != nil) != (demand&rule.EditDemandSuggestion != 0) {
				t.Fatalf("unexpected suggestions for demand %d", demand)
			}
		}
	}
}

func TestRstestResolvesContexts(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &impl.PreferExpectResolvesRule, nil, []rule_tester.InvalidTestCase{
		{Code: "test.concurrent('case', async ({ expect: check }) => { check(await pending).toBe(1); });", Errors: []rule_tester.InvalidTestCaseError{{MessageId: "expectResolves", Line: 1, Column: 62, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestExpectResolves", Output: "test.concurrent('case', async ({ expect: check }) => { await check(pending).resolves.toBe(1); });"}}}}},
	})
}
