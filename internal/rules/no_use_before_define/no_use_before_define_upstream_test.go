// TestNoUseBeforeDefineUpstream migrates the JavaScript half of the upstream
// valid/invalid suite from tests/lib/rules/no-use-before-define.js 1:1.
// Position assertions cover line/column for every invalid case. The
// TypeScript-parser half lives in no_use_before_define_upstream_typescript_test.go,
// and rslint-specific lock-in cases live in no_use_before_define_extras_test.go.
package no_use_before_define

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUseBeforeDefineUpstream(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoUseBeforeDefineRule,
		[]rule_tester.ValidTestCase{
			{Code: `unresolved`},
			{Code: `Array`},
			{Code: `function foo () { arguments; }`},
			{Code: `var a=10; alert(a);`},
			{Code: `function b(a) { alert(a); }`},
			{Code: `Object.hasOwnProperty.call(a);`},
			{Code: `function a() { alert(arguments);}`},
			{Code: `a(); function a() { alert(arguments); }`, Options: "nofunc"},
			{Code: `(() => { var a = 42; alert(a); })();`},
			{Code: `a(); try { throw new Error() } catch (a) {}`},
			{Code: `class A {} new A();`},
			{Code: `var a = 0, b = a;`},
			{Code: `var {a = 0, b = a} = {};`},
			{Code: `var [a = 0, b = a] = {};`},
			{Code: `function foo() { foo(); }`},
			{Code: `var foo = function() { foo(); };`},
			{Code: `var a; for (a in a) {}`},
			{Code: `var a; for (a of a) {}`},
			{Code: `let a; class C { static { a; } }`},
			{Code: `class C { static { let a; a; } }`},

			// ---- Block-level bindings ----
			{Code: `"use strict"; a(); { function a() {} }`},
			{Code: `"use strict"; { a(); function a() {} }`, Options: "nofunc"},
			{Code: `switch (foo) { case 1:  { a(); } default: { let a; }}`},
			{Code: `a(); { let a = function () {}; }`},

			// ---- object style options ----
			{Code: `a(); function a() { alert(arguments); }`, Options: map[string]any{"functions": false}},
			{Code: `"use strict"; { a(); function a() {} }`, Options: map[string]any{"functions": false}},
			{Code: `function foo() { new A(); } class A {};`, Options: map[string]any{"classes": false}},

			// ---- "variables" option ----
			{Code: `function foo() { bar; } var bar;`, Options: map[string]any{"variables": false}},
			{Code: `var foo = () => bar; var bar;`, Options: map[string]any{"variables": false}},
			{Code: `class C { static { () => foo; let foo; } }`, Options: map[string]any{"variables": false}},

			// ---- Tests related to class definition evaluation. These are not TDZ errors. ----
			{Code: `class C extends (class { method() { C; } }) {}`},
			{Code: `(class extends (class { method() { C; } }) {});`},
			{Code: `const C = (class extends (class { method() { C; } }) {});`},
			{Code: `class C extends (class { field = C; }) {}`},
			{Code: `(class extends (class { field = C; }) {});`},
			{Code: `const C = (class extends (class { field = C; }) {});`},
			{Code: `class C { [() => C](){} }`},
			{Code: `(class C { [() => C](){} });`},
			{Code: `const C = class { [() => C](){} };`},
			{Code: `class C { static [() => C](){} }`},
			{Code: `(class C { static [() => C](){} });`},
			{Code: `const C = class { static [() => C](){} };`},
			{Code: `class C { [() => C]; }`},
			{Code: `(class C { [() => C]; });`},
			{Code: `const C = class { [() => C]; };`},
			{Code: `class C { static [() => C]; }`},
			{Code: `(class C { static [() => C]; });`},
			{Code: `const C = class { static [() => C]; };`},
			{Code: `class C { method() { C; } }`},
			{Code: `(class C { method() { C; } });`},
			{Code: `const C = class { method() { C; } };`},
			{Code: `class C { static method() { C; } }`},
			{Code: `(class C { static method() { C; } });`},
			{Code: `const C = class { static method() { C; } };`},
			{Code: `class C { field = C; }`},
			{Code: `(class C { field = C; });`},
			{Code: `const C = class { field = C; };`},
			{Code: `class C { static field = C; }`},
			// `const C = class { static field = C; };` is a TDZ error — see the invalid suite.
			{Code: `(class C { static field = C; });`},
			{Code: `class C { static field = class { static field = C; }; }`},
			{Code: `(class C { static field = class { static field = C; }; });`},
			{Code: `class C { field = () => C; }`},
			{Code: `(class C { field = () => C; });`},
			{Code: `const C = class { field = () => C; };`},
			{Code: `class C { static field = () => C; }`},
			{Code: `(class C { static field = () => C; });`},
			{Code: `const C = class { static field = () => C; };`},
			{Code: `class C { field = class extends C {}; }`},
			{Code: `(class C { field = class extends C {}; });`},
			{Code: `const C = class { field = class extends C {}; }`},
			{Code: `class C { static field = class extends C {}; }`},
			// `const C = class { static field = class extends C {}; };` is a TDZ error.
			{Code: `(class C { static field = class extends C {}; });`},
			{Code: `class C { static field = class { [C]; }; }`},
			// `const C = class { static field = class { [C]; } };` is a TDZ error.
			{Code: `(class C { static field = class { [C]; }; });`},
			{Code: `const C = class { static field = class { field = C; }; };`},
			{Code: `class C { method() { a; } } let a;`, Options: map[string]any{"variables": false}},
			{Code: `class C { static method() { a; } } let a;`, Options: map[string]any{"variables": false}},
			// `class C { static field = a; } let a;` is a TDZ error.
			{Code: `class C { field = a; } let a;`, Options: map[string]any{"variables": false}},
			// `class C { static field = D; } class D {}` is a TDZ error.
			{Code: `class C { field = D; } class D {}`, Options: map[string]any{"classes": false}},
			// `class C { static field = class extends D {}; } class D {}` is a TDZ error.
			{Code: `class C { field = class extends D {}; } class D {}`, Options: map[string]any{"classes": false}},
			{Code: `class C { field = () => a; } let a;`, Options: map[string]any{"variables": false}},
			{Code: `class C { static field = () => a; } let a;`, Options: map[string]any{"variables": false}},
			{Code: `class C { field = () => D; } class D {}`, Options: map[string]any{"classes": false}},
			{Code: `class C { static field = () => D; } class D {}`, Options: map[string]any{"classes": false}},
			{Code: `class C { static field = class { field = a; }; } let a;`, Options: map[string]any{"variables": false}},
			// `const C = class { static { C; } }` is a TDZ error.
			{Code: `class C { static { C; } }`},
			{Code: `class C { static { C; } static {} static { C; } }`},
			{Code: `(class C { static { C; } })`},
			{Code: `class C { static { class D extends C {} } }`},
			{Code: `class C { static { (class { static { C } }) } }`},
			{Code: `class C { static { () => C; } }`},
			{Code: `(class C { static { () => C; } })`},
			{Code: `const C = class { static { () => C; } }`},
			{Code: `class C { static { () => D; } } class D {}`, Options: map[string]any{"classes": false}},
			{Code: `class C { static { () => a; } } let a;`, Options: map[string]any{"variables": false}},
			{Code: `const C = class C { static { C.x; } }`},

			// ---- "allowNamedExports" option ----
			{Code: `export { a }; const a = 1;`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `export { a as b }; const a = 1;`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `export { a, b }; let a, b;`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `export { a }; var a;`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `export { f }; function f() {}`, Options: map[string]any{"allowNamedExports": true}},
			{Code: `export { C }; class C {}`, Options: map[string]any{"allowNamedExports": true}},

			// ---- JSX ----
			{Code: `const App = () => <div/>; <App />;`, Tsx: true},
			{Code: `let Foo, Bar; <Foo><Bar /></Foo>;`, Tsx: true},
			{Code: `function App() { return <div/> } <App />;`, Tsx: true},
			{Code: `<App />; function App() { return <div/> }`, Tsx: true, Options: map[string]any{"functions": false}},
		},
		[]rule_tester.InvalidTestCase{
			// Upstream repeats this case three times, for `sourceType: module`,
			// `ecmaVersion: 6`, and the tester default. rslint models none of
			// those as scope-affecting, so one case covers all three.
			{
				Code:   `a++; var a=19;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Message: "'a' was used before it was defined.", Line: 1, Column: 1, EndLine: 1, EndColumn: 2}},
			},
			{
				Code:   `a(); var a=function() {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `alert(a[1]); var a=[1,3];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 7}},
			},
			{
				Code: `a(); function a() { alert(b); var b=10; a(); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 1},
					{MessageId: "usedBeforeDefined", Line: 1, Column: 27},
				},
			},
			{
				Code:    `a(); var a=function() {};`,
				Options: "nofunc",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `(() => { alert(a); var a = 42; })();`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 16}},
			},
			{
				Code:   `(() => a())(); function a() { }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 8}},
			},
			{
				// SKIP: rslint does not model ESLint's ecmaVersion-dependent
				// function-in-block hoisting. Upstream reports here only under
				// `ecmaVersion: 5`; the `ecmaVersion: 6` copy of this exact code
				// is a valid case above, and that is the behavior rslint has.
				Skip:   true,
				Code:   `"use strict"; a(); { function a() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `a(); try { throw new Error() } catch (foo) {var a;}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `var f = () => a; var a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `new A(); class A {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Message: "'A' was used before it was defined.", Line: 1, Column: 5}},
			},
			{
				Code:   `function foo() { new A(); } class A {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 22}},
			},
			{
				Code:   `new A(); var A = class {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 5}},
			},
			{
				Code:   `function foo() { new A(); } var A = class {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 22}},
			},

			// ---- Block-level bindings ----
			{
				Code:   `a++; { var a; }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:   `"use strict"; { a(); function a() {} }`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 17}},
			},
			{
				Code:   `{a; let a = 1}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 2}},
			},
			{
				Code:   "switch (foo) { case 1: a();\n default: \n let a;}",
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 24}},
			},
			{
				Code:   `if (true) { function foo() { a; } let a;}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 30}},
			},

			// ---- object style options ----
			{
				Code:    `a(); var a=function() {};`,
				Options: map[string]any{"functions": false, "classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},
			{
				Code:    `new A(); class A {};`,
				Options: map[string]any{"functions": false, "classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 5}},
			},
			{
				Code:    `new A(); var A = class {};`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 5}},
			},
			{
				Code:    `function foo() { new A(); } var A = class {};`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 22}},
			},

			// ---- invalid initializers ----
			{
				Code:   `var a = a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 9}},
			},
			{
				Code:   `let a = a + b;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 9}},
			},
			{
				Code:   `const a = foo(a);`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `function foo(a = a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
			{
				Code:   `var {a = a} = [];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `var [a = a] = [];`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `var {b = a, a} = {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `var [b = a, a] = {};`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `var {a = 0} = a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `var [a = 0] = a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `for (var a in a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `for (var a of a) {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},

			// ---- "variables" option ----
			{
				Code:    `function foo() { bar; var bar = 1; } var bar;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
			{
				Code:    `foo; var foo;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 1}},
			},

			// ---- https://github.com/eslint/eslint/issues/10227 ----
			{
				Code:   `for (let x = x;;); let x = 0`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 14}},
			},
			{
				Code:   `for (let x in xs); let xs = []`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `for (let x of xs); let xs = []`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 15}},
			},
			{
				Code:   `try {} catch ({message = x}) {} let x = ''`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 26}},
			},
			{
				Code:   `with (obj) x; let x = {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},

			// ---- WithStatements ----
			{
				Code:   `with (x); let x = {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 7}},
			},
			{
				Code:   `with (obj) { x } let x = {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 14}},
			},
			{
				Code:   `with (obj) { if (a) { x } } let x = {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 23}},
			},
			{
				Code:   `with (obj) { (() => { if (a) { x } })() } let x = {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 32}},
			},

			// ---- Tests related to class definition evaluation. These are TDZ errors. ----
			{
				Code:    `class C extends C {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 17}},
			},
			{
				Code:    `const C = class extends C {};`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 25}},
			},
			{
				Code:    `class C extends (class { [C](){} }) {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 27}},
			},
			{
				Code:    `const C = class extends (class { [C](){} }) {};`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 35}},
			},
			{
				Code:    `class C extends (class { static field = C; }) {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 41}},
			},
			{
				Code:    `const C = class extends (class { static field = C; }) {};`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 49}},
			},
			{
				Code:    `class C { [C](){} }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `(class C { [C](){} });`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 13}},
			},
			{
				Code:    `const C = class { [C](){} };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `class C { static [C](){} }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `(class C { static [C](){} });`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `const C = class { static [C](){} };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 27}},
			},
			{
				Code:    `class C { [C]; }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `(class C { [C]; });`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 13}},
			},
			{
				Code:    `const C = class { [C]; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `class C { [C] = foo; }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `(class C { [C] = foo; });`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 13}},
			},
			{
				Code:    `const C = class { [C] = foo; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `class C { static [C]; }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `(class C { static [C]; });`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `const C = class { static [C]; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 27}},
			},
			{
				Code:    `class C { static [C] = foo; }`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `(class C { static [C] = foo; });`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `const C = class { static [C] = foo; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 27}},
			},
			{
				Code:    `const C = class { static field = C; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 34}},
			},
			{
				Code:    `const C = class { static field = class extends C {}; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 48}},
			},
			{
				Code:    `const C = class { static field = class { [C]; } };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 43}},
			},
			{
				Code:    `const C = class { static field = class { static field = C; }; };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 57}},
			},
			{
				Code:    `class C extends D {} class D {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 17}},
			},
			{
				Code:    `class C extends (class { [a](){} }) {} let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 27}},
			},
			{
				Code:    `class C extends (class { static field = a; }) {} let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 41}},
			},
			{
				Code:    `class C { [a]() {} } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `class C { static [a]() {} } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `class C { [a]; } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `class C { static [a]; } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `class C { [a] = foo; } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 12}},
			},
			{
				Code:    `class C { static [a] = foo; } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 19}},
			},
			{
				Code:    `class C { static field = a; } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 26}},
			},
			{
				Code:    `class C { static field = D; } class D {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 26}},
			},
			{
				Code:    `class C { static field = class extends D {}; } class D {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 40}},
			},
			{
				Code:    `class C { static field = class { [a](){} } } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 35}},
			},
			{
				Code:    `class C { static field = class { static field = a; }; } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 49}},
			},
			{
				Code:    `const C = class { static { C; } };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 28}},
			},
			{
				Code:    `const C = class { static { (class extends C {}); } };`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 43}},
			},
			{
				Code:    `class C { static { a; } } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `class C { static { D; } } class D {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `class C { static { (class extends D {}); } } class D {}`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 35}},
			},
			{
				Code:    `class C { static { (class { [a](){} }); } } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 30}},
			},
			{
				Code:    `class C { static { (class { static field = a; }); } } let a;`,
				Options: map[string]any{"variables": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 44}},
			},
			{
				Code:    `(class C extends C {});`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 18}},
			},
			{
				Code:    `(class C extends (class { [C](){} }) {});`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 28}},
			},
			{
				Code:    `(class C extends (class { static field = C; }) {});`,
				Options: map[string]any{"classes": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 42}},
			},

			// ---- "allowNamedExports" option ----
			{
				Code:   `export { a }; const a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:    `export { a }; const a = 1;`,
				Options: map[string]any{},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:    `export { a }; const a = 1;`,
				Options: map[string]any{"allowNamedExports": false},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:    `export { a }; const a = 1;`,
				Options: "nofunc",
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `export { a as b }; const a = 1;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code: `export { a, b }; let a, b;`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 10},
					{MessageId: "usedBeforeDefined", Line: 1, Column: 13},
				},
			},
			{
				Code:   `export { a }; var a;`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `export { f }; function f() {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:   `export { C }; class C {}`,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 10}},
			},
			{
				Code:    `export const foo = a; const a = 1;`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 20}},
			},
			{
				Code:    `export default a; const a = 1;`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 16}},
			},
			{
				Code:    `export function foo() { return a; }; const a = 1;`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 32}},
			},
			{
				Code:    `export class C { foo() { return a; } }; const a = 1;`,
				Options: map[string]any{"allowNamedExports": true},
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "usedBeforeDefined", Line: 1, Column: 33}},
			},

			// ---- JSX ----
			{
				Code: `<App />; const App = () => <div />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Message: "'App' was used before it was defined.", Line: 1, Column: 2, EndLine: 1, EndColumn: 5},
				},
			},
			{
				Code: `function render() { return <Widget /> }; const Widget = () => <span />;`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 29, EndLine: 1, EndColumn: 35},
				},
			},
			{
				Code: `<Foo.Bar />; const Foo = { Bar: () => <div/> };`,
				Tsx:  true,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "usedBeforeDefined", Line: 1, Column: 2, EndLine: 1, EndColumn: 5},
				},
			},
		},
	)
}
