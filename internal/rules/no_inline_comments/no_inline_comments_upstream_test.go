// TestNoInlineCommentsUpstream migrates the full valid/invalid suite from
// upstream eslint/tests/lib/rules/no-inline-comments.js 1:1. Position
// assertions cover line/column for every invalid case. rslint-specific
// lock-in cases live in the no_inline_comments_extras_test.go file.
package no_inline_comments

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const unexpectedInlineCommentText = "Unexpected comment inline with code."

func TestNoInlineCommentsUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInlineCommentsRule,
		[]rule_tester.ValidTestCase{
			// ---- valid ----
			{Code: "// A valid comment before code\nvar a = 1;"},
			{Code: "var a = 2;\n// A valid comment after code"},
			{Code: "// A solitary comment"},
			{Code: "var a = 1; // eslint-disable-line no-debugger"},
			{Code: "var a = 1; /* eslint-disable-line no-debugger */"},
			{Code: "foo(); /* global foo */"},
			{Code: "foo(); /* globals foo */"},
			{Code: "var foo; /* exported foo */"},

			// JSX exception
			{Code: "var a = (\n            <div>\n            {/*comment*/}\n            </div>\n        )", Tsx: true},
			{Code: "var a = (\n            <div>\n            { /* comment */ }\n            <h1>Some heading</h1>\n            </div>\n        )", Tsx: true},
			{Code: "var a = (\n            <div>\n            {// comment\n            }\n            </div>\n        )", Tsx: true},
			{Code: "var a = (\n            <div>\n            { // comment\n            }\n            </div>\n        )", Tsx: true},
			{Code: "var a = (\n            <div>\n            {/* comment 1 */\n            /* comment 2 */}\n            </div>\n        )", Tsx: true},
			{Code: "var a = (\n            <div>\n            {/*\n              * comment 1\n              */\n             /*\n              * comment 2\n              */}\n            </div>\n        )", Tsx: true},
			{Code: "var a = (\n            <div>\n            {/*\n               multi\n               line\n               comment\n            */}\n            </div>\n        )", Tsx: true},
			{
				Code:            `import(/* webpackChunkName: "my-chunk-name" */ './locale/en');`,
				Options:         []any{map[string]any{"ignorePattern": `(?:webpackChunkName):\s.+`}},
				LanguageOptions: rule.LanguageOptions{ECMAVersion: 2020},
			},
			{
				Code:    "var foo = 2; // Note: This comment is legal.",
				Options: []any{map[string]any{"ignorePattern": "Note: "}},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- invalid ----
			{
				Code: "var a = 1; /*A block comment inline after code*/",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 1, Column: 12, EndLine: 1, EndColumn: 49},
				},
			},
			{
				Code: "/*A block comment inline before code*/ var a = 2;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 1, Column: 1, EndLine: 1, EndColumn: 39},
				},
			},
			{
				Code:    "/* something */ var a = 2;",
				Options: []any{map[string]any{"ignorePattern": "otherthing"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 1, Column: 1, EndLine: 1, EndColumn: 16},
				},
			},
			{
				Code: "var a = 3; //A comment inline with code",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 1, Column: 12, EndLine: 1, EndColumn: 40},
				},
			},
			{
				Code: "var a = 3; // someday use eslint-disable-line here",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 1, Column: 12, EndLine: 1, EndColumn: 51},
				},
			},
			{
				Code:    "var a = 3; // other line comment",
				Options: []any{map[string]any{"ignorePattern": "something"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 1, Column: 12, EndLine: 1, EndColumn: 33},
				},
			},
			{
				Code: "var a = 4;\n/**A\n * block\n * comment\n * inline\n * between\n * code*/ var foo = a;",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 1, EndLine: 7, EndColumn: 10},
				},
			},
			{
				Code: "var a = \n{/**/}",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 2, EndLine: 2, EndColumn: 6},
				},
			},

			// JSX
			{
				Code: "var a = (\n                <div>{/* comment */}</div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 23, EndLine: 2, EndColumn: 36},
				},
			},
			{
				Code: "var a = (\n                <div>{// comment\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 23, EndLine: 2, EndColumn: 33},
				},
			},
			{
				Code: "var a = (\n                <div>{/* comment */\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 23, EndLine: 2, EndColumn: 36},
				},
			},
			{
				Code: "var a = (\n                <div>{/*\n                       * comment\n                       */\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 23, EndLine: 4, EndColumn: 26},
				},
			},
			{
				Code: "var a = (\n                <div>{/*\n                       * comment\n                       */}\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 23, EndLine: 4, EndColumn: 26},
				},
			},
			{
				Code: "var a = (\n                <div>{/*\n                       * comment\n                       */}</div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 2, Column: 23, EndLine: 4, EndColumn: 26},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {/*\n                  * comment\n                  */}</div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 18, EndLine: 5, EndColumn: 21},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                 /*\n                  * comment\n                  */}</div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 18, EndLine: 6, EndColumn: 21},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                /* comment */}</div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 17, EndLine: 4, EndColumn: 30},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {b/* comment */}\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 19, EndLine: 3, EndColumn: 32},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {/* comment */b}\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 18, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {// comment\n                    b\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 18, EndLine: 3, EndColumn: 28},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {/* comment */\n                    b\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 18, EndLine: 3, EndColumn: 31},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {/*\n                  * comment\n                  */\n                    b\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 18, EndLine: 5, EndColumn: 21},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    b// comment\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 22, EndLine: 4, EndColumn: 32},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    /* comment */b\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 21, EndLine: 4, EndColumn: 34},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    b/* comment */\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 22, EndLine: 4, EndColumn: 35},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    b\n                /*\n                 * comment\n                 */}\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 5, Column: 17, EndLine: 7, EndColumn: 20},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    b\n                /* comment */}\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 5, Column: 17, EndLine: 5, EndColumn: 30},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    { /* this is an empty object literal, not braces for js code! */ }\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 23, EndLine: 4, EndColumn: 85},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    {// comment\n                    }\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 22, EndLine: 4, EndColumn: 32},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    {\n                    /* comment */}\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 5, Column: 21, EndLine: 5, EndColumn: 34},
				},
			},
			{
				Code: "var a = (\n                <div>\n                { /* two comments on the same line... */ /* ...are not allowed, same as with a non-JSX code */}\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 19, EndLine: 3, EndColumn: 57},
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 3, Column: 58, EndLine: 3, EndColumn: 111},
				},
			},
			{
				Code: "var a = (\n                <div>\n                {\n                    /* overlapping\n                    */ /*\n                       lines */\n                }\n                </div>\n            )",
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 4, Column: 21, EndLine: 5, EndColumn: 23},
					{MessageId: "unexpectedInlineComment", Message: unexpectedInlineCommentText, Line: 5, Column: 24, EndLine: 6, EndColumn: 32},
				},
			},
		},
	)
}
