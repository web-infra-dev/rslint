package ban_ts_comment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBanTsCommentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BanTsCommentRule, []rule_tester.ValidTestCase{
		// ========================
		// Edge cases: directive-like text in non-comment contexts
		// ========================
		// String literals should NOT be flagged
		{Code: "const c = \"// @ts-ignore\";"},
		{Code: "const c = \"/* @ts-expect-error */\";"},
		// Template literals should NOT be flagged
		{Code: "const c = `// @ts-ignore`;"},
		// Trailing comment after code (ts-ignore is default banned → tsIgnoreInsteadOfExpectError won't fire here because ts-ignore default is true and it suggests expect-error; but ts-expect-error default is allow-with-description, so with description it's valid)
		{Code: "const x = 1; // @ts-expect-error: suppress this"},

		// ========================
		// ts-expect-error: valid
		// ========================
		// Comment containing @ts-expect-error without directive formatting
		{Code: "// just a comment containing @ts-expect-error somewhere"},
		// Block comment with directive NOT on the last line → not a directive
		{Code: "/*\n @ts-expect-error running with long description in a block\n*/"},
		{Code: "/* @ts-expect-error not on the last line\n */"},
		{Code: "/**\n * @ts-expect-error not on the last line\n */"},
		{Code: "/* not on the last line\n * @ts-expect-error\n */"},
		{Code: "/* @ts-expect-error\n * not on the last line */"},
		// Disabled via option
		{Code: "// @ts-expect-error", Options: map[string]interface{}{"ts-expect-error": false}},
		// allow-with-description with sufficient description
		{Code: "// @ts-expect-error here is why the error is expected", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
		{Code: "/*\n * @ts-expect-error here is why the error is expected */", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
		// minimumDescriptionLength
		{Code: "// @ts-expect-error exactly 21 characters", Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-expect-error": "allow-with-description"}},
		{Code: "/*\n * @ts-expect-error exactly 21 characters*/", Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-expect-error": "allow-with-description"}},
		// descriptionFormat
		{Code: "// @ts-expect-error: TS1234 because xyz", Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},
		{Code: "/*\n * @ts-expect-error: TS1234 because xyz */", Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},
		// Emoji: 3 family emojis = 3 grapheme clusters ≥ 3 minimum
		{Code: "// @ts-expect-error 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},

		// ========================
		// ts-ignore: valid
		// ========================
		{Code: "// just a comment containing @ts-ignore somewhere"},
		// Disabled
		{Code: "// @ts-ignore", Options: map[string]interface{}{"ts-ignore": false}},
		// allow-with-description
		{Code: "// @ts-ignore I think that I am exempted from any need to follow the rules!", Options: map[string]interface{}{"ts-ignore": "allow-with-description"}},
		{Code: "/*\n @ts-ignore running with long description in a block\n*/", Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"}},
		// Block comment with directive NOT on last line → not a directive
		{Code: "/*\n @ts-ignore\n*/"},
		{Code: "/* @ts-ignore not on the last line\n */"},
		{Code: "/**\n * @ts-ignore not on the last line\n */"},
		{Code: "/* not on the last line\n * @ts-expect-error\n */"},
		{Code: "/* @ts-ignore\n * not on the last line */"},
		// descriptionFormat
		{Code: "// @ts-ignore: TS1234 because xyz", Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},
		// Emoji: 3 family emojis = 3 grapheme clusters ≥ 3 minimum
		{Code: "// @ts-ignore 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦", Options: map[string]interface{}{"ts-ignore": "allow-with-description"}},
		// Block comment with description on last line
		{Code: "/*\n * @ts-ignore here is why the error is expected */", Options: map[string]interface{}{"ts-ignore": "allow-with-description"}},
		// minimumDescriptionLength
		{Code: "// @ts-ignore exactly 21 characters", Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"}},
		{Code: "/*\n * @ts-ignore exactly 21 characters*/", Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-ignore": "allow-with-description"}},
		{Code: "/*\n * @ts-ignore: TS1234 because xyz */", Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},

		// ========================
		// ts-nocheck: valid
		// ========================
		{Code: "// just a comment containing @ts-nocheck somewhere"},
		// Disabled
		{Code: "// @ts-nocheck", Options: map[string]interface{}{"ts-nocheck": false}},
		// allow-with-description
		{Code: "// @ts-nocheck no doubt, people will put nonsense here from time to time just to get the rule to stop reporting, perhaps even long messages with other nonsense in them like other // @ts-nocheck or // @ts-ignore things", Options: map[string]interface{}{"ts-nocheck": "allow-with-description"}},
		{Code: "/*\n @ts-nocheck running with long description in a block\n*/", Options: map[string]interface{}{"minimumDescriptionLength": 21, "ts-nocheck": "allow-with-description"}},
		// descriptionFormat
		{Code: "// @ts-nocheck: TS1234 because xyz", Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},
		// Emoji: 3 family emojis = 3 grapheme clusters ≥ 3 minimum
		{Code: "// @ts-nocheck 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦", Options: map[string]interface{}{"ts-nocheck": "allow-with-description"}},
		// 4+ slashes: not a pragma comment
		{Code: "//// @ts-nocheck - pragma comments may contain 2 or 3 leading slashes"},
		// Block comments with ts-nocheck are NOT directives (pragma-only)
		{Code: "/**\n @ts-nocheck\n*/"},
		{Code: "/*\n @ts-nocheck\n*/"},
		{Code: "/** @ts-nocheck */"},
		{Code: "/* @ts-nocheck */"},
		// ts-nocheck after first statement: not effective, not reported
		{Code: "const a = 1;\n\n// @ts-nocheck - should not be reported\n\n// TS error is not actually suppressed\nconst b: string = a;"},

		// ========================
		// Default config full user flow: @ts-ignore → prefer → @ts-expect-error with desc → pass
		// ========================
		// Step 3: @ts-expect-error with sufficient description → valid (default: allow-with-description, minLength 3)
		{Code: "// @ts-expect-error: some valid reason here"},

		// ========================
		// ts-check: valid
		// ========================
		{Code: "// just a comment containing @ts-check somewhere"},
		// Default: ts-check is not banned
		{Code: "// @ts-check"},
		// Disabled
		{Code: "// @ts-check", Options: map[string]interface{}{"ts-check": false}},
		// allow-with-description
		{Code: "// @ts-check with a description and also with a no-op // @ts-ignore", Options: map[string]interface{}{"minimumDescriptionLength": 3, "ts-check": "allow-with-description"}},
		// descriptionFormat
		{Code: "// @ts-check: TS1234 because xyz", Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},
		// Emoji: 3 family emojis = 3 grapheme clusters ≥ 3 minimum
		{Code: "// @ts-check 👨‍👩‍👧‍👦👨‍👩‍👧‍👦👨‍👩‍👧‍👦", Options: map[string]interface{}{"ts-check": "allow-with-description"}},
		// 4+ slashes: not a pragma comment
		{Code: "//// @ts-check - pragma comments may contain 2 or 3 leading slashes", Options: map[string]interface{}{"ts-check": true}},
		// Block comments with ts-check are NOT directives (pragma-only)
		{Code: "/**\n @ts-check\n*/", Options: map[string]interface{}{"ts-check": true}},
		{Code: "/*\n @ts-check\n*/", Options: map[string]interface{}{"ts-check": true}},
		{Code: "/** @ts-check */", Options: map[string]interface{}{"ts-check": true}},
		{Code: "/* @ts-check */", Options: map[string]interface{}{"ts-check": true}},
	}, []rule_tester.InvalidTestCase{
		// ========================
		// ts-expect-error: invalid
		// ========================
		// Basic violation with ts-expect-error: true
		{
			Code:    "// @ts-expect-error",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		{
			Code:    "/* @ts-expect-error */",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// Block comment: directive on last line
		{
			Code:    "/*\n@ts-expect-error */",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		{
			Code:    "/** on the last line\n  @ts-expect-error */",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		{
			Code:    "/** on the last line\n * @ts-expect-error */",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// Block comment: description too short
		{
			Code:    "/**\n * @ts-expect-error: TODO */",
			Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		{
			Code:    "/**\n * @ts-expect-error: TS1234 because xyz */",
			Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Block comment: description format mismatch
		{
			Code:    "/**\n * @ts-expect-error: TS1234 */",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		{
			Code:    "/**\n * @ts-expect-error    : TS1234 */",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		// Block comment: emoji too short (1 family emoji = 1 grapheme < 3)
		{
			Code:    "/**\n * @ts-expect-error 👨‍👩‍👧‍👦 */",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// JSDoc-style block comment
		{
			Code:    "/** @ts-expect-error */",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// Single-line with description but banned completely
		{
			Code:    "// @ts-expect-error: Suppress next line",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// 5 slashes: still a directive for expect-error
		{
			Code:    "/////@ts-expect-error: Suppress next line",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// Nested in code
		{
			Code:    "if (false) {\n  // @ts-expect-error: Unreachable code error\n  console.log('hello');\n}",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// allow-with-description: no description
		{
			Code:    "// @ts-expect-error",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description too short for custom minimum
		{
			Code:    "// @ts-expect-error: TODO",
			Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		{
			Code:    "// @ts-expect-error: TS1234 because xyz",
			Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description format mismatch
		{
			Code:    "// @ts-expect-error: TS1234",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		{
			Code:    "// @ts-expect-error    : TS1234 because xyz",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		// Emoji: 1 family emoji = 1 grapheme cluster < 3
		{
			Code:    "// @ts-expect-error 👨‍👩‍👧‍👦",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},

		// ========================
		// Default config full user flow
		// ========================
		// Step 1: @ts-ignore with defaults → prefer ts-expect-error
		{
			Code:   "// @ts-ignore",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "// @ts-expect-error"}}}},
		},
		// Step 2: @ts-expect-error (no desc) with defaults → requires description
		{
			Code:   "// @ts-expect-error",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// (Step 3: @ts-expect-error with desc → valid, covered in valid section above)

		// ========================
		// ts-ignore: invalid — ts-expect-error config combinations
		// ========================
		// Both banned → tsDirectiveComment (no contradictory prefer)
		{
			Code:    "// @ts-ignore",
			Options: map[string]interface{}{"ts-expect-error": true, "ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// ts-expect-error: allow-with-description → prefer (makes sense)
		{
			Code:    "// @ts-ignore",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "// @ts-expect-error"}}}},
		},
		// ts-expect-error: descriptionFormat → prefer (makes sense, expect-error is allowed with format)
		{
			Code:    "// @ts-ignore",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+"}, "ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "// @ts-expect-error"}}}},
		},
		{
			Code:    "/* @ts-ignore */",
			Options: map[string]interface{}{"ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/* @ts-expect-error */"}}}},
		},
		// Block comment: directive on last line
		{
			Code:    "/*\n @ts-ignore */",
			Options: map[string]interface{}{"ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/*\n @ts-expect-error */"}}}},
		},
		// Block comment: duplicate @ts-ignore — suggestion must target the LAST one (on last line)
		{
			Code:    "/* @ts-ignore\n * @ts-ignore */",
			Options: map[string]interface{}{"ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/* @ts-ignore\n * @ts-expect-error */"}}}},
		},
		{
			Code:    "/** on the last line\n  @ts-ignore */",
			Options: map[string]interface{}{"ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/** on the last line\n  @ts-expect-error */"}}}},
		},
		{
			Code:    "/** on the last line\n * @ts-ignore */",
			Options: map[string]interface{}{"ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/** on the last line\n * @ts-expect-error */"}}}},
		},
		{
			Code:    "/** @ts-ignore */",
			Options: map[string]interface{}{"ts-expect-error": false, "ts-ignore": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/** @ts-expect-error */"}}}},
		},
		// Block comment: ts-ignore banned (default), with description
		{
			Code:    "/**\n * @ts-ignore: TODO */",
			Options: map[string]interface{}{"minimumDescriptionLength": 10, "ts-expect-error": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/**\n * @ts-expect-error: TODO */"}}}},
		},
		{
			Code:    "/**\n * @ts-ignore: TS1234 because xyz */",
			Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/**\n * @ts-expect-error: TS1234 because xyz */"}}}},
		},
		// Single-line with description, default ts-ignore: true
		{
			Code:   "// @ts-ignore: Suppress next line",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "// @ts-expect-error: Suppress next line"}}}},
		},
		// 5 slashes: still a directive for ts-ignore
		{
			Code:   "/////@ts-ignore: Suppress next line",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "/////@ts-expect-error: Suppress next line"}}}},
		},
		// Nested in code
		{
			Code:   "if (false) {\n  // @ts-ignore: Unreachable code error\n  console.log('hello');\n}",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsIgnoreInsteadOfExpectError", Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "replaceTsIgnoreWithTsExpectError", Output: "if (false) {\n  // @ts-expect-error: Unreachable code error\n  console.log('hello');\n}"}}}},
		},
		// allow-with-description: no description
		{
			Code:    "// @ts-ignore",
			Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// allow-with-description: only whitespace
		{
			Code:    "// @ts-ignore         ",
			Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// allow-with-description: description too short (1 char < 3)
		{
			Code:    "// @ts-ignore    .",
			Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description too short for custom minimum
		{
			Code:    "// @ts-ignore: TS1234 because xyz",
			Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description format mismatch
		{
			Code:    "// @ts-ignore: TS1234",
			Options: map[string]interface{}{"ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		{
			Code:    "// @ts-ignore    : TS1234 because xyz",
			Options: map[string]interface{}{"ts-ignore": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		// Emoji: 1 family emoji = 1 grapheme cluster < 3
		{
			Code:    "// @ts-ignore 👨‍👩‍👧‍👦",
			Options: map[string]interface{}{"ts-ignore": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},

		// ========================
		// ts-nocheck: invalid
		// ========================
		// Default: ts-nocheck is banned
		{
			Code:    "// @ts-nocheck",
			Options: map[string]interface{}{"ts-nocheck": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// Default options (ts-nocheck: true)
		{
			Code:   "// @ts-nocheck",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// With description, still banned
		{
			Code:   "// @ts-nocheck: Suppress next line",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// allow-with-description: no description
		{
			Code:    "// @ts-nocheck",
			Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description too short for custom minimum
		{
			Code:    "// @ts-nocheck: TS1234 because xyz",
			Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description format mismatch
		{
			Code:    "// @ts-nocheck: TS1234",
			Options: map[string]interface{}{"ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		{
			Code:    "// @ts-nocheck    : TS1234 because xyz",
			Options: map[string]interface{}{"ts-nocheck": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		// Emoji: 1 family emoji = 1 grapheme cluster < 3
		{
			Code:    "// @ts-nocheck 👨‍👩‍👧‍👦",
			Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Comment before first statement but offset column
		{
			Code:   " // @ts-nocheck\nconst a: true = false;",
			Errors: []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},

		// ========================
		// ts-check: invalid
		// ========================
		// Banned
		{
			Code:    "// @ts-check",
			Options: map[string]interface{}{"ts-check": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// With description, still banned
		{
			Code:    "// @ts-check: Suppress next line",
			Options: map[string]interface{}{"ts-check": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// Nested in code
		{
			Code:    "if (false) {\n  // @ts-check: Unreachable code error\n  console.log('hello');\n}",
			Options: map[string]interface{}{"ts-check": true},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveComment"}},
		},
		// allow-with-description: no description
		{
			Code:    "// @ts-check",
			Options: map[string]interface{}{"ts-check": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description too short for custom minimum
		{
			Code:    "// @ts-check: TS1234 because xyz",
			Options: map[string]interface{}{"minimumDescriptionLength": 25, "ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
		// Description format mismatch
		{
			Code:    "// @ts-check: TS1234",
			Options: map[string]interface{}{"ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		{
			Code:    "// @ts-check    : TS1234 because xyz",
			Options: map[string]interface{}{"ts-check": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"}},
		},
		// Emoji: 1 family emoji = 1 grapheme cluster < 3
		{
			Code:    "// @ts-check 👨‍👩‍👧‍👦",
			Options: map[string]interface{}{"ts-check": "allow-with-description"},
			Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "tsDirectiveCommentRequiresDescription"}},
		},
	})
}
