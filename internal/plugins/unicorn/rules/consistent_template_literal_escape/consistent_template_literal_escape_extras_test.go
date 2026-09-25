package consistent_template_literal_escape_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_template_literal_escape"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTemplateLiteralEscapeExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_template_literal_escape.ConsistentTemplateLiteralEscapeRule,
		[]rule_tester.ValidTestCase{
			{Code: "const tagged = tag`$\\{value}`;"},
			{Code: "const tagged = tag`${expr}$\\{value}`;"},
			{Code: "const tagged = tag`${a}$\\{x}${b}tail`;"},

		},
		[]rule_tester.InvalidTestCase{
			{Code: "const value = `prefix $\\{x} suffix`;", Output: []string{"const value = `prefix \\${x} suffix`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 15, EndLine: 1, EndColumn: 36}}},
			{Code: "const value = `${a}$\\{x}${b}`;", Output: []string{"const value = `${a}\\${x}${b}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27}}},
		},
	)
}
