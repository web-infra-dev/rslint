// TestPreferEnumInitializersUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/tests/rules/prefer-enum-initializers.test.ts
// 1:1. Position assertions cover line/column for every invalid case. Rslint-
// specific lock-in cases and tsgo edge-shape coverage live in
// prefer_enum_initializers_extras_test.go.
package prefer_enum_initializers

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferEnumInitializersUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &PreferEnumInitializersRule, []rule_tester.ValidTestCase{
		{Code: `
enum Direction {}
    `},
		{Code: `
enum Direction {
  Up = 1,
}
    `},
		{Code: `
enum Direction {
  Up = 1,
  Down = 2,
}
    `},
		{Code: `
enum Direction {
  Up = 'Up',
  Down = 'Down',
}
    `},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `
enum Direction {
  Up,
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 0,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 1,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 'Up',
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
enum Direction {
  Up,
  Down,
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 0,
  Down,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 1,
  Down,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 'Up',
  Down,
}
      `,
						},
					},
				},
				{
					MessageId: "defineInitializer",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up,
  Down = 1,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up,
  Down = 2,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up,
  Down = 'Down',
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
enum Direction {
  Up = 'Up',
  Down,
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      4,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 'Up',
  Down = 1,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 'Up',
  Down = 2,
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 'Up',
  Down = 'Down',
}
      `,
						},
					},
				},
			},
		},
		{
			Code: `
enum Direction {
  Up,
  Down = 'Down',
}
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "defineInitializer",
					Line:      3,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 0,
  Down = 'Down',
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 1,
  Down = 'Down',
}
      `,
						},
						{
							MessageId: "defineInitializerSuggestion",
							Output: `
enum Direction {
  Up = 'Up',
  Down = 'Down',
}
      `,
						},
					},
				},
			},
		},
	})
}
