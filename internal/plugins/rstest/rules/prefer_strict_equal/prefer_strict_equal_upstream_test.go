// TestPreferStrictEqualUpstream migrates the full valid/invalid suite from
// @vitest/eslint-plugin@v1.6.27 tests/prefer-strict-equal.test.ts 1:1.
// Position assertions cover line/column for every invalid case. Rstest source
// forms, assertion factories, chain semantics and edit boundaries live in
// prefer_strict_equal_extras_test.go.
package prefer_strict_equal

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/rstest/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStrictEqualUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&PreferStrictEqualRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(something).toStrictEqual(somethingElse);`},
			{Code: `expect(something).to.be.a("string");`},
			{Code: `a().toEqual('b')`},
			{Code: `expect(a);`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `expect(something).toEqual(somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Message: "Use `toStrictEqual()` instead",
					Line: 1, Column: 19, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(something).toStrictEqual(somethingElse);`}},
				}},
			},
			{
				Code: `expect(something).toEqual(somethingElse,);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Line: 1, Column: 19, EndLine: 1, EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(something).toStrictEqual(somethingElse,);`}},
				}},
			},
			{
				// Rslint preserves the authored accessor delimiter.
				Code: `expect(something)["toEqual"](somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Line: 1, Column: 19, EndLine: 1, EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(something)["toStrictEqual"](somethingElse);`}},
				}},
			},
		},
	)
}
