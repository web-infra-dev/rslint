// TestEolLastUpstream migrates the full valid/invalid suite from upstream
// packages/eslint-plugin/rules/eol-last/eol-last.test.ts 1:1. Position
// assertions cover line/column (and endLine/endColumn for 'never' cases) for
// every invalid case. rslint-specific lock-in cases (tsgo AST edge shapes,
// branch lock-ins, real-user shapes) live in eol_last_extras_test.go.
package eol_last_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/stylistic/rules/eol_last"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func optStr(s string) []any { return []any{s} }

func TestEolLastUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&eol_last.EolLastRule,
		[]rule_tester.ValidTestCase{
			// ---- default 'always' ----
			{Code: ``},
			{Code: "\n"},
			{Code: "var a = 123;\n"},
			{Code: "var a = 123;\n\n"},
			{Code: "var a = 123;\n   \n"},

			{Code: "\r\n"},
			{Code: "var a = 123;\r\n"},
			{Code: "var a = 123;\r\n\r\n"},
			{Code: "var a = 123;\r\n   \r\n"},

			// ---- 'never' ----
			{Code: `var a = 123;`, Options: optStr("never")},
			{Code: "var a = 123;\nvar b = 456;", Options: optStr("never")},
			{Code: "var a = 123;\r\nvar b = 456;", Options: optStr("never")},
		},
		[]rule_tester.InvalidTestCase{
			// ---- default 'always' — missing trailing newline ----
			{
				Code:   `var a = 123;`,
				Output: []string{"var a = 123;\n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 1, Column: 13},
				},
			},
			{
				Code:   "var a = 123;\n   ",
				Output: []string{"var a = 123;\n   \n"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "missing", Line: 2, Column: 4},
				},
			},

			// ---- 'never' — unexpected trailing newline ----
			{
				Code:    "var a = 123;\n",
				Output:  []string{`var a = 123;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code:    "var a = 123;\r\n",
				Output:  []string{`var a = 123;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13, EndLine: 2, EndColumn: 1},
				},
			},
			{
				Code:    "var a = 123;\r\n\r\n",
				Output:  []string{`var a = 123;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code:    "var a = 123;\nvar b = 456;\n",
				Output:  []string{"var a = 123;\nvar b = 456;"},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 13, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code:    "var a = 123;\r\nvar b = 456;\r\n",
				Output:  []string{"var a = 123;\r\nvar b = 456;"},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 13, EndLine: 3, EndColumn: 1},
				},
			},
			{
				Code:    "var a = 123;\n\n",
				Output:  []string{`var a = 123;`},
				Options: optStr("never"),
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 2, Column: 1, EndLine: 3, EndColumn: 1},
				},
			},
		},
	)
}
