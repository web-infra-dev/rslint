package no_self_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoSelfAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoSelfAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: `a = b`},
			{Code: `a = a + 1`},
			{Code: `a += a`},
			{Code: `a = +a`},
			{Code: `a = [a]`},
			{Code: `let a = a`},
			{Code: `a = a.b`},
			{Code: `a.b = a.c`},
			{Code: `a[0] = a[1]`},
			{Code: `a.b = a.b`, Options: map[string]interface{}{"props": false}},
			{Code: `a[0] = a[0]`, Options: map[string]interface{}{"props": false}},
			// Optional chaining on the right is not self-assignment
			{Code: `a.b = a?.b`},
			{Code: `a[0] = a?.[0]`},
			// Different identifiers in destructuring
			{Code: `[a, b] = [b, a]`},
			// Object destructuring with different values
			{Code: `({a} = {a: b})`},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code: `a = a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 5},
				},
			},
			{
				Code: `[a] = [a]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code: `[a, b] = [a, b]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 11},
					{MessageId: "selfAssignment", Line: 1, Column: 14},
				},
			},
			{
				Code: `[a, b] = [a, c]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 11},
				},
			},
			{
				Code:    `a.b = a.b`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			// Default props is true
			{
				Code: `a.b = a.b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code:    `this.x = this.x`,
				Options: map[string]interface{}{"props": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			{
				Code: `a[1] = a[1]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 8},
				},
			},
			{
				Code: `a["b"] = a["b"]`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 10},
				},
			},
			// Logical assignment operators
			{
				Code: `a &&= a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a ||= a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
			{
				Code: `a ??= a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "selfAssignment", Line: 1, Column: 7},
				},
			},
		},
	)
}
