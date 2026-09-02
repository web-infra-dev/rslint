// TestPreferToBeTruthyUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-to-be-truthy suite
// (tests/prefer-to-be-truthy.test.ts) 1:1. Position assertions cover
// line/column for every invalid case. The upstream cases written on
// `expectTypeOf` are kept as skipped cases: Rstest exposes no `expectTypeOf`,
// so there is no assertion for them to describe. Rstest API sources, accessor
// forms, fix boundaries and edit-demand coverage live in
// prefer_to_be_truthy_extras_test.go.
package prefer_to_be_truthy

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToBeTruthyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToBeTruthyRule,
		[]rule_tester.ValidTestCase{
			{Code: `[].push(true)`},
			{Code: `expect("something");`},
			{Code: `expect(value).to.be.a("boolean");`},
			{Code: `expect(true).toBeTrue();`},
			{Code: `expect(false).toBeTrue();`},
			{Code: `expect(fal,se).toBeFalse();`},
			{Code: `expect(true).toBeFalse();`},
			{Code: `expect(value).toEqual();`},
			{Code: `expect(value).not.toBeTrue();`},
			{Code: `expect(value).not.toEqual();`},
			{Code: `expect(value).toBe(undefined);`},
			{Code: `expect(value).not.toBe(undefined);`},
			{Code: `expect(true).toBe(false)`},
			{Code: `expect(value).toBe();`},
			{Code: `expect(true).toMatchSnapshot();`},
			{Code: `expect("a string").toMatchSnapshot(true);`},
			{Code: `expect("a string").not.toMatchSnapshot();`},
			{Code: `expect(something).toEqual('a string');`},
			{Code: `expect(true).toBe`},
			// SKIP: Rstest has no `expectTypeOf`, so this is an ordinary call
			// expression rather than an assertion.
			{Code: `expectTypeOf(true).toBe()`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(false).toBe(true);`,
				Output: []string{`expect(false).toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Message:   "Prefer using `toBeTruthy` to test value is `true`",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 19,
				}},
			},
			// SKIP: Rstest has no `expectTypeOf`.
			{
				Skip:   true,
				Code:   `expectTypeOf(false).toBe(true);`,
				Output: []string{`expectTypeOf(false).toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Line:      1,
					Column:    21,
				}},
			},
			{
				Code:   `expect(wasSuccessful).toEqual(true);`,
				Output: []string{`expect(wasSuccessful).toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Message:   "Prefer using `toBeTruthy` to test value is `true`",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 30,
				}},
			},
			{
				Code:   `expect(fs.existsSync('/path/to/file')).toStrictEqual(true);`,
				Output: []string{`expect(fs.existsSync('/path/to/file')).toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Message:   "Prefer using `toBeTruthy` to test value is `true`",
					Line:      1,
					Column:    40,
					EndLine:   1,
					EndColumn: 53,
				}},
			},
			{
				Code:   `expect("a string").not.toBe(true);`,
				Output: []string{`expect("a string").not.toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Message:   "Prefer using `toBeTruthy` to test value is `true`",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 28,
				}},
			},
			{
				Code:   `expect("a string").not.toEqual(true);`,
				Output: []string{`expect("a string").not.toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Message:   "Prefer using `toBeTruthy` to test value is `true`",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 31,
				}},
			},
			// SKIP: Rstest has no `expectTypeOf`.
			{
				Skip:   true,
				Code:   `expectTypeOf("a string").not.toStrictEqual(true);`,
				Output: []string{`expectTypeOf("a string").not.toBeTruthy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeTruthy",
					Line:      1,
					Column:    30,
				}},
			},
		},
	)
}
