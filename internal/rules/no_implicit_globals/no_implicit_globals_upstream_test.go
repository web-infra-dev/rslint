package no_implicit_globals

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// TestNoImplicitGlobalsUpstream migrates the full valid/invalid suite from
// upstream tests/lib/rules/no-implicit-globals.js 1:1. Position assertions
// cover line/column only where upstream itself asserts one; rslint's own
// position lock-ins (upstream leaves most declaration-report positions
// unchecked) live in no_implicit_globals_extras_test.go. rslint-specific
// augmentation lives in that same extras file.
func TestNoImplicitGlobalsUpstream(t *testing.T) {
	const globalNonLexicalBinding = "globalNonLexicalBinding"
	const globalLexicalBinding = "globalLexicalBinding"
	const redeclarationOfReadonlyGlobal = "redeclarationOfReadonlyGlobal"
	const assignmentToReadonlyGlobal = "assignmentToReadonlyGlobal"
	const globalVariableLeak = "globalVariableLeak"

	const varMessage = "Unexpected 'var' declaration in the global scope, wrap in an IIFE for a local variable, assign as global property for a global variable."
	const functionMessage = "Unexpected function declaration in the global scope, wrap in an IIFE for a local variable, assign as global property for a global variable."
	const constMessage = "Unexpected 'const' declaration in the global scope, wrap in a block or in an IIFE."
	const letMessage = "Unexpected 'let' declaration in the global scope, wrap in a block or in an IIFE."
	const classMessage = "Unexpected class declaration in the global scope, wrap in a block or in an IIFE."
	const readonlyRedeclarationMessage = "Unexpected redeclaration of read-only global variable."
	const readonlyAssignmentMessage = "Unexpected assignment to read-only global variable."
	const leakMessage = "Global variable leak, declare the variable if it is intended to be local."

	lexical := []any{map[string]any{"lexicalBindings": true}}

	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImplicitGlobalsRule,
		[]rule_tester.ValidTestCase{
			// ---- General ----

			// Recommended way to create a global variable in the browser
			{Code: `window.foo = 1;`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function foo() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function bar() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function*() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function *foo() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = async function() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = async function foo() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = async function*() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = async function *foo() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = class {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = class foo {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = class bar {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `self.foo = 1;`, Globals: map[string]any{"self": "readonly"}},
			{Code: `self.foo = function() {};`, Globals: map[string]any{"self": "readonly"}},

			// Another way to create a global variable. Not the best practice, but that isn't the responsibility of this rule.
			{Code: `this.foo = 1;`},
			{Code: `this.foo = function() {};`},
			{Code: `this.foo = function bar() {};`},

			// Test that the rule doesn't report global comments
			{Code: `/*global foo:readonly*/`},
			{Code: `/*global foo:writable*/`},
			{Code: `/*global Array:readonly*/`},
			{Code: `/*global Array:writable*/`},
			{Code: `/*global foo:readonly*/`, Globals: map[string]any{"foo": "readonly"}},
			{Code: `/*global foo:writable*/`, Globals: map[string]any{"foo": "readonly"}},
			{Code: `/*global foo:readonly*/`, Globals: map[string]any{"foo": "writable"}},
			{Code: `/*global foo:writable*/`, Globals: map[string]any{"foo": "writable"}},

			// ---- `var` and function declarations ----

			// Doesn't report function expressions
			{Code: `typeof function() {}`},
			{Code: `typeof function foo() {}`},
			{Code: `(function() {}) + (function foo() {})`},
			{Code: `typeof function *foo() {}`},
			{Code: `typeof async function foo() {}`},
			{Code: `typeof async function *foo() {}`},

			// Recommended way to create local variables
			{Code: `(function() { var foo = 1; })();`},
			{Code: `(function() { function foo() {} })();`},
			{Code: `(function() { function *foo() {} })();`},
			{Code: `(function() { async function foo() {} })();`},
			{Code: `(function() { async function *foo() {} })();`},
			{Code: `window.foo = (function() { var bar; function foo () {}; return function bar() {} })();`, Globals: map[string]any{"window": "readonly"}},

			// Different scoping
			{Code: `var foo = 1;`, FileName: "mjs/module-scoping-var.mjs", TSConfig: "tsconfig.allow-js.json"},
			{Code: `function foo() {}`, FileName: "mjs/module-scoping-function.mjs", TSConfig: "tsconfig.allow-js.json"},
			{Code: `function *foo() {}`, FileName: "mjs/module-scoping-generator.mjs", TSConfig: "tsconfig.allow-js.json"},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn (framework gap)
			{Code: `var foo = 1;`, Skip: true},
			{Code: `function foo() {}`, Skip: true},
			{Code: `var foo = 1;`, FileName: "cjs/scoping-var.cjs", TSConfig: "tsconfig.allow-js.json"},
			{Code: `function foo() {}`, FileName: "cjs/scoping-function.cjs", TSConfig: "tsconfig.allow-js.json"},

			// ---- `const`, `let` and class declarations ----

			// Test default option
			{Code: `const foo = 1; let bar; class Baz {}`},
			{Code: `const foo = 1; let bar; class Baz {}`, Options: []any{map[string]any{"lexicalBindings": false}}},

			// If the option is not set to true, even the redeclarations of read-only global variables are allowed.
			{Code: `const Array = 1; let Object; class Math {}`},
			{Code: `/*global foo:readonly, bar:readonly, Baz:readonly*/ const foo = 1; let bar; class Baz {}`},

			// Doesn't report class expressions
			{Code: `typeof class {}`, Options: lexical},
			{Code: `typeof class foo {}`, Options: lexical},

			// Recommended ways to create local variables
			{Code: `{ const foo = 1; let bar; class Baz {} }`, Options: lexical},
			{Code: `(function() { const foo = 1; let bar; class Baz {} })();`, Options: lexical},
			{Code: `window.foo = (function() { const bar = 1; let baz; class Quux {} return function () {} })();`, Options: lexical, Globals: map[string]any{"window": "readonly"}},

			// different scoping
			{Code: `const foo = 1; let bar; class Baz {}`, FileName: "mjs/module-lexical.mjs", TSConfig: "tsconfig.allow-js.json"},
			{Code: `const foo = 1; let bar; class Baz {}`, FileName: "cjs/commonjs-lexical.cjs", TSConfig: "tsconfig.allow-js.json"},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn (framework gap)
			{Code: `const foo = 1; let bar; class Baz {}`, Skip: true},

			// Regression tests
			{Code: `const foo = 1;`},
			{Code: `let foo = 1;`},
			{Code: `let foo = function() {};`},
			{Code: `const foo = function() {};`},
			{Code: `class Foo {}`},
			{Code: `(function() { let foo = 1; })();`},
			{Code: `(function() { const foo = 1; })();`},
			{Code: `let foo = 1;`, FileName: "mjs/module-let.mjs", TSConfig: "tsconfig.allow-js.json"},
			{Code: `const foo = 1;`, FileName: "mjs/module-const.mjs", TSConfig: "tsconfig.allow-js.json"},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn (framework gap)
			{Code: `let foo = 1;`, Skip: true},
			{Code: `const foo = 1;`, Skip: true},

			// ---- leaks ----

			// This rule doesn't report all undeclared variables, just leaks (assignments to an undeclared variable)
			{Code: `foo`},
			{Code: `foo + bar`},
			{Code: `foo(bar)`},
			{Code: `foo++`},
			{Code: `--foo`},
			{Code: `foo += 1`},
			{Code: `foo ||= 1`},
			{Code: `/* global foo: writable*/ foo = bar`},

			// Leaks are not possible in strict mode (explicit or implicit). Therefore, rule doesn't report assignments in strict mode.
			{Code: `'use strict';foo = 1;`},
			{Code: `(function() {'use strict'; foo = 1; })();`},
			{Code: `{ class Foo { constructor() { bar = 1; } baz() { bar = 1; } } }`},
			{Code: `foo = 1;`, FileName: "mjs/module-leak.mjs", TSConfig: "tsconfig.allow-js.json"},

			// This rule doesn't check the existence of the objects in property assignments. These are reference errors, not leaks. Note that the env is not set.
			{Code: `Foo.bar = 1;`},
			{Code: `Utils.foo = 1;`},
			{Code: `Utils.foo = function() {};`},
			{Code: `window.foo = 1;`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function foo() {};`, Globals: map[string]any{"window": "readonly"}},
			{Code: `self.foo = 1;`, Globals: map[string]any{"self": "readonly"}},
			{Code: `self.foo = function() {};`, Globals: map[string]any{"self": "readonly"}},

			// These are also just reference errors, thus not reported as leaks
			{Code: `++foo`},
			{Code: `foo--`},

			// Not a leak
			{Code: `foo = 1;`, Globals: map[string]any{"foo": "writable"}},
			{Code: `window.foo = function bar() { bar = 1; };`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function bar(baz) { baz = 1; };`, Globals: map[string]any{"window": "readonly"}},
			{Code: `window.foo = function bar() { var baz; function quux() { quux = 1; } };`, Globals: map[string]any{"window": "readonly"}},

			// ---- globals ----

			// Redeclarations of writable global variables are allowed
			{Code: `/*global foo:writable*/ var foo = 1;`},
			{Code: `function foo() {}`, Globals: map[string]any{"foo": "writable"}},
			{Code: `/*global foo:writable*/ function *foo() {}`},
			{Code: `/*global foo:writable*/ const foo = 1;`, Options: lexical},
			{Code: `/*global foo:writable*/ let foo;`, Options: lexical},
			{Code: `/*global Foo:writable*/ class Foo {}`, Options: lexical},

			// Assignments to writable global variables are allowed
			{Code: `/*global foo:writable*/ foo = 1;`},
			{Code: `foo = 1`, Globals: map[string]any{"foo": "writable"}},

			// This rule doesn't disallow assignments to properties of readonly globals
			{Code: `Array.from = 1;`},
			{Code: `Object['assign'] = 1;`},
			{Code: `/*global foo:readonly*/ foo.bar = 1;`},

			// This rule doesn't disallow updates of readonly globals
			{Code: `/*global foo:readonly*/ foo++;`},
			{Code: `/*global foo:readonly*/ --foo;`},
			{Code: `/*global foo:readonly*/ foo += 1;`},
			{Code: `/*global foo:readonly*/ foo ||= 1;`},

			// ---- exported ----

			// `var` and functions
			{Code: `/* exported foo */ var foo = 'foo';`},
			{Code: `/* exported foo */ function foo() {}`},
			{Code: `/* exported foo */ function *foo() {}`},
			{Code: `/* exported foo */ async function foo() {}`},
			{Code: `/* exported foo */ async function *foo() {}`},
			{Code: `/* exported foo */ var foo = function() {};`},
			{Code: `/* exported foo */ var foo = function foo() {};`},
			{Code: `/* exported foo */ var foo = function*() {};`},
			{Code: `/* exported foo */ var foo = function *foo() {};`},
			{Code: `/* exported foo, bar */ var foo = 1, bar = 2;`},

			// `const`, `let` and `class`
			{Code: `/* exported a */ const a = 1;`, Options: lexical},
			{Code: `/* exported a */ let a;`, Options: lexical},
			{Code: `/* exported a */ let a = 1;`, Options: lexical},
			{Code: `/* exported A */ class A {}`, Options: lexical},
			{Code: `/* exported a, b */ const a = 1; const b = 2;`, Options: lexical},
			{Code: `/* exported a, b */ const a = 1, b = 2;`, Options: lexical},
			{Code: `/* exported a, b */ let a, b = 1;`, Options: lexical},
			{Code: `/* exported a, b, C */ const a = 1; let b; class C {}`, Options: lexical},
			{Code: `/* exported a, b, c */ const [a, b, ...c] = [];`, Options: lexical},
			{Code: `/* exported a, b, c */ let { a, foo: b, bar: { c } } = {};`, Options: lexical},
		},
		[]rule_tester.InvalidTestCase{
			// ---- `var` and function declarations ----

			{
				Code:   `var foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `function *foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `async function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `async function *foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `var foo = function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `var foo = function foo() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `var foo = function*() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `var foo = function *foo() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code: `var foo = 1, bar = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalNonLexicalBinding, Message: varMessage},
					{MessageId: globalNonLexicalBinding, Message: varMessage},
				},
			},

			// ---- `const`, `let` and class declarations ----

			// Basic tests
			{
				Code:    `const a = 1;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: constMessage}},
			},
			{
				Code:    `let a;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: letMessage}},
			},
			{
				Code:    `let a = 1;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: letMessage}},
			},
			{
				Code:    `class A {}`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: classMessage, Line: 1, Column: 1}},
			},

			// Multiple and mixed tests
			{
				Code:    `const a = 1; const b = 2;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: globalLexicalBinding, Message: constMessage},
				},
			},
			{
				Code:    `const a = 1, b = 2;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: globalLexicalBinding, Message: constMessage},
				},
			},
			{
				Code:    `let a, b = 1;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: globalLexicalBinding, Message: letMessage},
				},
			},
			{
				Code:    `const a = 1; let b; class C {}`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: globalLexicalBinding, Message: classMessage},
				},
			},
			{
				Code:    `const [a, b, ...c] = [];`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: globalLexicalBinding, Message: constMessage},
				},
			},
			{
				Code:    `let { a, foo: b, bar: { c } } = {};`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: globalLexicalBinding, Message: letMessage},
				},
			},

			// ---- leaks ----

			// Basic tests
			{
				Code:   `foo = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `foo = function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `foo = function*() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:    `window.foo = function() { bar = 1; }`,
				Globals: map[string]any{"window": "readonly"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 27}},
			},
			{
				Code:   `(function() {}(foo = 1));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `for (foo in {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `for (foo of []);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},

			// Not implicit strict
			{
				Code:   `window.foo = { bar() { foo = 1 } }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 24}},
			},
			{
				Code:     `foo = 1`,
				FileName: "cjs/leak-commonjs.cjs",
				TSConfig: "tsconfig.allow-js.json",
				Errors:   []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			// SKIP: rslint does not support ESLint's parserOptions.ecmaFeatures.globalReturn (framework gap)
			{Code: `foo = 1;`, Skip: true},

			// Multiple and mixed
			{
				Code: `foo = 1, bar = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 1},
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 10},
				},
			},
			{
				Code: `foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 1},
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 7},
				},
			},
			{
				Code:   `/*global foo:writable*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 31}},
			},
			{
				Code:   `/*global bar:writable*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 25}},
			},
			{
				Code: `foo = 1; var bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalVariableLeak, Message: leakMessage},
					{MessageId: globalNonLexicalBinding, Message: varMessage},
				},
			},
			{
				Code: `var foo = bar = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalNonLexicalBinding, Message: varMessage},
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 11},
				},
			},
			{
				Code:   `/*global foo:writable*/ var foo = bar = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 35}},
			},
			{
				Code:   `/*global bar:writable*/ var foo = bar = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code: `[foo, bar] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 1},
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 1},
				},
			},
			{
				Code:   `/*global foo:writable*/ [foo, bar] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 25}},
			},
			{
				Code:   `/*global bar:writable*/ [foo, bar] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 25}},
			},

			// ---- globals ----

			// Basic assignment tests
			{
				Code:   `Array = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 1}},
			},
			{
				Code:    `window = 1;`,
				Globals: map[string]any{"window": "readonly"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `/*global foo:readonly*/ foo = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 25}},
			},
			{
				Code:    `foo = 1;`,
				Globals: map[string]any{"foo": "readonly"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 1}},
			},
			{
				Code:   `/*global foo:readonly*/ for (foo in {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 25}},
			},
			{
				Code:   `/*global foo:readonly*/ for (foo of []);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 25}},
			},

			// Basic redeclaration tests
			{
				Code:   `var Array = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `var Array = 1; Array = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ var foo`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ var foo = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ var foo; foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ for (var foo in obj);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ for (var foo in obj); foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ for (var foo of arr);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ for (var foo of arr); foo = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly*/ function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ const foo = 1`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ const foo = 1; foo = 2;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ let foo`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ let foo = 1`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ let foo; foo = 1;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global Foo:readonly*/ class Foo {}`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},

			// Multiple and mixed assignments
			{
				Code: `/*global foo:readonly, bar: readonly*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 40},
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 46},
				},
			},
			{
				Code:   `/*global foo:writable, bar: readonly*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 46}},
			},
			{
				Code:   `/*global foo:readonly, bar: writable*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 40}},
			},
			{
				Code: `/*global foo: readonly*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 26},
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 32},
				},
			},
			{
				Code: `/*global bar: readonly*/ foo = bar = 1`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 26},
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 32},
				},
			},
			{
				Code:   `/*global foo*/ [foo] = arr`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 16}},
			},
			{
				Code: `/*global foo, bar: readonly*/ [foo, bar] = arr`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 31},
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 31},
				},
			},
			{
				Code:   `/*global foo: readonly*/ ({ foo } = obj)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 27}},
			},
			{
				Code:   `/*global foo: readonly*/ ({ 'a': foo } = obj)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 27}},
			},
			{
				Code:   `/*global foo: readonly*/ ({ 'a': { 'b': [foo] } } = obj)`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 27}},
			},
			{
				Code: `/*global foo, bar: readonly*/ ({ foo, 'a': bar } = obj)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 32},
					{MessageId: assignmentToReadonlyGlobal, Message: readonlyAssignmentMessage, Line: 1, Column: 32},
				},
			},

			// Multiple and mixed redeclarations
			{
				Code: `/*global foo:readonly, bar: readonly*/ var foo, bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
				},
			},
			{
				Code:   `/*global foo:writable, bar: readonly*/ var foo, bar;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:   `/*global foo:readonly, bar: writable*/ var foo, bar;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code: `/*global foo:readonly*/ var foo, bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
					{MessageId: globalNonLexicalBinding, Message: varMessage},
				},
			},
			{
				Code: `/*global bar: readonly*/ var foo, bar;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalNonLexicalBinding, Message: varMessage},
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
				},
			},
			{
				Code:    `/*global foo:readonly, bar: readonly*/ const foo = 1, bar = 2;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
				},
			},
			{
				Code:    `/*global foo:writable, bar: readonly*/ const foo = 1, bar = 2;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly, bar: writable*/ const foo = 1, bar = 2;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ const foo = 1, bar = 2;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
					{MessageId: globalLexicalBinding, Message: constMessage},
				},
			},
			{
				Code:    `/*global bar: readonly*/ const foo = 1, bar = 2;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
				},
			},
			{
				Code:    `/*global foo:readonly, bar: readonly*/ let foo, bar;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
				},
			},
			{
				Code:    `/*global foo:writable, bar: readonly*/ let foo, bar;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly, bar: writable*/ let foo, bar;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage}},
			},
			{
				Code:    `/*global foo:readonly*/ let foo, bar;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
					{MessageId: globalLexicalBinding, Message: letMessage},
				},
			},
			{
				Code:    `/*global bar: readonly*/ let foo, bar;`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: redeclarationOfReadonlyGlobal, Message: readonlyRedeclarationMessage},
				},
			},

			// ---- exported ----

			// `var` and `function`
			{
				Code:   `/* exported bar */ var foo = 'text';`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `/* exported bar */ function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 20}},
			},
			{
				Code:   `/* exported bar */ function *foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 20}},
			},
			{
				Code:   `/* exported bar */ async function foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 20}},
			},
			{
				Code:   `/* exported bar */ async function *foo() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: functionMessage, Line: 1, Column: 20}},
			},
			{
				Code:   `/* exported bar */ var foo = function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `/* exported bar */ var foo = function foo() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `/* exported bar */ var foo = function*() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `/* exported bar */ var foo = function *foo() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},
			{
				Code:   `/* exported bar */ var foo = 1, bar = 2;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalNonLexicalBinding, Message: varMessage}},
			},

			// `let`, `const` and `class`
			{
				Code:    `/* exported b */ const a = 1;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: constMessage}},
			},
			{
				Code:    `/* exported b */ let a;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: letMessage}},
			},
			{
				Code:    `/* exported b */ let a = 1;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: letMessage}},
			},
			{
				Code:    `/* exported B */ class A {}`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: classMessage, Line: 1, Column: 18}},
			},
			{
				Code:    `/* exported a */ const a = 1; const b = 2;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: constMessage}},
			},
			{
				Code:    `/* exported a */ const a = 1, b = 2;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: constMessage}},
			},
			{
				Code:    `/* exported a */ let a, b = 1;`,
				Options: lexical,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalLexicalBinding, Message: letMessage}},
			},
			{
				Code:    `/* exported a */ const a = 1; let b; class C {}`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: globalLexicalBinding, Message: classMessage},
				},
			},
			{
				Code:    `/* exported a */ const [a, b, ...c] = [];`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: constMessage},
					{MessageId: globalLexicalBinding, Message: constMessage},
				},
			},
			{
				Code:    `/* exported a */ let { a, foo: b, bar: { c } } = {};`,
				Options: lexical,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: globalLexicalBinding, Message: letMessage},
					{MessageId: globalLexicalBinding, Message: letMessage},
				},
			},

			// Global variable leaks
			{
				Code:   `/* exported foo */ foo = 1`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `/* exported foo */ foo = function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `/* exported foo */ foo = function*() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:    `/* exported foo */ window.foo = function() { bar = 1; }`,
				Globals: map[string]any{"window": "readonly"},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage, Line: 1, Column: 46}},
			},
			{
				Code:   `/* exported foo */ (function() {}(foo = 1));`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `/* exported foo */ for (foo in {});`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
			{
				Code:   `/* exported foo */ for (foo of []);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: globalVariableLeak, Message: leakMessage}},
			},
		},
	)
}
