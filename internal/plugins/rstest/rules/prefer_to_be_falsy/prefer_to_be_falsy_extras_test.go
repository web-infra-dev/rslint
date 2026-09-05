// TestPreferToBeFalsyExtras locks in branches and edge shapes that the
// upstream test suite doesn't exercise: the matcher-shape and literal
// contract, every Rstest expect source, the accessor and trivia matrix around
// both edits, and the shapes where the fix is withheld while the diagnostic
// stands. Each case carries an inline comment pointing at the upstream branch,
// Dimension 4 row, or tsgo AST quirk it covers.
package prefer_to_be_falsy

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

func TestPreferToBeFalsyExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToBeFalsyRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Matcher shape ----
			// A matcher that is never called asserts nothing, so rewriting it
			// would turn a no-op into a live assertion.
			{Code: `expect(value).toBe;`},
			{Code: `expect(value).not.toEqual;`},
			// Already the target shape.
			{Code: `expect(value).toBeFalsy();`},
			// The matcher name is not one of the equality matchers.
			{Code: `expect(value).toBeFalse();`},
			{Code: `expect(value).toContain(false);`},
			{Code: `expect(fn).toHaveBeenCalledWith(false);`},
			{Code: `expect(value).toMatchObject(false);`},
			// Chai's falsy assertion is a property rather than a matcher
			// call, and its equality forms are named `equal`, so neither is
			// the shape this rule rewrites.
			{Code: `expect(value).to.be.false;`},
			{Code: `expect(value).to.equal(false);`},
			{Code: `expect(value).to.deep.equal(false);`},

			// ---- B. Argument count ----
			// Upstream requires args.length === 1; an uncalled matcher yields
			// an empty args list, and so does `toBe()`.
			{Code: `expect(value).toBe();`},
			{Code: `expect(value).toEqual(false, false);`},
			{Code: `expect(value).toStrictEqual(false, message);`},
			// The argument check reads the matcher's own call, so a trailing
			// call on the assertion's result cannot smuggle a literal in.
			{Code: `expect(value).toBe()(false);`},

			// ---- B. The literal ----
			// The opposite literal belongs to the sibling rule.
			{Code: `expect(value).toBe(true);`},
			{Code: `expect(value).not.toStrictEqual(true);`},
			// Not a boolean literal at all. tsgo spells `false` as a keyword
			// node, so a falsy value of another kind never matches.
			{Code: `expect(value).toBe(0);`},
			{Code: `expect(value).toBe('false');`},
			{Code: "expect(value).toBe(`false`);"},
			{Code: `expect(value).toEqual({ ok: false });`},
			{Code: `expect(value).toBe(flag);`},
			{Code: `expect(value).toBe(Boolean(0));`},
			// An operator applied to a literal is not a literal; ESTree has no
			// Literal node here either.
			{Code: `expect(value).toBe(!1);`},
			{Code: `expect(value).toBe(!!flag);`},
			// Locks in upstream getFirstMatcherArg() arm 1: a SpreadElement is
			// returned unwrapped and is never a Literal.
			{Code: `expect(value).toBe(...flags);`},
			// ---- Dimension 4: TS type-expression wrappers on the literal ----
			// followTypeAssertionChain stops at `satisfies` and at the
			// non-null assertion, so upstream leaves both alone.
			{Code: `expect(value).toBe(false satisfies boolean);`},
			{Code: `expect(value).toBe(flag!);`},

			// ---- C. Reverse sources ----
			{Code: `import { expect } from 'vitest'; expect(value).toBe(false);`},
			{Code: `import { expect } from '@jest/globals'; expect(value).toBe(false);`},
			{Code: `import { expect } from '@playwright/test'; expect(value).toBe(false);`},
			{Code: `import { expect } from 'chai'; expect(value).toBe(false);`},
			{Code: `const expect = createAssertionLibrary(); expect(value).toBe(false);`},
			{Code: `custom.expect(value).toBe(false);`},

			// ---- C. Shadowing ----
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(value).toBe(false); }`},
			{Code: `import { expect as check } from '@rstest/core'; function f(check: any) { check(value).toEqual(false); }`},
			{Code: `import * as core from '@rstest/core'; function f() { const core = helper(); core.expect(value).toBe(false); }`},

			// ---- D. Broken chains ----
			// A second promise modifier makes the chain unparseable, and the
			// rule refuses to guess what the author meant.
			{Code: `expect(value).resolves.rejects.toBe(false);`},
			// No assertion factory ran, so the chain asserts nothing about any
			// value. Renaming the matcher would only make it look right.
			{Code: `expect.toBe(false);`},
			{Code: `expect.not.toEqual(false);`},

			// ---- Dimension 4: computed identifier keys ----
			// The key names the matcher only at runtime, so rewriting it would
			// rename a variable reference instead of the member.
			{Code: `expect(value)[toBe](false);`},
			{Code: `const toBe = 'toEqual';
