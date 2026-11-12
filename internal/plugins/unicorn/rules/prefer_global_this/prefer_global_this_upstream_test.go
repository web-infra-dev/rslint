// Ported from eslint-plugin-unicorn v79.0.0 (MIT).
// Sources: test/prefer-global-this.js and docs/rules/prefer-global-this.md.
// cspell:ignore onresize jsdom onconnect onlanguagechange onoffline ononline onrejectionhandled onunhandledrejection onplay
package prefer_global_this_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_global_this"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferGlobalThisUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_this.PreferGlobalThisRule, []rule_tester.ValidTestCase{
		{Code: "globalThis", FileName: "case.js"},
		{Code: "globalThis.foo", FileName: "case.js"},
		{Code: "globalThis[foo]", FileName: "case.js"},
		{Code: "globalThis.foo()", FileName: "case.js"},
		{Code: "const { foo } = globalThis", FileName: "case.js"},
		{Code: "function foo (window) {}", FileName: "case.js"},
		{Code: "function foo (global) {}", FileName: "case.js"},
		{Code: "var foo = function foo (window) {}", FileName: "case.js"},
		{Code: "var foo = function foo (global) {}", FileName: "case.js"},
		{Code: "var window = {}", FileName: "case.js"},
		{Code: "let global = {}", FileName: "case.js"},
		{Code: "const global = {}", FileName: "case.js"},
		{Code: "function foo (window) {\n\twindow.foo();\n}", FileName: "case.js"},
		{Code: "var window = {};\nfunction foo () {\n\twindow.foo();\n}", FileName: "case.js"},
		{Code: "foo.window", FileName: "case.js"},
		{Code: "foo.global", FileName: "case.js"},
		{Code: "import window from \"xxx\"", FileName: "case.js"},
		{Code: "import * as window from \"xxx\"", FileName: "case.js"},
		{Code: "import window, {foo} from \"xxx\"", FileName: "case.js"},
		{Code: "export { window }  from \"xxx\"", FileName: "case.js"},
		{Code: "export * as window from \"xxx\";", FileName: "case.js"},
		{Code: "try {\n\n} catch (window) {}", FileName: "case.js"},
		{Code: "window.name = \"foo\"", FileName: "case.js"},
		{Code: "window.addEventListener", FileName: "case.js"},
		{Code: "window.innerWidth", FileName: "case.js"},
		{Code: "window.innerHeight", FileName: "case.js"},
		{Code: "window.addEventListener(\"resize\", () => {})", FileName: "case.js"},
		{Code: "window.addEventListener(`resize`, () => {})", FileName: "case.js"},
		{Code: "window.removeEventListener(`resize`, () => {})", FileName: "case.js"},
		{Code: "window.onresize = function () {}", FileName: "case.js"},
		{Code: "'open' in window; window.open(\"https://example.com\")", FileName: "case.js"},
		{Code: "`open` in window; window.open(\"https://example.com\")", FileName: "case.js"},
		{Code: "const {window} = jsdom()\nwindow.jQuery = jQuery;", FileName: "case.js"},
		{Code: "({ foo: window.name } =  {})", FileName: "case.js"},
		{Code: "[window.name] = []", FileName: "case.js"},
		{Code: "window[foo]", FileName: "case.js"},
		{Code: "window[title]", FileName: "case.js"},
		{Code: "window[\"foo\"]", FileName: "case.js"},
		{Code: "typeof window", FileName: "case.js"},
		{Code: "typeof window !== \"undefined\"", FileName: "case.js"},
		{Code: "typeof self !== \"undefined\"", FileName: "case.js"},
		{Code: "typeof global !== \"undefined\"", FileName: "case.js"},
		{Code: "typeof window === \"undefined\"", FileName: "case.js"},
		{Code: "typeof window === \"object\"", FileName: "case.js"},
		{Code: "type Window = typeof window;", FileName: "case.ts"},
		{Code: "let foo: typeof self;", FileName: "case.ts"},
		{Code: "type Foo = typeof window.foo;", FileName: "case.ts"},
		{Code: "type Foo = typeof window.foo.bar;", FileName: "case.ts"},
		{Code: "function foo(bar: typeof self.foo) {}", FileName: "case.ts"},
	}, []rule_tester.InvalidTestCase{
		{Code: "global", FileName: "case.js", Output: []string{"globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "self", FileName: "case.js", Output: []string{"globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "window", FileName: "case.js", Output: []string{"globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.foo", FileName: "case.js", Output: []string{"globalThis.foo"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.foo()", FileName: "case.js", Output: []string{"globalThis.foo()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window > 10", FileName: "case.js", Output: []string{"globalThis > 10"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "10 > window", FileName: "case.js", Output: []string{"10 > globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
		}},
		{Code: "window ?? 10", FileName: "case.js", Output: []string{"globalThis ?? 10"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "10 ?? window", FileName: "case.js", Output: []string{"10 ?? globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
		}},
		{Code: "window.foo = 123", FileName: "case.js", Output: []string{"globalThis.foo = 123"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window = 123", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "obj.a = window", FileName: "case.js", Output: []string{"obj.a = globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
		}},
		{Code: "function* gen() {\n  yield window\n}", FileName: "case.js", Output: []string{"function* gen() {\n  yield globalThis\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 15},
		}},
		{Code: "async function gen() {\n  await window\n}", FileName: "case.js", Output: []string{"async function gen() {\n  await globalThis\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 9, EndLine: 2, EndColumn: 15},
		}},
		{Code: "window ? foo : bar", FileName: "case.js", Output: []string{"globalThis ? foo : bar"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "foo ? window : bar", FileName: "case.js", Output: []string{"foo ? globalThis : bar"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
		}},
		{Code: "foo ? bar : window", FileName: "case.js", Output: []string{"foo ? bar : globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
		}},
		{Code: "function foo() {\n  return window\n}", FileName: "case.js", Output: []string{"function foo() {\n  return globalThis\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 10, EndLine: 2, EndColumn: 16},
		}},
		{Code: "new window()", FileName: "case.js", Output: []string{"new globalThis()"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
		}},
		{Code: "const obj = {\n\tfoo: window.foo,\n\tbar: window.bar,\n\twindow: window\n}", FileName: "case.js", Output: []string{"const obj = {\n\tfoo: globalThis.foo,\n\tbar: globalThis.bar,\n\twindow: globalThis\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 7, EndLine: 2, EndColumn: 13},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 3, Column: 7, EndLine: 3, EndColumn: 13},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 4, Column: 10, EndLine: 4, EndColumn: 16},
		}},
		{Code: "function sequenceTest() {\n\tlet x, y;\n\tx = (y = 10, y + 5, window);\n\tconsole.log(x, y);\n}", FileName: "case.js", Output: []string{"function sequenceTest() {\n\tlet x, y;\n\tx = (y = 10, y + 5, globalThis);\n\tconsole.log(x, y);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 3, Column: 22, EndLine: 3, EndColumn: 28},
		}},
		{Code: "window`Hello ${42} World`", FileName: "case.js", Output: []string{"globalThis`Hello ${42} World`"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "tag`Hello ${window.foo} World`", FileName: "case.js", Output: []string{"tag`Hello ${globalThis.foo} World`"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 13, EndLine: 1, EndColumn: 19},
		}},
		{Code: "var str = `hello ${window.foo} world!`", FileName: "case.js", Output: []string{"var str = `hello ${globalThis.foo} world!`"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
		}},
		{Code: "delete window.foo", FileName: "case.js", Output: []string{"delete globalThis.foo"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
		}},
		{Code: "++window", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
		}},
		{Code: "++window.foo", FileName: "case.js", Output: []string{"++globalThis.foo"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
		}},
		{Code: "for (var attr in window) {\n\n}", FileName: "case.js", Output: []string{"for (var attr in globalThis) {\n\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 24},
		}},
		{Code: "for (window.foo = 0; i < 10; window.foo++) {\n\n}", FileName: "case.js", Output: []string{"for (globalThis.foo = 0; i < 10; globalThis.foo++) {\n\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 36},
		}},
		{Code: "for (const item of window.foo) {\n}", FileName: "case.js", Output: []string{"for (const item of globalThis.foo) {\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
		}},
		{Code: "for (const item of window) {\n}", FileName: "case.js", Output: []string{"for (const item of globalThis) {\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 26},
		}},
		{Code: "switch (window) {}", FileName: "case.js", Output: []string{"switch (globalThis) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
		}},
		{Code: "switch (true) {\n\tcase window:\n\t\tbreak;\n}", FileName: "case.js", Output: []string{"switch (true) {\n\tcase globalThis:\n\t\tbreak;\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 7, EndLine: 2, EndColumn: 13},
		}},
		{Code: "switch (true) {\n\tcase window.foo:\n\t\tbreak;\n}", FileName: "case.js", Output: []string{"switch (true) {\n\tcase globalThis.foo:\n\t\tbreak;\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 2, Column: 7, EndLine: 2, EndColumn: 13},
		}},
		{Code: "while (window) {\n}", FileName: "case.js", Output: []string{"while (globalThis) {\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
		}},
		{Code: "do {} while (window) {}", FileName: "case.js", Output: []string{"do {} while (globalThis) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
		}},
		{Code: "if (window) {}", FileName: "case.js", Output: []string{"if (globalThis) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
		}},
		{Code: "throw window", FileName: "case.js", Output: []string{"throw globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 7, EndLine: 1, EndColumn: 13},
		}},
		{Code: "var foo = window", FileName: "case.js", Output: []string{"var foo = globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 11, EndLine: 1, EndColumn: 17},
		}},
		{Code: "function foo (name = window) {\n\n}", FileName: "case.js", Output: []string{"function foo (name = globalThis) {\n\n}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 28},
		}},
		{Code: "self.innerWidth", FileName: "case.js", Output: []string{"globalThis.innerWidth"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.innerHeight", FileName: "case.js", Output: []string{"globalThis.innerHeight"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.location", FileName: "case.js", Output: []string{"globalThis.location"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.navigator", FileName: "case.js", Output: []string{"globalThis.navigator"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.name", FileName: "case.js", Output: []string{"globalThis.name"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.postMessage(\"Hello\")", FileName: "case.js", Output: []string{"globalThis.postMessage(\"Hello\")"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.onconnect = function () {}", FileName: "case.js", Output: []string{"globalThis.onconnect = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.onerror = function () {}", FileName: "case.js", Output: []string{"globalThis.onerror = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.onlanguagechange = function () {}", FileName: "case.js", Output: []string{"globalThis.onlanguagechange = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.onoffline = function () {}", FileName: "case.js", Output: []string{"globalThis.onoffline = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.ononline = function () {}", FileName: "case.js", Output: []string{"globalThis.ononline = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.addEventListener(\"message\", () => {})", FileName: "case.js", Output: []string{"globalThis.addEventListener(\"message\", () => {})"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.removeEventListener(\"message\", () => {})", FileName: "case.js", Output: []string{"globalThis.removeEventListener(\"message\", () => {})"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.dispatchEvent(new Event(\"message\"))", FileName: "case.js", Output: []string{"globalThis.dispatchEvent(new Event(\"message\"))"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.self", FileName: "case.js", Output: []string{"globalThis.self"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.onrejectionhandled = function () {}", FileName: "case.js", Output: []string{"globalThis.onrejectionhandled = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "self.onunhandledrejection = function () {}", FileName: "case.js", Output: []string{"globalThis.onunhandledrejection = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "'postMessage' in self", FileName: "case.js", Output: []string{"'postMessage' in globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 18, EndLine: 1, EndColumn: 22},
		}},
		{Code: "window.crypto", FileName: "case.js", Output: []string{"globalThis.crypto"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.addEventListener(\"play\", () => {})", FileName: "case.js", Output: []string{"globalThis.addEventListener(\"play\", () => {})"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.onplay = function () {}", FileName: "case.js", Output: []string{"globalThis.onplay = function () {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "function greet({ name = window.foo }) {}", FileName: "case.js", Output: []string{"function greet({ name = globalThis.foo }) {}"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 31},
		}},
		{Code: "({ foo: window.foo } =  {})", FileName: "case.js", Output: []string{"({ foo: globalThis.foo } =  {})"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
		}},
		{Code: "[window.foo] = []", FileName: "case.js", Output: []string{"[globalThis.foo] = []"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
		}},
		{Code: "foo[window]", FileName: "case.js", Output: []string{"foo[globalThis]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
		}},
		{Code: "foo[window.foo]", FileName: "case.js", Output: []string{"foo[globalThis.foo]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 5, EndLine: 1, EndColumn: 11},
		}},
		{Code: "'foo' in window", FileName: "case.js", Output: []string{"'foo' in globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
		}},
		{Code: "'foo' in global", FileName: "case.js", Output: []string{"'foo' in globalThis"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 10, EndLine: 1, EndColumn: 16},
		}},
		{Code: "typeof window.something === \"function\"", FileName: "case.js", Output: []string{"typeof globalThis.something === \"function\""}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
		}},
		{Code: "typeof self.something === \"function\"", FileName: "case.js", Output: []string{"typeof globalThis.something === \"function\""}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 12},
		}},
		{Code: "typeof global.something === \"function\"", FileName: "case.js", Output: []string{"typeof globalThis.something === \"function\""}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 8, EndLine: 1, EndColumn: 14},
		}},
		{Code: "self[foo]", FileName: "case.js", Output: []string{"globalThis[foo]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "global[foo]", FileName: "case.js", Output: []string{"globalThis[foo]"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "const foo: typeof window = window;", FileName: "case.ts", Output: []string{"const foo: typeof window = globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 28, EndLine: 1, EndColumn: 34},
		}},
		{Code: "const foo = {window};", FileName: "case.js", Output: []string{"const foo = {window: globalThis};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
		}},
		{Code: "const foo = {window, self};", FileName: "case.js", Output: []string{"const foo = {window: globalThis, self: globalThis};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 22, EndLine: 1, EndColumn: 26},
		}},
		{Code: "foo({window});", FileName: "case.js", Output: []string{"foo({window: globalThis});"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
		}},
		{Code: "const foo = {[window]: 1};", FileName: "case.js", Output: []string{"const foo = {[globalThis]: 1};"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
		}},
		{Code: "({window} = foo);", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
		}},
		{Code: "({window = 1} = foo);", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
		}},
	})
}

func TestPreferGlobalThisGlobalsOff(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_this.PreferGlobalThisRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "global.global_did_not_declare_in_language_options", FileName: "case.js", Globals: map[string]any{"global": "off", "window": "off", "self": "off"}, Output: []string{"globalThis.global_did_not_declare_in_language_options"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.window_did_not_declare_in_language_options", FileName: "case.js", Globals: map[string]any{"global": "off", "window": "off", "self": "off"}, Output: []string{"globalThis.window_did_not_declare_in_language_options"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "self.self_did_not_declare_in_language_options", FileName: "case.js", Globals: map[string]any{"global": "off", "window": "off", "self": "off"}, Output: []string{"globalThis.self_did_not_declare_in_language_options"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
	})
}

func TestPreferGlobalThisWrites(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_this.PreferGlobalThisRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
		{Code: "window = 123;", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "self = 1;", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 5},
		}},
		{Code: "window += 1;", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "[window] = foo;", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 2, EndLine: 1, EndColumn: 8},
		}},
		{Code: "({window} = foo);", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 3, EndLine: 1, EndColumn: 9},
		}},
		{Code: "for (global of foo);", FileName: "case.js", Output: nil, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 6, EndLine: 1, EndColumn: 12},
		}},
		{Code: "window.foo = window;", FileName: "case.js", Output: []string{"globalThis.foo = globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 20},
		}},
	})
}

