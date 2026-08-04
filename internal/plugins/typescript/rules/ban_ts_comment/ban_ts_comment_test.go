package ban_ts_comment

import (
	"regexp"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/parser"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
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

func TestDirectiveMatchersMatchRegexOracle(t *testing.T) {
	singleLineDirectiveRegex := regexp.MustCompile(`^/{2,}\s*@ts-(expect-error|ignore)\b`)
	singleLinePragmaRegex := regexp.MustCompile(`^///?\s*@ts-(check|nocheck)\b`)
	multiLineLastLineRegex := regexp.MustCompile(`^\s*(?:[/*])*\s*@ts-(expect-error|ignore)\b`)

	directiveFromSuffix := func(suffix string) (directiveKind, bool) {
		switch suffix {
		case "expect-error":
			return directiveExpectError, true
		case "ignore":
			return directiveIgnore, true
		case "nocheck":
			return directiveNocheck, true
		case "check":
			return directiveCheck, true
		default:
			return 0, false
		}
	}
	oracleSingleLine := func(commentText string) (directiveKind, int, bool) {
		for _, expression := range []*regexp.Regexp{singleLineDirectiveRegex, singleLinePragmaRegex} {
			match := expression.FindStringSubmatchIndex(commentText)
			if match == nil {
				continue
			}
			directive, ok := directiveFromSuffix(commentText[match[2]:match[3]])
			return directive, match[1], ok
		}
		return 0, 0, false
	}
	oracleMultiLine := func(commentText string) (directiveKind, int, int, bool) {
		contentStart := 2
		contentEnd := len(commentText) - 2
		lastLineStart := contentStart
		if lineBreak := strings.LastIndexByte(commentText[contentStart:contentEnd], '\n'); lineBreak >= 0 {
			lastLineStart += lineBreak + 1
		}
		match := multiLineLastLineRegex.FindStringSubmatchIndex(commentText[lastLineStart:contentEnd])
		if match == nil {
			return 0, 0, 0, false
		}
		directive, ok := directiveFromSuffix(commentText[lastLineStart+match[2] : lastLineStart+match[3]])
		return directive, lastLineStart + match[1], contentEnd, ok
	}

	directives := []string{
		"@ts-expect-error",
		"@ts-ignore",
		"@ts-nocheck",
		"@ts-check",
		"@ts-unknown",
	}
	whitespace := []string{"", " ", "\t", "\n", "\f", "\r", "\v", " \t\f\r"}
	suffixes := []string{"", " ", ": why", "-why", "/", "_suffix", "0", "A", "é"}
	configs := directiveConfigs{}
	for directive := directiveKind(0); directive < directiveCount; directive++ {
		configs[directive].Enabled = true
	}

	for _, slashes := range []string{"", "/", "//", "///", "////", "/////", "//x"} {
		for _, space := range whitespace {
			for _, directiveText := range directives {
				for _, suffix := range suffixes {
					commentText := slashes + space + directiveText + suffix
					wantDirective, wantDescriptionStart, wantOK := oracleSingleLine(commentText)
					gotDirective, gotDescriptionStart, gotOK := matchSingleLineDirective(commentText)
					if gotOK != wantOK || gotDirective != wantDirective || gotDescriptionStart != wantDescriptionStart {
						t.Fatalf(
							"single-line match for %q = (%d, %d, %t), want (%d, %d, %t)",
							commentText,
							gotDirective,
							gotDescriptionStart,
							gotOK,
							wantDirective,
							wantDescriptionStart,
							wantOK,
						)
					}
					if wantOK && !containsEnabledDirective(commentText, &configs) {
						t.Fatalf("source-text gate rejected matching single-line directive %q", commentText)
					}
				}
			}
		}
	}

	lastLinePrefixes := []string{"", " ", "\t", "\f", "\r", "\v", "*", "**", "/", "/*", " */ ", "\t/*\r"}
	for _, previousLines := range []string{"", "header\n", "header\r\n", "\n"} {
		for _, prefix := range lastLinePrefixes {
			for _, directiveText := range directives {
				for _, suffix := range suffixes {
					commentText := "/*" + previousLines + prefix + directiveText + suffix + "*/"
					wantDirective, wantDescriptionStart, wantContentEnd, wantOK := oracleMultiLine(commentText)
					gotDirective, gotDescriptionStart, gotContentEnd, gotOK := matchMultiLineDirective(commentText)
					if gotOK != wantOK || gotDirective != wantDirective || gotDescriptionStart != wantDescriptionStart || gotContentEnd != wantContentEnd {
						t.Fatalf(
							"multi-line match for %q = (%d, %d, %d, %t), want (%d, %d, %d, %t)",
							commentText,
							gotDirective,
							gotDescriptionStart,
							gotContentEnd,
							gotOK,
							wantDirective,
							wantDescriptionStart,
							wantContentEnd,
							wantOK,
						)
					}
					if wantOK && !containsEnabledDirective(commentText, &configs) {
						t.Fatalf("source-text gate rejected matching multi-line directive %q", commentText)
					}
				}
			}
		}
	}
}

func TestBanTsCommentSuggestionDemand(t *testing.T) {
	const source = "// @ts-ignore"
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/ban-ts-comment-suggestion.ts",
		Path:     "/ban-ts-comment-suggestion.ts",
	}, source, core.ScriptKindTS)

	run := func(demand rule.EditDemand) rule.RuleDiagnostic {
		t.Helper()
		comments := rule.NewCommentStore(sourceFile)
		var diagnostics []rule.RuleDiagnostic
		ctx := rule.RuleContext{
			SourceFile:     sourceFile,
			Comments:       comments,
			DisableManager: rule.NewDisableManager(sourceFile, comments),
		}.WithDiagnosticConsumer(BanTsCommentRule.Name, rule.SeverityError, rule.DiagnosticConsumer{
			Demand: demand,
			Report: func(diagnostic rule.RuleDiagnostic) {
				diagnostics = append(diagnostics, diagnostic)
			},
		})
		BanTsCommentRule.Run(ctx, nil)
		if len(diagnostics) != 1 {
			t.Fatalf("demand %d produced %d diagnostics, want 1", demand, len(diagnostics))
		}
		return diagnostics[0]
	}

	for _, demand := range []rule.EditDemand{
		rule.EditDemandNone,
		rule.EditDemandAutofix,
		rule.EditDemandSuggestion,
		rule.EditDemandAll,
	} {
		diagnostic := run(demand)
		if diagnostic.Message.Id != "tsIgnoreInsteadOfExpectError" || diagnostic.Range.Pos() != 0 || diagnostic.Range.End() != len(source) {
			t.Fatalf("demand %d changed the diagnostic: %#v", demand, diagnostic)
		}
		if diagnostic.FixesPtr != nil {
			t.Fatalf("demand %d unexpectedly materialized an autofix", demand)
		}
		if demand&rule.EditDemandSuggestion == 0 {
			if diagnostic.Suggestions != nil {
				t.Fatalf("demand %d materialized suggestions", demand)
			}
			continue
		}
		if diagnostic.Suggestions == nil || len(*diagnostic.Suggestions) != 1 {
			t.Fatalf("demand %d suggestions = %#v, want one", demand, diagnostic.Suggestions)
		}
		suggestion := (*diagnostic.Suggestions)[0]
		if suggestion.Message.Id != "replaceTsIgnoreWithTsExpectError" || len(suggestion.Fixes()) != 1 {
			t.Fatalf("demand %d suggestion = %#v, want one replacement", demand, suggestion)
		}
		fix := suggestion.Fixes()[0]
		if fix.Range.Pos() != 3 || fix.Range.End() != len(source) || fix.Text != "@ts-expect-error" {
			t.Fatalf("demand %d fix = %#v, want @ts-ignore replacement", demand, fix)
		}
	}
}
