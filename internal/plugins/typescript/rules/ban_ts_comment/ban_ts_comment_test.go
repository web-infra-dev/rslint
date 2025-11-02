package ban_ts_comment

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestBanTsCommentRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &BanTsCommentRule, []rule_tester.ValidTestCase{
		// Comment containing @ts-expect-error without directive formatting
		{Code: "// Suppress ts-expect-error\nconst a = 0;"},

		// ts-expect-error - disabled
		{Code: "// @ts-expect-error\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": false}},

		// ts-expect-error - allow with description
		{Code: "// @ts-expect-error: Suppress error\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
		{Code: "// @ts-expect-error Suppress error\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
		{Code: "/* @ts-expect-error: Suppress error */\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
		{Code: "/* @ts-expect-error Suppress error */\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
		{Code: "/* @ts-expect-error\n * Multiline description\n */\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},

		// ts-expect-error - minimum description length
		{Code: "// @ts-expect-error: This is a very long description that exceeds minimum\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "minimumDescriptionLength": 10}},
		{Code: "// @ts-expect-error 0123456789012345678901\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "minimumDescriptionLength": 21}},

		// ts-expect-error - description format
		{Code: "// @ts-expect-error: TS1234 because reasons\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},
		{Code: "// @ts-expect-error: TS2345 because type mismatch\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}}},

		// ts-expect-error - Unicode/emoji descriptions
		{Code: "// @ts-expect-error: 💩💩💩💩\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "minimumDescriptionLength": 4}},

		// ts-ignore - disabled
		{Code: "// @ts-ignore\nconst a = 0;", Options: map[string]interface{}{"ts-ignore": false}},

		// ts-ignore - allow with description
		{Code: "// @ts-ignore: Suppress error\nconst a = 0;", Options: map[string]interface{}{"ts-ignore": "allow-with-description"}},
		{Code: "// @ts-ignore Suppress error\nconst a = 0;", Options: map[string]interface{}{"ts-ignore": "allow-with-description"}},
		{Code: "/* @ts-ignore: Suppress error */\nconst a = 0;", Options: map[string]interface{}{"ts-ignore": "allow-with-description"}},

		// ts-nocheck - disabled
		{Code: "// @ts-nocheck\nconst a = 0;", Options: map[string]interface{}{"ts-nocheck": false}},

		// ts-nocheck - allow with description
		{Code: "// @ts-nocheck: Suppress all errors\nconst a = 0;", Options: map[string]interface{}{"ts-nocheck": "allow-with-description"}},
		{Code: "// @ts-nocheck Suppress all errors\nconst a = 0;", Options: map[string]interface{}{"ts-nocheck": "allow-with-description"}},
		{Code: "/// @ts-nocheck: Suppress all errors\nconst a = 0;", Options: map[string]interface{}{"ts-nocheck": "allow-with-description"}},

		// ts-check - disabled (default)
		{Code: "// @ts-check\nconst a = 0;"},
		{Code: "// @ts-check\nconst a = 0;", Options: map[string]interface{}{"ts-check": false}},

		// ts-check - allow with description
		{Code: "// @ts-check: Enable checking\nconst a = 0;", Options: map[string]interface{}{"ts-check": "allow-with-description"}},

		// Multi-line comment with directive not on last line
		{Code: "/*\n@ts-expect-error\nSome other text\n*/\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": true}},

		// JSDoc-style comments
		{Code: "/**\n * @ts-expect-error: Description\n */\nconst a = 0;", Options: map[string]interface{}{"ts-expect-error": "allow-with-description"}},
	}, []rule_tester.InvalidTestCase{
		// ts-expect-error - basic violation
		{
			Code: "// @ts-expect-error\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveComment"},
			},
		},
		{
			Code: "/* @ts-expect-error */\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveComment"},
			},
		},

		// ts-expect-error - requires description
		{
			Code: "// @ts-expect-error\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentRequiresDescription"},
			},
		},
		{
			Code: "/* @ts-expect-error */\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentRequiresDescription"},
			},
		},

		// ts-expect-error - description too short
		{
			Code: "// @ts-expect-error: ab\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"},
			},
		},
		{
			Code: "// @ts-expect-error 0123456789012345678\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "minimumDescriptionLength": 21},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"},
			},
		},

		// ts-expect-error - description format mismatch
		{
			Code: "// @ts-expect-error: because reasons\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"},
			},
		},
		{
			Code: "// @ts-expect-error:TS1234 because reasons\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"},
			},
		},
		{
			Code: "// @ts-expect-error: TS because reasons\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": map[string]interface{}{"descriptionFormat": "^: TS\\d+ because .+$"}},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"},
			},
		},

		// ts-expect-error - Unicode/emoji too short
		{
			Code: "// @ts-expect-error: 💩💩💩\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": "allow-with-description", "minimumDescriptionLength": 4},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentDescriptionNotMatchPattern"},
			},
		},

		// ts-ignore - suggests ts-expect-error
		{
			Code: "// @ts-ignore\nconst a = 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsIgnoreInsteadOfExpectError"},
			},
		},
		{
			Code: "/* @ts-ignore */\nconst a = 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsIgnoreInsteadOfExpectError"},
			},
		},
		{
			Code: "/* @ts-ignore with description */\nconst a = 0;",
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsIgnoreInsteadOfExpectError"},
			},
		},

		// ts-nocheck - basic violation
		{
			Code: "// @ts-nocheck\nconst a = 0;",
			Options: map[string]interface{}{"ts-nocheck": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveComment"},
			},
		},
		{
			Code: "/// @ts-nocheck\nconst a = 0;",
			Options: map[string]interface{}{"ts-nocheck": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveComment"},
			},
		},

		// ts-nocheck - requires description
		{
			Code: "// @ts-nocheck\nconst a = 0;",
			Options: map[string]interface{}{"ts-nocheck": "allow-with-description"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentRequiresDescription"},
			},
		},

		// ts-check - basic violation
		{
			Code: "// @ts-check\nconst a = 0;",
			Options: map[string]interface{}{"ts-check": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveComment"},
			},
		},

		// ts-check - requires description
		{
			Code: "// @ts-check\nconst a = 0;",
			Options: map[string]interface{}{"ts-check": "allow-with-description"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveCommentRequiresDescription"},
			},
		},

		// Multi-line comments
		{
			Code: "/*\n@ts-expect-error\n*/\nconst a = 0;",
			Options: map[string]interface{}{"ts-expect-error": true},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "tsDirectiveComment"},
			},
		},
	})
}
