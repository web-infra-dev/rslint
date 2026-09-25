package consistent_template_literal_escape_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/consistent_template_literal_escape"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestConsistentTemplateLiteralEscapeDocs(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&consistent_template_literal_escape.ConsistentTemplateLiteralEscapeRule,
		[]rule_tester.ValidTestCase{
			{Code: "const template = `\\${variableName}`;"},
			{Code: "const message = `Use \\${variable} in your code`;"},
			{Code: "const regex = `/\\${pattern}/g`;"},
			{Code: "const name = 'Alice';\nconst greeting = `Hello ${name}, use \\${variable} for templates`;"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "const template = `$\\{variableName}`;", Output: []string{"const template = `\\${variableName}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
			{Code: "const template = `\\$\\{variableName}`;", Output: []string{"const template = `\\${variableName}`;"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "consistent-template-literal-escape"}}},
		},
	)
}
