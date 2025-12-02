package no_template_curly_in_string

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoTemplateCurlyInString(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoTemplateCurlyInString,
		// Valid cases - ported from ESLint
		[]rule_tester.ValidTestCase{
			{Code: "`Hello, ${name}`;"},
			{Code: "templateFunction`Hello, ${name}`;"},
			{Code: "`Hello, name`;"},
			{Code: "'Hello, name';"},
			{Code: "'Hello, ' + name;"},
			{Code: "`Hello, ${index + 1}`"},
			{Code: "`Hello, ${name + \" foo\"}`"},
			{Code: "`Hello, ${name || \"foo\"}`"},
			{Code: "`Hello, ${{foo: \"bar\"}.foo}`"},
			{Code: "'$2'"},
			{Code: "'${'"},
			{Code: "'$}'"},
			{Code: "'{foo}'"},
			{Code: `'{foo: \"bar\"}'`},
			{Code: "const number = 3"},
		},
		// Invalid cases - ported from ESLint
		[]rule_tester.InvalidTestCase{
			{
				Code: "'Hello, ${name}'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTemplateExpression", Line: 1, Column: 1},
				},
			},
			{
				Code: "'${greeting}, ${name}'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTemplateExpression", Line: 1, Column: 1},
				},
			},
			{
				Code: "'Hello, ${index + 1}'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTemplateExpression", Line: 1, Column: 1},
				},
			},
			{
				Code: "'Hello, ${name + \\\" foo\\\"}'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTemplateExpression", Line: 1, Column: 1},
				},
			},
			{
				Code: "'Hello, ${name || \\\"foo\\\"}'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTemplateExpression", Line: 1, Column: 1},
				},
			},
			{
				Code: "'Hello, ${{foo: \\\"bar\\\"}.foo}'",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unexpectedTemplateExpression", Line: 1, Column: 1},
				},
			},
		},
	)
}
