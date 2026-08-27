package no_unnecessary_template_expression

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessaryTemplateExpressionAutofix(t *testing.T) {
	errorWithMessage := func() rule_tester.InvalidTestCaseError {
		return rule_tester.InvalidTestCaseError{MessageId: "noUnnecessaryTemplateExpression"}
	}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUnnecessaryTemplateExpressionRule,
		[]rule_tester.ValidTestCase{
			{Code: "`trailing whitespace: ${' '}\r`;"},
			{Code: "`trailing whitespace: ${' '}\u2028`;"},
			{Code: "`trailing whitespace: ${' '}\u2029`;"},
		},
		[]rule_tester.InvalidTestCase{
			{
				Code:   "`${0o25}-${0b1010}-${0x25}-${1n}`;",
				Output: []string{"`21-10-37-1`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					errorWithMessage(), errorWithMessage(), errorWithMessage(), errorWithMessage(),
				},
			},
			{
				Code:   "const value = `back${'`'}tick`;",
				Output: []string{"const value = `back\\`tick`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			{
				Code:   "const value = `dollar${'${x}'}sign`;",
				Output: []string{"const value = `dollar\\${x}sign`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			{
				Code:   "const value = ` ${'$'}{} `;",
				Output: []string{"const value = ` \\${} `;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			{
				Code:   "const value = `use${`less`}`;",
				Output: []string{"const value = `useless`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			{
				Code:   "const value = `${'a'} ${true} ${/a/}`;",
				Output: []string{"const value = `a true /a/`;"},
				Errors: []rule_tester.InvalidTestCaseError{
					errorWithMessage(), errorWithMessage(), errorWithMessage(),
				},
			},
			{
				Code:   "type Value = `pre-${'suffix'}`;",
				Output: []string{"type Value = `pre-suffix`;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
			{
				Code:   "declare const value: string; `${value || ''}`.length;",
				Output: []string{"declare const value: string; (value || '').length;"},
				Errors: []rule_tester.InvalidTestCaseError{errorWithMessage()},
			},
		},
	)
}
