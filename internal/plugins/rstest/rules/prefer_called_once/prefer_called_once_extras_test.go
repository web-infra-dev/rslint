// TestPreferCalledOnceExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise: the matcher-shape and count-literal
// contract, every Rstest expect source, the accessor and trivia matrix around
// both edits, and the shapes where the fix is withheld while the diagnostic
// stands. Each case carries an inline comment pointing at the upstream branch,
// Dimension 4 row, or tsgo AST quirk it covers.
package prefer_called_once

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferCalledOnceExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferCalledOnceRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Matcher shape ----
			// Locks in upstream create() arm 1: parseVitestFnCall must resolve
			// a matcher at all. A matcher that is never called asserts
			// nothing, so rewriting it would turn a no-op into a live
			// assertion.
			{Code: `expect(fn).toHaveBeenCalledTimes;`},
			{Code: `expect(fn).not.toBeCalledTimes;`},
			// Already the target shape.
			{Code: `expect(fn).toHaveBeenCalledOnce();`},
			// Locks in upstream create() arm 2: the matcher name is neither
			// count matcher.
			{Code: `expect(fn).toHaveBeenCalled();`},
			{Code: `expect(fn).toHaveReturnedTimes(1);`},
			{Code: `expect(fn).toHaveBeenCalledExactlyOnceWith(1);`},
			// Chai's call-count assertions. `callCount(1)` is spelled
			// `calledOnce` rather than `toHaveBeenCalledOnce()`, so rewriting
			// it is a different edit than this rule makes.
			{Code: `expect(fn).callCount(1);`},
			{Code: `expect(fn).to.have.callCount(1);`},
			{Code: `expect(fn).calledOnce;`},

			// ---- B. Argument count ----
			// Locks in upstream create() arm 3: args.length === 1.
			{Code: `expect(fn).toHaveBeenCalledTimes();`},
			{Code: `expect(fn).toHaveBeenCalledTimes(1, 1);`},
			{Code: `expect(fn).toBeCalledTimes(1, extra);`},
			// The argument check reads the matcher's own call, so a trailing
			// call on the assertion's result cannot smuggle a count in.
			{Code: `expect(fn).toHaveBeenCalledTimes()(1);`},

			// ---- B. The count literal ----
			// Locks in upstream isOneLiteral(): a Literal whose value is not 1.
			{Code: `expect(fn).toHaveBeenCalledTimes(2);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(-1);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(1.5);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(10);`},
			{Code: `expect(fn).toHaveBeenCalledTimes('1');`},
			{Code: "expect(fn).toHaveBeenCalledTimes(`1`);"},
			{Code: `expect(fn).toHaveBeenCalledTimes(true);`},
			// Locks in upstream isOneLiteral(): not a Literal at all.
			{Code: `expect(fn).toHaveBeenCalledTimes(count);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(+1);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(2 - 1);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(expect.anything());`},
			// A BigInt is a distinct literal kind carrying a BigInt value, so
			// ESTree's `value === 1` is false for it too.
			{Code: `expect(fn).toHaveBeenCalledTimes(1n);`},
			// Locks in upstream getFirstMatcherArg() arm 1: a SpreadElement is
			// returned unwrapped and is never a Literal.
			{Code: `expect(fn).toHaveBeenCalledTimes(...counts);`},
			// ---- Dimension 4: TS type-expression wrappers on the count ----
			// followTypeAssertionChain stops at `satisfies` and at the
			// non-null assertion, so upstream keeps the count on both.
			{Code: `expect(fn).toHaveBeenCalledTimes(1 satisfies number);`},
			{Code: `expect(fn).toHaveBeenCalledTimes(one!);`},

			// ---- C. Reverse sources ----
			{Code: `import { expect } from 'vitest'; expect(fn).toHaveBeenCalledTimes(1);`},
			{Code: `import { expect } from '@jest/globals'; expect(fn).toHaveBeenCalledTimes(1);`},
			{Code: `import { expect } from '@playwright/test'; expect(fn).toHaveBeenCalledTimes(1);`},
			{Code: `import { expect } from 'chai'; expect(fn).toHaveBeenCalledTimes(1);`},
			{Code: `const expect = createAssertionLibrary(); expect(fn).toHaveBeenCalledTimes(1);`},
			{Code: `custom.expect(fn).toHaveBeenCalledTimes(1);`},

			// ---- C. Shadowing ----
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(fn).toHaveBeenCalledTimes(1); }`},
			{Code: `import { expect as check } from '@rstest/core'; function f(check: any) { check(fn).toHaveBeenCalledTimes(1); }`},
			{Code: `import * as core from '@rstest/core'; function f() { const core = helper(); core.expect(fn).toHaveBeenCalledTimes(1); }`},

			// ---- D. Broken chains ----
			// A second promise modifier makes the chain unparseable, and the
			// rule refuses to guess what the author meant.
			{Code: `expect(fn).resolves.rejects.toHaveBeenCalledTimes(1);`},
			// No assertion factory ran, so there is no value to count calls
			// on. The assertion is broken whichever matcher it names.
			{Code: `expect.toHaveBeenCalledTimes(1);`},
			{Code: `expect.not.toBeCalledTimes(1);`},

			// ---- Dimension 4: computed identifier keys ----
			// The key names the matcher only at runtime, so rewriting it would
			// rename a variable reference instead of the member.
			{Code: `expect(fn)[toHaveBeenCalledTimes](1);`},
			{Code: `const toHaveBeenCalledTimes = 'toHaveBeenCalledOnce';
