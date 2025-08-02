package ban_ts_comment

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestBanTsComment(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BanTsCommentRule,
		[]rule_tester.ValidTestCase{
			{Code: "// just a comment containing @ts-expect-error somewhere\nconst a = 1;"},
			{Code: `
/*
 @ts-expect-error running with long description in a block
*/
const a = 1;
			`},
			{Code: `
/* @ts-expect-error not on the last line
 */
const a = 1;
			`},
			{Code: `
/**
 * @ts-expect-error not on the last line
 */
const a = 1;
			`},
			{Code: `
/* not on the last line
 * @ts-expect-error
 */
const a = 1;
			`},
			{Code: `
/* @ts-expect-error
 * not on the last line */
const a = 1;
			`},
			{
				Code:    "// @ts-expect-error\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": false},
			},
			{
				Code:    "// @ts-expect-error here is why the error is expected\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-expect-error here is why the error is expected */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			},
			{
				Code:    "// @ts-expect-error exactly 21 characters\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-expect-error": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-expect-error exactly 21 characters*/
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-expect-error": "allow-with-description"},
			},
			{
				Code:    "// @ts-expect-error: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code: `
/*
 * @ts-expect-error: TS1234 because xyz */
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-expect-error 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			},
			
			// ts-ignore valid cases
			{Code: "// just a comment containing @ts-ignore somewhere\nconst a = 1;"},
			{
				Code:    "// @ts-ignore\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": false},
			},
			{
				Code:    "// @ts-ignore I think that I am exempted from any need to follow the rules!\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 @ts-ignore running with long description in a block
*/
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"},
			},
			{Code: `
/*
 @ts-ignore
*/
const a = 1;
			`},
			{Code: `
/* @ts-ignore not on the last line
 */
const a = 1;
			`},
			{Code: `
/**
 * @ts-ignore not on the last line
 */
const a = 1;
			`},
			{Code: `
/* not on the last line
 * @ts-expect-error
 */
const a = 1;
			`},
			{Code: `
/* @ts-ignore
 * not on the last line */
const a = 1;
			`},
			{
				Code:    "// @ts-ignore: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-ignore 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-ignore here is why the error is expected */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			},
			{
				Code:    "// @ts-ignore exactly 21 characters\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-ignore exactly 21 characters*/
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-ignore: TS1234 because xyz */
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			
			// ts-nocheck valid cases
			{Code: "// just a comment containing @ts-nocheck somewhere\nconst a = 1;"},
			{
				Code:    "// @ts-nocheck\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": false},
			},
			{
				Code:    "// @ts-nocheck no doubt, people will put nonsense here from time to time just to get the rule to stop reporting, perhaps even long messages with other nonsense in them like other // @ts-nocheck or // @ts-ignore things\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			},
			{
				Code: `
/*
 @ts-nocheck running with long description in a block
*/
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-nocheck": "allow-with-description"},
			},
			{
				Code:    "// @ts-nocheck: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-nocheck 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			},
			{Code: "//// @ts-nocheck - pragma comments may contain 2 or 3 leading slashes\nconst a = 1;"},
			{Code: `
/**
 @ts-nocheck
*/
const a = 1;
			`},
			{Code: `
/*
 @ts-nocheck
*/
const a = 1;
			`},
			{Code: "/** @ts-nocheck */\nconst a = 1;"},
			{Code: "/* @ts-nocheck */\nconst a = 1;"},
			{Code: `
const a = 1;

// @ts-nocheck - should not be reported

// TS error is not actually suppressed
const b: string = a;
			`},
			
			// ts-check valid cases
			{Code: "// just a comment containing @ts-check somewhere\nconst a = 1;"},
			{Code: `
/*
 @ts-check running with long description in a block
*/
const a = 1;
			`},
			{
				Code:    "// @ts-check\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": false},
			},
			{
				Code:    "// @ts-check with a description and also with a no-op // @ts-ignore\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 3, "ts-check": "allow-with-description"},
			},
			{
				Code:    "// @ts-check: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-check 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": "allow-with-description"},
			},
			{
				Code:    "//// @ts-check - pragma comments may contain 2 or 3 leading slashes\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code: `
/**
 @ts-check
*/
const a = 1;
				`,
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code: `
/*
 @ts-check
*/
const a = 1;
				`,
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code:    "/** @ts-check */\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code:    "/* @ts-check */\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ts-expect-error invalid cases
			{
				Code: "// @ts-expect-error\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "/* @ts-expect-error */\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `
/*
@ts-expect-error */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/** on the last line
  @ts-expect-error */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/** on the last line
 * @ts-expect-error */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/**
 * @ts-expect-error: TODO */
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/**
 * @ts-expect-error: TS1234 because xyz */
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/**
 * @ts-expect-error: TS1234 */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/**
 * @ts-expect-error    : TS1234 */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: `
/**
 * @ts-expect-error 👨‍👩‍👧‍👦 */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      2,
						Column:    1,
					},
				},
			},
			{
				Code: "/** @ts-expect-error */\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-expect-error: Suppress next line\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "/////@ts-expect-error: Suppress next line\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
				`,
				Options: map[string]interface{}{"ts-expect-error": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      3,
						Column:    3,
					},
				},
			},
			{
				Code: "// @ts-expect-error\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-expect-error: TODO\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-expect-error: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-expect-error: TS1234\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-expect-error    : TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-expect-error 👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},

			// ts-ignore invalid cases
			{
				Code: "// @ts-ignore\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": true, "ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore\nconst a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: "/* @ts-ignore */\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "/* @ts-expect-error */\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: `
/*
 @ts-ignore */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      2,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output: `
/*
 @ts-expect-error */
const a = 1;
				`,
							},
						},
					},
				},
			},
			{
				Code: `
/** on the last line
  @ts-ignore */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      2,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output: `
/** on the last line
  @ts-expect-error */
const a = 1;
				`,
							},
						},
					},
				},
			},
			{
				Code: `
/** on the last line
 * @ts-ignore */
const a = 1;
				`,
				Options: map[string]interface{}{"ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      2,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output: `
/** on the last line
 * @ts-expect-error */
const a = 1;
				`,
							},
						},
					},
				},
			},
			{
				Code: "/** @ts-ignore */\nconst a = 1;",
				Options: map[string]interface{}{"ts-expect-error": false, "ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "/** @ts-expect-error */\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: `
/**
 * @ts-ignore: TODO */
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      2,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output: `
/**
 * @ts-expect-error: TODO */
const a = 1;
				`,
							},
						},
					},
				},
			},
			{
				Code: `
/**
 * @ts-ignore: TS1234 because xyz */
const a = 1;
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      2,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output: `
/**
 * @ts-expect-error: TS1234 because xyz */
const a = 1;
				`,
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore: Suppress next line\nconst a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error: Suppress next line\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: "/////@ts-ignore: Suppress next line\nconst a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "/////@ts-expect-error: Suppress next line\nconst a = 1;",
							},
						},
					},
				},
			},
			{
				Code: `
if (false) {
  // @ts-ignore: Unreachable code error
  console.log('hello');
}
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      3,
						Column:    3,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output: `
if (false) {
  // @ts-expect-error: Unreachable code error
  console.log('hello');
}
				`,
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-ignore    .\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-ignore: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-ignore: TS1234\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-ignore    : TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-ignore 👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},

			// ts-nocheck invalid cases
			{
				Code: "// @ts-nocheck\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck\nconst a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck: Suppress next line\nconst a = 1;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck: TS1234\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck    : TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck 👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `
 // @ts-nocheck
const a: true = false;
				`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      2,
						Column:    2,
					},
				},
			},

			// ts-check invalid cases
			{
				Code: "// @ts-check\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-check: Suppress next line\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `
if (false) {
  // @ts-check: Unreachable code error
  console.log('hello');
}
				`,
				Options: map[string]interface{}{"ts-check": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      3,
						Column:    3,
					},
				},
			},
			{
				Code: "// @ts-check\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-check: TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-check: TS1234\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-check    : TS1234 because xyz\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentDescriptionNotMatchPattern",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-check 👨‍👩‍👧‍👦\nconst a = 1;",
				Options: map[string]interface{}{"ts-check": "allow-with-description"},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveCommentRequiresDescription",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}