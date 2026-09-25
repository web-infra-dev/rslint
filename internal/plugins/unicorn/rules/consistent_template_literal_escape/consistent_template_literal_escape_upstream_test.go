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
			{Code: "const foo = `$\\{a}`", Output: []string{"const foo = `\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `\\$\\{a}`", Output: []string{"const foo = `\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `$\\{a} and $\\{b}`", Output: []string{"const foo = `\\${a} and \\${b}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `\\\\$\\{a}`", Output: []string{"const foo = `\\\\\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `\\\\\\$\\{a}`", Output: []string{"const foo = `\\\\\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `$\\{a}${expr}`", Output: []string{"const foo = `\\${a}${expr}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `${expr}$\\{a}`", Output: []string{"const foo = `${expr}\\${a}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const foo = `$\\{a}${expr}$\\{b}`", Output: []string{"const foo = `\\${a}${expr}\\${b}`"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}, {MessageId: "consistent-template-literal-escape"}}},
		},
	)
}
