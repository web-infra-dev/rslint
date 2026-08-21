// TestBlockScopedVarUpstream migrates the full valid/invalid suite from
// ESLint v10.8.1 tests/lib/rules/block-scoped-var.js 1:1. rslint's parser
// does not gate syntax on languageOptions.ecmaVersion the way espree does, so
// that option is dropped from ported cases unless it also selects globals
// this rule cares about (it does not, so it never appears below).
// languageOptions.sourceType is dropped too: rslint infers module-ness from
// actual import/export syntax rather than an explicit flag, and every
// upstream case that sets it either has no such syntax (module-ness makes no
// observable difference here) or already has real import/export (module-ness
// is inferred anyway). rslint-specific lock-ins live in
// block_scoped_var_extras_test.go.
package block_scoped_var

import (
	"fmt"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// scopeErr builds the expected outOfScope error for a use reported at
// line/column, whose declaration data.name/definitionLine/definitionColumn
// point back at defLine/defColumn.
func scopeErr(name string, defLine, defColumn, line, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "outOfScope",
		Message:   fmt.Sprintf("'%s' declared on line %d column %d is used outside of binding context.", name, defLine, defColumn),
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name),
	}
}

func outOfScope(code string, errors ...rule_tester.InvalidTestCaseError) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{Code: code, Errors: errors}
}

func TestBlockScopedVarUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&BlockScopedVarRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: https://github.com/eslint/eslint/issues/2242 ----
			{Code: "function f() { } f(); var exports = { f: f };"},
			{Code: "var f = () => {}; f(); var exports = { f: f };"},

			// ---- upstream valid: basic function/var scoping ----
			{Code: "!function f(){ f; }"},
			{Code: "function f() { } f(); var exports = { f: f };"},
			{Code: "function f() { var a, b; { a = true; } b = a; }"},
			{Code: "var a; function f() { var b = a; }"},
			{Code: "function f(a) { }"},
			{Code: "!function(a) { };"},
			{Code: "!function f(a) { };"},
			{Code: "function f(a) { var b = a; }"},
			{Code: "!function f(a) { var b = a; };"},
			{Code: "function f() { var g = f; }"},
			{Code: "function f() { } function g() { var f = g; }"},
			{Code: "function f() { var hasOwnProperty; { hasOwnProperty; } }"},
			{Code: "function f(){ a; b; var a, b; }"},
			{Code: "function f(){ g(); function g(){} }"},
			{Code: "if (true) { var a = 1; a; }"},
			{Code: "var a; if (true) { a; }"},
			{Code: "for (var i = 0; i < 10; i++) { i; }"},
			{Code: "var i; for(i; i; i) { i; }"},
			{Code: `function myFunc(foo) {  "use strict";  var { bar } = foo;  bar.hello();}`},
			{Code: `function myFunc(foo) {  "use strict";  var [ bar ]  = foo;  bar.hello();}`},
			{Code: "function myFunc(...foo) {  return foo;}"},
			{Code: "var f = () => { var g = f; }"},
			{Code: "class Foo {}\nexport default Foo;"},
			{Code: "new Date"},
			{Code: "var eslint = require('eslint');"},
			{Code: "var fun = function({x}) {return x;};"},
			{Code: "var fun = function([,x]) {return x;};"},
			{Code: "function f(a) { return a.b; }"},
			{Code: `var a = { "foo": 3 };`},
			{Code: "var a = { foo: 3 };"},
			{Code: "var a = { foo: 3, bar: 5 };"},
			{Code: "var a = { set foo(a){}, get bar(){} };"},
			{Code: "function f(a) { return arguments[0]; }"},
			{Code: "function f() { }; var a = f;"},
			{Code: "var a = f; function f() { };"},
			{Code: "function f(){ for(var i; i; i) i; }"},
			{Code: "function f(){ for(var a=0, b=1; a; b) a, b; }"},
			{Code: "function f(){ for(var a in {}) a; }"},
			{Code: "function f(){ switch(2) { case 1: var b = 2; b; break; default: b; break;} }"},
			{Code: "a:;"},
			{Code: "foo: while (true) { bar: for (var i = 0; i < 13; ++i) {if (i === 7) break foo; } }"},
			{Code: "foo: while (true) { bar: for (var i = 0; i < 13; ++i) {if (i === 7) continue foo; } }"},
			{Code: `const React = require("react/addons");const cx = React.addons.classSet;`},
			{Code: "var v = 1;  function x() { return v; };"},
			{Code: `import * as y from "./other.js"; y();`},
			{Code: `import y from "./other.js"; y();`},
			{Code: `import {x as y} from "./other.js"; y();`},
			{Code: "var x; export {x};"},
			{Code: "var x; export {x as v};"},
			{Code: `export {x} from "./other.js";`},
			{Code: `export {x as v} from "./other.js";`},
			{Code: "class Test { myFunction() { return true; }}"},
			{Code: "class Test { get flag() { return true; }}"},
			{Code: "var Test = class { myFunction() { return true; }}"},
			{Code: "var doStuff; let {x: y} = {x: 1}; doStuff(y);"},
			{Code: "function foo({x: y}) { return y; }"},

			// ---- upstream valid: same coverage as no-undef ----
			{Code: "!function f(){}; f"},
			{Code: "var f = function foo() { }; foo(); var exports = { f: foo };"},
			{Code: "var f = () => { x; }"},
			{Code: "function f(){ x; }"},
			{Code: "var eslint = require('eslint');"},
			{Code: "function f(a) { return a[b]; }"},
			{Code: "function f() { return b.a; }"},
			{Code: "var a = { foo: bar };"},
			{Code: "var a = { foo: foo };"},
			{Code: "var a = { bar: 7, foo: bar };"},
			{Code: "var a = arguments;"},
			{Code: "function x(){}; var a = arguments;"},
			{Code: "function z(b){}; var a = b;"},
			{Code: "function z(){var b;}; var a = b;"},
			{Code: "function f(){ try{}catch(e){} e }"},
			{Code: "a:b;"},

			// ---- upstream valid: https://github.com/eslint/eslint/issues/2253 ----
			{Code: "/*global React*/ let {PropTypes, addons: {PureRenderMixin}} = React; let Test = React.createClass({mixins: [PureRenderMixin]});"},
			{Code: "/*global prevState*/ const { virtualSize: prevVirtualSize = 0 } = prevState;"},
			{Code: "const { dummy: { data, isLoading }, auth: { isLoggedIn } } = this.props;"},

			// ---- upstream valid: https://github.com/eslint/eslint/issues/2747 ----
			{Code: `function a(n) { return n > 0 ? b(n - 1) : "a"; } function b(n) { return n > 0 ? a(n - 1) : "b"; }`},

			// ---- upstream valid: https://github.com/eslint/eslint/issues/2967 ----
			{Code: "(function () { foo(); })(); function foo() {}"},

			// ---- upstream valid: class static blocks ----
			{Code: "class C { static { var foo; foo; } }"},
			{Code: "class C { static { foo; var foo; } }"},
			{Code: "class C { static { if (bar) { foo; } var foo; } }"},
			{Code: "var foo; class C { static { foo; } } "},
			{Code: "class C { static { foo; } } var foo;"},
			{Code: "var foo; class C { static {} [foo]; } "},
			{Code: "foo; class C { static {} } var foo; "},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: basic out-of-block reads ----
			outOfScope("function f(){ x; { var x; } }",
				scopeErr("x", 1, 24, 1, 15)),
			outOfScope("function f(){ { var x; } x; }",
				scopeErr("x", 1, 21, 1, 26)),
			outOfScope("function f() { var a; { var b = 0; } a = b; }",
				scopeErr("b", 1, 29, 1, 42)),
			outOfScope("function f() { try { var a = 0; } catch (e) { var b = a; } }",
				scopeErr("a", 1, 26, 1, 55)),
			outOfScope("function a() { for(var b in {}) { var c = b; } c; }",
				scopeErr("c", 1, 39, 1, 48)),
			outOfScope("function a() { for(var b of {}) { var c = b; } c; }",
				scopeErr("c", 1, 39, 1, 48)),
			outOfScope("function f(){ switch(2) { case 1: var b = 2; b; break; default: b; break;} b; }",
				scopeErr("b", 1, 39, 1, 76)),
			outOfScope("for (var a = 0;;) {} a;",
				scopeErr("a", 1, 10, 1, 22)),
			outOfScope("for (var a in []) {} a;",
				scopeErr("a", 1, 10, 1, 22)),
			outOfScope("for (var a of []) {} a;",
				scopeErr("a", 1, 10, 1, 22)),
			outOfScope("{ var a = 0; } a;",
				scopeErr("a", 1, 7, 1, 16)),
			outOfScope("if (true) { var a; } a;",
				scopeErr("a", 1, 17, 1, 22)),

			// ---- upstream invalid: sibling `var` declarations flag each other ----
			outOfScope("if (true) { var a = 1; } else { var a = 2; }",
				scopeErr("a", 1, 37, 1, 17),
				scopeErr("a", 1, 17, 1, 37)),
			outOfScope("for (var i = 0;;) {} for(var i = 0;;) {}",
				scopeErr("i", 1, 30, 1, 10),
				scopeErr("i", 1, 10, 1, 30)),

			outOfScope("class C { static { if (bar) { var foo; } foo; } }",
				scopeErr("foo", 1, 35, 1, 42)),

			outOfScope("{ var foo,\n  bar; } bar;",
				scopeErr("bar", 2, 3, 2, 10)),
			outOfScope("{ var { foo,\n  bar } = baz; } bar;",
				scopeErr("bar", 2, 3, 2, 18)),

			outOfScope("if (foo) { var a = 1; } else if (bar) { var a = 2; } else { var a = 3; }",
				scopeErr("a", 1, 45, 1, 16),
				scopeErr("a", 1, 65, 1, 16),
				scopeErr("a", 1, 16, 1, 45),
				scopeErr("a", 1, 65, 1, 45),
				scopeErr("a", 1, 16, 1, 65),
				scopeErr("a", 1, 45, 1, 65)),
		},
	)
}
