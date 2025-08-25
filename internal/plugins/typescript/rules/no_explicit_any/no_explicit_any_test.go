package no_explicit_any

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExplicitAnyRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExplicitAnyRule, []rule_tester.ValidTestCase{
		{Code: `const number: number = 1;`},
		{
			Code:    `function foo(...args: any[]) {}`,
			Options: []interface{}{map[string]interface{}{"ignoreRestArgs": true}},
		},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `const number: any = 1;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 18,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    `const number: unknown = 1;`,
						},
						{
							MessageId: "suggestNever",
							Output:    `const number: never = 1;`,
						},
					},
				},
			},
		},
		{
			Code: `type T = keyof any;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    16,
					EndLine:   1,
					EndColumn: 19,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestPropertyKey",
							Output:    `type T = PropertyKey;`,
						},
					},
				},
			},
		},
		{
			Code: `function foo(...args: any[]) {}`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    23,
					EndLine:   1,
					EndColumn: 26,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "suggestUnknown",
							Output:    `function foo(...args: unknown[]) {}`,
						},
						{
							MessageId: "suggestNever",
							Output:    `function foo(...args: never[]) {}`,
						},
					},
				},
			},
		},
		{
			Code:    `const number: any = 1;`,
			Options: []interface{}{map[string]interface{}{"fixToUnknown": true}},
			Output:  []string{`const number: unknown = 1;`},
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "unexpectedAny",
					Line:      1,
					Column:    15,
					EndLine:   1,
					EndColumn: 18,
				},
			},
		},
	})
}
