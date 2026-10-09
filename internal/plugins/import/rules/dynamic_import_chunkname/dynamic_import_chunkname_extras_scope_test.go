package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// strictModeRows: Strict-mode early errors in class bodies, which are strict code; method bodies are parsed although they never run.
// Checked against the pinned upstream rule and vm.runInNewContext.
var strictModeRows = []classSyntaxRow{
	{"class { m() { delete missing } }", false},
	{"class { m() { delete (missing) } }", false},
	{"class { static x = delete (missing) }", false},
	{"class { m() { 010 } }", false},
	{"class { static x = 010 }", false},
	{"class { static x = 08 }", false},
	{"class { static x = 0 }", true},
	{"class { m() { eval = 1 } }", false},
	{"class { m() { arguments = 1 } }", false},
	{"class { static x = (eval = 1) }", false},
	{"class { static x = (fallback = 1) }", false},
	{"class { m() { fallback = 1 } }", true},
	{"class { static { fallback = 1 } }", false},
	{"class { static x = (undefined = 1) }", false},
	{"class { static x = (NaN = 1) }", false},
	{"class { static x = (Object = 1) }", true},
	{"class { static x = (console = 1) }", true},
	{"class { static x = console }", true},
	{"class { static x = typeof console }", true},
	{"class { static x = (fallback += 1) }", false},
	{"class { static x = (fallback++) }", false},
	{"class { static x = ([fallback] = [1]) }", false},
	{"class { static x = ({a: fallback} = {a: 1}) }", false},
	{"class { static x = ([a] = [1]) }", false},
	{"class { static x = (1, fallback = 1) }", false},
	{"class { m(a, a) {} }", false},
	{"class { static x = (let = 1) }", false},
	{"class { m() { var let = 1 } }", false},
	{"class { m() { var yield } }", false},
	{"class { static x = '\\01' }", false},
	{"class { m() { with (a) {} } }", false},
	{"class { static x = (x = 1) }", false},
	{"class { static x = (static = 1) }", false},
}

// classNameRows: Reads and writes of a class's own name: unbound in the extends clause and computed keys, bound in static initializers and blocks, and immutable.
// Checked against the pinned upstream rule and vm.runInNewContext.
var classNameRows = []classSyntaxRow{
	{"class A { static x = A }", true},
	{"class A { [A]() {} }", false},
	{"class A { [A] = 1 }", false},
	{"class A extends A {}", false},
	{"class A { static x = (A = 1) }", false},
	{"class A { static { A } }", true},
	{"class A { m() { return A } }", true},
	{"class A { static m() { A = 1 } }", true},
	{"class A { static x = A; static y = missing }", false},
	{"class A { static [A] = 1 }", false},
	{"class A extends (class B { static x = A }) {}", false},
	{"class A { static x = class B { static y = A } }", true},
	{"class { static x = (static) }", false},
	{"class { m() { var x = static } }", false},
	{"class { m() { let = 1 } }", false},
	{"class { m() { yield = 1 } }", false},
	{"class { m() { implements = 1 } }", false},
	{"class { m(x) { var x } }", true},
	{"class { m() { function f(a, a) {} } }", false},
	{"class { m(a) { with (a) {} } }", false},
	{"class { m() { var x = 'a'; x = eval } }", true},
	{"class { static x = 0x10 }", true},
	{"class { static x = 0o10 }", true},
	{"class { static x = 1e3 }", true},
	{"class { static x = 00 }", false},
	{"class { static x = 0.5 }", true},
	{"class { static x = '\\0' }", true},
	{"class { static x = '\\8' }", false},
	{"class { m() { var o = {static: 1, let: 2} } }", true},
}

// staticBlockRows: Names declared inside a static block are in scope for it; let, const and class declarations in nested blocks are not visible outside them.
// Checked against the pinned upstream rule and vm.runInNewContext.
var staticBlockRows = []classSyntaxRow{
	{"class { static { let x = 1; x } }", true},
	{"class { static { let x = missing; } }", false},
	{"class { static { var x; x = 1 } }", true},
	{"class { static { x = 1 } }", false},
	{"class { static { function f() {} f() } }", true},
	{"class { static { { let y = 1 } y } }", false},
	{"class { static { const { a } = { a: 1 }; a } }", true},
	{"class { static { class C {} C } }", true},
	{"class { static { let x = 1; } static y = x }", false},
	{"class { static { let [a, b] = [1, 2]; a + b } }", true},
}

func TestDynamicImportChunknameClassScopeTables(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, rows := range [][]classSyntaxRow{strictModeRows, classNameRows, staticBlockRows} {
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
