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
					Column:    17,
					EndLine:   4,
					EndColumn: 19,
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

func TestPreferNullishCoalescingRuleIgnoreTernaryTests(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			// Should NOT flag when ignoreTernaryTests is true
			{
				Code: `
declare let x: string | null;
const result = (x || 'foo') ? null : null;`,
				Options:  map[string]any{"ignoreTernaryTests": true},
				FileName: "test.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Should flag when ignoreTernaryTests is false (explicit)
			{
				Code: `
declare let x: string | null;
const result = (x || 'foo') ? null : null;`,
				Options: map[string]any{"ignoreTernaryTests": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNullishOverOr",
					Line:      3,
					Column:    19,
					EndLine:   3,
					EndColumn: 21,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestNullish",
						Output: `
declare let x: string | null;
const result = (x ?? 'foo') ? null : null;`,
					}},
				}},
				FileName: "test.ts",
			},

			// Should flag OR in ternary consequent (not in test)
			{
				Code: `
declare let x: string | null;
declare let condition: boolean;
const result = condition ? (x || 'foo') : null;`,
				Options: map[string]any{"ignoreTernaryTests": true},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNullishOverOr",
					Line:      4,
					Column:    31,
					EndLine:   4,
					EndColumn: 33,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestNullish",
						Output: `
declare let x: string | null;
declare let condition: boolean;
const result = condition ? (x ?? 'foo') : null;`,
					}},
				}},
				FileName: "test.ts",
			},
		},
	)
}

func TestMemberAccessPattern(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			// Member access with non-nullable type
			{
				Code: `
declare let obj: { prop: string };
const result = obj.prop || 'default';`,
				FileName: "test.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Member access pattern: obj.prop ? obj.prop : 'default'
			{
				Code: `
declare let obj: { prop: string | null };
const result = obj.prop ? obj.prop : 'default';`,
				Options: map[string]any{"ignoreTernaryTests": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNullishOverTernary",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 47,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestNullish",
						Output: `
declare let obj: { prop: string | null };
const result = obj.prop ?? 'default';`,
					}},
				}},
				FileName: "test.ts",
			},
			// Negated member access pattern: !obj.prop ? 'default' : obj.prop
			{
				Code: `
declare let obj: { prop: string | null };
const result = !obj.prop ? 'default' : obj.prop;`,
				Options: map[string]any{"ignoreTernaryTests": false},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNullishOverTernary",
					Line:      3,
					Column:    16,
					EndLine:   3,
					EndColumn: 48,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestNullish",
						Output: `
declare let obj: { prop: string | null };
const result = obj.prop ?? 'default';`,
					}},
				}},
				FileName: "test.ts",
			},
		},
	)
}
