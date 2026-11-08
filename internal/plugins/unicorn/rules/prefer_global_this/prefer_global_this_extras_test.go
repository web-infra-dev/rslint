// Additional scope, ESTree, TypeScript, and fix-safety coverage.
// cspell:ignore onpageswap onbeforeunload onresize ndow
package prefer_global_this_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_global_this"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

func TestPreferGlobalThisExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_this.PreferGlobalThisRule, []rule_tester.ValidTestCase{
		// Additional scope, Unicode, and nested JSX boundaries.
		{Code: "import type {window, self, global} from 'pkg'; consume(window, self, global);", FileName: "case.ts"},
		{Code: "namespace window {} consume(window); type T = window;", FileName: "case.ts"},
		{Code: "typeof (/** @type {unknown} */ (window));", FileName: "case.js"},
		{Code: "window.n\\u0061me; window.\\u0061ddEventListener('resize', fn);", FileName: "case.ts"},
		{Code: "new window.addEventListener('play'); window.addEventListener`play`; (window?.addEventListener)?.('play');", FileName: "case.ts"},

		{Code: "const window = {}; const self = {}; const global = {}; window.foo; self.foo; global.foo;", FileName: "case.ts"},
		{Code: "function f(window, {self}, ...global) { return [window, self, global]; }", FileName: "case.ts"},
		{Code: "const fn = function self() { return self; }; const cls = class global { method() { return global; } };", FileName: "case.ts"},
		{Code: "class C { static { const window = {}; window.foo; } method(self) { return self; } }", FileName: "case.ts"},
		{Code: "try {} catch (window) { window.foo; } { let self; self; }", FileName: "case.ts"},
		{Code: "import {window as self} from 'pkg'; self.foo;", FileName: "case.ts"},
		{Code: "type T<window> = window; interface global { x: string };", FileName: "case.ts"},
		{Code: "type W = typeof window.foo.bar; type S = typeof self; type G = typeof global.foo;", FileName: "case.ts"},
		{Code: "typeof (((window as unknown) satisfies unknown)!); typeof self<string>;", FileName: "case.ts"},
		{Code: "window.name; window.close(); window.open(); window.screen; window.onpageswap; window.onbeforeunload = fn; 'navigation' in window; 'onresize' in window;", FileName: "case.ts"},
		{Code: "window.addEventListener('resize', fn); window.removeEventListener(`load`, fn); window.dispatchEvent('blur'); window.dispatchEvent;", FileName: "case.ts"},
		{Code: "(window?.addEventListener)('play'); (window?.addEventListener)('resize');", FileName: "case.ts"},
		{Code: "interface I extends window.innerWidth {} class C implements window.innerHeight {} class D extends window.innerWidth {}", FileName: "case.ts"},
		{Code: "let self = 1; self;", FileName: "case.js", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}},
		{Code: "<window />; <div window={1} />;", FileName: "case.tsx"},
		{Code: "/** @type {Window} */ (window).name;", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		// Additional scope, Unicode, and nested JSX boundaries.
		{Code: "type globalThis = unknown; window.foo;", FileName: "case.ts", Output: []string{"type globalThis = unknown; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 34},
		}},
		// A local globalThis makes the upstream replacement unsafe.
		{Code: "function f(globalThis = window) {}", FileName: "case.ts", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
		}},
		{Code: "function f(value = window) { let globalThis; }", FileName: "case.ts", Output: []string{"function f(value = globalThis) { let globalThis; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
		}},
		// A local globalThis makes the upstream replacement unsafe.
		{Code: "function f(globalThis = window) {}", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
		}},
		{Code: "namespace A.window {} consume(window);", FileName: "case.ts", Output: []string{"namespace A.window {} consume(globalThis);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
		}},
		{Code: "declare global { var window: object; } consume(window);", FileName: "case.ts", Output: []string{"declare global { var window: object; } consume(globalThis);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 54},
		}},
		{Code: "type globalThis = unknown; window;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"type globalThis = unknown; globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 34},
		}},
		{Code: "var {window = self, nested: {global}} = source; window;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"var {window = globalThis, nested: {global}} = source; globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 49, EndLine: 1, EndColumn: 55},
		}},
		{Code: "window; var window = self;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"globalThis; var window = globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
		}},
		{Code: "for (var {window = self} of values) { global.foo; }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"for (var {window = globalThis} of values) { globalThis.foo; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 24},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 45},
		}},
		{Code: "var window; function f(window) { return window; } window.foo;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"var window; function f(window) { return window; } globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 51, EndLine: 1, EndColumn: 57},
		}},
		{Code: "const obj = {\r\n  w\\u0069ndow /* keep */,\r\n  self\r\n};", FileName: "case.ts", Output: []string{"const obj = {\r\n  window: globalThis /* keep */,\r\n  self: globalThis\r\n};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 3, EndLine: 2, EndColumn: 14},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 3, Column: 3, EndLine: 3, EndColumn: 7},
		}},
		{Code: "const 文 = '😀'; window.foo; self.foo;", FileName: "case.ts", Output: []string{"const 文 = '😀'; globalThis.foo; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 23},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
		}},
		{Code: "<Outer render={<window.Widget />} />;", FileName: "case.jsx", Output: []string{"<Outer render={<globalThis.Widget />} />;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 23},
		}},
		{Code: "<Outer>{window.foo}</Outer>;", FileName: "case.jsx", Output: []string{"<Outer>{globalThis.foo}</Outer>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
		}},
		{Code: "<Outer window={window} />;", FileName: "case.tsx", Output: []string{"<Outer window={globalThis} />;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 22},
		}},
		// A local globalThis makes the upstream replacement unsafe.
		{Code: "function f(globalThis) { return <window.Widget />; }", FileName: "case.tsx", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 34, EndLine: 1, EndColumn: 40},
		}},
		{Code: "<window:Component></window:Component>;", FileName: "case.tsx", Output: []string{"<globalThis:Component></globalThis:Component>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 27},
		}},
		// A local globalThis makes the upstream replacement unsafe.
		{Code: "const globalThis = {}; window.foo;", FileName: "case.ts", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 24, EndLine: 1, EndColumn: 30},
		}},
		{Code: "const obj = {get global() { return window; }, self() { return self; }};", FileName: "case.ts", Output: []string{"const obj = {get global() { return globalThis; }, self() { return globalThis; }};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 42},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 63, EndLine: 1, EndColumn: 67},
		}},
		{Code: "function f() { return new.target === window; }", FileName: "case.ts", Output: []string{"function f() { return new.target === globalThis; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
		}},

		{Code: "([value as window] = array);", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 18},
		}},
		{Code: "type T = typeof foo<window>;", Output: []string{"type T = typeof foo<globalThis>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 27},
		}},
		{Code: "const obj = { window /* keep */ };", Output: []string{"const obj = { window: globalThis /* keep */ };"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
		}},
		{Code: "const 文 = '😀'; window;", Output: []string{"const 文 = '😀'; globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 23},
		}},
		{Code: "function f(value = window) { var window; return window; }", FileName: "case.ts", Output: []string{"function f(value = globalThis) { var window; return window; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
		}},
		{Code: "window?.[key]; (window)[key]; window[('x')]; window[window];", FileName: "case.ts", Output: []string{"window?.[key]; (window)[key]; window[('x')]; window[globalThis];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 53, EndLine: 1, EndColumn: 59},
		}},
		{Code: "class C { #name; method() { return window.#name; } }", FileName: "case.ts", Output: []string{"class C { #name; method() { return globalThis.#name; } }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 42},
		}},
		{Code: "window.addEventListener(); window.removeEventListener(name, fn); window.dispatchEvent(new Event('resize')); window.addEventListener(...events);", FileName: "case.ts", Output: []string{"globalThis.addEventListener(); globalThis.removeEventListener(name, fn); globalThis.dispatchEvent(new Event('resize')); globalThis.addEventListener(...events);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 34},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 66, EndLine: 1, EndColumn: 72},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 109, EndLine: 1, EndColumn: 115},
		}},
		{Code: "window.addEventListener('res' + 'ize', fn); window.addEventListener(`${event}`, fn);", FileName: "case.ts", Output: []string{"globalThis.addEventListener('res' + 'ize', fn); globalThis.addEventListener(`${event}`, fn);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 51},
		}},
		{Code: "window?.addEventListener('play', fn); window?.addEventListener('resize', fn); window.addEventListener?.('play', fn);", FileName: "case.ts", Output: []string{"globalThis?.addEventListener('play', fn); window?.addEventListener('resize', fn); globalThis.addEventListener?.('play', fn);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 79, EndLine: 1, EndColumn: 85},
		}},
		{Code: "(window as any).innerWidth; (self satisfies unknown)[key]; global![key];", FileName: "case.ts", Output: []string{"(globalThis as any).innerWidth; (globalThis satisfies unknown)[key]; globalThis![key];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 34},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 60, EndLine: 1, EndColumn: 66},
		}},
		{Code: "typeof window!.foo; typeof self[key]; typeof global.foo;", FileName: "case.ts", Output: []string{"typeof globalThis!.foo; typeof globalThis[key]; typeof globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 32},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 52},
		}},
		{Code: "('open' as string) in window; (name) in window; 'Open' in window; 'open' in self;", FileName: "case.ts", Output: []string{"('open' as string) in globalThis; (name) in globalThis; 'Open' in globalThis; 'open' in globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 41, EndLine: 1, EndColumn: 47},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 59, EndLine: 1, EndColumn: 65},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 77, EndLine: 1, EndColumn: 81},
		}},
		{Code: "self[window]; global[self]; window[self];", FileName: "case.ts", Output: []string{"globalThis[globalThis]; globalThis[globalThis]; window[globalThis];"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 40},
		}},
		{Code: "type window = unknown; const value = window; type global = unknown; global.foo;", FileName: "case.ts", Output: []string{"type window = unknown; const value = globalThis; type global = unknown; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 44},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 69, EndLine: 1, EndColumn: 75},
		}},
		{Code: "let value: window; let other: self; type T = global.Foo;", FileName: "case.ts", Output: []string{"let value: globalThis; let other: globalThis; type T = globalThis.Foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 18},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 35},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 52},
		}},
		{Code: "export {window, self as worker};", FileName: "case.ts", Output: []string{"export {globalThis as window, globalThis as worker};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 21},
		}},
		{Code: "const obj = {w\\u0069ndow, self}; const value = w\\u0069ndow.foo;", FileName: "case.ts", Output: []string{"const obj = {window: globalThis, self: globalThis}; const value = globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 25},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 27, EndLine: 1, EndColumn: 31},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 48, EndLine: 1, EndColumn: 59},
		}},
		{Code: "const text = '😀';\r\n  /* 注释 */ window.foo;", FileName: "case.ts", Output: []string{"const text = '😀';\r\n  /* 注释 */ globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 12, EndLine: 2, EndColumn: 18},
		}},
		{Code: "({window: {self = window}, ...global} = source);", FileName: "case.ts", Output: []string{"({window: {self = globalThis}, ...global} = source);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 25},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
		}},
		{Code: "[window = self] = source; for (global in obj) {}", FileName: "case.ts", Output: []string{"[window = globalThis] = source; for (global in obj) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 15},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 32, EndLine: 1, EndColumn: 38},
		}},
		{Code: "window ||= self; --global; self++;", FileName: "case.ts", Output: []string{"window ||= globalThis; --global; self++;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 12, EndLine: 1, EndColumn: 16},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 32},
		}},
		{Code: "const window = {}; window.foo;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"const window = {}; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
		}},
		{Code: "var window; window.foo;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"var window; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
		}},
		{Code: "const {window = 1, nested: [self]} = source; global.foo;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"const {window = 1, nested: [self]} = source; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 29, EndLine: 1, EndColumn: 33},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 46, EndLine: 1, EndColumn: 52},
		}},
		{Code: "for (var window of values) { window.foo; }", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Output: []string{"for (var window of values) { globalThis.foo; }"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
		}},
		{Code: "let self = 1; self;", FileName: "case.ts", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Output: []string{"let self = 1; globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 9},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 19},
		}},
		{Code: "<window.innerWidth />; <self.Widget />; <global.Component />;", FileName: "case.tsx", Output: []string{"<globalThis.innerWidth />; <globalThis.Widget />; <globalThis.Component />;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 29},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 42, EndLine: 1, EndColumn: 48},
		}},
		// Upstream offers an unsafe fix; keep the diagnostic without an edit.
		{Code: "<window.Widget></window.Widget>;", FileName: "case.jsx", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
		}},
		{Code: "<window.Widget></window.Widget>;", FileName: "case.tsx", Output: []string{"<globalThis.Widget></globalThis.Widget>;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
		}},
		{Code: "/* global window: off, self: writable, global: readonly */ window.foo; self.foo; global.foo;", FileName: "case.ts", Output: []string{"/* global window: off, self: writable, global: readonly */ globalThis.foo; globalThis.foo; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 60, EndLine: 1, EndColumn: 66},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 72, EndLine: 1, EndColumn: 76},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 82, EndLine: 1, EndColumn: 88},
		}},
		{Code: "delete window; void self; new global();", FileName: "case.ts", Output: []string{"delete globalThis; void globalThis; new globalThis();"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 25},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 37},
		}},
		{Code: "window.foo; self.foo; global.foo;", FileName: "case.ts", Globals: map[string]any{"window": "off", "self": "off", "global": "off"}, Output: []string{"globalThis.foo; globalThis.foo; globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 17},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 23, EndLine: 1, EndColumn: 29},
		}},
		{Code: "const {x} = window; const {y} = self; const {z} = global;", FileName: "case.ts", Output: []string{"const {x} = globalThis; const {y} = globalThis; const {z} = globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 37},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 51, EndLine: 1, EndColumn: 57},
		}},
		// Upstream offers an unsafe fix; keep the diagnostic without an edit.
		{Code: "function f() { let globalThis; return window; }", FileName: "case.ts", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 39, EndLine: 1, EndColumn: 45},
		}},
	})
}

