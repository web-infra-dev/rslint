package no_exports_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoExportsAssignRule(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule, []rule_tester.ValidTestCase{
		{Code: "module.exports.foo = 1;"},
		{Code: "exports.bar = 2;"},
		{Code: "module.exports = {};"},
		{Code: "module.exports = exports = {};"},
		{Code: "exports = module.exports = {};"},
		{Code: "exports = module.exports;"},
		{Code: "var exports = {};"},   // local variable
		{Code: "let exports = {};"},   // local variable
		{Code: "const exports = {};"}, // local variable
		// {Code: "function foo(exports) { exports = {}; }"}, // shadowed - TODO: implement scope/shadowing check
	}, []rule_tester.InvalidTestCase{
		{
			Code: "exports = {};",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExportsAssign",
					Line:      1,
					Column:    1,
				},
			},
		},
		{
			Code: "exports = 1;",
			Errors: []rule_tester.InvalidTestCaseError{
				{
					MessageId: "noExportsAssign",
					Line:      1,
					Column:    1,
				},
			},
		},
	})
}
