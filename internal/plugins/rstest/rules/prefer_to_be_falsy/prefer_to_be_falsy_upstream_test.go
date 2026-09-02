// TestPreferToBeFalsyUpstream migrates the complete
// @vitest/eslint-plugin@v1.6.27 prefer-to-be-falsy suite
// (tests/prefer-to-be-falsy.test.ts) 1:1. Position assertions cover
// line/column for every invalid case. The upstream cases written on
// `expectTypeOf` are kept as skipped cases: Rstest exposes no `expectTypeOf`,
// so there is no assertion for them to describe. Rstest API sources, accessor
// forms, fix boundaries and edit-demand coverage live in
// prefer_to_be_falsy_extras_test.go.
package prefer_to_be_falsy

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferToBeFalsyUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferToBeFalsyRule,
		[]rule_tester.ValidTestCase{
			{Code: `[].push(false)`},
			{Code: `expect("something");`},
			{Code: `expect(value).to.be.a("boolean");`},
			{Code: `expect(true).toBeTrue();`},
			{Code: `expect(false).toBeTrue();`},
			{Code: `expect(false).toBeFalsy();`},
			{Code: `expect(true).toBeFalsy();`},
			{Code: `expect(value).toEqual();`},
			{Code: `expect(value).not.toBeFalsy();`},
			{Code: `expect(value).not.toEqual();`},
			{Code: `expect(value).toBe(undefined);`},
			{Code: `expect(value).not.toBe(undefined);`},
			{Code: `expect(false).toBe(true)`},
			{Code: `expect(value).toBe();`},
			{Code: `expect(true).toMatchSnapshot();`},
			{Code: `expect("a string").toMatchSnapshot(false);`},
			{Code: `expect("a string").not.toMatchSnapshot();`},
			{Code: `expect(something).toEqual('a string');`},
			{Code: `expect(false).toBe`},
			// SKIP: Rstest has no `expectTypeOf`, so this is an ordinary
			// property access rather than an assertion.
			{Code: `expectTypeOf(false).toBe`, Skip: true},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `expect(true).toBe(false);`,
				Output: []string{`expect(true).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalsy",
					Message:   "Prefer using toBeFalsy()",
					Line:      1,
					Column:    14,
					EndLine:   1,
					EndColumn: 18,
				}},
			},
			{
				Code:   `expect(wasSuccessful).toEqual(false);`,
				Output: []string{`expect(wasSuccessful).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalsy",
					Message:   "Prefer using toBeFalsy()",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 30,
				}},
			},
			{
				Code:   `expect(fs.existsSync('/path/to/file')).toStrictEqual(false);`,
				Output: []string{`expect(fs.existsSync('/path/to/file')).toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalsy",
					Message:   "Prefer using toBeFalsy()",
					Line:      1,
					Column:    40,
					EndLine:   1,
					EndColumn: 53,
				}},
			},
			{
				Code:   `expect("a string").not.toBe(false);`,
				Output: []string{`expect("a string").not.toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalsy",
					Message:   "Prefer using toBeFalsy()",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 28,
				}},
			},
			{
				Code:   `expect("a string").not.toEqual(false);`,
				Output: []string{`expect("a string").not.toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalsy",
					Message:   "Prefer using toBeFalsy()",
					Line:      1,
					Column:    24,
					EndLine:   1,
					EndColumn: 31,
				}},
			},
			// SKIP: Rstest has no `expectTypeOf`.
			{
				Skip:   true,
				Code:   `expectTypeOf("a string").not.toEqual(false);`,
				Output: []string{`expectTypeOf("a string").not.toBeFalsy();`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferToBeFalsy",
					Line:      1,
					Column:    30,
				}},
			},
		},
	)
}
