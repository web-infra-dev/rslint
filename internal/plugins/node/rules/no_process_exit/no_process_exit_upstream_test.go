package no_process_exit_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_process_exit"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noProcessExitAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noProcessExit",
		Message:   "Don't use process.exit(); throw an error instead.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-process-exit.js
func TestNoProcessExitUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_process_exit.NoProcessExitRule,
		[]rule_tester.ValidTestCase{
			{Code: `Process.exit()`},
			{Code: `var exit = process.exit;`},
			{Code: `f(process.exit)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `process.exit(0);`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 1, 1, 16)},
			},
			{
				Code:   `process.exit(1);`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 1, 1, 16)},
			},
			{
				Code:   `f(process.exit(1));`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 3, 1, 18)},
			},
		},
	)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-process-exit.md
// The rule-enabling comments are omitted; the tester enables the rule.
func TestNoProcessExitDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_process_exit.NoProcessExitRule,
		[]rule_tester.ValidTestCase{
			{Code: `if (somethingBadHappened) {
    throw new Error("Something bad happened!");
}`},
			{Code: "Process.exit();\nvar exit = process.exit;"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `if (somethingBadHappened) {
    console.error("Something bad happened!");
    process.exit(1);
}`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(3, 5, 3, 20)},
			},
			{
				Code: "process.exit(1);\nprocess.exit(0);",
				Errors: []rule_tester.InvalidTestCaseError{
					noProcessExitAt(1, 1, 1, 16),
					noProcessExitAt(2, 1, 2, 16),
				},
			},
		},
	)
}
