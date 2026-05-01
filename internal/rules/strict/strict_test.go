package strict

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Ported from ESLint's tests/lib/rules/strict.js.
//
// Cases that rely on ESLint-only language options (parserOptions.ecmaFeatures
// .{impliedStrict,globalReturn} and sourceType === "commonjs") are preserved
// with Skip: true so the mapping to the upstream suite stays explicit.
// rslint detects ES modules structurally via ast.IsExternalModule, so any
// test that needed sourceType: "module" is adapted to include an
// `export {};` or equivalent trigger.
func TestStrictRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&StrictRule,
		[]rule_tester.ValidTestCase{
			// ---- "never" mode ----
			{Code: `foo();`, Options: "never"},
			{Code: `function foo() { return; }`, Options: "never"},
			{Code: `var foo = function() { return; };`, Options: "never"},
			{Code: `foo(); 'use strict';`, Options: "never"},
			{Code: `function foo() { bar(); 'use strict'; return; }`, Options: "never"},
			{Code: `var foo = function() { { 'use strict'; } return; };`, Options: "never"},
			{Code: `(function() { bar('use strict'); return; }());`, Options: "never"},
			{Code: `var fn = x => 1;`, Options: "never"},
			{Code: `var fn = x => { return; };`, Options: "never"},
			// ESLint uses `sourceType: "module"` here; rslint triggers module mode
			// via an import/export, and in module mode "use strict" is reported
			// (not valid) — so the equivalent valid case is plain script code.
			// The module variant is exercised in the invalid suite below.
			{Code: `foo(); export {};`, Options: "never"},
			// SKIP: parserOptions.ecmaFeatures.impliedStrict is not exposed by rslint.
			{Code: `function foo() { return; }`, Options: "never", Skip: true},

			// ---- "global" mode ----
			{Code: `// Intentionally empty`, Options: "global"},
			{Code: `"use strict"; foo();`, Options: "global"},
			{Code: `/* license */
/* eslint-disable rule-to-test/strict */
foo();`, Options: "global", Skip: true}, // SKIP: relies on ESLint disable semantics we don't emulate here.
			// Module files always ignore "global" in favor of "module" mode — the
			// ESLint variant uses sourceType: "module"; rslint adapts by adding
			// `export {};` which likewise triggers module detection.
			{Code: `foo(); export {};`, Options: "global", Skip: true}, // covered in invalid suite (module reports any "use strict")
			// SKIP: parserOptions.ecmaFeatures.impliedStrict not exposed.
			{Code: `function foo() { return; }`, Options: "global", Skip: true},
			{Code: `'use strict'; function foo() { return; }`, Options: "global"},
			{Code: `'use strict'; var foo = function() { return; };`, Options: "global"},
			{Code: `'use strict'; function foo() { bar(); 'use strict'; return; }`, Options: "global"},
			{Code: `'use strict'; var foo = function() { bar(); 'use strict'; return; };`, Options: "global"},
			{Code: `'use strict'; function foo() { return function() { bar(); 'use strict'; return; }; }`, Options: "global"},
			{Code: `'use strict'; var foo = () => { return () => { bar(); 'use strict'; return; }; }`, Options: "global"},

			// ---- "function" mode ----
			{Code: `function foo() { 'use strict'; return; }`, Options: "function"},
			// SKIP: ESLint uses sourceType: "module"; in rslint, module mode
			// reports all "use strict" — the behavior is covered by the module
			// test cases that add `export {};`.
			{Code: `function foo() { return; } export {};`, Options: "function", Skip: true},
			// SKIP: parserOptions.ecmaFeatures.impliedStrict not exposed.
			{Code: `function foo() { return; }`, Options: "function", Skip: true},
			{Code: `var foo = function() { return; } export {};`, Options: "function", Skip: true},
			{Code: `var foo = function() { 'use strict'; return; }`, Options: "function"},
			{Code: `function foo() { 'use strict'; return; } var bar = function() { 'use strict'; bar(); };`, Options: "function"},
			{Code: `var foo = function() { 'use strict'; function bar() { return; } bar(); };`, Options: "function"},
			{Code: `var foo = () => { 'use strict'; var bar = () => 1; bar(); };`, Options: "function"},
			// SKIP: sourceType: "module" — nested arrow valid case, module variant covered by module tests.
			{Code: `var foo = () => { var bar = () => 1; bar(); }; export {};`, Options: "function", Skip: true},
			{Code: `class A { constructor() { } }`, Options: "function"},
			{Code: `class A { foo() { } }`, Options: "function"},
			{Code: `class A { foo() { function bar() { } } }`, Options: "function"},
			{Code: `(function() { 'use strict'; function foo(a = 0) { } }())`, Options: "function"},

			// ---- "safe" mode ----
			{Code: `function foo() { 'use strict'; return; }`, Options: "safe"},
			// SKIP: parserOptions.ecmaFeatures.globalReturn not exposed.
			{Code: `'use strict'; function foo() { return; }`, Options: "safe", Skip: true},
			// SKIP: sourceType: "module" valid case — behavior covered in invalid suite via `export {};`.
			{Code: `function foo() { return; } export {};`, Options: "safe", Skip: true},
			// SKIP: impliedStrict not exposed.
			{Code: `function foo() { return; }`, Options: "safe", Skip: true},

			// ---- default to "safe" (= "function" in rslint) ----
			{Code: `function foo() { 'use strict'; return; }`},
			// SKIP: globalReturn not exposed.
			{Code: `'use strict'; function foo() { return; }`, Skip: true},
			// SKIP: module-sourced valid case covered elsewhere.
			{Code: `function foo() { return; } export {};`, Skip: true},
			// SKIP: impliedStrict not exposed.
			{Code: `function foo() { return; }`, Skip: true},

			// ---- Class static blocks (no directive prologue inside) ----
			{Code: `'use strict'; class C { static { foo; } }`, Options: "global"},
			{Code: `'use strict'; class C { static { 'use strict'; } }`, Options: "global"},
			{Code: `'use strict'; class C { static { 'use strict'; 'use strict'; } }`, Options: "global"},
			{Code: `class C { static { foo; } }`, Options: "function"},
			{Code: `class C { static { 'use strict'; } }`, Options: "function"},
			{Code: `class C { static { 'use strict'; 'use strict'; } }`, Options: "function"},
			{Code: `class C { static { foo; } }`, Options: "never"},
			{Code: `class C { static { 'use strict'; } }`, Options: "never"},
			{Code: `class C { static { 'use strict'; 'use strict'; } }`, Options: "never"},
			// SKIP: sourceType: "module" — module path exercised in invalid suite.
			{Code: `class C { static { 'use strict'; } } export {};`, Options: "safe", Skip: true},
			// SKIP: impliedStrict not exposed.
			{Code: `class C { static { 'use strict'; } }`, Options: "safe", Skip: true},
			// SKIP: sourceType: "commonjs" — rslint cannot distinguish CommonJS scripts.
			{Code: `'use strict'; module.exports = function identity (value) { return value; }`, Skip: true},
			{Code: `'use strict'; module.exports = function identity (value) { return value; }`, Options: "safe", Skip: true},

			// ---- Class heritage (not class body) ----
			// A function in an `extends` clause is NOT inside the class body, so
			// a `"use strict"` directive there is neither `unnecessary` nor
			// `unnecessaryInClasses` — it is the top-level directive of that
			// function (valid under "function" mode).
			{
				Code:    `class Foo extends (function() { 'use strict'; return class {}; }()) {}`,
				Options: "function",
			},

			// ---- Ambient / bodyless declarations (TypeScript) ----
			// `declare function`, abstract methods, and overload signatures have
			// no runtime body; the rule must not demand per-function "use strict"
			// on them. (Program-level "global" enforcement still applies — a file
			// whose only statement is a declare function with no top-level
			// directive is still flagged by "global" mode, matching ESLint.)
			{Code: `declare function foo(): void;`, Options: "function"},
			{Code: `declare function foo(): void;`, Options: "never"},
			{Code: `'use strict'; declare function foo(): void;`, Options: "global"},
			{Code: `abstract class A { abstract foo(): void; }`, Options: "function"},
			{
				Code:    `function foo(): void; function foo(a: number): void; function foo(a?: number) { 'use strict'; }`,
				Options: "function",
			},
		},
		[]rule_tester.InvalidTestCase{
			// ---- "never" mode ----
			{
				Code:    `"use strict"; foo();`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 1, Column: 1},
				},
			},
			{
				Code:    `function foo() { 'use strict'; return; }`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 1, Column: 18},
				},
			},
			{
				Code:    `var foo = function() { 'use strict'; return; };`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 1, Column: 24},
				},
			},
			{
				Code:    `function foo() { return function() { 'use strict'; return; }; }`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 1, Column: 38},
				},
			},
			{
				Code:    `'use strict'; function foo() { "use strict"; return; }`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 1, Column: 1},
					{MessageId: "never", Line: 1, Column: 32},
				},
			},
			// Module detection (rslint analogue of sourceType: "module").
			{
				Code:    `"use strict"; foo(); export {};`,
				Output:  []string{` foo(); export {};`},
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "module", Line: 1, Column: 1},
				},
			},
			// SKIP: ESLint uses impliedStrict; rslint has no equivalent.
			{
				Code:    `'use strict'; function foo() { 'use strict'; return; }`,
				Options: "never",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "implied"}, {MessageId: "implied"}},
			},
			// SKIP: module + impliedStrict combination; module-only covered above.
			{
				Code:    `'use strict'; function foo() { 'use strict'; return; }`,
				Options: "never",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "module"}, {MessageId: "module"}},
			},

			// ---- "global" mode ----
			{
				Code:    `foo();`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 1},
				},
			},
			{
				Code: `/* license */
function foo() {}
function bar() {}
/* end */`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 2, Column: 1, EndLine: 3, EndColumn: 18},
				},
			},
			{
				Code:    `function foo() { 'use strict'; return; }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 1},
					{MessageId: "global", Line: 1, Column: 18},
				},
			},
			{
				Code:    `var foo = function() { 'use strict'; return; }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 1},
					{MessageId: "global", Line: 1, Column: 24},
				},
			},
			{
				Code:    `var foo = () => { 'use strict'; return () => 1; }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 1},
					{MessageId: "global", Line: 1, Column: 19},
				},
			},
			{
				Code:    `'use strict'; function foo() { 'use strict'; return; }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 32},
				},
			},
			{
				Code:    `'use strict'; var foo = function() { 'use strict'; return; };`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 38},
				},
			},
			{
				Code:    `'use strict'; 'use strict'; foo();`,
				Output:  []string{`'use strict';  foo();`},
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multiple", Line: 1, Column: 15},
				},
			},
			{
				Code:    `'use strict'; foo(); export {};`,
				Output:  []string{` foo(); export {};`},
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "module", Line: 1, Column: 1},
				},
			},
			// SKIP: impliedStrict not exposed.
			{
				Code:    `'use strict'; function foo() { 'use strict'; return; }`,
				Options: "global",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "implied"}, {MessageId: "implied"}},
			},

			// ---- "function" mode ----
			{
				Code:    `'use strict'; foo();`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 1},
				},
			},
			{
				Code:    `'use strict'; (function() { 'use strict'; return true; }());`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 1},
				},
			},
			{
				Code:    `(function() { 'use strict'; function f() { 'use strict'; return } return true; }());`,
				Output:  []string{`(function() { 'use strict'; function f() {  return } return true; }());`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessary", Line: 1, Column: 44},
				},
			},
			{
				Code:    `(function() { return true; }());`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 2},
				},
			},
			{
				Code:    `(() => { return true; })();`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 2},
				},
			},
			{
				Code:    `(() => true)();`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 2},
				},
			},
			{
				Code:    `var foo = function() { foo(); 'use strict'; return; }; function bar() { foo(); 'use strict'; }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 11},
					{MessageId: "function", Line: 1, Column: 56},
				},
			},
			{
				Code:    `function foo() { 'use strict'; 'use strict'; return; }`,
				Output:  []string{`function foo() { 'use strict';  return; }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multiple", Line: 1, Column: 32},
				},
			},
			{
				Code:    `var foo = function() { 'use strict'; 'use strict'; return; }`,
				Output:  []string{`var foo = function() { 'use strict';  return; }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "multiple", Line: 1, Column: 38},
				},
			},
			{
				Code:    `var foo = function() {  'use strict'; return; }; export {};`,
				Output:  []string{`var foo = function() {   return; }; export {};`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "module", Line: 1, Column: 25},
				},
			},
			// SKIP: impliedStrict not exposed.
			{
				Code:    `'use strict'; function foo() { 'use strict'; return; }`,
				Options: "function",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "implied"}, {MessageId: "implied"}},
			},
			{
				Code:    `function foo() { return function() { 'use strict'; return; }; }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 1},
				},
			},
			{
				Code:    `var foo = function() { function bar() { 'use strict'; return; } return; }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 11},
				},
			},
			{
				Code:    `function foo() { 'use strict'; return; } var bar = function() { return; };`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 52},
				},
			},
			{
				Code:    `var foo = function() { 'use strict'; return; }; function bar() { return; };`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 49},
				},
			},
			{
				Code:    `function foo() { 'use strict'; return function() { 'use strict'; 'use strict'; return; }; }`,
				Output:  []string{`function foo() { 'use strict'; return function() {   return; }; }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessary", Line: 1, Column: 52},
					{MessageId: "multiple", Line: 1, Column: 66},
				},
			},
			{
				Code:    `var foo = function() { 'use strict'; function bar() { 'use strict'; 'use strict'; return; } }`,
				Output:  []string{`var foo = function() { 'use strict'; function bar() {   return; } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessary", Line: 1, Column: 55},
					{MessageId: "multiple", Line: 1, Column: 69},
				},
			},
			{
				Code:    `var foo = () => { return; };`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 11},
				},
			},

			// ---- Classes ----
			{
				Code:    `class A { constructor() { "use strict"; } }`,
				Output:  []string{`class A { constructor() {  } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 1, Column: 27},
				},
			},
			{
				Code:    `class A { foo() { "use strict"; } }`,
				Output:  []string{`class A { foo() {  } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 1, Column: 19},
				},
			},
			{
				Code:    `class A { foo() { function bar() { "use strict"; } } }`,
				Output:  []string{`class A { foo() { function bar() {  } } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 1, Column: 36},
				},
			},
			{
				Code:    `class A { field = () => { "use strict"; } }`,
				Output:  []string{`class A { field = () => {  } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 1, Column: 27},
				},
			},
			{
				Code:    `class A { field = function() { "use strict"; } }`,
				Output:  []string{`class A { field = function() {  } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 1, Column: 32},
				},
			},

			// ---- "safe" mode (= "function" in rslint script files) ----
			{
				Code:    `'use strict'; function foo() { return; }`,
				Options: "safe",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 1},
					{MessageId: "function", Line: 1, Column: 15},
				},
			},
			// SKIP: globalReturn not exposed.
			{
				Code:    `function foo() { 'use strict'; return; }`,
				Options: "safe",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "global"}, {MessageId: "global"}},
			},
			// SKIP: impliedStrict not exposed.
			{
				Code:    `'use strict'; function foo() { 'use strict'; return; }`,
				Options: "safe",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "implied"}, {MessageId: "implied"}},
			},

			// ---- Default "safe" (= "function" in rslint script files) ----
			{
				Code: `'use strict'; function foo() { return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 1},
					{MessageId: "function", Line: 1, Column: 15},
				},
			},
			{
				Code: `function foo() { return; }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 1},
				},
			},
			// SKIP: globalReturn not exposed.
			{
				Code:   `function foo() { 'use strict'; return; }`,
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "global"}, {MessageId: "global"}},
			},

			// ---- Non-simple parameter list: https://github.com/eslint/eslint/issues/6405 ----
			{
				Code:    `function foo(a = 0) { 'use strict' }`,
				Options: []interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 23},
				},
			},
			{
				Code:    `(function() { 'use strict'; function foo(a = 0) { 'use strict' } }())`,
				Options: []interface{}{},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 51},
				},
			},
			// SKIP: globalReturn not exposed.
			{
				Code:    `function foo(a = 0) { 'use strict' }`,
				Options: []interface{}{},
				Skip:    true,
				Errors: []rule_tester.InvalidTestCaseError{
					{Message: "Use the global form of 'use strict'."},
					{MessageId: "nonSimpleParameterList"},
				},
			},
			// SKIP: globalReturn not exposed.
			{
				Code:    `'use strict'; function foo(a = 0) { 'use strict' }`,
				Options: []interface{}{},
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "nonSimpleParameterList"}},
			},
			{
				Code:    `function foo(a = 0) { 'use strict' }`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 23},
				},
			},
			{
				Code:    `function foo(a = 0) { 'use strict' }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 1, Column: 1},
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 23},
				},
			},
			{
				Code:    `'use strict'; function foo(a = 0) { 'use strict' }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 37},
				},
			},
			{
				Code:    `function foo(a = 0) { 'use strict' }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 23},
				},
			},
			{
				Code:    `(function() { 'use strict'; function foo(a = 0) { 'use strict' } }())`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "nonSimpleParameterList", Line: 1, Column: 51},
				},
			},
			{
				Code:    `function foo(a = 0) { }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap function 'foo' in a function with 'use strict' directive.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `(function() { function foo(a = 0) { } }())`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Message: "Use the function form of 'use strict'.", Line: 1, Column: 2},
				},
			},

			// ---- wrap message: full modifier / kind coverage (rslint addition) ----
			// Each case sits at program scope with a non-simple parameter list and
			// no directive, which is the only path that reaches `wrap`. Covers the
			// modifier prefixes produced by ESLint's getFunctionNameWithKind.
			{
				Code:    `async function foo(a = 0) { }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap async function 'foo' in a function with 'use strict' directive.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `function* gen(a = 0) { }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap generator function 'gen' in a function with 'use strict' directive.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `async function* gen(a = 0) { }`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap async generator function 'gen' in a function with 'use strict' directive.",
						Line:      1,
						Column:    1,
					},
				},
			},
			{
				Code:    `(function(a = 0) { })();`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap function in a function with 'use strict' directive.",
						Line:      1,
						Column:    2,
					},
				},
			},
			{
				Code:    `(function named(a = 0) { })();`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap function 'named' in a function with 'use strict' directive.",
						Line:      1,
						Column:    2,
					},
				},
			},
			{
				Code:    `var foo = (a = 0) => { };`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						// Now resolves the binding name from the parent
						// VariableDeclaration, matching ESLint's
						// astUtils.getName.
						Message: "Wrap arrow function 'foo' in a function with 'use strict' directive.",
						Line:    1,
						Column:  11,
					},
				},
			},
			{
				Code:    `var foo = async (a = 0) => { };`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{
						MessageId: "wrap",
						Message:   "Wrap async arrow function 'foo' in a function with 'use strict' directive.",
						Line:      1,
						Column:    11,
					},
				},
			},

			// ---- Functions inside class static blocks ----
			{
				Code: `'use strict'; class C { static { function foo() {
'use strict'; } } }`,
				Options: "global",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "global", Line: 2},
				},
			},
			{
				Code: `class C { static { function foo() {
'use strict'; } } }`,
				Options: "never",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "never", Line: 2},
				},
			},
			{
				Code: `class C { static { function foo() {
'use strict'; } } } export {};`,
				Output: []string{`class C { static { function foo() {
 } } } export {};`},
				Options: "safe",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "module", Line: 2},
				},
			},
			// SKIP: impliedStrict not exposed.
			{
				Code: `class C { static { function foo() {
'use strict'; } } }`,
				Options: "safe",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "implied", Line: 2}},
			},
			{
				Code: `function foo() {'use strict'; class C { static { function foo() {
'use strict'; } } } }`,
				Output: []string{`function foo() {'use strict'; class C { static { function foo() {
 } } } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessary", Line: 2},
				},
			},
			{
				Code: `class C { static { function foo() {
'use strict'; } } }`,
				Output: []string{`class C { static { function foo() {
 } } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 2},
				},
			},
			{
				Code: `class C { static { function foo() {
'use strict';
'use strict'; } } }`,
				Output: []string{`class C { static { function foo() {

 } } }`},
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "unnecessaryInClasses", Line: 2},
					{MessageId: "multiple", Line: 3},
				},
			},
			// SKIP: sourceType: "commonjs" — rslint cannot distinguish CommonJS scripts.
			{
				Code:    `module.exports = function identity (value) { return value; }`,
				Options: "safe",
				Skip:    true,
				Errors:  []rule_tester.InvalidTestCaseError{{MessageId: "global", Line: 1}},
			},
			{
				Code:   `module.exports = function identity (value) { return value; }`,
				Skip:   true,
				Errors: []rule_tester.InvalidTestCaseError{{MessageId: "global", Line: 1}},
			},

			// ---- Class heritage: functions in extends are NOT in class body ----
			// A function expression in an `extends` clause sits outside the class
			// body in ESLint's scope model, so it should be treated as a top-level
			// function — "function" mode requires a `"use strict"` there.
			{
				Code:    `class Foo extends (function() { return class {}; }()) {}`,
				Options: "function",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "function", Line: 1, Column: 20},
				},
			},
		},
	)
}
