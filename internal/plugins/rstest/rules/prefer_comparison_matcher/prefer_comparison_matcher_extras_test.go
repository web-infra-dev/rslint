// Rstest semantics and edge shapes beyond prefer_comparison_matcher_upstream_test.go.
package prefer_comparison_matcher

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferComparisonMatcherExtras(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// Locks in subject, operator, matcher and boolean-argument gates.
		{Code: `expect(a > b).toBe();`}, {Code: `expect(a > b).toBe(flag);`}, {Code: `expect(a > b).toBe(...flags);`},
		{Code: `expect(a + b).toBe(true);`}, {Code: `expect(a > b).toContain(true);`},
		{Code: `expect(a > b).toBe(true satisfies boolean);`}, {Code: `expect(a > b).toBe(true!);`},
		// ---- Dimension 4: receiver and computed-name boundaries ----
		{Code: `expect(a > b).toBe;`}, {Code: `expect(a > b)[toBe](true);`}, {Code: `expect(a > b)[0](true);`},
		{Code: `expect(a > b).to.be.true;`}, {Code: `expect(a > b).equal(true);`},
		{Code: `expect(a > b)!.toBe(true);`}, {Code: `(expect(a > b) as any).toBe(true);`},
		{Code: `(expect(a > b) satisfies Assertion).toBe(true);`},
		{Code: `expect(a > b).not.not.toBe(true);`},
		// Relational operators always produce booleans, not promises or locators.
		{Code: `expect(a > b).resolves.toBe(true);`}, {Code: `expect(a > b).rejects.toBe(false);`},
		{Code: `expect.poll(() => a > b).toBe(true);`}, {Code: `expect.poll(a > b).toBe(true);`},
		{Code: `expect.element(a > b).toBe(true);`}, {Code: `expect.element(locator).toBe(true);`},
		// Proven non-numeric operands cannot use number/bigint matchers.
		{Code: `expect(null < 1).toBe(true);`}, {Code: `expect(true > 0).toBe(true);`},
		{Code: `expect([] < 1).toBe(true);`}, {Code: `expect({ valueOf: () => 2 } > 1).toBe(true);`},
		{Code: `expect(a > /x/).toBe(false);`}, {Code: `expect(a > void 0).toBe(false);`},
		{Code: `expect(("a" as string) > b).toBe(true);`}, {Code: "expect(a > (`${b}` satisfies string)).toBe(true);"},
		// Foreign imports and shadowed expect are not Rstest assertions.
		{Code: `import { expect } from 'vitest'; expect(a > b).toBe(true);`},
		{Code: `import { expect } from '@jest/globals'; expect(a > b).toBe(true);`},
		{Code: `import { expect } from 'chai'; expect(a > b).toBe(true);`},
		{Code: `const expect = makeExpect(); expect(a > b).toBe(true);`},
		{Code: `function check(expect: any) { expect(a > b).toBe(true); }`},
	}
	var invalid []rule_tester.InvalidTestCase
	for _, test := range []struct{ code, output, matcher string }{
		// ---- Dimension 4: parentheses, TS, optional chains, static accessor spellings ----
		{`expect(((a) > (b))).toBe(true);`, "", "toBeGreaterThan"},
		{`expect((f(), a) > b).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(a > (f(), b)).toBe(true);`, "", "toBeGreaterThan"},
		{`expect((2 as number) > (1 as number)).toBe((true as const));`, `expect(2 as number).toBeGreaterThan((1 as number));`, "toBeGreaterThan"},
		{`expect(2 > 1).toBe(<const>true);`, `expect(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`expect(2 > 1)?.toBe(false);`, `expect(2)?.not.toBeGreaterThan(1);`, "not.toBeGreaterThan"},
		{`expect(2 > 1)?.not.toBe(false);`, `expect(2)?.toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`expect(2 > 1).toBe?.(true);`, `expect(2).toBeGreaterThan?.(1);`, "toBeGreaterThan"},
		{`expect(2 > 1)[("toBe")](true);`, `expect(2)[("toBeGreaterThan")](1);`, "toBeGreaterThan"},
		{"expect(2 > 1)[`toBe`](false);", "expect(2).not[`toBeGreaterThan`](1);", "not.toBeGreaterThan"},
		{`expect(2 > 1).to.be.toEqual(false);`, `expect(2).to.be.not.toBeGreaterThan(1);`, "not.toBeGreaterThan"},
		// The original boolean remains the subject of later Chai assertions, so do not rewrite a multi-matcher chain.
		{`expect(a > b).toBe(true).and.toEqual(true);`, "", "toBeGreaterThan"},
		{`expect(a > b).toBe(true).toString();`, "", "toBeGreaterThan"},
		// ---- NaN, bigints and negative comparisons ----
		{`expect(NaN > 1).toBe(false);`, "", "not.toBeGreaterThan"},
		{`expect(2n <= 1n).not.toStrictEqual(false);`, `expect(2n).toBeLessThanOrEqual(1n);`, "toBeLessThanOrEqual"},
		{`expect(-2 >= +1).toEqual(true);`, `expect(-2).toBeGreaterThanOrEqual(+1);`, "toBeGreaterThanOrEqual"},
		// ---- Rstest factories, sources and TestContext ----
		{`expect.soft(2 > 1, 'limit').toBe(false);`, `expect.soft(2, 'limit').not.toBeGreaterThan(1);`, "not.toBeGreaterThan"},
		{`import { expect as check } from '@rstest/core'; check(2 > 1).toBe(true);`, `import { expect as check } from '@rstest/core'; check(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`import * as core from '@rstest/core'; core.expect(2 > 1).toBe(true);`, `import * as core from '@rstest/core'; core.expect(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`const { expect } = require('@rstest/core'); expect(2 > 1).toBe(true);`, `const { expect } = require('@rstest/core'); expect(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`import { expect } from '@rstest/playwright'; expect(2 > 1).toBe(true);`, `import { expect } from '@rstest/playwright'; expect(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`import.meta.rstest.expect(2 > 1).toBe(true);`, `import.meta.rstest.expect(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`const api = import.meta.rstest; api.expect(2 > 1).toBe(true);`, `const api = import.meta.rstest; api.expect(2).toBeGreaterThan(1);`, "toBeGreaterThan"},
		{`test('limit', ({ expect: check }) => check(2 > 1).toBe(true));`, `test('limit', ({ expect: check }) => check(2).toBeGreaterThan(1));`, "toBeGreaterThan"},
		{`test('limit', ctx => ctx.expect(2 > 1).toBe(true));`, `test('limit', ctx => ctx.expect(2).toBeGreaterThan(1));`, "toBeGreaterThan"},
		// ---- Real-user: empty expect regression (jest-community/eslint-plugin-jest#2000) ----
		{`expect().toBe(true); expect(a > b).toBe(true);`, "", "toBeGreaterThan"},
		// ---- Real-user: comments and trailing commas in formatted assertions ----
		{`expect(2 > 1, 'message').not /* keep */ .toBe(false,);`, `expect(2, 'message') /* keep */ .toBeGreaterThan(1,);`, "toBeGreaterThan"},
		{`expect(2 /* keep */ > 1).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(a > b).toBe(true /* keep */ as const);`, "", "toBeGreaterThan"},
		{`expect(2 > 1).toBe(/* keep */ true,);`, `expect(2).toBeGreaterThan(/* keep */ 1,);`, "toBeGreaterThan"},
		// Moving operands past expect() changes evaluation order or delays object coercion.
		{`expect(1 > expect.getState().assertionCalls).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(a > limits.maximum).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(a > nextLimit()).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(a > counter++).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(a > (ready ? upper : lower)).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(objectWithValueOf > 1).toBe(true);`, "", "toBeGreaterThan"},
		// Type arguments describe the original boolean subject or matcher argument.
		{`expect<boolean>(1 > 0).toBe(true);`, "", "toBeGreaterThan"},
		{`expect(1 > 0).toBe<boolean>(true);`, "", "toBeGreaterThan"},
		// N/A: declaration names and private identifiers do not participate in matcher selection.
	} {
		expectedError := comparisonError(test.matcher, test.output)
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: test.code, Errors: []rule_tester.InvalidTestCaseError{expectedError}})
	}
	for _, test := range []struct {
		code, output            string
		line, column, endColumn int
	}{
		{"expect(a > b)\n  .toBe(true);", "", 2, 4, 8},
		{`expect(a > b, "😀").toBe(true);`, "", 1, 21, 25},
	} {
		expectedError := comparisonError("toBeGreaterThan", test.output)
		expectedError.Line, expectedError.Column, expectedError.EndLine, expectedError.EndColumn = test.line, test.column, test.line, test.endColumn
		invalid = append(invalid, rule_tester.InvalidTestCase{Code: test.code, Errors: []rule_tester.InvalidTestCaseError{expectedError}})
	}
	// Adjacent constant suggestions must only rewrite their own assertion.
	code := `expect(2 > 1).toBe(true); expect(1 < 2).toBe(false);`
	invalid = append(invalid, rule_tester.InvalidTestCase{Code: code, Errors: []rule_tester.InvalidTestCaseError{
		comparisonError("toBeGreaterThan", strings.Replace(code, "expect(2 > 1).toBe(true)", "expect(2).toBeGreaterThan(1)", 1)),
		comparisonError("not.toBeLessThan", strings.Replace(code, "expect(1 < 2).toBe(false)", "expect(1).not.toBeLessThan(2)", 1)),
	}})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferComparisonMatcherRule, valid, invalid)
}
