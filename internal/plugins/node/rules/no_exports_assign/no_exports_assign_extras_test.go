package no_exports_assign

import (
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/typescript/rules/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// These additional cases were checked against eslint-plugin-n v18.3.0 with
// ESLint 10.9.0 and @typescript-eslint/parser 8.65.0.

// Globals and source types.
func TestNoExportsAssignGlobals(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "exports = {}"},
			{Code: "exports = {}", Globals: map[string]any{"exports": "off", "module": "readonly"}},
			{Code: "/* global exports: off */ exports = {}", Globals: exportsGlobals},
			{Code: "/* global module: readonly */ exports = module.exports = {}", Globals: map[string]any{"exports": "writable"}},
			{Code: "var exports; exports = {}", Globals: exportsGlobals},
			{Code: "var exports; exports = {}", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}},
			// tsgo's implicit exports symbol is not an authored local binding.
			{Code: "exports = module.exports = factory(); exports.foo = 1;", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}},
			{Code: "var module; exports = module.exports = {}", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "exports = {}", Globals: map[string]any{"exports": "readonly"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			{Code: "/* global exports: writable */ exports = {}", Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 32, 1, 44)}},
			{Code: "exports = module.exports = {}", Globals: map[string]any{"exports": "writable"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 30)}},
			{Code: "module.exports = exports = {}", Globals: map[string]any{"exports": "writable", "module": "off"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 18, 1, 30)}},
			{Code: "var exports; exports = {}", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 14, 1, 26)}},
			{Code: "let exports; exports = {}", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: map[string]any{"exports": "off"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 14, 1, 26)}},
			{Code: "exports = {}", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			{Code: "exports = {}; module.exports = {};", FileName: "input.cjs", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 13)}},
			{Code: "/** @typedef {object} exports */ exports = {};", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 34, 1, 46)}},
			{Code: "var exports; exports = {}", FileName: "input.cts", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 14, 1, 26)}},
			{Code: "var module; exports = module.exports = {}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 42)}},
		},
	)
}

// Lexical scope ownership, including upstream syntactic lookup.
func TestNoExportsAssignScopes(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "{ let exports; exports = {}; }", Globals: exportsGlobals},
			{Code: "function f() { exports = {}; var exports; }", Globals: exportsGlobals},
			{Code: "function f(value = exports = {}) { var exports; }", Globals: exportsGlobals},
			{Code: "const f = (exports) => { exports = {}; };", Globals: exportsGlobals},
			{Code: "const f = function exports() { exports = {}; };", Globals: exportsGlobals},
			{Code: "try {} catch (exports) { exports = {}; }", Globals: exportsGlobals},
			{Code: "class exports { static value = (exports = {}); }", Globals: exportsGlobals},
			{Code: "class Example { static { let exports; exports = {}; } }", Globals: exportsGlobals},
			{Code: "function f() { interface exports {} exports = {}; }", Globals: exportsGlobals},
			{Code: "import exports from \"pkg\"; exports = {};", Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "function f(module) { exports = module.exports = {}; }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 22, 1, 51)}},
			{Code: "function f(module) { module.exports = exports = {}; }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 39, 1, 51)}},
			{Code: "class Example { [exports = {}](exports) {} }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 18, 1, 30)}},
			{Code: "function f() { type module = object; exports = module.exports = {}; }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 38, 1, 67)}},
		},
	)
}

// ESTree wrappers and chained assignments.
func TestNoExportsAssignExpressions(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "module.exports = ((exports) = {});", Globals: exportsGlobals},
			{Code: "(exports) = ((module).exports = {});", Globals: exportsGlobals},
			{Code: "exports = ((module.exports) = {});", Globals: exportsGlobals},
			{Code: "module.exports += exports ||= {};", Globals: exportsGlobals},
			{Code: "exports &&= module.exports ||= {};", Globals: exportsGlobals},
			{Code: "this.exports = {}; exports[\"x\"] = {};", Globals: exportsGlobals},
			{Code: "(exports as any) = {}; exports! = {}; (exports satisfies object) = {};", Globals: exportsGlobals},
			{Code: "module.exports = /** @type {object} */ (exports = {});", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "  ((exports) = {});", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 4, 1, 18)}},
			{Code: "exports = (0, module.exports = {});", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 35)}},
			{Code: "module.exports = (0, exports = {});", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 22, 1, 34)}},
			{Code: "exports = exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 23), forbiddenAt(1, 11, 1, 23)}},
			{Code: "module.exports = exports = exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 28, 1, 40)}},
			{Code: "exports = module[\"exports\"] = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 33)}},
			{Code: "module[\"exports\"] = exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 21, 1, 33)}},
			{Code: "exports = other.exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 29)}},
			{Code: "module.foo = exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 14, 1, 26)}},
			{Code: "exports = module.exports;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 25)}},
			{Code: "exports = module?.exports;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 26)}},
			{Code: "class Example { #exports; f() { exports = this.#exports = {}; } }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 33, 1, 61)}},
			{Code: "exports = (module.exports = {}) as object;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 42)}},
			{Code: "module.exports = ((exports = {}) as object);", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 20, 1, 32)}},
			{Code: "(module as any).exports = exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 27, 1, 39)}},
			{Code: "/** @type {object} */ (exports) = {};", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 23, 1, 37)}},
		},
	)
}

// Assignment patterns and nested runtime assignments.
func TestNoExportsAssignPatterns(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "exports++; --exports; [exports] = values; ({exports} = value);", Globals: exportsGlobals},
			{Code: "[exports = {}] = values; ({value: exports = {}} = input);", Globals: exportsGlobals},
			{Code: "for ([exports = {}] of values) {}", Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "({ [exports = {}]: target } = input);", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 5, 1, 17)}},
			{Code: "[target = (exports = {})] = values;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 12, 1, 24)}},
			// The enclosing default is an AssignmentPattern, not an allowed
			// module.exports AssignmentExpression.
			{Code: "[module.exports = (exports = {})] = values;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 20, 1, 32)}},
			{Code: "({value: module.exports = exports = {}} = input);", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 27, 1, 39)}},
			{Code: "({ target = exports = {} } = input);", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 13, 1, 25)}},
			{Code: "const view = <Widget value={(exports = {})} />;", FileName: "input.tsx", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 30, 1, 42)}},
		},
	)
}

