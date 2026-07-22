// TestNoRedeclareUpstream migrates the full valid/invalid suite from ESLint
// v10.7.0 tests/lib/rules/no-redeclare.js 1:1. Position assertions cover
// line/column/endLine/endColumn for every invalid case that rslint can execute.
// Unsupported parser/scope configuration cases remain in place with Skip and
// an explicit reason. rslint-specific lock-ins live in no_redeclare_extras_test.go.
package no_redeclare

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRedeclareUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRedeclareRule,
		[]rule_tester.ValidTestCase{
			// ---- upstream valid: basic non-redeclarations ----
			{Code: "var a = 3; var b = function() { var a = 10; };"},
			{Code: "var a = 3; a = 10;"},
			{Code: "if (true) {\n    let b = 2;\n} else {    \nlet b = 3;\n}"},

			// ---- upstream valid: class static block scope boundaries ----
			{Code: "var a; class C { static { var a; } }"},
			{Code: "class C { static { var a; } } var a; "},
			{Code: "function a(){} class C { static { var a; } }"},
			{Code: "var a; class C { static { function a(){} } }"},
			{Code: "class C { static { var a; } static { var a; } }"},
			{Code: "class C { static { function a(){} } static { function a(){} } }"},
			{Code: "class C { static { var a; { function a(){} } } }"},
			{Code: "class C { static { function a(){}; { function a(){} } } }"},
			{Code: "class C { static { var a; { let a; } } }"},
			{Code: "class C { static { let a; { let a; } } }"},
			{Code: "class C { static { { let a; } { let a; } } }"},

			// ---- upstream valid: builtinGlobals option ----
			{Code: "var Object = 0;", Options: map[string]interface{}{"builtinGlobals": false}},
			// SKIP: rslint does not support ESLint's sourceType override without import/export syntax.
			{Code: "var Object = 0;", Options: map[string]interface{}{"builtinGlobals": true}, Skip: true},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn.
			{Code: "var Object = 0;", Options: map[string]interface{}{"builtinGlobals": true}, Skip: true},
			{Code: "var top = 0;", Options: map[string]interface{}{"builtinGlobals": true}},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn.
			{Code: "var top = 0;", Options: map[string]interface{}{"builtinGlobals": true}, Globals: map[string]bool{"top": true}, Skip: true},
			// SKIP: rslint does not support ESLint's sourceType override plus browser globals.
			{Code: "var top = 0;", Options: map[string]interface{}{"builtinGlobals": true}, Skip: true},
			{Code: "var self = 1", Options: map[string]interface{}{"builtinGlobals": true}},
			// SKIP: rslint does not model ESLint's ecmaVersion-specific builtin global set.
			{Code: "var globalThis = foo", Options: map[string]interface{}{"builtinGlobals": true}, Skip: true},
			// SKIP: rslint does not model ESLint's ecmaVersion-specific builtin global set.
			{Code: "var globalThis = foo", Options: map[string]interface{}{"builtinGlobals": true}, Skip: true},

			// ---- upstream valid: directive comments and configured globals ----
			{Code: "/*globals Array */", Options: map[string]interface{}{"builtinGlobals": false}},
			{Code: "/*globals a */", Options: map[string]interface{}{"builtinGlobals": false}, Globals: map[string]bool{"a": true}},
			{Code: "/*globals a */", Options: map[string]interface{}{"builtinGlobals": false}, Globals: map[string]bool{"a": true}},
			{Code: "/*globals a:off */", Options: map[string]interface{}{"builtinGlobals": true}, Globals: map[string]bool{"a": true}},
			{Code: "/*globals a */", Options: map[string]interface{}{"builtinGlobals": true}, Globals: map[string]bool{"a": false}},
		},
		[]rule_tester.InvalidTestCase{
			// ---- upstream invalid: basic var/function redeclarations ----
			invalidRedeclared("var a = 3; var a = 10;", "a", 1, 16),
			invalidRedeclared("switch(foo) { case a: var b = 3;\ncase b: var b = 4}", "b", 2, 13),
			invalidRedeclared("var a = 3; var a = 10;", "a", 1, 16),
			invalidRedeclared("var a = {}; var a = [];", "a", 1, 17),
			invalidRedeclared("var a; function a() {}", "a", 1, 17),
			invalidRedeclared("function a() {} function a() {}", "a", 1, 26),
			invalidRedeclared("var a = function() { }; var a = function() { }", "a", 1, 29),
			invalidRedeclared("var a = function() { }; var a = new Date();", "a", 1, 29),
			{
				Code: "var a = 3; var a = 10; var a = 15;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 16),
					redeclaredError("a", 1, 28),
				},
			},
			// SKIP: rslint does not support ESLint's sourceType override without import/export syntax.
			{
				Code: "var a; var a;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 12),
				},
				Skip: true,
			},
			invalidRedeclared("export var a; var a;", "a", 1, 19),

			// ---- upstream invalid: var redeclaration in class static blocks ----
			invalidRedeclared("class C { static { var a; var a; } }", "a", 1, 31),
			invalidRedeclared("class C { static { var a; { var a; } } }", "a", 1, 33),
			invalidRedeclared("class C { static { { var a; } var a; } }", "a", 1, 35),
			invalidRedeclared("class C { static { { var a; } { var a; } } }", "a", 1, 37),

			// ---- upstream invalid: builtinGlobals ----
			invalidBuiltin("var Object = 0;", "Object", 1, 5),
			{Code: "var top = 0;", Options: map[string]interface{}{"builtinGlobals": true}, Globals: map[string]bool{"top": true}, Errors: []rule_tester.InvalidTestCaseError{builtinError("top", 1, 5)}},
			{
				Code:    "var a; var {a = 0, b: Object = 0} = {};",
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 13),
					builtinError("Object", 1, 23),
				},
			},
			// SKIP: rslint does not support ESLint's sourceType override without import/export syntax.
			{
				Code:    "var a; var {a = 0, b: Object = 0} = {};",
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 13),
				},
				Skip: true,
			},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn.
			{
				Code:    "var a; var {a = 0, b: Object = 0} = {};",
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 13),
				},
				Skip: true,
			},
			{
				Code:    "var a; var {a = 0, b: Object = 0} = {};",
				Options: map[string]interface{}{"builtinGlobals": false},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 13),
				},
			},
			invalidBuiltin("var globalThis = 0;", "globalThis", 1, 5),
			{
				Code:    "var a; var {a = 0, b: globalThis = 0} = {};",
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 13),
					builtinError("globalThis", 1, 23),
				},
			},

			// ---- upstream invalid: directive comments ----
			{
				Code:    "/*global b:false*/ var b = 1;",
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredBySyntaxError("b", 1, 10),
				},
			},
			{
				Code:    "/*global b:true*/ var b = 1;",
				Options: map[string]interface{}{"builtinGlobals": true},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredBySyntaxError("b", 1, 10),
				},
			},

			// ---- upstream invalid: function and for scopes ----
			invalidRedeclared("function f() { var a; var a; }", "a", 1, 27),
			invalidRedeclared("function f(a) { var a; }", "a", 1, 21),
			invalidRedeclared("function f() { var a; if (test) { var a; } }", "a", 1, 39),
			invalidRedeclared("for (var a, a;;);", "a", 1, 13),

			// ---- upstream invalid: default options and browser globals ----
			invalidBuiltin("var Object = 0;", "Object", 1, 5),
			{Code: "var top = 0;", Globals: map[string]bool{"top": true}, Errors: []rule_tester.InvalidTestCaseError{builtinError("top", 1, 5)}},

			// ---- upstream invalid: directive comments and configured globals ----
			invalidBuiltin("/*globals Array */", "Array", 1, 11),
			invalidBuiltin("/*globals parseInt */", "parseInt", 1, 11),
			invalidBuiltin("/*globals foo, Array */", "Array", 1, 16),
			invalidBuiltin("/* globals foo, Array, baz */", "Array", 1, 17),
			invalidBuiltin("/*global foo, Array, baz*/", "Array", 1, 15),
			invalidBuiltin("/*global array, Array*/", "Array", 1, 17),
			invalidBuiltin("/*globals a,Array*/", "Array", 1, 13),
			invalidBuiltin("/*globals a:readonly, Array:writable */", "Array", 1, 23),
			invalidBuiltin("\n/*globals Array */", "Array", 2, 11),
			invalidBuiltin("/*globals\nArray */", "Array", 2, 1),
			invalidBuiltin("\n/*globals\n\nArray*/", "Array", 4, 1),
			invalidBuiltin("/*globals foo,\n    Array */", "Array", 2, 5),
			{
				Code:    "/*globals a */",
				Options: map[string]interface{}{"builtinGlobals": true},
				Globals: map[string]bool{"a": true},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("a", 1, 11),
				},
			},
			{
				Code:    "/*globals a */",
				Options: map[string]interface{}{"builtinGlobals": true},
				Globals: map[string]bool{"a": true},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("a", 1, 11),
				},
			},
			{
				Code: "/*globals a */ /*globals a */",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 26),
				},
			},
			{
				Code:    "/*globals a */ /*globals a */ var a = 0",
				Options: map[string]interface{}{"builtinGlobals": true},
				Globals: map[string]bool{"a": true},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("a", 1, 11),
					builtinError("a", 1, 26),
					builtinError("a", 1, 35),
				},
			},
		},
	)
}
func invalidRedeclared(code string, name string, line int, column int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			redeclaredError(name, line, column),
		},
	}
}

func invalidBuiltin(code string, name string, line int, column int) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code,
		Errors: []rule_tester.InvalidTestCaseError{
			builtinError(name, line, column),
		},
	}
}

func redeclaredError(name string, line int, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "redeclared",
		Message:   "'" + name + "' is already defined.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name),
	}
}

func redeclaredBySyntaxError(name string, line int, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "redeclaredBySyntax",
		Message:   "'" + name + "' is already defined by a variable declaration.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name),
	}
}

func builtinError(name string, line int, column int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "redeclaredAsBuiltin",
		Message:   "'" + name + "' is already defined as a built-in global variable.",
		Line:      line,
		Column:    column,
		EndLine:   line,
		EndColumn: column + len(name),
	}
}
