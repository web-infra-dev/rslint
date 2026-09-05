// TestPreferStrictBooleanMatchersExtras locks in branches and edge shapes that
// the upstream test suite doesn't exercise: the matcher-shape contract, every
// Rstest expect source, the accessor and trivia matrix around both edits, and
// the shapes where the fix is withheld while the diagnostic stands. Each case
// carries an inline comment pointing at the upstream branch, Dimension 4 row,
// or tsgo AST quirk it covers.
package prefer_strict_boolean_matchers

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

func TestPreferStrictBooleanMatchersExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferStrictBooleanMatchersRule,
		[]rule_tester.ValidTestCase{
			// ---- A. Matcher shape ----
			// A matcher that is never called asserts nothing, so rewriting it
			// would turn a no-op into a live assertion.
			{Code: `expect(value).toBeTruthy;`},
			{Code: `expect(value).not.toBeFalsy;`},
			// Already the target shape.
			{Code: `expect(value).toBe(true);`},
			{Code: `expect(value).toBe(false);`},
			// Other truthiness-adjacent matchers are outside the pair this
			// rule rewrites; there is no strict literal that replaces them.
			{Code: `expect(value).toBeDefined();`},
			{Code: `expect(value).toBeNull();`},
			{Code: `expect(value).toBeTrue();`},
			{Code: `expect(value).toBeFalse();`},
			{Code: `expect(value).toBeUndefined();`},
			// Chai's truthiness assertions are properties rather than matcher
			// calls, and their strict form is `equal(true)`.
			{Code: `expect(value).to.be.ok;`},
			{Code: `expect(value).to.be.true;`},
			{Code: `expect(value).to.be.false;`},

			// ---- B. The matcher takes no arguments ----
			// Upstream inserts the literal without looking at what is already
			// between the parentheses, welding the two together into one
			// run-together identifier — still valid code, asserting something else.
			{Code: `expect(value).toBeTruthy(flag);`},
			{Code: `expect(value).toBeFalsy(0);`},
			{Code: `expect(value).toBeTruthy(...flags);`},
			// The argument check reads the matcher's own call, so a trailing
			// call on the assertion's result does not count as one.
			{Code: `expect(value).toBeTruthy(1)();`},

			// ---- C. Reverse sources ----
			{Code: `import { expect } from 'vitest'; expect(value).toBeTruthy();`},
			{Code: `import { expect } from '@jest/globals'; expect(value).toBeFalsy();`},
			{Code: `import { expect } from '@playwright/test'; expect(value).toBeTruthy();`},
			{Code: `import { expect } from 'chai'; expect(value).toBeTruthy();`},
			{Code: `const expect = createAssertionLibrary(); expect(value).toBeTruthy();`},
			{Code: `custom.expect(value).toBeFalsy();`},

			// ---- C. Shadowing ----
			{Code: `import { expect } from '@rstest/core'; function f(expect: any) { expect(value).toBeTruthy(); }`},
			{Code: `import { expect as check } from '@rstest/core'; function f(check: any) { check(value).toBeFalsy(); }`},
			{Code: `import * as core from '@rstest/core'; function f() { const core = helper(); core.expect(value).toBeTruthy(); }`},

			// ---- D. Broken chains ----
			// A second promise modifier makes the chain unparseable, and the
			// rule refuses to guess what the author meant.
			{Code: `expect(value).resolves.rejects.toBeTruthy();`},
			// No assertion factory ran, so the chain asserts nothing about any
			// value. Renaming the matcher would only make it look right.
			{Code: `expect.toBeTruthy();`},
			{Code: `expect.not.toBeFalsy();`},

			// ---- Dimension 4: computed identifier keys ----
			// The key names the matcher only at runtime, so rewriting it would
			// rename a variable reference instead of the member.
			{Code: `expect(value)[toBeTruthy]();`},
			{Code: `const toBeTruthy = 'toBeFalsy';
expect(value)[toBeTruthy]();`},

			// ---- Dimension 4: TS wrappers on the receiver are analysis boundaries ----
			{Code: `expect(value)!.toBeTruthy();`},
			{Code: `(expect(value) as any).toBeTruthy();`},
			{Code: `(expect(value) satisfies Assertion).toBeFalsy();`},

			// ---- Dimension 4: graceful degradation ----
			{Code: `broken.expect?.(value).toBeTruthy();`},
			// N/A: declaration and container variants, function kinds, class
			// members, and overload signatures are unrelated to this rule's
			// single call-expression listener over analysis.ParseExpectCall.
			// N/A: numeric-literal and private-identifier accessor keys cannot
			// name a matcher; GetMemberEntries never stores them as one.
			// N/A: this rule reads no matcher argument, so the literal-kind and
			// type-assertion rows of Dimension 4 have nothing to apply to.
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: empty arguments list on the assertion factory ----
			// `expect()` still ran the factory, so the chain parses and the
			// matcher is rewritten; whether the assertion has a value to check
			// is not this rule's question.
			{
				Code:   `expect().toBeTruthy();`,
				Output: []string{`expect().toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 10}},
			},

			// ---- Dimension 4: accessor forms; the written form survives both edits ----
			{
				Code:   `expect(value)['toBeTruthy']();`,
				Output: []string{`expect(value)['toBe'](true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value)["toBeFalsy"]();`,
				Output: []string{`expect(value)["toBe"](false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value)[`toBeTruthy`]();",
				Output: []string{"expect(value)[`toBe`](true);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value)?.toBeTruthy();`,
				Output: []string{`expect(value)?.toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 16}},
			},
			{
				Code:   `expect(value).toBeFalsy?.();`,
				Output: []string{`expect(value).toBe?.(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 15}},
			},
			{
				Code:   `(expect(value)).toBeTruthy();`,
				Output: []string{`(expect(value)).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 17}},
			},

			// ---- Trailing calls on the assertion's result ----
			// The literal goes into the matcher's own argument list, not into
			// whichever call happens to enclose the assertion.
			{
				Code:   `expect(value).toBeTruthy()();`,
				Output: []string{`expect(value).toBe(true)();`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 15}},
			},
			{
				Code:   `expect(value)['toBeFalsy']()(2);`,
				Output: []string{`expect(value)['toBe'](false)(2);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 15}},
			},

			// ---- A parenthesized optional chain is still an assertion ----
			// ESTree wraps an optional chain in a ChainExpression, which
			// upstream's chain walk does not enter, so upstream misses these.
			{
				Code:   `(expect(value)?.toBeTruthy)();`,
				Output: []string{`(expect(value)?.toBe)(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 17}},
			},
			{
				Code:   `((expect(value)?.toBeFalsy))();`,
				Output: []string{`((expect(value)?.toBe))(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 18}},
			},

			// ---- Dimension 3: trivia around the inserted literal ----
			// Upstream computes the insertion point as `matcher.range[1] + 1`,
			// assuming the `(` follows the matcher name; the parenthesis is
			// scanned for instead, so quoting, comments and line breaks between
			// the name and the parentheses all survive.
			{
				Code:   "expect(value).\n  toBeTruthy();",
				Output: []string{"expect(value).\n  toBe(true);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 2, Column: 3}},
			},
			{
				Code:   "expect(value). /* keep */ toBeTruthy();",
				Output: []string{"expect(value). /* keep */ toBe(true);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 27}},
			},
			{
				Code:   "expect(value).toBeFalsy /* keep */ ();",
				Output: []string{"expect(value).toBe /* keep */ (false);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value)[/* keep */ 'toBeTruthy']();",
				Output: []string{"expect(value)[/* keep */ 'toBe'](true);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 26}},
			},
			// A comment already between the parentheses is kept: the literal is
			// inserted before it rather than written over the whole span.
			{
				Code:   "expect(value).toBeTruthy(/* the flag */);",
				Output: []string{"expect(value).toBe(true/* the flag */);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 15}},
			},
			{
				Code:   "expect(value).toBeFalsy(\n);",
				Output: []string{"expect(value).toBe(false\n);"},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 15}},
			},

			// ---- Assertion factories, the second expect argument, modifiers ----
			// Nothing outside the matcher is rewritten, so the factory, its
			// message argument and the modifier chain all survive.
			{
				Code:   `expect.soft(value).toBeTruthy();`,
				Output: []string{`expect.soft(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 20}},
			},
			{
				Code:   `await expect.poll(() => value).toBeTruthy();`,
				Output: []string{`await expect.poll(() => value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 32}},
			},
			{
				Code:   `expect(value, 'the flag is set').toBeFalsy();`,
				Output: []string{`expect(value, 'the flag is set').toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 34}},
			},
			{
				Code:   `await expect(promise).resolves.toBeTruthy();`,
				Output: []string{`await expect(promise).resolves.toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 32}},
			},
			{
				Code:   `await expect(promise).rejects.not.toBeFalsy();`,
				Output: []string{`await expect(promise).rejects.not.toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 35}},
			},

			// ---- Rstest expect sources ----
			{
				Code: `import { expect } from '@rstest/core';
expect(value).toBeTruthy();`,
				Output: []string{`import { expect } from '@rstest/core';
expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 2, Column: 15}},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(value).toBeFalsy();`,
				Output: []string{`import { expect as check } from '@rstest/core';
check(value).toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 2, Column: 14}},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(value).toBeTruthy();`,
				Output: []string{`const { expect } = require('@rstest/core');
expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 2, Column: 15}},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(value).toBeTruthy();`,
				Output: []string{`import * as core from '@rstest/core';
