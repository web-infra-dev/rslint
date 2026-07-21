// TestNoRedeclareExtras locks in branches and edge shapes that the upstream
// test suite doesn't exercise. Each case carries an inline comment pointing at
// the specific branch / Dimension 4 row / tsgo AST quirk it covers, so future
// refactors can't silently regress them without breaking a named lock-in.
//
// Dimension 4 walk (rows that don't apply to no-redeclare, with reasons):
//   - N/A receiver / expression wrappers ((X).y, X!.y, X as T, X satisfies T,
//     X?.y, X?.()): the rule inspects declarations, not receiver expressions.
//   - N/A access / key forms (identifier, string, numeric, private, computed,
//     element access): property keys are not declaration names for this rule.
//   - N/A autofix boundaries: the rule does not provide an autofix.
//   - N/A body-absent function scope: overload / abstract / declare members
//     have no runtime body to traverse; TypeScript declaration behavior is
//     covered by the extension rule's suite.
package no_redeclare

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoRedeclareExtras(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoRedeclareRule,
		[]rule_tester.ValidTestCase{
			// ---- Dimension 4: declaration/container forms ----
			{Code: "function outer() { var a; function inner() { var a; } }"},
			{Code: "const fn = () => { var a; }; var a;"},
			{Code: "class C { method() { var a; } } var a;"},
			{Code: "class C { get value() { var a; return a; } set value(a) {} }"},
			{Code: "class C { static { var a; } static { var a; } }"},
			{Code: "import './foo';\nvar foo;"},
			{Code: "var a; const text = '/*global a*/';"},

			// ---- Dimension 4: block and switch boundaries ----
			{Code: "{ let a; } { let a; }"},
			{Code: "switch (foo) { case 1: let a; break; case 2: { let a; } }"},
			{Code: "for (let x in obj) { let x; }"},
			{Code: "for (let x of obj) { let x; }"},
			// Annex B functions nested under `if` / labels bind to the nearest
			// block, while `var` inside that block still belongs to the program.
			{Code: "{ var a; if (ok) function a() {} }"},

			// ---- Dimension 4: graceful degradation around empty forms ----
			{Code: "{}"},
			{Code: "function f({}) {}"},
			{Code: "function f([]) {}"},
			{Code: "const {} = obj; const [] = arr;"},

			// ---- Dimension 4: TypeScript wrappers in initializers do not matter ----
			{Code: "var a = (foo as any); function f() { var a = foo satisfies string; }"},

			// ---- Real-user: browser-global names remain option-dependent ----
			{Code: "var top = 0;", Options: map[string]interface{}{"builtinGlobals": true}},

			// ---- Real-user: generated declarations may repeat names across branches ----
			{Code: "function generated() { var chunk; if (ok) { let chunk; } }"},

			// Locks in upstream Program() special-scope arm: import/export syntax makes this a module.
			{Code: "export {};\nvar Object = 0;", Options: map[string]interface{}{"builtinGlobals": true}},

			// Locks in upstream iterateDeclarations() arm 1: builtinGlobals false disables builtin reports.
			{Code: "var Object = 0;", Options: map[string]interface{}{"builtinGlobals": false}},

			// ---- Real-user: config and inline global precedence ----
			{Code: "/* globals custom */"},
			{Code: "/* globals a */ /* globals a:off */"},
			{Code: "/* globals Object:off */ var Object = 0;"},
			{Code: "var Object = 0;", Globals: map[string]bool{"Object": false}},
			{Code: "/* globals a:off */ var a = 0;", Globals: map[string]bool{"a": true}},
			{Code: "/* globals a, a */"},

			// A module-level declaration and an outer global declaration do not
			// share a scope, even when they have the same name.
			{Code: "export {};\n/* globals a */ var a = 0;"},

			// A named function expression has a separate name scope. Its body may
			// declare the same spelling without leaking it into the program scope.
			{Code: "const fn = function self() { var self; };"},

			// Catch bindings use their own scope. ESLint's scope model permits the
			// catch body to contain a same-spelled `var` when no outer var exists.
			{Code: "try {} catch (error) { var error; }"},

			// Computed property keys and declarations inside default-value closures
			// are references / child scopes, not bindings in the current pattern.
			{Code: "var key; var {[key]: value} = source;"},
			{Code: "var [item = (() => { var item; return item; })()] = items;"},

			// eslint/eslint#19141: a configured application global can deliberately
			// be redeclared when builtinGlobals is disabled.
			{
				Code:    "const chatgpt = {};",
				Options: map[string]interface{}{"builtinGlobals": false},
				Globals: map[string]bool{"chatgpt": true},
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- Annex B: nested function declarations use their true scope ----
			{
				Code: "var a;\nif (ok) function a() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 2, 18),
				},
			},
			{
				Code: "var a; if (ok) use(); else function a() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 37),
				},
			},
			{
				Code: "var a; outer: inner: function a() {}",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 31),
				},
			},
			{
				Code: "function outer() {\n  var a;\n  if (ok) function a() {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 3, 20),
				},
			},
			{
				Code: "{\n  let a;\n  outer: if (ok) function a() {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 3, 27),
				},
			},
			{
				Code: "switch (value) {\ncase 0:\n  let a;\n  if (ok) function a() {}\n}",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 4, 20),
				},
			},

			// Deeply nested `var` declarations still resolve to the program's
			// variable scope rather than any intervening statement or block.
			invalidRedeclared("var deep;\ntry { label: if (ok) { switch (value) { case 0: var deep; } } } catch (error) {}", "deep", 2, 53),

			// ---- Dimension 4: destructuring binding names all participate ----
			{
				Code: "var {a, b: a} = obj;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 12),
				},
			},
			{
				Code: "var [a, a] = arr;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 9),
				},
			},
			invalidRedeclared("var {a: {value}, b: [value]} = source;", "value", 1, 22),
			invalidRedeclared("var {[key]: value, nested: {value}} = source;", "value", 1, 29),
			invalidRedeclared("var [head, {tail: [head]}] = rows;", "head", 1, 20),
			invalidRedeclared("var {a: [first = (() => 1)()], b: {first = 2}} = source;", "first", 1, 36),
			invalidRedeclared("var {a, ...a} = source;", "a", 1, 12),
			{
				Code: "var {a} = obj;\nvar a;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 2, 5),
				},
			},
			{
				Code: "var [a] = arr;\nvar a;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 2, 5),
				},
			},
			{
				Code: "function f({a}) { var a; }",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 1, 23),
				},
			},

			// ---- Dimension 4: async / generator containers have independent scopes ----
			invalidRedeclared("async function f() {\n  var a;\n  var a;\n}", "a", 3, 7),
			invalidRedeclared("function* f() {\n  var a;\n  var a;\n}", "a", 3, 7),

			// Every tsgo function-like kind maps to an upstream function scope.
			invalidRedeclared("const fn = function () { var local; var local; };", "local", 1, 41),
			invalidRedeclared("const fn = () => { var local; var local; };", "local", 1, 35),
			invalidRedeclared("class C { method() { var local; var local; } }", "local", 1, 37),
			invalidRedeclared("class C { constructor(value) { var value; } }", "value", 1, 36),
			invalidRedeclared("class C { get value() { var local; var local; return local; } }", "local", 1, 40),
			invalidRedeclared("class C { set value(input) { var input; } }", "input", 1, 34),

			// A static block is a var-scope boundary, but arbitrary statements and
			// nested blocks inside it do not hide a hoisted duplicate.
			invalidRedeclared("class C { static { var state; if (ready) { try { var state; } finally {} } } }", "state", 1, 54),

			// ---- Dimension 4: switch shares one block scope for lexical declarations ----
			invalidRedeclared("switch (foo) {\ncase 1: let a; break;\ncase 2: let a;\n}", "a", 3, 13),

			// ---- Dimension 4: import binding forms reuse the shared import helper ----
			invalidRedeclared("import a from './foo';\nvar a;", "a", 2, 5),
			invalidRedeclared("import { a as b } from './foo';\nconst b = 1;", "b", 2, 7),
			invalidRedeclared("import * as ns from './foo';\nlet ns;", "ns", 2, 5),
			{
				Code: "import a, { b } from './foo';\nvar a;\nvar b;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("a", 2, 5),
					redeclaredError("b", 3, 5),
				},
			},
			invalidRedeclared("import a = require('./foo');\nvar a;", "a", 2, 5),

			// ---- Dimension 4: for-of destructuring initializer is its own scope ----
			invalidRedeclared("for (let {a, b: a} of xs) {}", "a", 1, 17),
			invalidRedeclared("for (let [a, a] of xs) {}", "a", 1, 14),
			// The sibling ForIn listener has the same lexical-initializer contract.
			invalidRedeclared("for (let {left, right: left} in records) {}", "left", 1, 24),
			// ForAwaitOf is represented by tsgo as a ForOfStatement as well.
			invalidRedeclared("async function f() { for await (const [item, item] of stream) {} }", "item", 1, 46),

			// Explicit resource-management declarations are block scoped and use
			// the same VariableDeclarationList path as let / const.
			invalidRedeclared("using resource = acquire(); using resource = acquire();", "resource", 1, 35),
			invalidRedeclared("async function f() { await using resource = acquire(); await using resource = acquire(); }", "resource", 1, 68),

			// ---- Real-user: built-in globals report each user declaration ----
			{
				Code: "var Object;\nvar Object;",
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("Object", 1, 5),
					builtinError("Object", 2, 5),
				},
			},

			// ---- Real-user: duplicate generated function declarations ----
			invalidRedeclared("function init() {}\nfunction init() {}", "init", 2, 10),

			// Generated declaration lists often interleave several names. Reports
			// remain source ordered even though declarations are grouped by symbol.
			{
				Code: "var alpha, beta; var beta, alpha;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("beta", 1, 22),
					redeclaredError("alpha", 1, 28),
				},
			},

			// Locks in upstream iterateDeclarations() arm 2: syntax declarations after the first report as plain redeclarations.
			invalidRedeclared("let a;\nlet a;", "a", 2, 5),

			// Locks in upstream findVariablesInScope() detail arm: builtin declaration is first, so user syntax reports builtin-specific message.
			invalidBuiltin("var Array = 0;", "Array", 1, 5),
			// An omitted property in an explicitly supplied empty option object
			// retains upstream's builtinGlobals: true default.
			{
				Code:    "var Object = 0;",
				Options: map[string]interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("Object", 1, 5),
				},
			},
			// ESLint core treats parser-provided type declarations as variables;
			// only the TypeScript extension excludes pure type-space declarations.
			invalidBuiltin("interface Object {}", "Object", 1, 11),
			invalidBuiltin("type Array = unknown;", "Array", 1, 6),

			// Locks in upstream checkForBlock() arm: non-function blocks are checked as their own lexical scope.
			invalidRedeclared("{\n  const a = 1;\n  const a = 2;\n}", "a", 3, 9),

			// Locks in upstream ForStatement listener.
			invalidRedeclared("for (let i = 0, i = 1; ; ) {}", "i", 1, 17),

			// ---- Real-user: config and inline global declaration ordering ----
			invalidRedeclared("/* globals a:off */ /* globals a */", "a", 1, 32),
			{
				Code:    "/* globals Object */ var Object = 0;",
				Globals: map[string]bool{"Object": false},
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredBySyntaxError("Object", 1, 12),
				},
			},
			{
				Code:    "/* globals a */ var a = 0;",
				Globals: map[string]bool{"a": true},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("a", 1, 12),
					builtinError("a", 1, 21),
				},
			},
			{
				Code: "/* globals a:off */ /* globals a */ var a = 0;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredBySyntaxError("a", 1, 12),
					redeclaredBySyntaxError("a", 1, 32),
				},
			},
			{
				Code: "/* globals a, a */ var a;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredBySyntaxError("a", 1, 12),
				},
			},

			// Global comments still share the outer scope in a module.
			invalidBuiltin("export {};\n/* globals Array */", "Array", 2, 12),
			invalidRedeclared("export {};\n/* globals a */ /* globals a */", "a", 2, 28),
			{
				Code:    "export {};\n/* globals app */",
				Globals: map[string]bool{"app": true},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("app", 2, 12),
				},
			},

			// eslint/eslint#19141: config-declared application globals participate
			// in the implicit-global branch when builtinGlobals uses its default.
			{
				Code:    "const chatgpt = {};",
				Globals: map[string]bool{"chatgpt": true},
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("chatgpt", 1, 7),
				},
			},

			// eslint/eslint#12334: directive diagnostics must cover exactly the
			// name, including non-zero end locations across CRLF line endings.
			{
				Code: "/*globals foo,\r\n    Array */",
				Errors: []rule_tester.InvalidTestCaseError{
					builtinError("Array", 2, 5),
				},
			},

			// Diagnostic coordinates are UTF-16, matching ESLint for BMP and
			// astral identifiers as well as astral text before a declaration.
			{
				Code: "var 变量; var 变量;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redeclared",
						Message:   "'变量' is already defined.",
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			{
				Code: "const emoji = \"😀\"; var name; var name;",
				Errors: []rule_tester.InvalidTestCaseError{
					redeclaredError("name", 1, 35),
				},
			},
			{
				Code: "var 𐊧; var 𐊧;",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "redeclared",
						Message:   "'𐊧' is already defined.",
						Line:      1,
						Column:    13,
						EndLine:   1,
						EndColumn: 15,
					},
				},
			},
			invalidRedeclared("var item;\r\nvar item;", "item", 2, 5),

			// Export and type-only import wrappers must not hide their declarations.
			invalidRedeclared("export default function initialize() {} function initialize() {}", "initialize", 1, 50),
			invalidRedeclared("import type {Model} from './types'; type Model = {};", "Model", 1, 42),
		},
	)
}

