// TestJsxFunctionCallNewlineUpstream migrates the full valid/invalid suite from
// upstream packages/eslint-plugin/rules/jsx-function-call-newline/
// jsx-function-call-newline.test.ts 1:1. Upstream asserts only messageId on its
// invalid cases; the line/column/endLine/endColumn and exact message text here
// are computed from the source each case carries (the diagnostic spans the JSX
// argument, from its `<` to just past its closing `>`). rslint-specific lock-in
// cases live in jsx_function_call_newline_extras_test.go.
package jsx_function_call_newline

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestJsxFunctionCallNewlineUpstream(t *testing.T) {
	always := []interface{}{"always"}
	const msg = "Missing line break around JSX"

	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &JsxFunctionCallNewlineRule, []rule_tester.ValidTestCase{
		{Code: "fn(<div />)", Tsx: true},
		{Code: "fn(<div />, <div />)", Tsx: true},
		{Code: "fn(<div />,\n<div />)", Tsx: true},
		{Code: "fn(\n<div />, <div />)", Tsx: true},
		{Code: "fn(\n<div />, <div />\n)", Tsx: true},
		{Code: "fn(\n<div />\n)", Tsx: true, Options: always},
		{Code: "fn(<div />, \n<div \n style={{ color: 'red' }}\n />\n)", Tsx: true},
		{Code: "fn(<div />, <div />, <div />)", Tsx: true},
		{Code: "fn(<div />, <div />\n, <div />)", Tsx: true},
		{Code: "fn(\n<div />\n,\n<div />\n,\n<div />\n)", Tsx: true},
		{Code: "fn(\n<div />\n,\n<div />\n,\n<div />\n)", Tsx: true, Options: always},
		{Code: "fn(\n<div />\n,\n<div ></div>)", Tsx: true},
		{Code: "fn((<div style={{}} />), <div />, <div />)", Tsx: true},
		{Code: "new OBJ((<div style={{}} />), <div />, <div />)", Tsx: true},
		{Code: "new OBJ(<div />, <div />, <div />)", Tsx: true},
		{Code: "new OBJ(<div />, <div />\n, <div />)", Tsx: true},
		{Code: "new OBJ(\n<div />\n,\n<div />\n,\n<div />\n)", Tsx: true},
		{Code: "new OBJ(\n<div />\n,\n<div />\n,\n<div />\n)", Tsx: true, Options: always},
		{Code: "new OBJ(\n<div />\n,\n<div ></div>)", Tsx: true},
	}, []rule_tester.InvalidTestCase{
		{
			Code:   "fn(<div\n        />)",
			Tsx:    true,
			Output: []string{"fn(\n<div\n        />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 11},
			},
		},
		{
			Code:   "new OBJ(<div\n        />)",
			Tsx:    true,
			Output: []string{"new OBJ(\n<div\n        />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 9, EndLine: 2, EndColumn: 11},
			},
		},
		{
			Code:    "fn(<div />)",
			Tsx:     true,
			Options: always,
			Output:  []string{"fn(\n<div />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 1, EndColumn: 11},
			},
		},
		{
			Code:    "fn(\n<div />,<div />,\n<div />)",
			Tsx:     true,
			Options: always,
			Output:  []string{"fn(\n<div />,\n<div />,\n<div />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 2, Column: 9, EndLine: 2, EndColumn: 16},
				{MessageId: missingLineBreak, Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 8},
			},
		},
		{
			Code:    "new OBJ(\n<div />,<div />,\n<div />)",
			Tsx:     true,
			Options: always,
			Output:  []string{"new OBJ(\n<div />,\n<div />,\n<div />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 2, Column: 9, EndLine: 2, EndColumn: 16},
				{MessageId: missingLineBreak, Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 8},
			},
		},
		{
			Code:    "fn((\n<div />),<div />,\n<div />)",
			Tsx:     true,
			Options: always,
			Output:  []string{"fn((\n<div />\n),\n<div />,\n<div />\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 2, Column: 1, EndLine: 2, EndColumn: 8},
				{MessageId: missingLineBreak, Message: msg, Line: 2, Column: 10, EndLine: 2, EndColumn: 17},
				{MessageId: missingLineBreak, Message: msg, Line: 3, Column: 1, EndLine: 3, EndColumn: 8},
			},
		},
		{
			Code:   "fn(<div />, <span>\n</span>)",
			Tsx:    true,
			Output: []string{"fn(<div />, \n<span>\n</span>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 13, EndLine: 2, EndColumn: 8},
			},
		},
		{
			Code:   "fn(<div \n />, <span>\n</span>)",
			Tsx:    true,
			Output: []string{"fn(\n<div \n />, \n<span>\n</span>\n)"},
			Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: missingLineBreak, Message: msg, Line: 1, Column: 4, EndLine: 2, EndColumn: 4},
				{MessageId: missingLineBreak, Message: msg, Line: 2, Column: 6, EndLine: 3, EndColumn: 8},
			},
		},
	})
}
