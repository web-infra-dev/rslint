package no_confusing_non_null_assertion

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestNoConfusingNonNullAssertion(t *testing.T) {
	validTestCases := []rule_tester.ValidTestCase{
		{
			Code: `a == b!;`,
		},
		{
			Code: `a = b!;`,
		},
		{
			Code: `a !== b;`,
		},
		{
			Code: `a != b;`,
		},
		{
			Code: `(a + b!) == c;`,
		},
		{
			Code: `(a + b!) = c;`,
		},
		{
			Code: `(a + b!) in c;`,
		},
		{
			Code: `(a || b!) instanceof c;`,
		},
		{
			Code: `(a)! == b;`,
		},
		{
			Code: `(a)! = b;`,
		},
		{
			Code: `(a)! in b;`,
		},
		{
			Code: `(a)! instanceof b;`,
		},
		{
			Code: `a!! == b;`,
		},
		{
			Code: `a!! = b;`,
		},
		{
			Code: `const getName = () => otherCondition ? 'a' : 'b';`,
		},
		{
			Code: `const foo = otherCondition ? 'a' : 'b';`,
		},
	}

	invalidTestCases := []rule_tester.InvalidTestCase{
		// Assignment operator cases
		{
			Code: `a! = b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingAssign",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `(a!) = b;`,
						},
					},
				},
			},
		},
		{
			Code: `a! = +b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingAssign",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `(a!) = +b;`,
						},
					},
				},
			},
		},
		{
			Code: `
a!
= b;
`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingAssign",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output: `
(a!)
= b;
`,
						},
					},
				},
			},
		},
		{
			Code: `(obj = new new OuterObj().InnerObj).Name! = c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingAssign",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `((obj = new new OuterObj().InnerObj).Name!) = c;`,
						},
					},
				},
			},
		},
		// Equality operator cases
		{
			Code: `a! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `(a!) == b;`,
						},
					},
				},
			},
		},
		{
			Code: `a! === b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `(a!) === b;`,
						},
					},
				},
			},
		},
		// in operator cases
		{
			Code: `a! in b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `(a!) in b;`,
						},
					},
				},
			},
		},
		{
			Code: `
a !in b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output: `
(a !)in b;
      `,
						},
					},
				},
			},
		},
		// instanceof operator cases
		{
			Code: `a! instanceof b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output:    `(a!) instanceof b;`,
						},
					},
				},
			},
		},
		{
			Code: `
a !instanceof b;
      `,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{
							MessageId: "wrapUpLeft",
							Output: `
(a !)instanceof b;
      `,
						},
					},
				},
			},
		},
	}

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingNonNullAssertionRule, validTestCases, invalidTestCases)
}