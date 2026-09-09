package no_process_exit_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_process_exit"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoProcessExitExtras(t *testing.T) {
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

			// The upstream selector matches property.name, not string values.
			{Code: `process["exit"](1)`},

			{Code: `const key = "exit"; process[key](1)`}, // valid

			{Code: `process.exit`},       // just reading, not calling
			{Code: `new process.exit()`}, // NewExpression, not CallExpression

			{Code: `getProcess().exit(1)`}, // not `process` identifier
			{Code: `foo.process.exit(1)`},  // nested access, `process` is not the top-level identifier
			{Code: `this.process.exit(1)`}, // this.process

			// Parentheses terminate the optional chain, making the callee a ChainExpression.
			{Code: `(process?.exit)(1)`},
			{Code: `(process?.exit)?.(1)`},

			// Authored TypeScript assertions remain visible in ESTree.
			{Code: `(process as any).exit(1)`},
			{Code: `(process.exit as any)(1)`},
			{Code: `process[exit as any](1)`},
		},
		[]rule_tester.InvalidTestCase{
			// The JSX member name is not a call; the expression container contains one.
			{
				Code:   `const element = <process.exit code={process.exit(1)} />;`,
				Tsx:    true,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 37, 1, 52)},
			},
			// Optional calls and members still match unless a parenthesis ends the chain.
			{
				Code:   `process.exit?.(1)`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 1, 1, 18)},
			},
			{
				Code: `process?.exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 17,
					},
				},
			},
			// Ordinary parentheses are transparent to ESTree.
			{
				Code: `(process.exit)(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			{
				Code: `(process).exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 18,
					},
				},
			},
			{
				Code: `process.exit()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    1,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code: `if (err) process.exit(1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      1,
						Column:    10,
						EndLine:   1,
						EndColumn: 25,
					},
				},
			},
			{
				Code: "setTimeout(function() {\n  process.exit(1)\n}, 1000)",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "noProcessExit",
						Line:      2,
						Column:    3,
						EndLine:   2,
						EndColumn: 18,
					},
				},
			},
			// Upstream does not filter computed properties or private identifiers.
			{
				Code:   `process[exit](1)`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 1, 1, 17)},
			},
			{
				Code:   `process[(exit)](1)`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 1, 1, 19)},
			},
			{
				Code:   `class C { #exit() {} f(process) { process.#exit(1); } }`,
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 35, 1, 51)},
			},
			// JSDoc casts in JS synthesize wrappers that ESTree does not expose.
			{
				Code:     `/** @type {any} */ (process).exit(1)`,
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 20, 1, 37)},
			},
			{
				Code:     `(/** @type {any} */ (process.exit))(1)`,
				FileName: "input.js",
				TSConfig: "tsconfig.allowJs.json",
				Errors:   []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 1, 1, 39)},
			},
			// Diagnostic ranges use UTF-16 columns and span the complete call.
			{
				Code:   "\"😀\"; process.exit(\n  1,\n)",
				Errors: []rule_tester.InvalidTestCaseError{noProcessExitAt(1, 7, 3, 2)},
			},
		},
	)
}
