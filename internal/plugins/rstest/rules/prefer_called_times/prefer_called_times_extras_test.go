// TestPreferCalledTimesExtras locks in the Rstest-only augmentation required by
// the port spec: the matcher-shape contract, every Rstest expect source, the
// accessor and trivia matrix around both edits, and the shapes where the fix
// is withheld while the diagnostic stands.
package prefer_called_times

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

func TestPreferCalledTimesExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferCalledTimesRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Matcher shape ----
			// A matcher that is never called asserts nothing; rewriting it
			// would turn a no-op into a live assertion.
			{Code: `expect(fn).toHaveBeenCalledOnce;`},
			{Code: `expect(fn).not.toHaveBeenCalledOnce;`},
			// Already the target shape.
			{Code: `expect(fn).toHaveBeenCalledTimes(1);`},
			// `toHaveBeenCalledOnce` takes no arguments, so anything already
			// written between the parentheses is a different assertion.
			{Code: `expect(fn).toHaveBeenCalledOnce(1);`},
			{Code: `expect(fn).toHaveBeenCalledOnce(...args);`},
			// The argument check reads the matcher's own call, so a trailing
			// call on the assertion's result cannot hide it.
			{Code: `expect(fn).toHaveBeenCalledOnce(1)();`},
			// Chai's call-count property. Its `toHaveBeenCalledTimes(1)`
			// equivalent is `callCount(1)`, so this rule leaves it alone.
			{Code: `expect(fn).calledOnce;`},
			{Code: `expect(fn).to.have.been.calledOnce;`},

			// ---- A. Other matchers ----
			{Code: `expect(fn).toHaveBeenCalled();`},
			{Code: `expect(fn).toHaveReturnedOnce();`},
			{Code: `expect(fn).toHaveBeenCalledOnceWith(1);`},

			// ---- B. Matchers Rstest does not have ----
			// `toBeCalledOnce` is absent from @vitest/expect@4.1.10, so code
			// calling it is broken rather than improvable.
			{Code: `expect(fn).toBeCalledOnce();`},
			{Code: `expect(fn).rejects.not.toBeCalledOnce();`},

			// ---- C. Reverse sources ----
			{Code: `import { expect } from 'vitest'; expect(fn).toHaveBeenCalledOnce();`},
			{Code: `import { expect } from '@jest/globals'; expect(fn).toHaveBeenCalledOnce();`},
			{Code: `import { expect } from '@playwright/test'; expect(fn).toHaveBeenCalledOnce();`},
			{Code: `import { expect } from 'chai'; expect(fn).toHaveBeenCalledOnce();`},
			{Code: `const expect = createAssertionLibrary(); expect(fn).toHaveBeenCalledOnce();`},
			{Code: `custom.expect(fn).toHaveBeenCalledOnce();`},

			// ---- C. Shadowing ----
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(fn).toHaveBeenCalledOnce(); }`},
			{Code: `import { expect as check } from '@rstest/core'; function f(check: any) { check(fn).toHaveBeenCalledOnce(); }`},
			{Code: `import * as core from '@rstest/core'; function f() { const core = helper(); core.expect(fn).toHaveBeenCalledOnce(); }`},

			// ---- D. Broken chains ----
			// A second promise modifier makes the chain unparseable, and the
			// rule refuses to guess what the author meant.
			{Code: `expect(fn).resolves.rejects.toHaveBeenCalledOnce();`},
			// No assertion factory ran, so there is no value to count calls
			// on. The assertion is broken whichever matcher it names.
			{Code: `expect.toHaveBeenCalledOnce();`},
			{Code: `expect.not.toHaveBeenCalledOnce();`},

			// ---- D. Computed identifier keys ----
			// The key names the matcher only at runtime, so the alias tables
			// cannot be consulted and rewriting it would rename a variable.
			{Code: `expect(fn)[toHaveBeenCalledOnce]();`},
			{Code: `const toHaveBeenCalledOnce = 'toHaveBeenCalledTimes';
