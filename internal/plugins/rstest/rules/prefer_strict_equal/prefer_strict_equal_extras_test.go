// TestPreferStrictEqualExtras locks in branches and edge shapes that the
// upstream suite does not exercise. Each group names the Rstest source,
// matcher-chain boundary, Dimension 4 row or real-user shape it covers.
package prefer_strict_equal

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func strictEqualError(line int, column int, output string) rule_tester.InvalidTestCaseError {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "useToStrictEqual",
		Line:      line,
		Column:    column,
	}
	if output != "" {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "suggestReplaceWithStrictEqual",
			Output:    output,
		}}
	}
	return diagnostic
}

func TestPreferStrictEqualExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferStrictEqualRule,
		[]rule_tester.ValidTestCase{
			// Locks in upstream create() arm 1: only a resolved Rstest expect call is considered.
			{Code: `helper(value).toEqual(expected);`},
			{Code: `import { expect } from 'vitest'; expect(value).toEqual(expected);`},
			{Code: `import { expect } from '@jest/globals'; expect(value).toEqual(expected);`},
			{Code: `const expect = createAssertionLibrary(); expect(value).toEqual(expected);`},
			{Code: `custom.expect(value).toEqual(expected);`},

			// Locks in upstream create() arm 2: only the first matcher is inspected.
			{Code: `expect(value).to.equal(expected);`},
			{Code: `expect(value).to.equal(expected).and.toEqual(other);`},
			{Code: `expect(value).ok.and.toEqual(other);`},

			// Call-style matching excludes incomplete, static and dynamic chains.
			{Code: `expect(value).toEqual;`},
			{Code: `expect.toEqual(expected);`},
			{Code: `expect.not.toEqual(expected);`},
			{Code: `const matcher = 'toEqual'; expect(value)[matcher](expected);`},
			{Code: `expect(value)[matchers.toEqual](expected);`},
			{Code: `expect(value)[getMatcher()](expected);`},
			{Code: `expect(value)[0](expected);`},

			// ---- Dimension 4: TS expression wrappers are parser boundaries ----
			{Code: `expect(value)!.toEqual(expected);`},
			{Code: `(expect(value) as any).toEqual(expected);`},
			{Code: `(expect(value) satisfies Assertion).toEqual(expected);`},

			// ---- Dimension 4: dynamic element-access keys do not match ----
			// Covered above by identifier, property-access, call and numeric keys.
			// N/A: private identifiers cannot be used as access expressions here.

			// ---- Dimension 4: graceful degradation ----
			{Code: `broken.expect?.(value).toEqual(expected);`},
			// N/A: declaration/container forms and function kinds are outside this
			// rule's single call-expression listener.
			// N/A: spread/rest, empty declaration bodies and overload signatures do
			// not alter an expect call's matcher chain.
		},
		[]rule_tester.InvalidTestCase{
			// Locks in upstream create() arm 2: a first toEqual matcher reports.
			{
				Code:   `expect(value).toEqual(expected).and.equal(other);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 15, `expect(value).toStrictEqual(expected).and.equal(other);`)},
			},

			// ---- Dimension 4: receiver parentheses and optional chains ----
			{
				Code:   `(expect(value)).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 17, `(expect(value)).toStrictEqual(expected);`)},
			},
			{
				Code:   `((expect(value))).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 19, `((expect(value))).toStrictEqual(expected);`)},
			},
			{
				Code:   `expect(value)?.toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 16, `expect(value)?.toStrictEqual(expected);`)},
			},
			{
				Code:   `expect(value).toEqual?.(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 15, `expect(value).toStrictEqual?.(expected);`)},
			},

			// ---- Dimension 4: static accessor forms ----
			{
				Code:   `expect(value)['toEqual'](expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 15, `expect(value)['toStrictEqual'](expected);`)},
			},
			{
				Code:   `expect(value)["toEqual"](expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 15, `expect(value)["toStrictEqual"](expected);`)},
			},
			{
				Code:   "expect(value)[`toEqual`](expected);",
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 15, "expect(value)[`toStrictEqual`](expected);")},
			},

			// ---- Accessor trivia: only the matcher value is replaced ----
			{
				Code:   "expect(value). /* keep */ toEqual(expected);",
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 27, "expect(value). /* keep */ toStrictEqual(expected);")},
			},
			{
				Code: "expect(value).\n  toEqual(expected);",
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Message: "Use `toStrictEqual()` instead",
					Line: 2, Column: 3, EndLine: 2, EndColumn: 10,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: "expect(value).\n  toStrictEqual(expected);"}},
				}},
			},

			// ---- Dimension 4: nesting stays scoped to the inner assertion ----
			{
				Code:   `consume(expect(value).toEqual(expected));`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 23, `consume(expect(value).toStrictEqual(expected));`)},
			},
			// ---- Dimension 4: empty argument lists degrade without special cases ----
			{
				Code:   `expect().toEqual();`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 10, `expect().toStrictEqual();`)},
			},

			// ---- Diagnostic ranges: UTF-16 columns and repeated assertions ----
			{
				Code:   `expect("😀").toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 14, `expect("😀").toStrictEqual(expected);`)},
			},
			{
				Code: `expect(first).toEqual(one);
expect(second).toEqual(two);`,
				Errors: []rule_tester.InvalidTestCaseError{
					strictEqualError(1, 15, `expect(first).toStrictEqual(one);
expect(second).toEqual(two);`),
					strictEqualError(2, 16, `expect(first).toEqual(one);
expect(second).toStrictEqual(two);`),
				},
			},

			// ---- Assertion factories, modifiers and root spelling ----
			{
				Code:   `expect.soft(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 20, `expect.soft(value).toStrictEqual(expected);`)},
			},
			{
				Code:   `await expect.poll(load).resolves.not.toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 38, `await expect.poll(load).resolves.not.toStrictEqual(expected);`)},
			},
			{
				// No Entry gate: keep the syntax contract for browser assertions.
				Code:   `expect.element(locator).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 25, `expect.element(locator).toStrictEqual(expected);`)},
			},

			// ---- Rstest expect source matrix ----
			{
				Code: `import { expect } from '@rstest/core';
expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 15, `import { expect } from '@rstest/core';
expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `import { expect as check } from '@rstest/core';
check(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 14, `import { expect as check } from '@rstest/core';
check(value).toStrictEqual(expected);`)},
			},
			{
				Code: `const { expect } = require('@rstest/core');
expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 15, `const { expect } = require('@rstest/core');
expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `const { expect: check } = require('@rstest/core');
check(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 14, `const { expect: check } = require('@rstest/core');
check(value).toStrictEqual(expected);`)},
			},
			{
				Code: `import * as core from '@rstest/core';
core.expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 20, `import * as core from '@rstest/core';
core.expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `import * as core from '@rstest/core';
core['expect'](value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 23, `import * as core from '@rstest/core';
core['expect'](value).toStrictEqual(expected);`)},
			},
			{
				Code: `const core = require('@rstest/core');
core.expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 20, `const core = require('@rstest/core');
core.expect(value).toStrictEqual(expected);`)},
			},
			{
				Code:   `import.meta.rstest.expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 34, `import.meta.rstest.expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `const { expect } = import.meta.rstest;
expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 15, `const { expect } = import.meta.rstest;
expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `const api = import.meta.rstest;
api.expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 19, `const api = import.meta.rstest;
api.expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `import { expect } from 'rstack/test';
expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 15, `import { expect } from 'rstack/test';
expect(value).toStrictEqual(expected);`)},
			},
			{
				Code: `import { expect } from '@rstest/playwright';
expect(value).toEqual(expected);`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(2, 15, `import { expect } from '@rstest/playwright';
expect(value).toStrictEqual(expected);`)},
			},
			{
				Code:   `test('compares values', ctx => ctx.expect(value).toEqual(expected));`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 50, `test('compares values', ctx => ctx.expect(value).toStrictEqual(expected));`)},
			},
			{
				Code:   `test('compares values', ({ expect }) => expect(value).toEqual(expected));`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 55, `test('compares values', ({ expect }) => expect(value).toStrictEqual(expected));`)},
			},

			// ---- Real-user: vitest-eslint#738 soft assertion ----
			{
				Code:   `test('compares softly', () => { expect.soft(actual).toEqual(expected); });`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 53, `test('compares softly', () => { expect.soft(actual).toStrictEqual(expected); });`)},
			},
			// ---- Real-user: vitest-eslint#738 poll assertion ----
			{
				Code:   `await expect.poll(readState).toEqual({ ready: true });`,
				Errors: []rule_tester.InvalidTestCaseError{strictEqualError(1, 30, `await expect.poll(readState).toStrictEqual({ ready: true });`)},
			},
		},
	)
}
