package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Reserved words as property keys in binding patterns, which are names rather than identifiers, and the built-in names that a fresh vm context defines.
// Each row is an expression placed after `x:` in a magic comment that also
// names the chunk. Validity was recorded from the pinned upstream rule and
// checked against Node.js's vm.runInNewContext, which the rule uses.

var bindingPropertyNameRows = []classSyntaxRow{
	{"class { m({static: value}) {} }", true},
	{"class { m() { const {static: value} = {}; } }", true},
	{"class { m() { const {let: a, yield: b} = {}; } }", true},
	{"class { m({static}) {} }", false},
	{"class { m({a: static}) {} }", false},
	{"class { m() { var {static} = {} } }", false},
	{"class { m([static]) {} }", false},
	{"class { m(static) {} }", false},
	{"class { static x = ({static: 1}).static }", true},
	{"class { static x = ({let}) }", false},
	{"class { m({static: {let: x}}) {} }", true},
	{"class { m({static: {let}}) {} }", false},
	{"class { m({implements: a = 1}) {} }", true},
	{"class { m({a = static}) {} }", false},
	{"class { m({static = 1}) {} }", false},
	{"class { m({[static]: a}) {} }", false},
	{"class { m() { return {static: 1, let: 2, yield: 3, interface: 4, package: 5, private: 6, protected: 7, public: 8, implements: 9} } }", true},
	{"class { m() { return a.static + a.let + a.yield } }", true},
	{"class { m() { return a?.static } }", true},
	{"class { static x = {static() {}, let() {}, get yield() { return 1 }, set package(v) {}} }", true},
	{"class { static() {} let() {} yield() {} }", true},
	{"class { static static() {} }", true},
	{"class { let = 1; yield = 2; }", true},
	{"class { get static() { return 1 } set static(v) {} }", true},
	{"class { m() { static: while (false) {} } }", false},
	{"class { m() { var x = {static} } }", false},
	{"class { m() { function static() {} } }", false},
	{"class { m() { class static {} } }", false},
	{"class { m() { try {} catch (static) {} } }", false},
	{"class { m() { try {} catch ({static: s}) {} } }", true},
	{"class { m() { for (const {static: s} of []) {} } }", true},
	{"class { m() { for (const {static} of []) {} } }", false},
	{"class { m() { ({static: value} = {}) } }", true},
	{"class { m() { ({static} = {}) } }", false},
	{"class { m() { [static] = [] } }", false},
}

var builtinNameRows = []classSyntaxRow{
	{"Iterator", true},
	{"Float16Array", true},
	{"DisposableStack", true},
	{"AsyncDisposableStack", true},
	{"SuppressedError", true},
	{"Temporal", true},
	{"typeof Temporal", true},
	{"Iterator.prototype", true},
	{"Iterators", false},
	{"iterator", false},
	{"Object", true},
	{"Reflect", true},
	{"Atomics", true},
	{"WebAssembly", true},
	{"console", true},
	{"globalThis", true},
	{"escape", true},
	{"queueMicrotask", false},
	{"setTimeout", false},
	{"process", false},
	{"require", false},
	{"module", false},
	{"Buffer", false},
	{"window", false},
	{"self", false},
	{"structuredClone", false},
	{"TextEncoder", false},
	{"URL", false},
	{"fetch", false},
}

func TestDynamicImportChunknameBindingNameTables(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, rows := range [][]classSyntaxRow{bindingPropertyNameRows, builtinNameRows} {
		for _, row := range rows {
			code := "import(\n  /* webpackChunkName: \"a\", x: " + row.expression + " */\n  'm',\n)"
			if row.valid {
				valid = append(valid, rule_tester.ValidTestCase{Code: code})
				continue
			}
			invalid = append(invalid, rule_tester.InvalidTestCase{
				Code: code,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: classInvalidSyntaxMessage, Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			})
		}
	}
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule, valid, invalid)
}