expect(fn)[toHaveBeenCalledOnce]();`},

			// ---- D. TS wrappers are analysis boundaries ----
			{Code: `expect(fn)!.toHaveBeenCalledOnce();`},
			{Code: `(expect(fn) as any).toHaveBeenCalledOnce();`},
			{Code: `(expect(fn) satisfies Assertion).toHaveBeenCalledOnce();`},

			// ---- D. Dimension 4 / graceful degradation ----
			{Code: `broken.expect?.(fn).toHaveBeenCalledOnce();`},
			// N/A: declaration and container variants, function kinds, class
			// members, and overload signatures are unrelated to this rule's
			// single call-expression listener over analysis.ParseExpectCall.
		},
		[]rule_tester.InvalidTestCase{
			// ---- E. Accessor forms; the written form survives both edits ----
			{
				Code:   `expect(fn)['toHaveBeenCalledOnce']();`,
				Output: []string{`expect(fn)['toHaveBeenCalledTimes'](1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn)["toHaveBeenCalledOnce"]();`,
				Output: []string{`expect(fn)["toHaveBeenCalledTimes"](1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[`toHaveBeenCalledOnce`]();",
				Output: []string{"expect(fn)[`toHaveBeenCalledTimes`](1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn)?.toHaveBeenCalledOnce();`,
				Output: []string{`expect(fn)?.toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 13}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledOnce?.();`,
				Output: []string{`expect(fn).toHaveBeenCalledTimes?.(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   `(expect(fn)).toHaveBeenCalledOnce();`,
				Output: []string{`(expect(fn)).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 14}},
			},

			// ---- E. Trailing calls on the assertion's result ----
			// The count belongs to the matcher's own argument list, not to
			// whichever call happens to enclose the assertion.
			{
				Code:   `expect(fn).toHaveBeenCalledOnce()();`,
				Output: []string{`expect(fn).toHaveBeenCalledTimes(1)();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledOnce()()();`,
				Output: []string{`expect(fn).toHaveBeenCalledTimes(1)()();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledOnce()(1);`,
				Output: []string{`expect(fn).toHaveBeenCalledTimes(1)(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn)['toHaveBeenCalledOnce']()();`,
				Output: []string{`expect(fn)['toHaveBeenCalledTimes'](1)();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},

			// ---- E. A parenthesized optional chain is still an assertion ----
			// ESTree wraps an optional chain in a ChainExpression, which
			// upstream's chain walk does not enter, so upstream misses these.
			// The matcher and its receiver are unambiguous here and the
			// rewrite is exact, so this port reports them.
			{
				Code:   `(expect(fn)?.toHaveBeenCalledOnce)();`,
				Output: []string{`(expect(fn)?.toHaveBeenCalledTimes)(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 14}},
			},
			{
				Code:   `((expect(fn)?.toHaveBeenCalledOnce))();`,
				Output: []string{`((expect(fn)?.toHaveBeenCalledTimes))(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 15}},
			},
			{
				Code:   `(expect(fn)?.toHaveBeenCalledOnce)?.();`,
				Output: []string{`(expect(fn)?.toHaveBeenCalledTimes)?.(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 14}},
			},

			// ---- E. Trivia between the matcher name and its parentheses ----
			// Upstream inserts at `matcher.range[1] + 1`, which lands inside
			// the comment or on the wrong character in every case below.
			{
				Code:   "expect(fn).\n  toHaveBeenCalledOnce();",
				Output: []string{"expect(fn).\n  toHaveBeenCalledTimes(1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(fn). /* keep */ toHaveBeenCalledOnce();",
				Output: []string{"expect(fn). /* keep */ toHaveBeenCalledTimes(1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 24}},
			},
			{
				Code:   "expect(fn).toHaveBeenCalledOnce /* keep */ ();",
				Output: []string{"expect(fn).toHaveBeenCalledTimes /* keep */ (1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn).toHaveBeenCalledOnce(/* keep */);",
				Output: []string{"expect(fn).toHaveBeenCalledTimes(1/* keep */);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn).toHaveBeenCalledOnce(\n);",
				Output: []string{"expect(fn).toHaveBeenCalledTimes(1\n);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[/* keep */ 'toHaveBeenCalledOnce']();",
				Output: []string{"expect(fn)[/* keep */ 'toHaveBeenCalledTimes'](1);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 23}},
			},

			// ---- F. Assertion factories and the second expect argument ----
			// Nothing outside the matcher is rewritten, so the factory, its
			// message argument and the modifier chain all survive.
			{
				Code:   `expect.soft(fn).toHaveBeenCalledOnce();`,
				Output: []string{`expect.soft(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 17}},
			},
			{
				Code:   `expect(fn, 'spy was called once').toHaveBeenCalledOnce();`,
				Output: []string{`expect(fn, 'spy was called once').toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 35}},
			},
			{
				Code:   `await expect(fn).resolves.not.toHaveBeenCalledOnce();`,
				Output: []string{`await expect(fn).resolves.not.toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 31}},
			},
			{
				Code:   `await expect(fn).rejects.toHaveBeenCalledOnce();`,
				Output: []string{`await expect(fn).rejects.toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 26}},
			},

			// ---- G. Rstest expect sources ----
			{
				Code: `import { expect } from '@rstest/core';
expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`import { expect } from '@rstest/core';
expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 12}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(fn).toHaveBeenCalledOnce();`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 11}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 12}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 17}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`const core = require('@rstest/core');
core.expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 17}},
			},
			{
				Code:   `import.meta.rstest.expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`import.meta.rstest.expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 31}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`const { expect } = import.meta.rstest;
expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 12}},
			},
			{
				Code: `const api = import.meta.rstest;
api.expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`const api = import.meta.rstest;
api.expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 16}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(fn).toHaveBeenCalledOnce();`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(fn).toHaveBeenCalledTimes(1);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 2, Column: 12}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(fn).toHaveBeenCalledOnce());`,
				Output: []string{`test('x', ctx => ctx.expect(fn).toHaveBeenCalledTimes(1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 33}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(fn).toHaveBeenCalledOnce());`,
				Output: []string{`test('x', ({ expect }) => expect(fn).toHaveBeenCalledTimes(1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 38}},
			},
			{
				Code:   `test('x', { timeout: 1 }, ({ expect: check }) => check(fn).toHaveBeenCalledOnce());`,
				Output: []string{`test('x', { timeout: 1 }, ({ expect: check }) => check(fn).toHaveBeenCalledTimes(1));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 1, Column: 60}},
			},

			// ---- H. Real-user shapes ----
			{
				Code: `import { expect, test, rs } from '@rstest/core';

test('notifies once', () => {
  const listener = rs.fn();
  emitter.emit('ready');
  expect(listener).toHaveBeenCalledOnce();
});`,
				Output: []string{`import { expect, test, rs } from '@rstest/core';

test('notifies once', () => {
  const listener = rs.fn();
  emitter.emit('ready');
  expect(listener).toHaveBeenCalledTimes(1);
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 6, Column: 20}},
			},
			{
				Code: `test.for([1, 2])('case %i', (row, { expect }) => {
  expect(handler)
    .toHaveBeenCalledOnce();
});`,
				Output: []string{`test.for([1, 2])('case %i', (row, { expect }) => {
  expect(handler)
    .toHaveBeenCalledTimes(1);
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledTimes", Line: 3, Column: 6}},
			},
		},
	)
}

// TestPreferCalledTimesEditDemand locks in that the diagnostic is identical
// under every edit demand and that only the fixes come and go.
func TestPreferCalledTimesEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(fn).toHaveBeenCalledOnce();
expect(fn).not['toHaveBeenCalledOnce']();`,
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
					Name:     PreferCalledTimesRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferCalledTimesRule.Run(ctx, nil)
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

	for _, diagnostic := range diagnosticsOnly {
		if diagnostic.FixesPtr != nil {
			t.Fatal("EditDemandNone unexpectedly materialized fixes")
		}
	}
	for _, diagnostic := range suggestionOnly {
		if diagnostic.FixesPtr != nil {
			t.Fatal("EditDemandSuggestion unexpectedly materialized fixes")
		}
	}
	for _, diagnostics := range [][]rule.RuleDiagnostic{diagnosticsOnly, autofixOnly, suggestionOnly, allEdits} {
		for _, diagnostic := range diagnostics {
			if diagnostic.Suggestions != nil {
				t.Fatal("prefer-called-times unexpectedly materialized suggestions")
			}
		}
	}

	for index := range allEdits {
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("assertion %d produced inconsistent fixes between autofix and all demands", index)
		}
		// The matcher rename and the inserted `1` are always one fix each.
		if got := len(*autofixOnly[index].FixesPtr); got != 2 {
			t.Fatalf("assertion %d produced %d fixes, want 2", index, got)
		}
	}
}
