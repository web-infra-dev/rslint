// TestPreferStrictEqualUpstream migrates the full valid/invalid suite from
// eslint-plugin-jest@v29.16.0 src/rules/__tests__/prefer-strict-equal.test.ts
// 1:1. Position assertions cover line/column for every invalid case. The
// existing rslint accessor and trivia lock-ins live in
// prefer_strict_equal_extras_test.go.
package prefer_strict_equal_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/jest/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/rules/prefer_strict_equal"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferStrictEqualUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&prefer_strict_equal.PreferStrictEqualRule,
		[]rule_tester.ValidTestCase{
			{Code: `expect(something).toStrictEqual(somethingElse);`},
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
				// Rslint keeps the authored delimiter instead of normalizing it.
				Code: `expect(something)["toEqual"](somethingElse);`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "useToStrictEqual", Line: 1, Column: 19, EndLine: 1, EndColumn: 28,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestReplaceWithStrictEqual", Output: `expect(something)["toStrictEqual"](somethingElse);`}},
				}},
			},
		},
	)
}