func TestPreferGlobalThisArtifactsFollowDemand(t *testing.T) {
	for _, tc := range []struct {
		code, output, filename string
		count                  int
	}{
		{"window.foo; const obj = {self}; global = 1;", "globalThis.foo; const obj = {self: globalThis}; global = 1;", "demand.ts", 3},
		{"function f(globalThis) { return window; }", "function f(globalThis) { return window; }", "demand.ts", 1},
		{"<window.Widget></window.Widget>;", "<window.Widget></window.Widget>;", "demand.jsx", 1},
		{"// eslint-disable-next-line unicorn/prefer-global-this\nwindow;", "// eslint-disable-next-line unicorn/prefer-global-this\nwindow;", "demand.ts", 0},
	} {
		t.Run(tc.code, func(t *testing.T) {
			helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
			program, source, err := helper.CreateTestProgram(tc.code, tc.filename, "tsconfig.json")
			if err != nil {
				t.Fatal(err)
			}
			results := make(map[rule.EditDemand][]rule.RuleDiagnostic)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
				linter.LintSingleFile(linter.LintSingleFileOptions{
					Program: lintprogram.NewFromCompiler(program), File: source.FileName(),
					GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
						return []rule.ConfiguredRule{{
							Name: prefer_global_this.PreferGlobalThisRule.Name, Severity: rule.SeverityError,
							Run: func(ctx rule.RuleContext) rule.RuleListeners {
								// Reuse a complete graph, as when another rule has
								// already requested references to every name.
								scopeanalysis.References(ctx, nil)
								return prefer_global_this.PreferGlobalThisRule.Run(ctx, nil)
							},
						}}
					},
					Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { results[demand] = append(results[demand], d) }},
				})
				if len(results[demand]) != tc.count {
					t.Fatalf("demand %d: got %d diagnostics, want %d", demand, len(results[demand]), tc.count)
				}
			}
			for demand, diagnostics := range results {
				for index, diagnostic := range diagnostics {
					baseline, all := results[rule.EditDemandNone][index], results[rule.EditDemandAll][index]
					if diagnostic.Range != baseline.Range || !reflect.DeepEqual(diagnostic.Message, baseline.Message) || diagnostic.RuleName != baseline.RuleName || diagnostic.Severity != baseline.Severity {
						t.Errorf("demand %d changed diagnostic identity", demand)
					}
					wantFix := all.FixesPtr != nil && (demand == rule.EditDemandAutofix || demand == rule.EditDemandAll)
					if (diagnostic.FixesPtr != nil) != wantFix {
						t.Errorf("demand %d: unexpected autofix artifacts", demand)
					}
					if wantFix && !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
						t.Errorf("demand %d changed autofix artifacts", demand)
					}
					if diagnostic.Suggestions != nil {
						t.Errorf("demand %d produced suggestions", demand)
					}
				}
				wantOutput := tc.code
				if demand == rule.EditDemandAutofix || demand == rule.EditDemandAll {
					wantOutput = tc.output
				}
				output, _, _ := linter.ApplyRuleFixes(tc.code, diagnostics)
				if output != wantOutput {
					t.Errorf("demand %d: got %q, want %q", demand, output, wantOutput)
				}
			}
		})
	}
}
