// TestPreferStrictBooleanMatchersUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-strict-boolean-matchers suite
// (tests/prefer-strict-boolean-matchers.test.ts) 1:1. Position assertions cover
// line/column for every invalid case. The upstream cases written on
// `expectTypeOf` are kept as skipped cases: Rstest exposes no `expectTypeOf`,
// so there is no assertion for them to describe. Rstest API sources, accessor
// forms, fix boundaries and edit-demand coverage live in
// prefer_strict_boolean_matchers_extras_test.go.
package prefer_strict_boolean_matchers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStrictBooleanMatchersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferStrictBooleanMatchersRule,
		[]rule_tester.ValidTestCase{
			{Code: `[].push(true)`},
			{Code: `[].push(false)`},
			{Code: `expect("something");`},
			{Code: `expect(true).toBe(true);`},
			{Code: `expect(true).toBe(false);`},
			{Code: `expect(false).toBe(true);`},
			{Code: `expect(false).toBe(false);`},
			{Code: `expect(fal,se).toBe(true);`},
			{Code: `expect(fal,se).toBe(false);`},
			{Code: `expect(value).toEqual();`},
			{Code: `expect(value).not.toBe(true);`},
			{Code: `expect(value).not.toBe(false);`},
			{Code: `expect(value).not.toEqual();`},
			{Code: `expect(value).toBe(undefined);`},
			{Code: `expect(value).not.toBe(undefined);`},
			{Code: `expect(value).toBe();`},
			{Code: `expect(true).toMatchSnapshot();`},
			{Code: `expect("a string").toMatchSnapshot(true);`},
			{Code: `expect("a string").toMatchSnapshot(false);`},
			{Code: `expect("a string").not.toMatchSnapshot();`},
			{Code: `expect(something).toEqual('a string');`},
			{Code: `expect(true).toBe`},
			{Code: `expect(value).to.be.a("boolean");`},
			// SKIP: Rstest has no `expectTypeOf`, so this is an ordinary call
			// expression rather than an assertion.
			{Code: `expectTypeOf(true).toBe()`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(false).toBeTruthy();`,
				Output: []string{`expect(false).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTrue",
					Message:   "Prefer using `toBe(true)` to test value is `true`",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 25,
				}},
			},
			{
				Code:   `expect(false).toBeFalsy();`,
				Output: []string{`expect(false).toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalse",
					Message:   "Prefer using `toBe(false)` to test value is `false`",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 24,
				}},
			},
			// SKIP: Rstest has no `expectTypeOf`.
			{
				Skip:   true,
				Code:   `expectTypeOf(false).toBeTruthy();`,
				Output: []string{`expectTypeOf(false).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTrue",
					Line:      1,
					Column:    21,
				}},
			},
			// SKIP: Rstest has no `expectTypeOf`.
			{
				Skip:   true,
				Code:   `expectTypeOf(false).toBeFalsy();`,
				Output: []string{`expectTypeOf(false).toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalse",
					Line:      1,
					Column:    21,
				}},
			},
			{
				Code:   `expect(wasSuccessful).toBeTruthy();`,
				Output: []string{`expect(wasSuccessful).toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTrue",
					Message:   "Prefer using `toBe(true)` to test value is `true`",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
			{
				Code:   `expect(wasSuccessful).toBeFalsy();`,
				Output: []string{`expect(wasSuccessful).toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalse",
					Message:   "Prefer using `toBe(false)` to test value is `false`",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 32,
				}},
			},
			{
				Code:   `expect("a string").not.toBeTruthy();`,
				Output: []string{`expect("a string").not.toBe(true);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTrue",
					Message:   "Prefer using `toBe(true)` to test value is `true`",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 34,
				}},
			},
			{
				Code:   `expect("a string").not.toBeFalsy();`,
				Output: []string{`expect("a string").not.toBe(false);`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalse",
					Message:   "Prefer using `toBe(false)` to test value is `false`",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 33,
				}},
			},
		},
	)
}
