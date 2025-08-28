package prefer_nullish_coalescing

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBrandedTypeWithIgnorePrimitives(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferNullishCoalescingRule,
		[]rule_tester.ValidTestCase{
			// Branded type with ignorePrimitives.string = true
			// This should NOT trigger the rule
			{
				Code: `
declare let x: (string & { __brand?: any }) | undefined;
declare let y: string;
const result = x || y;`,
				Options: []any{map[string]any{
					"ignorePrimitives": map[string]any{
						"string": true,
					},
				}},
				FileName: "test.ts",
			},
			// Regular string with ignorePrimitives.string = true
			// This should also NOT trigger
			{
				Code: `
declare let x: string | undefined;
declare let y: string;
const result = x || y;`,
				Options: []any{map[string]any{
					"ignorePrimitives": map[string]any{
						"string": true,
					},
				}},
				FileName: "test.ts",
			},
		},
		[]rule_tester.InvalidTestCase{
			// Without ignorePrimitives, branded type should trigger
			{
				Code: `
declare let x: (string & { __brand?: any }) | undefined;
declare let y: string;
const result = x || y;`,
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "preferNullishOverOr",
					Line:      4,
					Column:    18,
					EndLine:   4,
					EndColumn: 20,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{{
						MessageId: "suggestNullish",
						Output: `
declare let x: (string & { __brand?: any }) | undefined;
declare let y: string;
const result = x ?? y;`,
					}},
				}},
				FileName: "test.ts",
			},
		},
	)
}
