package eqeqeq

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestEqeqeqRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&EqeqeqRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			// Default "always" mode - strict equality is valid
			{Code: `a === b`},
			{Code: `a !== b`},
			{Code: `a === null`},
			{Code: `null !== a`},

			// "smart" mode
			{Code: `typeof a == 'number'`, Options: "smart"},
			{Code: `'string' != typeof a`, Options: "smart"},
			{Code: `typeof a == typeof b`, Options: "smart"},
			{Code: `null == a`, Options: "smart"},
			{Code: `a != null`, Options: "smart"},
			{Code: `'hello' == 'world'`, Options: "smart"},
			{Code: `0 == 0`, Options: "smart"},
			{Code: `true == true`, Options: "smart"},
			{Code: `null == null`, Options: "smart"},
			{Code: `a === b`, Options: "smart"},

			// "always" with null:"ignore"
			{Code: `a == null`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `null == a`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `a != null`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `null != a`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},
			{Code: `a === b`, Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}}},

			// "allow-null" (same as ["always", {"null": "ignore"}])
			{Code: `a == null`, Options: "allow-null"},
			{Code: `null == a`, Options: "allow-null"},
			{Code: `a != null`, Options: "allow-null"},
			{Code: `null != a`, Options: "allow-null"},
			{Code: `a === b`, Options: "allow-null"},

			// "always" with null:"never" - loose null checks are valid
			{Code: `a == null`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `null == a`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
			{Code: `a != null`, Options: []interface{}{"always", map[string]interface{}{"null": "never"}}},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			// Default "always" mode
			{
				Code: `a == b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code: `a != b`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code: `typeof a == 'number'`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 10},
				},
			},
			{
				Code: `true == true`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			{
				Code: `a == null`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code: `null != a`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},

			// "smart" mode - non-exempted loose equality
			{
				Code:    `a == b`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code:    `a != b`,
				Options: "smart",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},

			// "always" with null:"ignore" - non-null loose equality still flagged
			{
				Code:    `a == b`,
				Options: []interface{}{"always", map[string]interface{}{"null": "ignore"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},

			// "always" with null:"never" - strict null checks flagged
			{
				Code:    `a === null`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
			{
				Code:    `null !== a`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 6},
				},
			},
			// "always" with null:"never" - non-null loose equality still flagged
			{
				Code:    `a == b`,
				Options: []interface{}{"always", map[string]interface{}{"null": "never"}},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},

			// "allow-null" - non-null loose equality flagged
			{
				Code:    `a == b`,
				Options: "allow-null",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpected", Line: 1, Column: 3},
				},
			},
		},
	)
}
