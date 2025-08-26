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
				Code:     `const foo: string | null = bar || baz;`,
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: "preferNullishOverOr", Line: 1, Column: 33, EndLine: 1, EndColumn: 43}},
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
