package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

const classInvalidSyntaxMessage = `dynamic imports require a "webpack" comment with valid syntax`

type classSyntaxRow struct {
	expression string
	valid      bool
}

// classSyntaxRows are expressions placed after `x:` in a magic comment that
// also names the chunk. Each row's validity was checked against the pinned
// upstream rule and against Node.js's vm.runInNewContext: they agree on every
// row. They cover class members that use TypeScript-only syntax (modifiers,
// optional and definite markers, property types, bodiless members, parameter
// properties and this parameters), and type arguments on calls, new
// expressions and tagged templates, which JavaScript does not accept.
var classSyntaxRows = []classSyntaxRow{
	{"class { readonly x = 1 }", false},
	{"class { private x = 1 }", false},
	{"class { public x = 1 }", false},
	{"class { protected x = 1 }", false},
	{"class { declare x = 1 }", false},
	{"class { abstract m() {} }", false},
	{"class { override m() {} }", false},
	{"class { x? }", false},
	{"class { x! }", false},
	{"class { x: number }", false},
	{"class { x: number = 1 }", false},
	{"class { m?() {} }", false},
	{"class { m(): void {} }", false},
	{"class { m<T>() {} }", false},
	{"class { m(a?) {} }", false},
	{"class { m(): void; }", false},
	{"class { constructor(); }", false},
	{"class { get a(): number { return 1 } }", false},
	{"class { static x = 1 }", true},
	{"class { static m() {} }", true},
	{"class { async m() {} }", true},
	{"class { *g() {} }", true},
	{"class { #p = 1 }", true},
	{"class { static { } }", true},
	{"class { [k] = 1 }", false},
	{"class { 'a' = 1 }", true},
	{"class { x = 1; m() { return 1 } }", true},
	{"class { constructor(a) {} }", true},
	{"class { constructor(public a) {} }", false},
	{"class { m(this) {} }", false},
	{"class { m(this: X) {} }", false},
	{"class extends B {}", false},
	{"class extends B<T> {}", false},
	{"class implements I {}", false},
	{"class { static readonly x = 1 }", false},
	{"class { override readonly x = 1 }", false},
	{"class { private static x = 1 }", false},
	{"class { declare readonly x }", false},
	{"class { accessor x = 1 }", false},
	{"class { static async *g() {} }", true},
	{"class { x = 1 }", true},
	{"class { 1 = 2 }", true},
	{"class { static x? }", false},
	{"class { get x() {} }", true},
	{"class { set x(v) {} }", true},
	{"class { set x(v?) {} }", false},
	{"class { m() }", false},
	{"class { get x() }", false},
	{"class { constructor() }", false},
	{"class { x; y }", true},
	{"class { x }", true},
	{"f<number>()", false},
	{"new C<number>()", false},
	{"tag<T>`x`", false},
	{"f?.<number>()", false},
	{"a<b>", false},
	{"a < b > c", false},
	{"f<number>", false},
	{"f(<number>x)", false},
	{"f<number>(1)[0]", false},
	{"(f<number>)()", false},
	{"x?.y<number>()", false},
	{"new C<T>", false},
	{"class { m(...a: number[]) {} }", false},
	{"class { m(a: number) {} }", false},
	{"class { [k]?: number }", false},
	{"class { 'b'?() {} }", false},
	{"class { #p?: number }", false},
	{"class { static m(): void; }", false},
}

func TestDynamicImportChunknameClassSyntaxTable(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, row := range classSyntaxRows {
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule, valid, invalid)
}

// classReadRows are class expressions whose validity depends on what runs
// when the class is defined: the extends clause, computed member names, and
// static initializers and blocks. Instance initializers and method bodies do
// not run at definition. Checked against the pinned upstream rule and
// vm.runInNewContext as above.
var classReadRows = []classSyntaxRow{
	{"class { static x = missing }", false},
	{"class { x = missing }", true},
	{"class { m() { return missing } }", true},
	{"class { [missing]() {} }", false},
	{"class extends missing {}", false},
	{"class A extends missing {}", false},
	{"class A {}", true},
	{"class { static { missing } }", false},
	{"class { static { } }", true},
	{"class { static m() { missing } }", true},
	{"class { static [missing] = 1 }", false},
	{"class { static [k]() {} }", false},
	{"class { [k] = 1 }", false},
	{"class { static x = (() => missing) }", true},
	{"class { x = missing; static y = 1 }", true},
	{"class { static x = 1; y = missing }", true},
	{"class extends (missing) {}", false},
	{"class extends fallback || Base {}", false},
	{"class extends (0 || missing) {}", false},
	{"class { get [missing]() { return 1 } }", false},
	{"class { static get x() { return missing } }", true},
	{"class { 'a' = missing }", true},
	{"class { static async m() { missing } }", true},
	{"class A { static x = A }", true},
	{"class { static x = typeof missing }", true},
	{"class { static x = delete missing }", false},
	{"class { static x = missing.p }", false},
}

func TestDynamicImportChunknameClassReadsTable(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, row := range classReadRows {
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
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &dynamic_import_chunkname.DynamicImportChunknameRule, valid, invalid)
}