func TestPreferGlobalThisTypeScript(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_this.PreferGlobalThisRule, []rule_tester.ValidTestCase{
		{Code: "typeof (window as unknown);", FileName: "case.ts"},
		{Code: "typeof (window satisfies unknown);", FileName: "case.ts"},
		{Code: "typeof (<unknown>window);", FileName: "case.ts"},
		{Code: "typeof window!;", FileName: "case.ts"},
		{Code: "typeof (self as unknown);", FileName: "case.ts"},
		{Code: "typeof (self satisfies unknown);", FileName: "case.ts"},
		{Code: "typeof (<unknown>self);", FileName: "case.ts"},
		{Code: "typeof self!;", FileName: "case.ts"},
		{Code: "typeof (global as unknown);", FileName: "case.ts"},
		{Code: "typeof (global satisfies unknown);", FileName: "case.ts"},
		{Code: "typeof (<unknown>global);", FileName: "case.ts"},
		{Code: "typeof global!;", FileName: "case.ts"},
	}, []rule_tester.InvalidTestCase{
		{Code: "const value = (window as unknown);", FileName: "case.ts", Output: []string{"const value = (globalThis as unknown);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 22},
		}},
		{Code: "typeof (window as unknown).value;", FileName: "case.ts", Output: []string{"typeof (globalThis as unknown).value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
		}},
		{Code: "const value = (self as unknown);", FileName: "case.ts", Output: []string{"const value = (globalThis as unknown);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 20},
		}},
		{Code: "typeof (self as unknown).value;", FileName: "case.ts", Output: []string{"typeof (globalThis as unknown).value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `self`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 13},
		}},
		{Code: "const value = (global as unknown);", FileName: "case.ts", Output: []string{"const value = (globalThis as unknown);"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 16, EndLine: 1, EndColumn: 22},
		}},
		{Code: "typeof (global as unknown).value;", FileName: "case.ts", Output: []string{"typeof (globalThis as unknown).value;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 9, EndLine: 1, EndColumn: 15},
		}},
	})
}