func TestNoRedeclareParseOptionsMatrix(t *testing.T) {
	tests := []struct {
		name string
		raw  []any
		want options
	}{
		{name: "omitted", want: coreDefaults()},
		{name: "empty options array", raw: []any{}, want: coreDefaults()},
		{name: "empty object", raw: []any{map[string]any{}}, want: coreDefaults()},
		{
			name: "builtin globals enabled",
			raw:  []any{map[string]any{"builtinGlobals": true}},
			want: coreDefaults(),
		},
		{
			name: "builtin globals disabled",
			raw:  []any{map[string]any{"builtinGlobals": false}},
			want: options{builtinGlobals: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := parseOptionsWith(test.raw, coreDefaults(), false)
			if got != test.want {
				t.Fatalf("parseOptionsWith() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestNoRedeclareSchemaMatrix(t *testing.T) {
	if NoRedeclareRule.Schema == nil {
		t.Fatal("no-redeclare must declare its upstream option schema")
	}

	valid := []struct {
		name    string
		options []any
	}{
		{name: "omitted"},
		{name: "empty options array", options: []any{}},
		{name: "empty object", options: []any{map[string]any{}}},
		{name: "builtin globals true", options: []any{map[string]any{"builtinGlobals": true}}},
		{name: "builtin globals false", options: []any{map[string]any{"builtinGlobals": false}}},
	}
	for _, test := range valid {
		t.Run("valid/"+test.name, func(t *testing.T) {
			if err := NoRedeclareRule.Schema.Validate(test.options); err != nil {
				t.Fatalf("unexpected schema error: %v", err)
			}
		})
	}

	invalid := []struct {
		name    string
		options []any
	}{
		{name: "non-object option", options: []any{true}},
		{name: "null option", options: []any{nil}},
		{name: "non-boolean value", options: []any{map[string]any{"builtinGlobals": "true"}}},
		{name: "unknown property", options: []any{map[string]any{"ignoreDeclarationMerge": true}}},
		{name: "too many options", options: []any{map[string]any{}, map[string]any{}}},
	}
	for _, test := range invalid {
		t.Run("invalid/"+test.name, func(t *testing.T) {
			if err := NoRedeclareRule.Schema.Validate(test.options); err == nil {
				t.Fatal("expected schema validation to reject options")
			}
		})
	}
}
