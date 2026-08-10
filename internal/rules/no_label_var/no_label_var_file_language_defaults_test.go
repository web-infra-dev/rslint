package no_label_var

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoLabelVarFileLanguageDefaults(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.allow-js.json",
		t,
		&NoLabelVarRule,
		[]rule_tester.ValidTestCase{
			{Code: `arguments: while (false) { break arguments; }`, FileName: "plain-label.js"},
			{Code: `require: while (false) { break require; }`, FileName: "plain-require.js"},
			{
				Code:     `require: while (false) { break require; }`,
				FileName: "require-off.cjs",
				Globals:  map[string]any{"require": "off"},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:     `function f() { arguments: while (false) { break arguments; } }`,
				FileName: "function-arguments.js",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 16},
				},
			},
			{
				Code:     `arguments: while (false) { break arguments; }`,
				FileName: "arguments-label.cjs",
				Globals:  map[string]any{"arguments": "off"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 1},
				},
			},
			{
				Code:     `require: while (false) { break require; }`,
				FileName: "require-label.cjs",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "identifierClashWithLabel", Line: 1, Column: 1},
				},
			},
		},
	)
}