expect(fn)[toHaveBeenCalledTimes](1);`},

			// ---- Dimension 4: TS wrappers on the receiver are analysis boundaries ----
			{Code: `expect(fn)!.toHaveBeenCalledTimes(1);`},
			{Code: `(expect(fn) as any).toHaveBeenCalledTimes(1);`},
			{Code: `(expect(fn) satisfies Assertion).toHaveBeenCalledTimes(1);`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `broken.expect?.(fn).toHaveBeenCalledTimes(1);`},
			// N/A: declaration and container variants, function kinds, class
			// members, and overload signatures are unrelated to this rule's
			// single call-expression listener over analysis.ParseExpectCall.
			// N/A: numeric-literal and private-identifier accessor keys cannot
			// name a matcher; GetMemberEntries never stores them as one.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: empty arguments list on the assertion factory ----
			// `expect()` still ran the factory, so the chain parses and the
			// matcher is rewritten; whether the assertion has anything to
			// count is not this rule's question.
			{
				Code:   `expect().toHaveBeenCalledTimes(1);`,
				Output: []string{`expect().toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 10}},
			},

			// ---- Dimension 4: accessor forms; the written form survives both edits ----
			{
				Code:   `expect(fn)['toBeCalledTimes'](1);`,
				Output: []string{`expect(fn)['toHaveBeenCalledOnce']();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn)["toHaveBeenCalledTimes"](1);`,
				Output: []string{`expect(fn)["toHaveBeenCalledOnce"]();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[`toBeCalledTimes`](1);",
				Output: []string{"expect(fn)[`toHaveBeenCalledOnce`]();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn)?.toHaveBeenCalledTimes(1);`,
				Output: []string{`expect(fn)?.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 13}},
			},
			{
				Code:   `expect(fn).toBeCalledTimes?.(1);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce?.();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `(expect(fn)).toHaveBeenCalledTimes(1);`,
				Output: []string{`(expect(fn)).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 14}},
			},

			// ---- Dimension 4: parenthesized count ----
			// ESTree has no ParenthesizedExpression, so upstream's isOneLiteral
			// sees the literal directly; the port has to skip the wrappers.
			{
				Code:   `expect(fn).toHaveBeenCalledTimes((1));`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(((1)));`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},

			// Locks in upstream getFirstMatcherArg() arm 2: followTypeAssertionChain
			// unwraps `as` and `<T>x`, and removing the argument takes the
			// whole assertion with it.
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1 as number);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toBeCalledTimes(1 as const as number);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},

			// ---- Dimension 4: numeric-literal spellings of one ----
			// tsgo normalizes a NumericLiteral's text at parse time, the same
			// value ESTree's `value === 1` compares.
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1.0);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(0x1);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1e0);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},

			// ---- Trailing calls on the assertion's result ----
			// The count belongs to the matcher's own argument list, not to
			// whichever call happens to enclose the assertion.
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1)();`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce()();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1)(2);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce()(2);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   `expect(fn)['toBeCalledTimes'](1)();`,
				Output: []string{`expect(fn)['toHaveBeenCalledOnce']()();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},

			// ---- A parenthesized optional chain is still an assertion ----
			// ESTree wraps an optional chain in a ChainExpression, which
			// upstream's chain walk does not enter, so upstream misses these.
			{
				Code:   `(expect(fn)?.toHaveBeenCalledTimes)(1);`,
				Output: []string{`(expect(fn)?.toHaveBeenCalledOnce)();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 14}},
			},
			{
				Code:   `((expect(fn)?.toBeCalledTimes))(1);`,
				Output: []string{`((expect(fn)?.toHaveBeenCalledOnce))();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 15}},
			},

			// ---- Dimension 3: trivia and the removed argument list ----
			// Upstream removes only the argument node, leaving a trailing
			// comma and any whitespace behind; the whole span between the
			// parentheses goes instead.
			{
				Code:   `expect(fn).toHaveBeenCalledTimes(1,);`,
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn).toHaveBeenCalledTimes(\n  1,\n);",
				Output: []string{`expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn).\n  toBeCalledTimes(1);",
				Output: []string{"expect(fn).\n  toHaveBeenCalledOnce();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(fn). /* keep */ toHaveBeenCalledTimes(1);",
				Output: []string{"expect(fn). /* keep */ toHaveBeenCalledOnce();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 24}},
			},
			{
				Code:   "expect(fn).toHaveBeenCalledTimes /* keep */ (1);",
				Output: []string{"expect(fn).toHaveBeenCalledOnce /* keep */ ();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn)[/* keep */ 'toBeCalledTimes'](1);",
				Output: []string{"expect(fn)[/* keep */ 'toHaveBeenCalledOnce']();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 23}},
			},

			// ---- Dimension 3: a comment inside the parentheses withholds the fix ----
			// Emptying the argument list would delete the comment with it, so
			// the diagnostic stands alone and the rewrite is left to the author.
			{
				Code:   "expect(fn).toHaveBeenCalledTimes(/* exactly one */ 1);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn).toBeCalledTimes(1 /* exactly one */);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},
			{
				Code:   "expect(fn).toHaveBeenCalledTimes(\n  // exactly one\n  1,\n);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 12}},
			},

			// ---- Assertion factories and the second expect argument ----
			// Nothing outside the matcher is rewritten, so the factory, its
			// message argument and the modifier chain all survive.
			{
				Code:   `expect.soft(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`expect.soft(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 17}},
			},
			{
				Code:   `expect(fn, 'spy was called once').toBeCalledTimes(1);`,
				Output: []string{`expect(fn, 'spy was called once').toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 35}},
			},
			{
				Code:   `await expect(fn).resolves.not.toHaveBeenCalledTimes(1);`,
				Output: []string{`await expect(fn).resolves.not.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 31}},
			},
			{
				Code:   `await expect(fn).rejects.toBeCalledTimes(1);`,
				Output: []string{`await expect(fn).rejects.toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 26}},
			},

			// ---- Rstest expect sources ----
			{
				Code: `import { expect } from '@rstest/core';
expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`import { expect } from '@rstest/core';
expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 12}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(fn).toBeCalledTimes(1);`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 11}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 12}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 17}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`const core = require('@rstest/core');
core.expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 17}},
			},
			{
				Code:   `import.meta.rstest.expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`import.meta.rstest.expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 31}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`const { expect } = import.meta.rstest;
expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 12}},
			},
			{
				Code: `const api = import.meta.rstest;
