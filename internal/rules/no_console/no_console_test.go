package no_console

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoConsoleRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConsoleRule,
		[]rule_tester.ValidTestCase{
			{
				Code:    `console.warn("test");`,
				Options: map[string]interface{}{"allow": []interface{}{"warn"}},
			},
			{
				Code:    `console.error("test");`,
				Options: map[string]interface{}{"allow": []interface{}{"error"}},
			},
			{
				Code:    `console.warn("test"); console.error("test");`,
				Options: map[string]interface{}{"allow": []interface{}{"warn", "error"}},
			},
			{Code: `var x = { console: 1 };`},
			// Shadowed console should not be reported
			{Code: `function f(console: any) { console.log("x"); }`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `console.log("test");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `console.warn("test");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `console.error("test");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code: `console.info("test");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `console.log("test");`,
				Options: map[string]interface{}{"allow": []interface{}{"warn"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Computed property access: console["log"]
			{
				Code: `console["log"]("test");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Chained member access: console.log.bind should still report console.log
			{
				Code: `console.log.bind(null)("test");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
