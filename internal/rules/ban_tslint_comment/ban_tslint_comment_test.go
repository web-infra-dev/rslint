package ban_tslint_comment

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestBanTslintComment(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BanTslintCommentRule,
		[]rule_tester.ValidTestCase{
			{Code: "let a: readonly any[] = [];"},
			{Code: "let a = new Array();"},
			{Code: "// some other comment"},
			{Code: "// TODO: this is a comment that mentions tslint"},
			{Code: "/* another comment that mentions tslint */"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "/* tslint:disable */",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "",
							},
						},
					},
				},
			},
			{
				Code: "/* tslint:enable */",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "",
							},
						},
					},
				},
			},
			{
				Code: "/* tslint:disable:rule1 rule2 rule3... */",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "",
							},
						},
					},
				},
			},
			{
				Code: "/* tslint:enable:rule1 rule2 rule3... */",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "",
							},
						},
					},
				},
			},
			{
				Code: "// tslint:disable-next-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "",
							},
						},
					},
				},
			},
			{
				Code: "someCode(); // tslint:disable-line",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    13,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "someCode();",
							},
						},
					},
				},
			},
			{
				Code: "// tslint:disable-next-line:rule1 rule2 rule3...",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output:    "",
							},
						},
					},
				},
			},
			{
				Code: `
const woah = doSomeStuff();
// tslint:disable-line
console.log(woah);
`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "commentDetected",
						Line:      3,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "removeTslintComment",
								Output: `
const woah = doSomeStuff();
console.log(woah);
`,
							},
						},
					},
				},
			},
		},
	)
}