api.expect(fn).toBeCalledTimes(1);`,
				Output: []string{`const api = import.meta.rstest;
api.expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 16}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(fn).toHaveBeenCalledTimes(1);`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(fn).toHaveBeenCalledOnce();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 2, Column: 12}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(fn).toHaveBeenCalledTimes(1));`,
				Output: []string{`test('x', ctx => ctx.expect(fn).toHaveBeenCalledOnce());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 33}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(fn).toBeCalledTimes(1));`,
				Output: []string{`test('x', ({ expect }) => expect(fn).toHaveBeenCalledOnce());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 38}},
			},
			{
				Code:   `test('x', { timeout: 1 }, ({ expect: check }) => check(fn).toHaveBeenCalledTimes(1));`,
				Output: []string{`test('x', { timeout: 1 }, ({ expect: check }) => check(fn).toHaveBeenCalledOnce());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 1, Column: 60}},
			},

			// ---- Real-user: a spy asserted once inside a normal test ----
			{
				Code: `import { expect, test, rs } from '@rstest/core';

test('notifies once', () => {
  const listener = rs.fn();
  emitter.emit('ready');
  expect(listener).toHaveBeenCalledTimes(1);
});`,
				Output: []string{`import { expect, test, rs } from '@rstest/core';

test('notifies once', () => {
  const listener = rs.fn();
  emitter.emit('ready');
  expect(listener).toHaveBeenCalledOnce();
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 6, Column: 20}},
			},
			// ---- Real-user: a chain broken across lines inside test.for ----
			{
				Code: `test.for([1, 2])('case %i', (row, { expect }) => {
  expect(handler)
    .toBeCalledTimes(1);
});`,
				Output: []string{`test.for([1, 2])('case %i', (row, { expect }) => {
  expect(handler)
    .toHaveBeenCalledOnce();
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferCalledOnce", Line: 3, Column: 6}},
			},
		},
	)
}

// TestPreferCalledOnceEditDemand locks in that the diagnostic is identical
// under every edit demand and that only the fixes come and go.
func TestPreferCalledOnceEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(fn).toHaveBeenCalledTimes(1);
expect(fn).not['toBeCalledTimes'](1);`,
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
			Program:     lintprogram.NewFromCompiler(program),
			File:        sourceFile.FileName(),
			HasTypeInfo: true,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     PreferCalledOnceRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferCalledOnceRule.Run(ctx, nil)
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
				t.Fatal("prefer-called-once unexpectedly materialized suggestions")
			}
		}
	}

	for index := range allEdits {
		if autofixOnly[index].FixesPtr == nil ||
			!reflect.DeepEqual(autofixOnly[index].FixesPtr, allEdits[index].FixesPtr) {
			t.Fatalf("assertion %d produced inconsistent fixes between autofix and all demands", index)
		}
		// The matcher rename and the emptied argument list are always one fix
		// each.
		if got := len(*autofixOnly[index].FixesPtr); got != 2 {
			t.Fatalf("assertion %d produced %d fixes, want 2", index, got)
		}
	}
}
