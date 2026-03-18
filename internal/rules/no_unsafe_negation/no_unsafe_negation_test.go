package no_unsafe_negation

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnsafeNegationRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnsafeNegationRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Normal relational expressions
			{Code: `a in b`},
			{Code: `a instanceof b`},

			// Parenthesized negation is intentional and allowed
			{Code: `(!a) in b`},
			{Code: `(!a) instanceof b`},

			// Negation of the whole expression is fine
			{Code: `!(a in b)`},
			{Code: `!(a instanceof b)`},

			// Ordering relations are NOT checked by default
			{Code: `!a < b`},
			{Code: `!a > b`},
			{Code: `!a <= b`},
			{Code: `!a >= b`},

			// Parenthesized negation with ordering operators (with option enabled)
			{
				Code:    `(!a) < b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `(!a) > b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `(!a) <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
			{
				Code:    `(!a) >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Negation with 'in' operator
			{
				Code: `!a in b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			// Negation with 'instanceof' operator
			{
				Code: `!a instanceof b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},

			// Ordering relations with enforceForOrderingRelations: true
			{
				Code:    `!a < b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `!a > b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `!a <= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
			{
				Code:    `!a >= b`,
				Options: map[string]interface{}{"enforceForOrderingRelations": true},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 1},
				},
			},
		},
	)
}
