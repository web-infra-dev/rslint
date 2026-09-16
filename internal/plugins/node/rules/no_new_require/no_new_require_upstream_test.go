package no_new_require_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/node/rules/no_new_require"
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func noNewRequireAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "noNewRequire",
		Message:   "Unexpected use of new with require.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-new-require.js
func TestNoNewRequireUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_require.NoNewRequireRule,
		[]rule_tester.ValidTestCase{
			{Code: `var appHeader = require('app-header')`},
			{Code: `var AppHeader = new (require('app-header'))`},
			{Code: `var AppHeader = new (require('headers').appHeader)`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var appHeader = new require('app-header')`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 17, 1, 42)},
			},
			{
				Code:   `var appHeader = new require('headers').appHeader`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 17, 1, 39)},
			},
		},
	)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-new-require.md
// The duplicate incorrect example is included once; the tester enables the rule.
func TestNoNewRequireDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_new_require.NoNewRequireRule,
		[]rule_tester.ValidTestCase{
			{Code: `var appHeader = require('app-header');`},
			{Code: `var appHeader = new (require('app-header'));`},
			{Code: "var AppHeader = require('app-header');\nvar appHeader = new AppHeader();"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   `var appHeader = new require('app-header');`,
				Errors: []rule_tester.InvalidTestCaseError{noNewRequireAt(1, 17, 1, 42)},
			},
		},
	)
}
