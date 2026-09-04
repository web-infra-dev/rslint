// TestNoProcessExitExtras locks in branches and edge shapes that the upstream test suite
// doesn't exercise. Each case carries an inline comment pointing at the specific branch /
// Dimension 4 row / tsgo AST quirk it covers.
package no_process_exit_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/n/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/no_process_exit"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoProcessExitExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_process_exit.NoProcessExitRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: optional chain ----
			// N/A: process?.exit(1) is still reported because the upstream ESLint rule also
			// catches it — the CSS selector `CallExpression > MemberExpression.callee` matches
			// optional MemberExpressions in ESLint's AST too.

			// ---- Dimension 4: element access ----
			{Code: `process["exit"](1)`}, // valid because this is ElementAccessExpression, not PropertyAccessExpression

			// ---- Dimension 4: computed key ----
			{Code: `const key = "exit"; process[key](1)`}, // valid

			// ---- Dimension 4: spread / empty ----
			{Code: `process.exit`},       // just reading, not calling
			{Code: `new process.exit()`}, // NewExpression, not CallExpression

			// ---- Real-user: non-identifier object ----
			{Code: `getProcess().exit(1)`}, // not `process` identifier
			{Code: `foo.process.exit(1)`},  // nested access, `process` is not the top-level identifier
			{Code: `this.process.exit(1)`}, // this.process
		},
		[]rule_tester.InvalidTestCase{
			// ---- Dimension 4: optional chain ----
			{
				Code: `process?.exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
					},
				},
			},
			// ---- Dimension 4: parenthesized callee ----
			{
				Code: `(process.exit)(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
					},
				},
			},
			// ---- Dimension 4: parenthesized process ----
			{
				Code: `(process).exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
					},
				},
			},
			// ---- Dimension 4: no arguments ----
			{
				Code: `process.exit()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
					},
				},
			},
			// Locks in upstream create() arm: process.exit call inside expression
			{
				Code: `if (err) process.exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    10,
					},
				},
			},
			// ---- Real-user: process.exit in nested callback ----
			{
				Code: "setTimeout(function() {\n  process.exit(1)\n}, 1000)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      2,
						Column:    3,
					},
				},
			},
		},
	)
}