core.expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 2, Column: 20}},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(value).toBeFalsy();`,
				Output: []string{`const core = require('@rstest/core');
core.expect(value).toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 2, Column: 20}},
			},
			{
				Code:   `import.meta.rstest.expect(value).toBeTruthy();`,
				Output: []string{`import.meta.rstest.expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 34}},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(value).toBeTruthy();`,
				Output: []string{`const { expect } = import.meta.rstest;
expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 2, Column: 15}},
			},
			{
				Code: `const api = import.meta.rstest;
api.expect(value).toBeFalsy();`,
				Output: []string{`const api = import.meta.rstest;
api.expect(value).toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 2, Column: 19}},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(value).toBeTruthy();`,
				Output: []string{`import { expect } from '@rstest/playwright';
expect(value).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 2, Column: 15}},
			},
			{
				Code:   `test('x', ctx => ctx.expect(value).toBeTruthy());`,
				Output: []string{`test('x', ctx => ctx.expect(value).toBe(true));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 36}},
			},
			{
				Code:   `test('x', ({ expect }) => expect(value).toBeFalsy());`,
				Output: []string{`test('x', ({ expect }) => expect(value).toBe(false));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 1, Column: 41}},
			},
			{
				Code:   `test('x', { timeout: 1 }, ({ expect: check }) => check(value).toBeTruthy());`,
				Output: []string{`test('x', { timeout: 1 }, ({ expect: check }) => check(value).toBe(true));`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 1, Column: 63}},
			},

			// ---- Real-user: a predicate asserted inside a normal test ----
			{
				Code: `import { expect, test } from '@rstest/core';

test('accepts a complete configuration', () => {
  expect(isValidConfig(config)).toBeTruthy();
});`,
				Output: []string{`import { expect, test } from '@rstest/core';

test('accepts a complete configuration', () => {
  expect(isValidConfig(config)).toBe(true);
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeTrue", Line: 4, Column: 33}},
			},
			// ---- Real-user: a chain broken across lines inside test.for ----
			{
				Code: `test.for([1, 2])('case %i', (row, { expect }) => {
  expect(row.ok)
    .toBeFalsy();
});`,
				Output: []string{`test.for([1, 2])('case %i', (row, { expect }) => {
  expect(row.ok)
    .toBe(false);
});`},
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "preferToBeFalse", Line: 3, Column: 6}},
			},
		},
	)
}

// TestPreferStrictBooleanMatchersEditDemand locks in that the diagnostic is
// identical under every edit demand and that only the fixes come and go.
func TestPreferStrictBooleanMatchersEditDemand(t *testing.T) {
	t.Parallel()

	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(
		`expect(value).toBeTruthy();
expect(value).not['toBeFalsy']();`,
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
					Name:     PreferStrictBooleanMatchersRule.Name,
					Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return PreferStrictBooleanMatchersRule.Run(ctx, nil)
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
				t.Fatal("prefer-strict-boolean-matchers unexpectedly materialized suggestions")
			}
		}
	}
}
