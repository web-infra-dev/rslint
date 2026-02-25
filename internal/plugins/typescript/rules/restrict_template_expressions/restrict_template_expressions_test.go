package restrict_template_expressions

import (
	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"testing"
)

func TestRestrictTemplateExpressionsRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&RestrictTemplateExpressionsRule,
		[]rule_tester.ValidTestCase{
			// String literal in template
			{Code: "const msg = `arg = ${'foo'}`;"},
			// String variable in template
			{Code: `
const arg = 'foo';
const msg = ` + "`arg = ${arg}`" + `;
`},
			// Number with allowNumber option
			{
				Code: "const msg = `arg = ${123}`;",
				Options: map[string]interface{}{
					"allowNumber": true,
				},
			},
		},
		[]rule_tester.InvalidTestCase{
			// TODO: Add invalid test cases once type checking is implemented
			// The rule implementation has a TODO for type checking (see line 102)
		},
	)
}
