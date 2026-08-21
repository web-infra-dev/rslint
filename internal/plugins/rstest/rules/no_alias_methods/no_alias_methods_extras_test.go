// TestNoAliasMethodsExtras locks in the Rstest-only augmentation required by
// the port spec: first-matcher contract, strict computed-identifier fallback,
// source matrices, trivia-preserving fixes, real-user shapes, and edit-demand
// parity.
package no_alias_methods

import (
	"reflect"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAliasMethodsExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoAliasMethodsRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Canonical names stay valid ----
			{Code: `expect(fn).toHaveBeenCalledWith(1)`},
			{Code: `expect(fn).toHaveBeenCalledTimes(2)`},
			{Code: `expect(fn).toHaveBeenNthCalledWith(1, 2)`},
			{Code: `expect(fn).toHaveBeenLastCalledWith(3)`},
			{Code: `expect(fn).toHaveReturned()`},
			{Code: `expect(fn).toHaveReturnedTimes(2)`},
			{Code: `expect(fn).toHaveReturnedWith(3)`},
			{Code: `expect(fn).toHaveLastReturnedWith(4)`},
			{Code: `expect(fn).toHaveNthReturnedWith(1, 5)`},
			{Code: `expect(fn).toThrow()`},
			{Code: `expect.soft(fn).toHaveBeenCalled()`},

			// ---- B. Locks in Vitest first-matcher contract ----
			{Code: `expect(spy).toHaveBeenCalled().and.toReturn()`},
			{Code: `expect(spy).equal(other).and.toBeCalled()`},
			{Code: `expect(spy).ok.and.toBeCalled()`},
			{Code: `expect(spy).a("function").and.toBeCalled()`},

			// ---- C. strict computed-identifier fallback fail-closed boundaries ----
			{Code: `expect(spy).to.have[toBeCalled]()`},
			{Code: `expect(spy).equal(other)[toBeCalled]()`},
			{Code: `expect(spy).ok[toBeCalled]()`},
			{Code: `expect(spy).unknown[toBeCalled]()`},
			{Code: `expect(spy)[toBeCalled]().and.toReturn()`},
			{Code: `expect[toBeCalled]()`},
			{Code: `expect(spy)[aliases.toBeCalled]()`},
			{Code: `expect(spy)[getMatcher()]()`},
			{Code: `expect(spy)[0]()`},

			// ---- D. Reverse sources: foreign/local expects stay out of scope ----
			{Code: `import { expect } from 'vitest'; expect(fn).toBeCalled()`},
			{Code: `import { expect } from '@jest/globals'; expect(fn).toBeCalled()`},
			{Code: `import { expect } from '@playwright/test'; expect(fn).toBeCalled()`},
			{Code: `const expect = createAssertionLibrary(); expect(fn).toBeCalled()`},
			{Code: `custom.expect(fn).toBeCalled()`},
			{Code: `const check = createAssertionLibrary(); check(fn).toBeCalled()`},

			// ---- D. Shadowing should prevent matches ----
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(fn).toBeCalled(); }`},
			{Code: `import { expect as check } from '@rstest/core'; function f(check: any) { check(fn).toBeCalled(); }`},
			{Code: `import * as core from '@rstest/core'; function f() { const core = helper(); core.expect(fn).toBeCalled(); }`},

			// ---- E. TS wrappers are analysis boundaries, not rule-owned unwrapping ----
			{Code: `expect(fn)!.toBeCalled()`},
			{Code: `(expect(fn) as any).toBeCalled()`},
			{Code: `(expect(fn) satisfies Assertion).toBeCalled()`},

			// ---- E. element boundary ----
			{Code: `expect.element(locator).toHaveCount(1)`},

			// ---- F. Dimension 4 / graceful degradation ----
			{Code: `broken.expect?.(fn).toBeCalled()`},

			// ---- F. Dimension 4 N/A markers ----
			// N/A: declaration/container variants, function kinds, class members,
			// spread/rest, and overload signatures are unrelated to this rule's
			// single call-expression listener plus analysis.ParseExpectCall.
		},
		[]rule_tester.InvalidTestCase{
			// ---- A. 11 aliases complete coverage ----
			{
				Code:   `expect(fn).toBeCalledWith(1)`,
				Output: []string{`expect(fn).toHaveBeenCalledWith(1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).lastCalledWith(1)`,
				Output: []string{`expect(fn).toHaveBeenLastCalledWith(1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).nthCalledWith(1, 2)`,
				Output: []string{`expect(fn).toHaveBeenNthCalledWith(1, 2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   `expect.soft(fn).toReturn()`,
				Output: []string{`expect.soft(fn).toHaveReturned()`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 17}},
			},
			{
				Code:   `expect.poll(fn).toReturnTimes(2)`,
				Output: []string{`expect.poll(fn).toHaveReturnedTimes(2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 17}},
			},
			{
				Code:   `expect(fn).toReturnWith(1)`,
				Output: []string{`expect(fn).toHaveReturnedWith(1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).lastReturnedWith(1)`,
				Output: []string{`expect(fn).toHaveLastReturnedWith(1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).nthReturnedWith(1, 2)`,
				Output: []string{`expect(fn).toHaveNthReturnedWith(1, 2)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).not.toThrowError()`,
				Output: []string{`expect(fn).not.toThrow()`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 16}},
			},
			{
				Code:   `expect(fn).resolves.toBeCalledTimes(1)`,
				Output: []string{`expect(fn).resolves.toHaveBeenCalledTimes(1)`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 21}},
			},
			{
				Code:   `expect(fn).rejects.toReturn()`,
				Output: []string{`expect(fn).rejects.toHaveReturned()`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 20}},
			},

			// ---- B. Locks in Vitest first-matcher contract ----
			{
				Code:   `expect(spy).toBeCalled().and.toReturn()`,
				Output: []string{`expect(spy).toHaveBeenCalled().and.toReturn()`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 13}},
			},
			{
				Code:   `expect(spy).to.have.been.toBeCalled()`,
				Output: []string{`expect(spy).to.have.been.toHaveBeenCalled()`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 26}},
			},

			// ---- C. strict computed-identifier fallback hit ----
			{
				Code: `const toBeCalled = "toBeCalled";
expect(spy)[toBeCalled]();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 13}},
			},
			{
				Code: `const toBeCalled = "toBeCalled";
expect(spy)[(toBeCalled)]();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 14}},
			},
			{
				Code: `const toThrowError = "toThrowError";
expect(spy).not[toThrowError]();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 17}},
			},
			{
				Code: `const toBeCalled = "toBeCalled";
expect.soft(spy)[toBeCalled]();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 18}},
			},

			// ---- D. 15 real Rstest expect sources ----
			{
				Code: `import { expect } from '@rstest/core';
expect(fn).toBeCalled();`,
				Output: []string{`import { expect } from '@rstest/core';
expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 12}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(fn).toBeCalled();`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 11}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(fn).toBeCalled();`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 12}},
			},
			{
				Code: `const { expect: check } = require('@rstest/core');
check(fn).toBeCalled();`,
				Output: []string{`const { expect: check } = require('@rstest/core');
check(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 11}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(fn).toBeCalled();`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 17}},
			},
			{
				Code: `import * as core from '@rstest/core';
core['expect'](fn).toBeCalled();`,
				Output: []string{`import * as core from '@rstest/core';
core['expect'](fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 20}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(fn).toBeCalled();`,
				Output: []string{`const core = require('@rstest/core');
core.expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 17}},
			},
			{
				Code:   `import.meta.rstest.expect(fn).toBeCalled();`,
				Output: []string{`import.meta.rstest.expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 31}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(fn).toBeCalled();`,
				Output: []string{`const { expect } = import.meta.rstest;
expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 12}},
			},
			{
				Code: `const api = import.meta.rstest;
api.expect(fn).toBeCalled();`,
				Output: []string{`const api = import.meta.rstest;
api.expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 16}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(fn).toBeCalled());`,
				Output: []string{`test('x', ctx => ctx.expect(fn).toHaveBeenCalled());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 33}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(fn).toBeCalled());`,
				Output: []string{`test('x', ({ expect }) => expect(fn).toHaveBeenCalled());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 38}},
			},
			{
				Code:   `test('x', { timeout: 1 }, ({ expect: check }) => check(fn).toBeCalled());`,
				Output: []string{`test('x', { timeout: 1 }, ({ expect: check }) => check(fn).toHaveBeenCalled());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 60}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(fn).toBeCalled();`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(fn).toHaveBeenCalled();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 12}},
			},

			// ---- E. fix/trivia/accessor matrix ----
			{
				Code:   "expect(fn).\n  toBeCalled()",
				Output: []string{"expect(fn).\n  toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(fn). /* keep */ toBeCalled()",
				Output: []string{"expect(fn). /* keep */ toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 24}},
			},
			{
				Code:   "expect(fn). // keep\n  toBeCalled()",
				Output: []string{"expect(fn). // keep\n  toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(fn)['toBeCalled']()",
				Output: []string{"expect(fn)['toHaveBeenCalled']()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[\"toBeCalled\"]()",
				Output: []string{"expect(fn)[\"toHaveBeenCalled\"]()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[`toBeCalled`]()",
				Output: []string{"expect(fn)[`toHaveBeenCalled`]()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[/* keep */ 'toBeCalled']()",
				Output: []string{"expect(fn)[/* keep */ 'toHaveBeenCalled']()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 23}},
			},
			{
				Code:   "expect(fn)[\n  'toBeCalled'\n]()",
				Output: []string{"expect(fn)[\n  'toHaveBeenCalled'\n]()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(fn)?.toBeCalled()",
				Output: []string{"expect(fn)?.toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 13}},
			},
			{
				Code:   "(expect(fn)).toBeCalled()",
				Output: []string{"(expect(fn)).toHaveBeenCalled()"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 14}},
			},

			// ---- F. alias should match only the expect chain it belongs to ----
			{
				Code:   `foo(expect(fn).toBeCalled())`,
				Output: []string{`foo(expect(fn).toHaveBeenCalled())`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 16}},
			},
			{
				Code:   `results[expect(fn).toBeCalled()]`,
				Output: []string{`results[expect(fn).toHaveBeenCalled()]`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 1, Column: 20}},
			},

			// ---- G. Real-user: Jest-style mock assertions in an Rstest helper ----
			{
				Code: `function assertSpy(spy: any) {
  expect(spy).toHaveReturned();
  expect(spy).toBeCalledTimes(1);
}`,
				Output: []string{`function assertSpy(spy: any) {
  expect(spy).toHaveReturned();
  expect(spy).toHaveBeenCalledTimes(1);
}`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 3, Column: 15}},
			},

			// ---- G. Real-user: renamed TestContext expect in a parameterized test ----
			{
				Code: `test.for([1])('case', (row, { expect: check }) => {
  check(row).toBeCalled();
});`,
				Output: []string{`test.for([1])('case', (row, { expect: check }) => {
  check(row).toHaveBeenCalled();
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "replaceAlias", Line: 2, Column: 14}},
			},
		},
	)
}

func TestNoAliasMethodsEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(fn).toBeCalled();
const toThrowError = "toThrowError";
expect(fn).not[toThrowError]();`,
		"edit-demand.ts",
		"tsconfig.json",
	)
	if err != nil {
		t.Fatal(err)
	}

	run := func(demand rule.EditDemand) []rule.RuleDiagnostic {
		t.Helper()

		var diagnostics []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program:      lintprogram.NewFromCompiler(program),
			File:         sourceFile.FileName(),
			HasTypeInfo:  true,
			ExcludePaths: []string{},
			GetRulesForFile: func(*ast.SourceFile) []linter.ConfiguredRule {
				return []linter.ConfiguredRule{{
					Name:     NoAliasMethodsRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return NoAliasMethodsRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{
				Demand: demand,
				Report: func(diagnostic rule.RuleDiagnostic) {
					diagnostics = append(diagnostics, diagnostic)
				},
			},
		})
		if len(diagnostics) != 2 {
			t.Fatalf("demand %d: diagnostics = %d, want 2", demand, len(diagnostics))
		}
		return diagnostics
	}

	diagnosticsOnly := run(rule.EditDemandNone)
	autofixOnly := run(rule.EditDemandAutofix)
	suggestionOnly := run(rule.EditDemandSuggestion)
	allEdits := run(rule.EditDemandAll)

	withoutEdits := func(diagnostic rule.RuleDiagnostic) rule.RuleDiagnostic {
		diagnostic.FixesPtr = nil
		diagnostic.Suggestions = nil
		return diagnostic
	}
	for index := range allEdits {
		want := withoutEdits(allEdits[index])
		for demand, diagnostics := range map[rule.EditDemand][]rule.RuleDiagnostic{
			rule.EditDemandNone:       diagnosticsOnly,
			rule.EditDemandAutofix:    autofixOnly,
			rule.EditDemandSuggestion: suggestionOnly,
		} {
			if got := withoutEdits(diagnostics[index]); !reflect.DeepEqual(got, want) {
				t.Errorf("demand %d diagnostic %d changed:\ngot  %#v\nwant %#v", demand, index, got, want)
			}
		}
	}

	if diagnosticsOnly[0].FixesPtr != nil || diagnosticsOnly[1].FixesPtr != nil {
		t.Fatal("EditDemandNone unexpectedly materialized fixes")
	}
	if suggestionOnly[0].FixesPtr != nil || suggestionOnly[1].FixesPtr != nil {
		t.Fatal("EditDemandSuggestion unexpectedly materialized fixes")
	}
	if diagnosticsOnly[0].Suggestions != nil || diagnosticsOnly[1].Suggestions != nil ||
		autofixOnly[0].Suggestions != nil || autofixOnly[1].Suggestions != nil ||
		suggestionOnly[0].Suggestions != nil || suggestionOnly[1].Suggestions != nil ||
		allEdits[0].Suggestions != nil || allEdits[1].Suggestions != nil {
		t.Fatal("no-alias-methods unexpectedly materialized suggestions")
	}

	if autofixOnly[0].FixesPtr == nil || !reflect.DeepEqual(autofixOnly[0].FixesPtr, allEdits[0].FixesPtr) {
		t.Fatal("fixable alias produced inconsistent fixes between autofix and all demands")
	}
	if len(*autofixOnly[0].FixesPtr) != 1 {
		t.Fatalf("fixable alias produced %d fixes, want 1", len(*autofixOnly[0].FixesPtr))
	}
	if autofixOnly[1].FixesPtr != nil || allEdits[1].FixesPtr != nil {
		t.Fatal("computed-identifier fallback must report without fixes under every demand")
	}
}
