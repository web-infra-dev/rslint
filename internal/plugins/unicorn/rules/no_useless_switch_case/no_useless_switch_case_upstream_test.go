package no_useless_switch_case_test

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_switch_case"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const (
	noUselessSwitchCaseMessage      = "Useless case in switch statement."
	noUselessSwitchCaseMessageID    = "no-useless-switch-case/error"
	noUselessSwitchCaseSuggestionID = "no-useless-switch-case/suggestion"
)

// TestNoUselessSwitchCaseUpstream migrates the full valid/invalid suite from
// upstream test/no-useless-switch-case.js 1:1. Position assertions cover
// line/column for every invalid case. rslint-specific lock-in cases live in
// the no_useless_switch_case_extras_test.go file.
func TestNoUselessSwitchCaseUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&no_useless_switch_case.NoUselessSwitchCaseRule,
		[]rule_tester.ValidTestCase{
			// ---- test.snapshot valid ----
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\tcase b:",
				"\t\thandleDefaultCase();",
				"\t\tbreak;",
				"}",
			)),
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\t\thandleCaseA();",
				"\t\tbreak;",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"\t\tbreak;",
				"}",
			)),
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\t\thandleCaseA();",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"\t\tbreak;",
				"}",
			)),
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\t\tbreak;",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"\t\tbreak;",
				"}",
			)),
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\t\thandleCaseA();",
				"\t\t// Fallthrough",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"\t\tbreak;",
				"}",
			)),
			jsValid(lines(
				"switch (foo) {",
				"\tcase a:",
				"\tdefault:",
				"\t\thandleDefaultCase();",
				"\t\tbreak;",
				"\tcase b:",
				"\t\thandleCaseB();",
				"\t\tbreak;",
				"}",
			)),
			jsValid(lines(
				"switch (1) {",
				"\t\t// This is not useless",
				"\t\tcase 1:",
				"\t\tdefault:",
				"\t\t\t\tconsole.log('1')",
				"\t\tcase 1:",
				"\t\t\t\tconsole.log('2')",
				"}",
			)),
		},
		[]rule_tester.InvalidTestCase{
			// ---- test.snapshot invalid ----
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 17, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase null:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 12, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
				expectedError(3, 2, 3, 17, lines(
					"switch (foo) {",
					"\tcase a:",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase undefined:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 17, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			tsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a: {",
					"\t}",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a: {",
					"\t\t;;",
					"\t\t{",
					"\t\t\t;;",
					"\t\t\t{",
					"\t\t\t\t;;",
					"\t\t\t}",
					"\t\t}",
					"\t}",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\tcase (( b ))         :",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tcase (( b ))         :",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
				expectedError(3, 2, 3, 24, lines(
					"switch (foo) {",
					"\tcase a:",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\tcase b:",
					"\t\thandleCaseAB();",
					"\t\tbreak;",
					"\tcase d:",
					"\tcase d:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(6, 2, 6, 9, lines(
					"switch (foo) {",
					"\tcase a:",
					"\tcase b:",
					"\t\thandleCaseAB();",
					"\t\tbreak;",
					"\t",
					"\tcase d:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
				expectedError(7, 2, 7, 9, lines(
					"switch (foo) {",
					"\tcase a:",
					"\tcase b:",
					"\t\thandleCaseAB();",
					"\t\tbreak;",
					"\tcase d:",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\tcase b:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\tcase b:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
				expectedError(3, 2, 3, 9, lines(
					"switch (foo) {",
					"\tcase a:",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\t// eslint-disable-next-line",
					"\tcase a:",
					"\tcase b:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(4, 2, 4, 9, lines(
					"switch (foo) {",
					"\t// eslint-disable-next-line",
					"\tcase a:",
					"\t",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
			jsInvalid(
				lines(
					"switch (foo) {",
					"\tcase a:",
					"\t// eslint-disable-next-line",
					"\tcase b:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				),
				expectedError(2, 2, 2, 9, lines(
					"switch (foo) {",
					"\t",
					"\t// eslint-disable-next-line",
					"\tcase b:",
					"\tdefault:",
					"\t\thandleDefaultCase();",
					"\t\tbreak;",
					"}",
				)),
			),
		},
	)
}

func lines(parts ...string) string {
	return strings.Join(parts, "\n")
}

func jsValid(code string) rule_tester.ValidTestCase {
	return rule_tester.ValidTestCase{Code: code, FileName: "file.js"}
}

func jsInvalid(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "file.js", Errors: errors}
}

func tsInvalid(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, FileName: "file.ts", Errors: errors}
}

func expectedError(line, column, endLine, endColumn int, suggestionOutput string) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: noUselessSwitchCaseMessageID,
		Message:   noUselessSwitchCaseMessage,
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
		Suggestions: []rule_tester.InvalidTestCaseSuggestion{
			{
				MessageId: noUselessSwitchCaseSuggestionID,
				Output:    suggestionOutput,
			},
		},
	}
}
