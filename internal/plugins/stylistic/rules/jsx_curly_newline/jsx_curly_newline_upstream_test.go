// TestJsxCurlyNewlineUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/jsx-curly-newline/
// jsx-curly-newline.test.ts 1:1. Upstream asserts only messageId on its invalid
// cases; the line/column/endLine/endColumn and exact message text here are
// computed from the source each case carries. rslint-specific lock-in cases
// live in jsx_curly_newline_extras_test.go.
package jsx_curly_newline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxCurlyNewlineUpstream(t *testing.T) {
	consistent := []interface{}{"consistent"}
	never := []interface{}{"never"}
	multilineRequire := []interface{}{map[string]interface{}{"singleline": "consistent", "multiline": "require"}}

	const (
		expectedAfter    = "Expected newline after '{'."
		expectedBefore   = "Expected newline before '}'."
		unexpectedAfter  = "Unexpected newline after '{'."
		unexpectedBefore = "Unexpected newline before '}'."
	)

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxCurlyNewlineRule, []rule_tester.ValidTestCase{
		// ---- consistent (default) ----
		{Code: "<div>{foo}</div>", Tsx: true, Options: consistent},
		{Code: "<div>\n          {\n            foo\n          }\n        </div>", Tsx: true, Options: consistent},
		{Code: "<div>\n          { foo &&\n            foo.bar }\n        </div>", Tsx: true, Options: consistent},
		{Code: "<div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>", Tsx: true, Options: consistent},
		{Code: "<div foo={\n          bar\n        } />", Tsx: true, Options: consistent},
		// ---- { singleline: consistent, multiline: require } ----
		{Code: "<div>{foo}</div>", Tsx: true, Options: multilineRequire},
		{Code: "<div foo={bar} />", Tsx: true, Options: multilineRequire},
		{Code: "<div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>", Tsx: true, Options: multilineRequire},
		{Code: "<div>\n          {\n            foo\n          }\n        </div>", Tsx: true, Options: multilineRequire},
		// ---- never ----
		{Code: "<div>{foo}</div>", Tsx: true, Options: never},
		{Code: "<div foo={bar} />", Tsx: true, Options: never},
		{Code: "<div>\n          { foo &&\n            foo.bar }\n        </div>", Tsx: true, Options: never},
	}, []rule_tester.InvalidTestCase{
		// ---- consistent: newline before } but not after { ----
		{
			Code:    "<div>\n          { foo \n}\n        </div>",
			Tsx:     true,
			Options: consistent,
			Output:  []string{"<div>\n          { foo}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 3, Column: 1, EndLine: 3, EndColumn: 2},
			},
		},
		{
			Code:    "<div>\n          { foo &&\n            foo.bar \n}\n        </div>",
			Tsx:     true,
			Options: consistent,
			Output:  []string{"<div>\n          { foo &&\n            foo.bar}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		{
			Code:    "<div>\n          { foo &&\n            bar\n          }\n        </div>",
			Tsx:     true,
			Options: consistent,
			Output:  []string{"<div>\n          { foo &&\n            bar}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 4, Column: 11, EndLine: 4, EndColumn: 12},
			},
		},
		// ---- { multiline: require }: newline before } unexpected on single-line expr ----
		{
			Code:    "<div>{foo\n}</div>",
			Tsx:     true,
			Options: multilineRequire,
			Output:  []string{"<div>{foo}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 2, Column: 1, EndLine: 2, EndColumn: 2},
			},
		},
		{
			Code:    "<div>{\nfoo}</div>",
			Tsx:     true,
			Options: multilineRequire,
			Output:  []string{"<div>{\nfoo\n}</div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 2, Column: 4, EndLine: 2, EndColumn: 5},
			},
		},
		{
			Code:    "<div>\n          { foo &&\n            bar }\n        </div>",
			Tsx:     true,
			Options: multilineRequire,
			Output:  []string{"<div>\n          {\n foo &&\n            bar \n}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedAfter", Message: expectedAfter, Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
				{MessageId: "expectedBefore", Message: expectedBefore, Line: 3, Column: 17, EndLine: 3, EndColumn: 18},
			},
		},
		{
			Code:    "<div style={foo &&\n          foo.bar\n        } />",
			Tsx:     true,
			Options: multilineRequire,
			Output:  []string{"<div style={\nfoo &&\n          foo.bar\n        } />"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "expectedAfter", Message: expectedAfter, Line: 1, Column: 12, EndLine: 1, EndColumn: 13},
			},
		},
		// ---- never: newlines on both sides unexpected ----
		{
			Code:    "<div>\n          {\nfoo\n}\n        </div>",
			Tsx:     true,
			Options: never,
			Output:  []string{"<div>\n          {foo}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 4, Column: 1, EndLine: 4, EndColumn: 2},
			},
		},
		{
			Code:    "<div>\n          {\n            foo &&\n            foo.bar\n          }\n        </div>",
			Tsx:     true,
			Options: never,
			Output:  []string{"<div>\n          {foo &&\n            foo.bar}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 5, Column: 11, EndLine: 5, EndColumn: 12},
			},
		},
		{
			Code:    "<div>\n          { foo &&\n            foo.bar\n          }\n        </div>",
			Tsx:     true,
			Options: never,
			Output:  []string{"<div>\n          { foo &&\n            foo.bar}\n        </div>"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 4, Column: 11, EndLine: 4, EndColumn: 12},
			},
		},
		// ---- never: comment in the gap suppresses the fix (output unchanged) ----
		{
			Code:    "<div>\n          { /* not fixed due to comment */\n            foo }\n        </div>",
			Tsx:     true,
			Options: never,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedAfter", Message: unexpectedAfter, Line: 2, Column: 11, EndLine: 2, EndColumn: 12},
			},
		},
		{
			Code:    "<div>\n          { foo\n            /* not fixed due to comment */}\n        </div>",
			Tsx:     true,
			Options: never,
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "unexpectedBefore", Message: unexpectedBefore, Line: 3, Column: 43, EndLine: 3, EndColumn: 44},
			},
		},
	})
}
