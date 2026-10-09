package dynamic_import_chunkname_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/import/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/import/rules/dynamic_import_chunkname"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Statements in class static blocks: only statements that are certain to run are searched, and the names a block declares are scoped to it. Not covered on purpose, because they need loop, switch and exception flow or a temporal dead zone to decide: a do-while body, a switch case, a try block, an if condition that is not a constant, and a let read inside its own initializer.
// Each row is an expression placed after `x:` in a magic comment that also
// names the chunk. Validity was recorded from the pinned upstream rule and
// checked against Node.js's vm.runInNewContext, which the rule uses.

var staticBlockStatementRows = []classSyntaxRow{
	{"class { static { if (false) missing; } }", true},
	{"class { static { while (false) missing; } }", true},
	{"class { static { if (true) {} else missing } }", true},
	{"class { static { for (;false;) missing } }", true},
	{"class { static { do { break; missing } while (false) } }", true},
	{"class { static { L: { break L; missing } } }", true},
	{"class { static { { let a = 1; a; } } }", true},
	{"class { static { for (let i = 0; i < 1; i++) {} } }", true},
	{"class { static { for (const x of [1]) x } }", true},
	{"class { static { for (var k in {a: 1}) k } }", true},
	{"class { static { try { } catch (e) { e } } }", true},
	{"class { static { try { throw 1 } catch (e) { e } } }", true},
	{"class { static { switch (1) { case 1: let a = 1; a } } }", true},
	{"class { static { if (true) { let a = 1; a } } }", true},
	{"class { static { if (missing) {} } }", false},
	{"class { static { if (true) missing } }", false},
	{"class { static { if (false) {} else missing } }", false},
	{"class { static { missing } }", false},
	{"class { static { let a = 1; { a; missing } } }", false},
	{"class { static { { let a = 1 } a } }", false},
	{"class { static { var a; if (false) { var b } b } }", true},
	{"class { static { if (false) { var b = 1 } b } }", true},
	{"class { static { throw 1 } }", false},
	{"class { static { function f() { missing } } }", true},
	{"class { static { class C extends missing {} } }", false},
	{"class { static { class C { static x = missing } } }", false},
	{"class { static { class C { static x = C } } }", true},
	{"class { static { class C { x = missing } } }", true},
	{"class { static { let f = () => missing; } }", true},
	{"class { static { let a = missing; } }", false},
	{"class { static { const a = 1, b = a; b } }", true},
	{"class { static { let a = 1; let b = a; b } }", true},
	{"class { static { let {a} = {a: 1}; a } }", true},
	{"class { static { let [a] = [1]; a } }", true},
	{"class { static { { function f() {} f() } } }", true},
	{"class { static { for (let i = 0, j = i; j < 1; j++) {} } }", true},
	{"class { static { for (const [k, v] of [[1, 2]]) k + v } }", true},
	{"class { static { for (let i = 0; i < 1; i++) { i } i } }", false},
	{"class { static { try { } finally { missing } } }", false},
	{"class { static { ; } }", true},
	{"class { static { a: ; } }", true},
	{"class { static { var a = 1; var a = 2; a } }", true},
	{"class { static { let a; { let a; a } a } }", true},
	{"class { static { { { let a = 1; a } } } }", true},
	{"class { static { { var a = 1 } a } }", true},
	{"class { static { if (false) var a; a } }", true},
	{"class { static { return } }", false},
	{"class { static { arguments } }", false},
	{"class { static { await } }", false},
	{"class { static { yield } }", false},
	{"class { static { () => arguments } }", false},
	{"class { static { function f() { return arguments } } }", true},
	{"class { static { break } }", false},
	{"class { static { continue } }", false},
	{"class { static { L: { break M } } }", false},
	{"class { static { for (;;) { break } } }", true},
	{"class { static { while (true) break; } }", true},
	{"class { static { do {} while (false) } }", true},
	{"class { static { debugger } }", true},
	{"class { static { 'use strict' } }", true},
	{"class { static { new.target } }", true},
	{"class { static { this } }", true},
	{"class { static { super.x } }", true},
	{"class { static { yield } }", false},
	{"class { static { await } }", false},
	{"class { static { await 1 } }", false},
	{"class { static { async function f() { await 1 } } }", true},
	{"class { static { for (var i = 0; i < 1; i++) {} i } }", true},
	{"class { static { for (let i of missing) {} } }", false},
	{"class { static { for (let i in missing) {} } }", false},
	{"class { static { for (missing;;) { break } } }", false},
	{"class { static { for (;missing;) {} } }", false},
	{"class { static { while (missing) {} } }", false},
	{"class { static { switch (missing) {} } }", false},
	{"class { static { try { missing } catch (e) {} } }", true},
	{"class { static { try { } catch (e) { missing } } }", true},
	{"class { static { try { missing } catch (e) { } finally { } } }", true},
	{"class { static { if (1) { missing } } }", false},
	{"class { static { if (0) { missing } else { missing2 } } }", false},
	{"class { static { if (0) { missing } else { 1 } } }", true},
	{"class { static { if (undefined) missing } }", true},
	{"class { static { if (!1) missing } }", true},
	{"class { static { if (null ?? 1) missing } }", false},
	{"class { static { if (x) missing } }", false},
	{"class { static { let a = 1; a = missing } }", false},
	{"class { static { var a = missing } }", false},
	{"class { static { let a = () => missing; a } }", true},
	{"class { static { let a = function () { missing }; a } }", true},
	{"class { static { let a = class { m() { missing } }; a } }", true},
	{"class { static { let a = class { static x = missing }; } }", false},
	{"class { static { class C { static { missing } } } }", false},
	{"class { static { class C { static { var q; q } } } }", true},
	{"class { static { static_ = 1 } }", false},
	{"class { static { let x; { x = 1 } } }", true},
	{"class { static { let x; { var x2 = x } x2 } }", true},
	{"class { static { { var inner = 1 } inner } }", true},
	{"class { static { { function f() {} } f } }", false},
	{"class { static { function f() {} f } }", true},
	{"class { static { { class C {} } C } }", false},
	{"class { static { { class C {} C } } }", true},
	{"class { static { L: for (;;) { break L } } }", true},
	{"class { static { const a = 1; { const a = 2; a } a } }", true},
	{"class { static { { let a = 1; { a } } } }", true},
	{"class { static { for (let i = 0; i < 1; i++) { let i2 = i } } }", true},
	{"class { static { for (let i = 0;; ) { break } i } }", false},
	{"class { static { for (const x of []) {} x } }", false},
	{"class { static { for (var v of []) {} v } }", true},
	{"class { static { for (var v in {}) {} v } }", true},
	{"class { static { try {} catch (e) {} e } }", false},
	{"class { static { try {} catch { } } }", true},
	{"class { static { switch (1) { case 1: let a; } a } }", false},
	{"class { static { label: function f() {} } }", false},
	{"class { static { ({ missing }) } }", false},
	{"class { static { ({ a: missing }) } }", false},
	{"class { static { [missing] } }", false},
	{"class { static { `${missing}` } }", false},
	{"class { static { missing`x` } }", false},
	{"class { static { void missing } }", false},
	{"class { static { typeof missing } }", true},
	{"class { static { delete missing } }", false},
	{"class { static { missing++ } }", false},
	{"class { static { missing = 1 } }", false},
	{"class { static { missing += 1 } }", false},
	{"class { static { ({ missing } = {}) } }", false},
	{"class { static { [missing] = [] } }", false},
	{"class { static { ({ a: missing } = {}) } }", false},
}

func TestDynamicImportChunknameStaticBlockTables(t *testing.T) {
	var valid []rule_tester.ValidTestCase
	var invalid []rule_tester.InvalidTestCase
	for _, rows := range [][]classSyntaxRow{staticBlockStatementRows} {
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
