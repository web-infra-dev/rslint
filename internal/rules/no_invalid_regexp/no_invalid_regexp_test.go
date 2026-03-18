// cspell:ignore dgimsuy
package no_invalid_regexp

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoInvalidRegexpRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoInvalidRegexpRule,
		// Valid cases
		[]rule_tester.ValidTestCase{
			{Code: `RegExp('.')`},
			{Code: `new RegExp('.')`},
			{Code: `new RegExp('.', 'im')`},
			{Code: `new RegExp('.', 'gmi')`},
			{Code: `new RegExp('.', 'dgimsuy')`},
			{Code: `new RegExp(pattern, 'g')`}, // non-literal pattern, skip
			{Code: `new RegExp('.', flags)`},   // non-literal flags, skip
			{Code: `RegExp('')`},               // empty pattern is valid
			{Code: `RegExp('a|b')`},            // alternation
			{Code: `new RegExp('\\\\d+')`},     // escaped digits
			{Code: `new RegExp('[abc]')`},      // character class
			{Code: `new RegExp('(?:a)')`},      // non-capturing group
			{Code: `RegExp('a{1,2}')`},         // quantifier
			{Code: `new RegExp('.', 'v')`},     // v flag alone is valid
			{Code: `new RegExp('.', 'u')`},     // u flag alone is valid
			// allowConstructorFlags option
			{
				Code:    `new RegExp('.', 'z')`,
				Options: map[string]interface{}{"allowConstructorFlags": "z"},
			},
			{
				Code:    `new RegExp('.', 'az')`,
				Options: map[string]interface{}{"allowConstructorFlags": "az"},
			},
			// allowConstructorFlags as array with multi-char string
			{
				Code:    `new RegExp('.', 'az')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"az"}},
			},
			// allowConstructorFlags as array with single-char strings
			{
				Code:    `new RegExp('.', 'az')`,
				Options: map[string]interface{}{"allowConstructorFlags": []interface{}{"a", "z"}},
			},
		},
		// Invalid cases
		[]rule_tester.InvalidTestCase{
			{
				Code: `RegExp('.', 'z');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'aa');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp('.', 'uv');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('.', 'gg');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp('[');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `RegExp('(');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			{
				Code: `new RegExp('\\');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
			// Non-literal flags with invalid pattern: still reported
			{
				Code: `new RegExp('[', flags);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "regexMessage", Line: 1, Column: 1},
				},
			},
		},
	)
}
