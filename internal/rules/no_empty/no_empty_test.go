package no_empty

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoEmptyRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoEmptyRule,
		[]rule_tester.ValidTestCase{
			{Code: `if (foo) { bar() }`},
			{Code: `while (foo) { bar() }`},
			{Code: `for (;;) { bar() }`},
			{Code: `try { foo() } catch (e) { bar() }`},
			{Code: `switch (foo) { case 1: break; }`},
			{Code: `function foo() {}`},
			{Code: `var foo = function() {}`},
			{Code: `var foo = () => {}`},
			{Code: `if (foo) { /* comment */ }`},
			{Code: `while (foo) { /* comment */ }`},
			{Code: `try { foo() } catch (e) { /* comment */ }`},
			{
				Code:    `try { foo() } catch (e) {}`,
				Options: map[string]interface{}{"allowEmptyCatch": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `if (foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code: `while (foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 13},
				},
			},
			{
				Code: `for (;;) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code: `try {} catch (e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 5},
					{MessageId: "unexpected", Line: 1, Column: 18},
				},
			},
			{
				Code: `try { foo() } catch (e) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 25},
				},
			},
			{
				Code: `switch (foo) {}`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