func TestPreferGlobalThisDocumentation(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_global_this.PreferGlobalThisRule, []rule_tester.ValidTestCase{
		{Code: "globalThis;", FileName: "case.js"},
		{Code: "globalThis.foo;", FileName: "case.js"},
		{Code: "globalThis;", FileName: "case.js"},
		{Code: "globalThis.foo;", FileName: "case.js"},
		{Code: "const {foo} = globalThis;", FileName: "case.js"},
		{Code: "globalThis.navigator;", FileName: "case.js"},
		{Code: "globalThis.location;", FileName: "case.js"},
		{Code: "window.innerWidth;", FileName: "case.js"},
		{Code: "window.innerHeight;", FileName: "case.js"},
		{Code: "typeof window === 'undefined';", FileName: "case.js"},
		{Code: "typeof window === 'object';", FileName: "case.js"},
		{Code: "globalThis.addEventListener('click', () => {});", FileName: "case.js"},
		{Code: "window.addEventListener('resize', () => {});", FileName: "case.js"},
		{Code: "window.addEventListener('load', () => {});", FileName: "case.js"},
		{Code: "window.addEventListener('unload', () => {});", FileName: "case.js"},
		{Code: "window[foo]", FileName: "case.js"},
		{Code: "typeof window === 'undefined'", FileName: "case.js"},
	}, []rule_tester.InvalidTestCase{
		{Code: "window;", FileName: "case.js", Output: []string{"globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.foo;", FileName: "case.js", Output: []string{"globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "global;", FileName: "case.js", Output: []string{"globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "global.foo;", FileName: "case.js", Output: []string{"globalThis.foo;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `global`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "const {foo} = window;", FileName: "case.js", Output: []string{"const {foo} = globalThis;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 15, EndLine: 1, EndColumn: 21},
		}},
		{Code: "window.navigator;", FileName: "case.js", Output: []string{"globalThis.navigator;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.location;", FileName: "case.js", Output: []string{"globalThis.location;"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
		{Code: "window.addEventListener('click', () => {});", FileName: "case.js", Output: []string{"globalThis.addEventListener('click', () => {});"}, Errors: []rule_tester.InvalidTestCaseError{
			{MessageId: "prefer-global-this/error", Message: "Prefer `globalThis` over `window`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 7},
		}},
	})
}
