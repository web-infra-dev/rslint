// cspell:ignore wibble
package no_implied_eval

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoImpliedEvalRule(t *testing.T) {
	rule_tester.RunRuleTester(
		fixtures.GetRootDir(),
		"tsconfig.json",
		t,
		&NoImpliedEvalRule,
		[]rule_tester.ValidTestCase{
			// ---- Direct references without a call (global name is not invoked as eval) ----
			{Code: `setTimeout();`},
			{Code: `setTimeout;`},
			{Code: `setTimeout = foo;`},
			{Code: `window.setTimeout;`},
			{Code: `window.setTimeout = foo;`},
			{Code: `window['setTimeout'];`},
			{Code: `window['setTimeout'] = foo;`},
			{Code: `global.setTimeout;`},
			{Code: `global.setTimeout = foo;`},
			{Code: `global['setTimeout'];`},
			{Code: `global['setTimeout'] = foo;`},
			{Code: `globalThis['setTimeout'] = foo;`},

			// ---- Calls where the first argument is not evaluated as a string ----
			{Code: `setTimeout(function() { x = 1; }, 100);`},
			{Code: `setInterval(function() { x = 1; }, 100);`},
			{Code: `execScript(function() { x = 1; }, 100);`},
			{Code: `window.setTimeout(function() { x = 1; }, 100);`},
			{Code: `window.setInterval(function() { x = 1; }, 100);`},
			{Code: `window.execScript(function() { x = 1; }, 100);`},
			{Code: `window.setTimeout(foo, 100);`},
			{Code: `window.setInterval(foo, 100);`},
			{Code: `window.execScript(foo, 100);`},
			{Code: `global.setTimeout(function() { x = 1; }, 100);`},
			{Code: `global.setInterval(function() { x = 1; }, 100);`},
			{Code: `global.execScript(function() { x = 1; }, 100);`},
			{Code: `global.setTimeout(foo, 100);`},
			{Code: `global.setInterval(foo, 100);`},
			{Code: `global.execScript(foo, 100);`},
			{Code: `globalThis.setTimeout(foo, 100);`},

			// ---- Non-global receiver: only top-level / known-global receivers are checked ----
			{Code: `foo.setTimeout('hi')`},

			// ---- Identifier / function-expression arguments are safe ----
			{Code: `setTimeout(foo, 10)`},
			{Code: `setInterval(1, 10)`},
			{Code: `execScript(2)`},
			{Code: `setTimeout(function() {}, 10)`},
			{Code: `foo.setInterval('hi')`},
			{Code: `setInterval(foo, 10)`},
			{Code: `setInterval(function() {}, 10)`},
			{Code: `foo.execScript('hi')`},
			{Code: `execScript(foo)`},
			{Code: `execScript(function() {})`},

			// ---- Binary `+` of non-strings does not guarantee a string ----
			{Code: `setTimeout(foo + bar, 10)`},

			// ---- Only the first argument is checked ----
			{Code: `setTimeout(foobar, 'buzz')`},
			{Code: `setTimeout(foobar, foo + 'bar')`},

			// ---- Only the immediate subtree of the argument is inspected ----
			{Code: `setTimeout(function() { return 'foobar'; }, 10)`},

			// ---- Prefix-match names that are not actually setTimeout/etc. ----
			// https://github.com/eslint/eslint/issues/7821
			{Code: `setTimeoutFooBar('Foo Bar')`},

			// ---- Non-global intermediate receivers ----
			{Code: `foo.window.setTimeout('foo', 100);`},
			{Code: `foo.global.setTimeout('foo', 100);`},

			// ---- Shadowing by variable declaration ----
			{Code: `var window; window.setTimeout('foo', 100);`},
			{Code: `var global; global.setTimeout('foo', 100);`},

			// ---- Shadowing by function parameter ----
			{Code: `function foo(window) { window.setTimeout('foo', 100); }`},
			{Code: `function foo(global) { global.setTimeout('foo', 100); }`},

			// ---- Not a call — passed as an argument ----
			{Code: `foo('', window.setTimeout);`},
			{Code: `foo('', global.setTimeout);`},

			// ---- Shadowing by function declaration at the top level ----
			// https://github.com/eslint/eslint/issues/19923
			{Code: `
			function execScript(string) {
				console.log("This is not your grandparent's execScript().");
			}

			execScript('wibble');
			`},
			{Code: `
			function setTimeout(string) {
				console.log("This is not your grandparent's setTimeout().");
			}

			setTimeout('wibble');
			`},
			{Code: `
			function setInterval(string) {
				console.log("This is not your grandparent's setInterval().");
			}

			setInterval('wibble');
			`},

			// ---- Shadowing in a nested scope ----
			{Code: `
			function outer() {
				function setTimeout(string) {
					console.log("Shadowed setTimeout");
				}
				setTimeout('code');
			}
			`},
			{Code: `
			function outer() {
				function setInterval(string) {
					console.log("Shadowed setInterval");
				}
				setInterval('code');
			}
			`},
			{Code: `
			function outer() {
				function execScript(string) {
					console.log("Shadowed execScript");
				}
				execScript('code');
			}
			`},
			{Code: `
			{
				const setTimeout = function(string) {
					console.log("Block-scoped setTimeout");
				};
				setTimeout('code');
			}
			`},
			{Code: `
			{
				const setInterval = function(string) {
					console.log("Block-scoped setInterval");
				};
				setInterval('code');
			}
			`},

			// ---- Template literal receivers are not static names in our set ----
			{Code: "window[`SetTimeOut`]('foo', 100);"},
			{Code: "global[`SetTimeOut`]('foo', 100);"},
			{Code: "global[`setTimeout${foo}`]('foo', 100);"},
			{Code: "globalThis[`setTimeout${foo}`]('foo', 100);"},
			{Code: "self[`SetTimeOut`]('foo', 100);"},
			{Code: "self[`setTimeout${foo}`]('foo', 100);"},

			// ---- self as global candidate ----
			{Code: `self.setTimeout;`},
			{Code: `self.setTimeout = foo;`},
			{Code: `self['setTimeout'];`},
			{Code: `self['setTimeout'] = foo;`},
			{Code: `self.setTimeout(function() { x = 1; }, 100);`},
			{Code: `self.setInterval(function() { x = 1; }, 100);`},
			{Code: `self.execScript(function() { x = 1; }, 100);`},
			{Code: `self.setTimeout(foo, 100);`},
			{Code: `foo.self.setTimeout('foo', 100);`},
			{Code: `var self; self.setTimeout('foo', 100);`},
			{Code: `function foo(self) { self.setTimeout('foo', 100); }`},
			{Code: `foo('', self.setTimeout);`},
			{Code: `
			function outer() {
				function self() {
					console.log("Shadowed self");
				}
				self.setTimeout('code');
			}`},

			// ---- Cross-candidate chains: upstream only walks same-name chains ----
			{Code: `window.global.setTimeout('code');`},
			{Code: `self.window.setTimeout('code');`},
			{Code: `globalThis.window.setTimeout('code');`},
			{Code: `window['global']['setTimeout']('code');`},
			{Code: `self['globalThis']['setTimeout']('code');`},
			// `this` is not a global candidate, so `this.setTimeout('code')` is not flagged.
			{Code: `this.setTimeout('code');`},
			{Code: `this['setTimeout']('code');`},

			// ---- TS outer expressions on callee / receiver: upstream uses
			// `isSpecificId` / `isSpecificMemberAccess`, which reject
			// TSNonNullExpression / TSAsExpression / TSTypeAssertion /
			// TSSatisfiesExpression (they are not Identifier / MemberExpression).
			// Strict alignment: not flagged.
			{Code: `setTimeout!('code')`},
			{Code: `(setTimeout as Function)('code')`},
			{Code: `(window.setTimeout!)('code')`},
			{Code: `(window.setTimeout as Function)('code')`},
			{Code: `window!.setTimeout('code')`},
			{Code: `(window as any).setTimeout('code')`},
			{Code: `(window as any).execScript('code')`},

			// ---- let / var with writes: not resolvable → not flagged ----
			{Code: `let s = 'x'; s = 'y'; setTimeout(s);`},
			{Code: `var s = 'x'; s = 'y'; setTimeout(s);`},

			// ---- Conditional with unresolvable cond → not flagged ----
			{Code: `setTimeout(c ? 'x' : 'y');`},

			// ---- Logical / nullish with non-string short-circuit result → not flagged ----
			{Code: `setTimeout(null && 'x');`},
			{Code: `setTimeout(1 ?? 'b');`}, // 1 is not nullish → returns 1 (number)
			// Purely numeric binary produces a number, not a string.
			{Code: `setTimeout(1 + 2);`},
			{Code: `const a = 1; const b = 2; setTimeout(a + b);`},

			// ---- String() with unresolvable arg → not flagged ----
			{Code: `setTimeout(String(x));`},
			{Code: `setTimeout(String(x + y));`},
			{Code: `setTimeout(Number('x'));`},   // Number() produces a number
			{Code: `setTimeout(Boolean('x'));`},  // Boolean() produces a boolean
			{Code: `function String(){} setTimeout(String('x'));`},

			// ---- String.raw with unresolvable sub → not flagged ----
			{Code: "setTimeout(String.raw`x${y}`);"},

			// ---- typeof on unresolvable operand → not flagged ----
			{Code: `setTimeout(typeof x);`},
			{Code: `setTimeout(void 0);`},

			// ---- const of number / undefined / null → not a string ----
			{Code: `const n = 1; setTimeout(n);`},
			{Code: `const u = undefined; setTimeout(u);`},
			{Code: `const z = null; setTimeout(z);`},

			// ---- Property access on unresolvable receiver → not flagged ----
			{Code: `setTimeout(o.x);`},
			{Code: `let o = { x: 'y' }; o = null; setTimeout(o.x);`},
			{Code: `setTimeout({ x: y }.x);`}, // property value unresolvable
			{Code: `const o = { x: { y: z } }; setTimeout(o.x.y);`},
			{Code: `setTimeout({ [k]: 'y' }.x);`}, // computed dynamic key

			// ---- String methods that eslint-utils does NOT fold → not flagged ----
			{Code: `setTimeout('x'.repeat(2));`},
			{Code: `setTimeout('x'.replace('a', 'b'));`},
			{Code: `setTimeout('x'.replaceAll('a', 'b'));`},
			{Code: `setTimeout('x'.split('a'));`},
			{Code: `setTimeout('x'.charCodeAt(0));`},
			{Code: `setTimeout('x'.indexOf('a'));`},
			{Code: `setTimeout('x'.startsWith('a'));`},
			{Code: `setTimeout('x'.toLocaleString());`},

			// ---- Array with unresolvable element → array unresolvable ----
			{Code: `setTimeout([x].toString());`},
			{Code: `setTimeout([x].join(','));`},

			// ---- Method on unresolvable receiver → not flagged ----
			{Code: `setTimeout(foo().toString());`},
			{Code: `setTimeout((typeof x).toUpperCase());`}, // typeof of unresolvable
			{Code: `setTimeout(Array.from('x'));`}, // Array.from not in whitelist
			{Code: "setTimeout(tag`x`);"}, // unknown tagged template

			// SKIP upstream cases that depend on ESLint's `languageOptions.globals` /
			// `sourceType` configuration, which rslint does not model. rslint's
			// `IsShadowed` treats any undeclared reference as global, so these
			// upstream-valid cases become false positives here.
			//
			// Covered upstream cases:
			//   "window.setTimeout('foo')"          { sourceType: "commonjs" }
			//   "window.setInterval('foo')"         { sourceType: "commonjs" }
			//   "window['setTimeout']('foo')"       { sourceType: "commonjs" }
			//   "window['setInterval']('foo')"      { sourceType: "commonjs" }
			//   "global.setTimeout('foo')"          { globals: browser }
			//   "global.setInterval('foo')"         { globals: browser }
			//   "global['setTimeout']('foo')"       { globals: browser }
			//   "global['setInterval']('foo')"      { globals: browser }
			//   "setTimeout('code');"               { globals: {} }
			//   "setInterval('code');"              { globals: {} }
			//   "execScript('code');"               { globals: {} }
			//   "window.setTimeout('code');"        { globals: {} }
			//   "self.setTimeout('code');"          { globals: {} }
		},
		[]rule_tester.InvalidTestCase{
			// ---- Direct calls with string literal ----
			{
				Code: `setTimeout("x = 1;");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout("x = 1;", 100);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setInterval("x = 1;");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `execScript("x = 1;");`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- Const resolution and String(...) constructor ----
			{
				Code: `const s = 'x=1'; setTimeout(s, 100);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 18},
				},
			},
			{
				Code: `setTimeout(String('x=1'), 100);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- window.* member access ----
			{
				Code: `window.setTimeout('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.setInterval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.execScript('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `window['setTimeout']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window['setInterval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "window[`setInterval`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window['execScript']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: "window[`execScript`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- Chained globals: window.window.* ----
			{
				Code: `window.window['setInterval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.window['execScript']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- global.* member access ----
			{
				Code: `global.setTimeout('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.setInterval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.execScript('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `global['setTimeout']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global['setInterval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "global[`setInterval`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global['execScript']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: "global[`execScript`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.global['setInterval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.global['execScript']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- globalThis.* member access ----
			{
				Code: `globalThis.setTimeout('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis.setInterval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis.execScript('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- Template literals as first argument ----
			{
				Code: "setTimeout(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "window.setTimeout(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "window.window.setTimeout(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "global.global.setTimeout(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- String concatenation as first argument ----
			{
				Code: `setTimeout('foo' + bar)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(foo + 'bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "setTimeout(`foo` + bar)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.setTimeout('foo' + bar)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.setTimeout(foo + 'bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "window.setTimeout(`foo` + bar)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window.window.setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.setTimeout('foo' + bar)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.setTimeout(foo + 'bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "global.setTimeout(`foo` + bar)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `global.global.setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `globalThis.setTimeout('foo' + bar)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- Correct node / line reporting when nested ----
			{
				Code: "setTimeout('foo' + (function() {\n   setTimeout(helper);\n   execScript('str');\n   return 'bar';\n})())",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
					{MessageId: "execScript", Line: 3, Column: 4},
				},
			},
			{
				Code: "window.setTimeout('foo' + (function() {\n   setTimeout(helper);\n   window.execScript('str');\n   return 'bar';\n})())",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
					{MessageId: "execScript", Line: 3, Column: 4},
				},
			},
			{
				Code: "global.setTimeout('foo' + (function() {\n   setTimeout(helper);\n   global.execScript('str');\n   return 'bar';\n})())",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
					{MessageId: "execScript", Line: 3, Column: 4},
				},
			},

			// ---- Optional chaining ----
			{
				Code: `window?.setTimeout('code', 0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `(window?.setTimeout)('code', 0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `window?.execScript('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `(window?.execScript)('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- self.* member access ----
			{
				Code: `self.setTimeout('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.setInterval('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.execScript('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `self['setTimeout']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self['setInterval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "self[`setInterval`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self['execScript']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: "self[`execScript`]('foo')",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.self['setInterval']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.self['execScript']('foo')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: "self.setTimeout(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "self.self.setTimeout(`foo${bar}`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.setTimeout('foo' + bar)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.setTimeout(foo + 'bar')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "self.setTimeout(`foo` + bar)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self.self.setTimeout(1 + ';' + 1)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self?.setTimeout('code', 0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `(self?.setTimeout)('code', 0)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `self?.execScript('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},
			{
				Code: `(self?.execScript)('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- Multi-line: position targets the CallExpression ----
			{
				Code: "window.setTimeout(\n  'code',\n  0\n);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1, EndLine: 4, EndColumn: 2},
				},
			},

			// ---- Parenthesized callee ----
			{
				Code: `(setTimeout)('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `((setTimeout))('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `(window).setTimeout('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- Parenthesized first argument ----
			{
				Code: `setTimeout(('code'))`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- No-substitution template as first argument ----
			{
				Code: "setTimeout(`code`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "window.setTimeout(`code`)",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- Optional call: setTimeout?.('code') ----
			{
				Code: `setTimeout?.('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `execScript?.('code')`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "execScript", Line: 1, Column: 1},
				},
			},

			// ---- TS outer expressions on first argument ----
			{
				Code: `setTimeout('code' as any)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout('code'!)`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- Class member context (global name not shadowed) ----
			{
				Code: `class A { x = setTimeout('code'); }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 15},
				},
			},
			{
				Code: `class A { static { window.setTimeout('code'); } }`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 20},
				},
			},

			// ---- IIFE ----
			{
				Code: `(function() { setTimeout('code'); })()`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 15},
				},
			},

			// ---- Multiple errors in one file ----
			{
				Code: "setTimeout('a');\nsetInterval('b');\nexecScript('c');",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
					{MessageId: "impliedEval", Line: 2, Column: 1},
					{MessageId: "execScript", Line: 3, Column: 1},
				},
			},

			// ---- Binary + with identifier resolution (const / let-no-writes / var-no-writes) ----
			{
				Code: `const a = 'x'; const b = 'y'; setTimeout(a + b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 31},
				},
			},
			{
				Code: `const a = 1; const b = 'y'; setTimeout(a + b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 29},
				},
			},
			{
				Code: `const a = 'x'; const b = 'y'; const c = a + b; setTimeout(c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 48},
				},
			},
			// Logical result that IS a string (non-nullish left) should flag.
			{
				Code: `setTimeout('a' ?? 'b');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `const a = 'x'; const b = a; setTimeout(b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 29},
				},
			},

			// ---- let / var with no writes ----
			{
				Code: `let s = 'x'; setTimeout(s);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 14},
				},
			},
			{
				Code: `var s = 'x'; setTimeout(s);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 14},
				},
			},

			// ---- Conditional: cond resolves, both branches (or the chosen one) are string ----
			{
				Code: `setTimeout(true ? 'x' : 'y');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(false ? 1 : 'y');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `const flag = true; setTimeout(flag ? 'x' : 'y');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 20},
				},
			},

			// ---- Logical ||, &&, ?? with resolvable left ----
			{
				Code: `setTimeout('a' || 'b');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout('' || 'fallback');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout('a' && 'b');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(null ?? 'x');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(undefined ?? 'x');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- typeof on resolvable operand always produces a string ----
			{
				Code: `setTimeout(typeof 'x');`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `const n = 1; setTimeout(typeof n);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 14},
				},
			},

			// ---- String() with resolvable arg (of any type), or no args ----
			{
				Code: `setTimeout(String(5));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(String(undefined));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(String(null));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(String(true));`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `setTimeout(String());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},

			// ---- String.raw tagged templates ----
			{
				Code: "setTimeout(String.raw`x`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: "const y = 'z'; setTimeout(String.raw`x${y}`);",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 16},
				},
			},

			// ---- Deep nesting: conditional inside binary + with resolvable operand ----
			{
				Code: `const b = 'z'; setTimeout((true ? 'x' : 'y') + b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 16},
				},
			},
			// ---- Deep nesting: logical inside conditional inside binary + ----
			{
				Code: `const b = 'z'; setTimeout((true ? ('a' || 'c') : 'y') + b);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 16},
				},
			},

			// ---- Property access on const / let / var object literal ----
			{
				Code: `const o = { x: 'y' }; setTimeout(o.x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 23},
				},
			},
			{
				Code: `const o = { x: 'y' }; setTimeout(o['x']);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 23},
				},
			},
			{
				Code: `let o = { x: 'y' }; setTimeout(o.x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 21},
				},
			},
			{
				Code: `const o = { a: { b: { c: 'y' } } }; setTimeout(o.a.b.c);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 37},
				},
			},
			{
				Code: `setTimeout({ ['x']: 'y' }.x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			// Shorthand property: key resolves to the outer binding.
			{
				Code: `const x = 'y'; setTimeout({ x }.x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 16},
				},
			},

			// ---- String.prototype method whitelist ----
			{Code: `setTimeout('x'.toString());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.toUpperCase());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.toLowerCase());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout(' x'.trim());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout(' x'.trimStart());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x '.trimEnd());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.concat('y'));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.slice(0));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('xy'.substring(0, 1));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.padStart(3));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.padEnd(3));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.charAt(0));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.at(0));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout('x'.normalize());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},

			// ---- Method chains ----
			{
				Code: `setTimeout('x'.toUpperCase().toLowerCase());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 1},
				},
			},
			{
				Code: `const s = 'x'; setTimeout(s.toUpperCase());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 16},
				},
			},
			{
				Code: `const o = { x: 'y' }; setTimeout(o.x.toUpperCase());`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 23},
				},
			},

			// ---- Number.prototype method whitelist ----
			{Code: `setTimeout((1).toString());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout((1).toFixed(2));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout((1).toExponential());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout((1).toPrecision(3));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},

			// ---- Array literal + method ----
			{Code: `setTimeout([1, 2].toString());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout([].toString());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout([1, 2].join(','));`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},
			{Code: `setTimeout([1, 2].slice(0).toString());`, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "impliedEval", Line: 1, Column: 1}}},

			// ---- TS non-null on const string via member ----
			{
				Code: `const o = { x: 'y' }; setTimeout(o.x!);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 23},
				},
			},

			// ---- Computed property key resolved through const ----
			{
				Code: `const key = 'x'; setTimeout({ [key]: 'y' }.x);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 18},
				},
			},
			{
				Code: `const key = 'x'; const o = { x: 'y' }; setTimeout(o[key]);`,
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "impliedEval", Line: 1, Column: 40},
				},
			},
		},
	)
}
