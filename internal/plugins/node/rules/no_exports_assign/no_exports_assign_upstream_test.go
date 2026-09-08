package no_exports_assign

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

var exportsGlobals = map[string]any{"exports": "writable", "module": "readonly"}

func forbiddenAt(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "forbidden",
		Message:   "Unexpected assignment to 'exports' variable. Use 'module.exports' instead.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

// Every case from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-exports-assign.js
func TestNoExportsAssignUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "module.exports.foo = 1", Globals: exportsGlobals},
			{Code: "exports.bar = 1", Globals: exportsGlobals},
			{Code: "module.exports = exports = {}", Globals: exportsGlobals},
			{Code: "exports = module.exports = {}", Globals: exportsGlobals},
			{Code: "function f(exports) { exports = {} }", Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "exports = {}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
		},
	)
}

// Every JavaScript example from https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/no-exports-assign.md
// The documentation's rule-enabling comments are omitted; the tester enables the rule.
func TestNoExportsAssignDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: `module.exports.foo = 1
exports.bar = 2

module.exports = {}

// allows ` + "`exports = {}`" + ` if along with ` + "`module.exports =`" + `
module.exports = exports = {}
exports = module.exports = {}`, Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "// This assigned object is not exported.\n// You need to use `module.exports = { ... }`.\nexports = {\n    foo: 1\n}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(3, 1, 5, 2)}},
			{Code: "exports = {}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
		},
	)
}
