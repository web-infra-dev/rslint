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
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const value = `prefix $\\{x} suffix`;", Output: []string{"const value = `prefix \\${x} suffix`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
		},
	)
}
