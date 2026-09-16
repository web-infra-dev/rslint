// TestNoCallbackLiteralUpstream migrates all upstream tests and documentation examples from v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-callback-literal.js
// rslint-specific edge shapes and branch lock-ins live in no_callback_literal_extras_test.go.
// cspell:ignore snork
package no_callback_literal

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoCallbackLiteralUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoCallbackLiteralRule,
		[]rule_tester.ValidTestCase{
			// ---- random stuff ----
			{Code: "horse()"},
			{Code: "sort(null)"},
			{Code: "require(\"zyx\")"},
			{Code: "require(\"zyx\", data)"},
			// ---- callback() ----
			{Code: "callback()"},
			{Code: "callback(undefined)"},
			{Code: "callback(null)"},
			{Code: "callback(x)"},
			{Code: "callback(new Error(\"error\"))"},
			{Code: "callback(friendly, data)"},
			{Code: "callback(undefined, data)"},
			{Code: "callback(null, data)"},
			{Code: "callback(x, data)"},
			{Code: "callback(new Error(\"error\"), data)"},
			{Code: "callback(x = obj, data)"},
			{Code: "callback((1, a), data)"},
			{Code: "callback(a || b, data)"},
			{Code: "callback(a ? b : c, data)"},
			{Code: "callback(a ? 1 : c, data)"},
			{Code: "callback(a ? b : 1, data)"},
			// ---- cb() ----
			{Code: "cb()"},
			{Code: "cb(undefined)"},
			{Code: "cb(null)"},
			{Code: "cb(undefined, \"super\")"},
			{Code: "cb(null, \"super\")"},
			// https://github.com/eslint-community/eslint-plugin-n/issues/162
			{Code: "cb(e as Error)", FileName: "input.ts"},
			// ---- Documentation: correct example (rule directive omitted) ----
			{Code: "cb(undefined);\ncb(null, 5);\ncallback(new Error('some error'));\ncallback(someVariable);"},
		},
		[]rule_tester.InvalidTestCase{
			// ---- callback ----
			{Code: "callback(false, \"snork\")", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 25)}},
			{Code: "callback(\"help\")", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 17)}},
			{Code: "callback(\"help\", data)", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 23)}},
			// ---- cb ----
			{Code: "cb(false)", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 10)}},
			{Code: "cb(\"help\")", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 11)}},
			{Code: "cb(\"help\", data)", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 17)}},
			{Code: "cb({ a: 1 })", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 13)}},
			{Code: "cb([])", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 7)}},
			{Code: "cb(`message ${value}`)", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 23)}},
			{Code: "callback((a, 1), data)", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 23)}},
			// ---- Documentation: incorrect example (rule directive omitted) ----
			{Code: "cb('this is an error string');\ncb({ a: 1 });\ncallback(0);", Errors: []rule_tester.InvalidTestCaseError{unexpectedLiteralAt(1, 1, 1, 30), unexpectedLiteralAt(2, 1, 2, 13), unexpectedLiteralAt(3, 1, 3, 12)}},
		},
	)
}

func unexpectedLiteralAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unexpectedLiteral",
		Message:   "Unexpected literal in error position of callback.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}
