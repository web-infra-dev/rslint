package consistent_template_literal_escape_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_template_literal_escape"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTemplateLiteralEscapeUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_template_literal_escape.ConsistentTemplateLiteralEscapeRule,
		[]rule_tester.ValidTestCase{
			{Code: "const foo = `\\${a}`"},
			{Code: "const foo = `hello`"},
			{Code: "const foo = `$`"},
			{Code: "const foo = `{`"},
			{Code: "const foo = ``"},
			{Code: "const foo = `${a}`"},
			{Code: "const foo = `${a}${b}`"},
			{Code: "const foo = String.raw`$\\{a}`"},
			{Code: "const foo = html`$\\{a}`"},
			{Code: "const foo = `\\\\\\${a}`"},
			{Code: "const foo = '$\\{a}'"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const foo = `$\\{a}`", Output: []string{"const foo = `\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 20}}},
			{Code: "const foo = `\\$\\{a}`", Output: []string{"const foo = `\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
			{Code: "const foo = `$\\{a} and $\\{b}`", Output: []string{"const foo = `\\${a} and \\${b}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 30}}},
			{Code: "const foo = `\\\\$\\{a}`", Output: []string{"const foo = `\\\\\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 22}}},
			{Code: "const foo = `\\\\\\$\\{a}`", Output: []string{"const foo = `\\\\\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 23}}},
			{Code: "const foo = `$\\{a}${expr}`", Output: []string{"const foo = `\\${a}${expr}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}}},
			{Code: "const foo = `${expr}$\\{a}`", Output: []string{"const foo = `${expr}\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 20, EndLine: 1, EndColumn: 27}}},
			{Code: "const foo = `$\\{a}${expr}$\\{b}`", Output: []string{"const foo = `\\${a}${expr}\\${b}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 13, EndLine: 1, EndColumn: 21}, {MessageId: "consistent-template-literal-escape", Message: "Use `\\${` instead of `$\\{` to escape in template literals.", Line: 1, Column: 25, EndLine: 1, EndColumn: 32}}},
		},
	)
}
