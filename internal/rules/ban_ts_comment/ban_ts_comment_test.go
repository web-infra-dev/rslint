package ban_ts_comment

import (
	"testing"

	"github.com/typescript-eslint/rslint/internal/rule_tester"
	"github.com/typescript-eslint/rslint/internal/rules/fixtures"
)

func TestBanTsComment(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BanTsCommentRule,
		[]rule_tester.ValidTestCase{
			{Code: "// just a comment containing @ts-expect-error somewhere"},
			{Code: `
/*
 @ts-expect-error running with long description in a block
*/
			`},
			{Code: `
/* @ts-expect-error not on the last line
 */
			`},
			{Code: `
/**
 * @ts-expect-error not on the last line
 */
			`},
			{Code: `
/* not on the last line
 * @ts-expect-error
 */
			`},
			{Code: `
/* @ts-expect-error
 * not on the last line */
			`},
			{
				Code:    "// @ts-expect-error",
				Options: map[string]interface{}{"ts-expect-error": false},
			},
			{
				Code:    "// @ts-expect-error here is why the error is expected",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-expect-error here is why the error is expected */
				`,
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			},
			{
				Code:    "// @ts-expect-error exactly 21 characters",
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-expect-error": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-expect-error exactly 21 characters*/
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-expect-error": "allow-with-description"},
			},
			{
				Code:    "// @ts-expect-error: TS1234 because xyz",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code: `
/*
 * @ts-expect-error: TS1234 because xyz */
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-expect-error 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			},
			
			// ts-ignore valid cases
			{Code: "// just a comment containing @ts-ignore somewhere"},
			{
				Code:    "// @ts-ignore",
				Options: map[string]interface{}{"ts-ignore": false},
			},
			{
				Code:    "// @ts-ignore I think that I am exempted from any need to follow the rules!",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 @ts-ignore running with long description in a block
*/
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"},
			},
			{Code: `
/*
 @ts-ignore
*/
			`},
			{Code: `
/* @ts-ignore not on the last line
 */
			`},
			{Code: `
/**
 * @ts-ignore not on the last line
 */
			`},
			{Code: `
/* not on the last line
 * @ts-expect-error
 */
			`},
			{Code: `
/* @ts-ignore
 * not on the last line */
			`},
			{
				Code:    "// @ts-ignore: TS1234 because xyz",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-ignore 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦",
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-ignore here is why the error is expected */
				`,
				Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			},
			{
				Code:    "// @ts-ignore exactly 21 characters",
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-ignore exactly 21 characters*/
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"},
			},
			{
				Code: `
/*
 * @ts-ignore: TS1234 because xyz */
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			
			// ts-nocheck valid cases
			{Code: "// just a comment containing @ts-nocheck somewhere"},
			{
				Code:    "// @ts-nocheck",
				Options: map[string]interface{}{"ts-nocheck": false},
			},
			{
				Code:    "// @ts-nocheck no doubt, people will put nonsense here from time to time just to get the rule to stop reporting, perhaps even long messages with other nonsense in them like other // @ts-nocheck or // @ts-ignore things",
				Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			},
			{
				Code: `
/*
 @ts-nocheck running with long description in a block
*/
				`,
				Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-nocheck": "allow-with-description"},
			},
			{
				Code:    "// @ts-nocheck: TS1234 because xyz",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-nocheck 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦",
				Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			},
			{Code: "//// @ts-nocheck - pragma comments may contain 2 or 3 leading slashes"},
			{Code: `
/**
 @ts-nocheck
*/
			`},
			{Code: `
/*
 @ts-nocheck
*/
			`},
			{Code: "/** @ts-nocheck */"},
			{Code: "/* @ts-nocheck */"},
			{Code: `
const a = 1;

// @ts-nocheck - should not be reported

// TS error is not actually suppressed
const b: string = a;
			`},
			
			// ts-check valid cases
			{Code: "// just a comment containing @ts-check somewhere"},
			{Code: `
/*
 @ts-check running with long description in a block
*/
			`},
			{
				Code:    "// @ts-check",
				Options: map[string]interface{}{"ts-check": false},
			},
			{
				Code:    "// @ts-check with a description and also with a no-op // @ts-ignore",
				Options: map[string]interface{}{"minimumDescriptionLength": 3, "ts-check": "allow-with-description"},
			},
			{
				Code:    "// @ts-check: TS1234 because xyz",
				Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+"}},
			},
			{
				Code:    "// @ts-check 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦",
				Options: map[string]interface{}{"ts-check": "allow-with-description"},
			},
			{
				Code:    "//// @ts-check - pragma comments may contain 2 or 3 leading slashes",
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code: `
/**
 @ts-check
*/
				`,
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code: `
/*
 @ts-check
*/
				`,
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code:    "/** @ts-check */",
				Options: map[string]interface{}{"ts-check": true},
			},
			{
				Code:    "/* @ts-check */",
				Options: map[string]interface{}{"ts-check": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ts-expect-error invalid cases
			{
				Code: "// @ts-expect-error",
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
				Code: "/* @ts-expect-error */",
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
				Code: "/** @ts-expect-error */",
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
				Code: "// @ts-expect-error: Suppress next line",
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
				Code: "/////@ts-expect-error: Suppress next line",
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
				Code: "// @ts-expect-error",
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
				Code: "// @ts-expect-error: TODO",
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
				Code: "// @ts-expect-error: TS1234 because xyz",
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
				Code: "// @ts-expect-error: TS1234",
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
				Code: "// @ts-expect-error    : TS1234 because xyz",
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
				Code: "// @ts-expect-error 👨‍👩‍👧‍👦",
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
				Code: "// @ts-ignore",
				Options: map[string]interface{}{"ts-expect-error": true, "ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error",
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore",
				Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error",
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error",
							},
						},
					},
				},
			},
			{
				Code: "/* @ts-ignore */",
				Options: map[string]interface{}{"ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "/* @ts-expect-error */",
							},
						},
					},
				},
			},
			{
				Code: `
/*
 @ts-ignore */
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
				`,
							},
						},
					},
				},
			},
			{
				Code: "/** @ts-ignore */",
				Options: map[string]interface{}{"ts-expect-error": false, "ts-ignore": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "/** @ts-expect-error */",
							},
						},
					},
				},
			},
			{
				Code: `
/**
 * @ts-ignore: TODO */
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
				`,
							},
						},
					},
				},
			},
			{
				Code: "// @ts-ignore: Suppress next line",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "// @ts-expect-error: Suppress next line",
							},
						},
					},
				},
			},
			{
				Code: "/////@ts-ignore: Suppress next line",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsIgnoreInsteadOfExpectError",
						Line:      1,
						Column:    1,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{
								MessageId: "replaceTsIgnoreWithTsExpectError",
								Output:    "/////@ts-expect-error: Suppress next line",
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
				Code: "// @ts-ignore",
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
				Code: "// @ts-ignore    .",
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
				Code: "// @ts-ignore: TS1234 because xyz",
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
				Code: "// @ts-ignore: TS1234",
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
				Code: "// @ts-ignore    : TS1234 because xyz",
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
				Code: "// @ts-ignore 👨‍👩‍👧‍👦",
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
				Code: "// @ts-nocheck",
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
				Code: "// @ts-nocheck",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck: Suppress next line",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "tsDirectiveComment",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "// @ts-nocheck",
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
				Code: "// @ts-nocheck: TS1234 because xyz",
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
				Code: "// @ts-nocheck: TS1234",
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
				Code: "// @ts-nocheck    : TS1234 because xyz",
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
				Code: "// @ts-nocheck 👨‍👩‍👧‍👦",
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
				Code: "// @ts-check",
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
				Code: "// @ts-check: Suppress next line",
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
				Code: "// @ts-check",
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
				Code: "// @ts-check: TS1234 because xyz",
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
				Code: "// @ts-check: TS1234",
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
				Code: "// @ts-check    : TS1234 because xyz",
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
				Code: "// @ts-check 👨‍👩‍👧‍👦",
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