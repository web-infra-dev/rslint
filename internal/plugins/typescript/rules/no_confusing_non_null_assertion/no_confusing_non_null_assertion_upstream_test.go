// TestNoConfusingNonNullAssertionUpstream migrates the full valid/invalid
// suite from upstream
// https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-confusing-non-null-assertion.test.ts
// 1:1. Position assertions cover line/column for every invalid case.
// rslint-specific lock-in cases live in
// no_confusing_non_null_assertion_extras_test.go.
package no_confusing_non_null_assertion

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConfusingNonNullAssertionUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoConfusingNonNullAssertionRule, []rule_tester.ValidTestCase{
		{Code: `a == b!;`},
		{Code: `a = b!;`},
		{Code: `a !== b;`},
		{Code: `a != b;`},
		{Code: `(a + b!) == c;`},
		{Code: `(a + b!) = c;`},
		{Code: `(a + b!) in c;`},
		{Code: `(a || b!) instanceof c;`},
	}, []rule_tester.InvalidTestCase{
		{
			Code: `a! == b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `a == b;`},
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
						{MessageId: "notNeedInEqualTest", Output: `a === b;`},
					},
				},
			},
		},
		{
			Code: `a + b! == c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "wrapUpLeft", Output: `(a + b!) == c;`},
					},
				},
			},
		},
		{
			Code: `(obj = new new OuterObj().InnerObj).Name! == c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `(obj = new new OuterObj().InnerObj).Name == c;`},
					},
				},
			},
		},
		{
			Code: `(a==b)! ==c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingEqual",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInEqualTest", Output: `(a==b) ==c;`},
					},
				},
			},
		},
		{
			Code: `a! = b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingAssign",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInAssign", Output: `a = b;`},
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
						{MessageId: "notNeedInAssign", Output: `(obj = new new OuterObj().InnerObj).Name = c;`},
					},
				},
			},
		},
		{
			Code: `(a=b)! =c;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingAssign",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInAssign", Output: `(a=b) =c;`},
					},
				},
			},
		},
		{
			Code: `a! in b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `a in b;`},
						{MessageId: "wrapUpLeft", Output: `(a!) in b;`},
					},
				},
			},
		},
		{
			Code: "\na !in b;\n      ",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      2,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: "\na in b;\n      "},
						{MessageId: "wrapUpLeft", Output: "\n(a !)in b;\n      "},
					},
				},
			},
		},
		{
			Code: `a! instanceof b;`,
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "confusingOperator",
					Line:      1,
					Column:    1,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "notNeedInOperator", Output: `a instanceof b;`},
						{MessageId: "wrapUpLeft", Output: `(a!) instanceof b;`},
					},
				},
			},
		},
	})
}