// Multiline and UTF-16 diagnostic ranges.
func TestNoExportsAssignRanges(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "// eslint-disable-next-line\nexports = {};", Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "/* 😀 */\n  exports = {\n    café: \"😀\"\n  };", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(2, 3, 4, 4)}},
			{Code: "const 文 = \"😀\"; exports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 17, 1, 29)}},
			{Code: "/* eslint-disable */\nexports = {};\n/* eslint-enable */\nexports = {};", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(4, 1, 4, 13)}},
		},
	)
}

// Every assignment operator.
func TestNoExportsAssignOperators(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{},
		[]rule_tester.InvalidTestCase{
			{Code: "exports = value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 16)}},
			{Code: "exports += value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports -= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports *= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports /= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports %= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports **= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
			{Code: "exports <<= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
			{Code: "exports >>= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
			{Code: "exports >>>= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 19)}},
			{Code: "exports |= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports ^= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports &= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 17)}},
			{Code: "exports &&= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
			{Code: "exports ||= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
			{Code: "exports ??= value;", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
		},
	)
}

// Escaped names, TypeScript declaration spaces, decorators and runtime
// assignments nested in patterns were compared with the pinned upstream rule.
func TestNoExportsAssignAdditionalSyntax(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoExportsAssignRule,
		[]rule_tester.ValidTestCase{
			{Code: "exports = " + strings.ReplaceAll("module.exports", "o", "\\u006f") + " = {};", Globals: exportsGlobals},
			{Code: "for (let exports of values) { exports = {}; }", Globals: exportsGlobals},
			{Code: "namespace A { let exports; exports = {}; }", Globals: exportsGlobals},
			{Code: "enum Example { exports = (exports = {}) }", Globals: exportsGlobals},
			{Code: "class Example<exports> { method() { exports = {}; } }", Globals: exportsGlobals},
			{Code: "function f<exports>() { exports = {}; }", Globals: exportsGlobals},
			{Code: "export {}; declare global { var exports: object; } exports = {};"},
			{Code: "class Example { method(@decorate(exports = {}) exports) {} }", Globals: exportsGlobals},
		},
		[]rule_tester.InvalidTestCase{
			{Code: strings.ReplaceAll("exports = {};", "o", "\\u006f"), Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 1, 1, 18)}},
			{Code: "{ var exports; } exports = {};", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 18, 1, 30)}},
			{Code: "function exports() { exports = {}; }", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 22, 1, 34)}},
			{Code: "namespace A.B { exports = {}; }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 17, 1, 29)}},
			{Code: "namespace A { export const module = {}; exports = module.exports = {}; }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 41, 1, 70)}},
			{Code: "enum Example { Value = (exports = {}) }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 25, 1, 37)}},
			{Code: "@decorate(exports = {}) class Example {}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 11, 1, 23)}},
			{Code: "class Example { @decorate(exports = {}) method(exports) {} }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 27, 1, 39)}},
			{Code: "with (object) { exports = {}; }", FileName: "input.js", TSConfig: "tsconfig.allowJs.json", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 17, 1, 29)}},
			{Code: "[exports] = [(exports = {})];", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 15, 1, 27)}},
			{Code: "function f({ value = (exports = {}) }) {}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 23, 1, 35)}},
			{Code: "for (exports = {};;) { break; }", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 6, 1, 18)}},
			{Code: "for ([module.exports = exports = {}] of values) {}", Globals: exportsGlobals, Errors: []rule_tester.InvalidTestCaseError{forbiddenAt(1, 24, 1, 36)}},
		},
	)
}