expect(value)[toBe](false);`},

			// ---- Dimension 4: TS wrappers on the receiver are analysis boundaries ----
			{Code: `expect(value)!.toBe(false);`},
			{Code: `(expect(value) as any).toBe(false);`},
			{Code: `(expect(value) satisfies Assertion).toBe(false);`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `broken.expect?.(value).toBe(false);`},
			// N/A: declaration and container variants, function kinds, class
			// members, and overload signatures are unrelated to this rule's
			// single call-expression listener over analysis.ParseExpectCall.
			// N/A: numeric-literal and private-identifier accessor keys cannot
			// name a matcher; GetMemberEntries never stores them as one.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: empty arguments list on the assertion factory ----
			// `expect()` still ran the factory, so the chain parses and the
			// matcher is rewritten; whether the assertion has a value to check
			// is not this rule's question.
			{
				Code:   `expect().toBe(false);`,
				Output: []string{`expect().toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 10}},
			},

			// ---- Dimension 4: accessor forms; the written form survives both edits ----
			{
				Code:   `expect(value)['toBe'](false);`,
				Output: []string{`expect(value)['toBeFalsy']();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value)["toEqual"](false);`,
				Output: []string{`expect(value)["toBeFalsy"]();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value)[`toStrictEqual`](false);",
				Output: []string{"expect(value)[`toBeFalsy`]();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value)?.toBe(false);`,
				Output: []string{`expect(value)?.toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 16}},
			},
			{
				Code:   `expect(value).toBe?.(false);`,
				Output: []string{`expect(value).toBeFalsy?.();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   `(expect(value)).toBe(false);`,
				Output: []string{`(expect(value)).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 17}},
			},

			// ---- Dimension 4: parenthesized literal ----
			// ESTree has no ParenthesizedExpression, so upstream sees the
			// literal directly; the port has to skip the wrappers.
			{
				Code:   `expect(value).toBe((false));`,
				Output: []string{`expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).toEqual(((false)));`,
				Output: []string{`expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},

			// Locks in upstream getFirstMatcherArg() arm 2: followTypeAssertionChain
			// unwraps `as` and `<T>x`, and emptying the argument list takes the
			// whole wrapper with it.
			{
				Code:   `expect(value).toBe(false as boolean);`,
				Output: []string{`expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value).toStrictEqual(false as const as boolean);`,
				Output: []string{`expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},

			// ---- Trailing calls on the assertion's result ----
			// The literal belongs to the matcher's own argument list, not to
			// whichever call happens to enclose the assertion.
			{
				Code:   `expect(value).toBe(false)();`,
				Output: []string{`expect(value).toBeFalsy()();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value)['toEqual'](false)(2);`,
				Output: []string{`expect(value)['toBeFalsy']()(2);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},

			// ---- A parenthesized optional chain is still an assertion ----
			// ESTree wraps an optional chain in a ChainExpression, which
			// upstream's chain walk does not enter, so upstream misses these.
			{
				Code:   `(expect(value)?.toBe)(false);`,
				Output: []string{`(expect(value)?.toBeFalsy)();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 17}},
			},
			{
				Code:   `((expect(value)?.toEqual))(false);`,
				Output: []string{`((expect(value)?.toBeFalsy))();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 18}},
			},

			// ---- Dimension 3: trivia and the emptied argument list ----
			// Upstream removes only the argument node, leaving a trailing
			// comma and any whitespace behind; the whole span between the
			// parentheses goes instead, so no `(,)` is ever produced.
			{
				Code:   `expect(value).toBe(false,);`,
				Output: []string{`expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value).toEqual(\n  false,\n);",
				Output: []string{`expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value).\n  toBe(false);",
				Output: []string{"expect(value).\n  toBeFalsy();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(value). /* keep */ toBe(false);",
				Output: []string{"expect(value). /* keep */ toBeFalsy();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 27}},
			},
			{
				Code:   "expect(value).toBe /* keep */ (false);",
				Output: []string{"expect(value).toBeFalsy /* keep */ ();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value)[/* keep */ 'toEqual'](false);",
				Output: []string{"expect(value)[/* keep */ 'toBeFalsy']();"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 26}},
			},

			// ---- Dimension 3: a comment inside the parentheses withholds the fix ----
			// Emptying the argument list would delete the comment with it, so
			// the diagnostic stands alone and the rewrite is left to the author.
			{
				Code:   "expect(value).toBe(/* the flag */ false);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value).toEqual(false /* the flag */);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value).toBe(\n  // the flag\n  false,\n);",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 15}},
			},

			// ---- Assertion factories, the second expect argument, modifiers ----
			// Nothing outside the matcher is rewritten, so the factory, its
			// message argument and the modifier chain all survive.
			{
				Code:   `expect.soft(value).toBe(false);`,
				Output: []string{`expect.soft(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 20}},
			},
			{
				Code:   `await expect.poll(() => value).toBe(false);`,
				Output: []string{`await expect.poll(() => value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 32}},
			},
			{
				Code:   `expect(value, 'the flag is set').toEqual(false);`,
				Output: []string{`expect(value, 'the flag is set').toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 34}},
			},
			{
				Code:   `await expect(promise).resolves.toBe(false);`,
				Output: []string{`await expect(promise).resolves.toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 32}},
			},
			{
				Code:   `await expect(promise).rejects.not.toStrictEqual(false);`,
				Output: []string{`await expect(promise).rejects.not.toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 35}},
			},

			// ---- Rstest expect sources ----
			{
				Code: `import { expect } from '@rstest/core';
expect(value).toBe(false);`,
				Output: []string{`import { expect } from '@rstest/core';
expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 15}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(value).toEqual(false);`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 14}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(value).toBe(false);`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 15}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(value).toBe(false);`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 20}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(value).toBe(false);`,
				Output: []string{`const core = require('@rstest/core');
core.expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 20}},
			},
			{
				Code:   `import.meta.rstest.expect(value).toBe(false);`,
				Output: []string{`import.meta.rstest.expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 34}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(value).toBe(false);`,
				Output: []string{`const { expect } = import.meta.rstest;
expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 15}},
			},
			{
				Code: `const api = import.meta.rstest;
api.expect(value).toEqual(false);`,
				Output: []string{`const api = import.meta.rstest;
api.expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 19}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(value).toBe(false);`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(value).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 2, Column: 15}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(value).toBe(false));`,
				Output: []string{`test('x', ctx => ctx.expect(value).toBeFalsy());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 36}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(value).toEqual(false));`,
				Output: []string{`test('x', ({ expect }) => expect(value).toBeFalsy());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 41}},
			},
			{
				Code:   `test('x', { timeout: 1 }, ({ expect: check }) => check(value).toBe(false));`,
				Output: []string{`test('x', { timeout: 1 }, ({ expect: check }) => check(value).toBeFalsy());`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 1, Column: 63}},
			},

			// ---- Real-user: a predicate asserted inside a normal test ----
			{
				Code: `import { expect, test } from '@rstest/core';

test('rejects an incomplete configuration', () => {
  expect(isValidConfig(config)).toBe(false);
});`,
				Output: []string{`import { expect, test } from '@rstest/core';

test('rejects an incomplete configuration', () => {
  expect(isValidConfig(config)).toBeFalsy();
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 4, Column: 33}},
			},
			// ---- Real-user: a chain broken across lines inside test.for ----
			{
				Code: `test.for([1, 2])('case %i', (row, { expect }) => {
  expect(row.ok)
    .toStrictEqual(false);
});`,
				Output: []string{`test.for([1, 2])('case %i', (row, { expect }) => {
  expect(row.ok)
    .toBeFalsy();
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalsy", Line: 3, Column: 6}},
			},
		},
	)
}

// TestPreferToBeFalsyEditDemand locks in that the diagnostic is identical
// under every edit demand and that only the fixes come and go.
func TestPreferToBeFalsyEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(value).toBe(false);
expect(value).not['toEqual'](false);`,
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
			HasTypeInfo: false,
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{
					Name:     PreferToBeFalsyRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferToBeFalsyRule.Run(ctx, nil)
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
				t.Fatal("prefer-to-be-falsy unexpectedly materialized suggestions")
			}
		}
	}
}
