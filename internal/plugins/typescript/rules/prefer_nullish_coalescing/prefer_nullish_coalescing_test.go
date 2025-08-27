package prefer_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferNullishCoalescingRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			// Valid cases using nullish coalescing
			{Code: `const foo = bar ?? baz;`},
			{Code: `foo ??= bar;`},
			// Non-nullable types should not trigger
			{Code: `const foo = bar || baz;`},
		},
		[]rule_tester.InvalidTestCase{
			// Basic logical OR cases with nullable types
			{
				Code: `
declare const bar: string | null;
declare const baz: string;
const foo = bar || baz;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNullishOverOr",
					Line:      4,
					Column:    13,
					EndLine:   4,
					EndColumn: 23,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestNullish",
						Output: `
declare const bar: string | null;
declare const baz: string;
const foo = bar ?? baz;`,
					}},
				}},
				FileName: "test.ts",
			},
		},
	)
}

func TestPreferNullishCoalescingRuleWithOptions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `const foo = bar || baz;`,
				Options: map[string]any{"ignorePrimitives": true},
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}

func TestPreferNullishCoalescingRuleStrictNullChecks(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `const foo = bar || baz;`,
				Options: map[string]any{"allowRuleToRunWithoutStrictNullChecksIKnowWhatIAmDoing": true},
			},
		},
		[]rule_tester.InvalidTestCase{},
	)
}
