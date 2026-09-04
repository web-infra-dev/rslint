// TestNoProcessExitUpstream migrates the full valid/invalid suite from upstream
// eslint-plugin-n no-process-exit rule 1:1. rslint-specific lock-in cases live in
// the no_process_exit_extras_test.go file.
package no_process_exit_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/n/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/n/rules/no_process_exit"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoProcessExitUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_process_exit.NoProcessExitRule,
		[]rule_tester.ValidTestCase{
			{Code: `process.stderr.write("message\n")`},
			{Code: `throw new Error("an error occurred")`},
			{Code: `var exit = process.exit; exit(1)`},
			{Code: `Process.exit(1)`},
			{Code: `process.Exit(1)`},
			{Code: `f(process.exit)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `process.exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Message:   "Don't use process.exit(); throw an error instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `process.exit(0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Message:   "Don't use process.exit(); throw an error instead.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: `f(process.exit(1))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    3,
					},
				},
			},
		},
	)
}
