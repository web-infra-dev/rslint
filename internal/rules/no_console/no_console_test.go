package no_console

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoConsoleRule,
		[]rule_tester.ValidTestCase{
			{
				Code: "Console.log('foo')",
			},
			{
				Code: "console.log('foo')",
				Options: map[string]interface{}{
					"allow": []interface{}{"log"},
				},
			},
			{
				Code: "console.error('foo')",
				Options: map[string]interface{}{
					"allow": []interface{}{"error"},
				},
			},
			{
				Code: "console['log']('foo')",
				Options: map[string]interface{}{
					"allow": []interface{}{"log"},
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: "console.log('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "console.error('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "console.log",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "console.log = foo",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "console['log']",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code: "console.warn('foo')",
				Options: map[string]interface{}{
					"allow": []interface{}{"log"},
				},
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "unexpected",
						Line:      1,
						Column:    1,
					},
				},
			},
		},
	)
}
