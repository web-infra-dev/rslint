// Additional AST and naming cases checked against eslint-plugin-unicorn v76.0.0.
// Documented suggestion safety differences keep generated code valid.
// cspell:ignore ßeta İstanbul i̇stanbul İstanbul
package no_anonymous_default_export_test

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_anonymous_default_export"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoAnonymousDefaultExportExtras(t *testing.T) {
	rule_tester.RunRuleTester(filenameTestRoot(), "filename-tests.json", t, &no_anonymous_default_export.NoAnonymousDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{Code: "export default (function named() {});", FileName: "/path/to/foo.ts"},
			{Code: "export default ((class Named {}));", FileName: "/path/to/foo.ts"},
			{Code: "export default { method() {}, field: () => {} };", FileName: "/path/to/foo.ts"},
			{Code: "export { named as default }; function named() {}", FileName: "/path/to/foo.ts"},
			{Code: "module['exports'] = () => {};", FileName: "/path/to/foo.ts"},
			{Code: "module.exports.extra = () => {};", FileName: "/path/to/foo.ts"},
			{Code: "exports.extra = () => {};", FileName: "/path/to/foo.ts"},
			{Code: "module.exports = function named() {};", FileName: "/path/to/foo.ts"},
			{Code: "exports = class Named {};", FileName: "/path/to/foo.ts"},
			{Code: "foo = (exports = () => {});", FileName: "/path/to/foo.ts"},
			{Code: "(exports = () => {}, done());", FileName: "/path/to/foo.ts"},
			{Code: "void (exports = () => {});", FileName: "/path/to/foo.ts"},
			{Code: "for (exports = () => {}; false;) {}", FileName: "/path/to/foo.ts"},
			{Code: "export default (() => {}) as Function;", FileName: "/path/to/foo.ts"},
			{Code: "export default (() => {}) satisfies Function;", FileName: "/path/to/foo.ts"},
			{Code: "export default (() => {})!;", FileName: "/path/to/foo.ts"},
			{Code: "export = () => {};", FileName: "/path/to/foo.ts"},
			{Code: "export default function named<T>(value: T): T { return value; }", FileName: "/path/to/foo.ts"},
			{Code: "export default function (value: string): void;", FileName: "/path/to/foo.ts"},
			{Code: "class Holder { #exports; set() { this.#exports = () => {}; } }", FileName: "/path/to/foo.ts"},
		}, []rule_tester.InvalidTestCase{
			// CommonJS supplies an implicit exports binding; scripts do not.
			{Code: "module.exports = () => {};", FileName: "/path/to/exports.cjs", LanguageOptions: rule.LanguageOptions{SourceType: "commonjs"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const exports_ = () => {};\nmodule.exports = exports_;"}}},
			}},
			{Code: "module.exports = () => {};", FileName: "/path/to/exports.cjs", LanguageOptions: rule.LanguageOptions{SourceType: "script"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const exports = () => {};\nmodule.exports = exports;"}}},
			}},
			// Suggestion safety: documented differences from v76.0.0
			{Code: "export default function<T extends (() => void)>() {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo<T extends (() => void)>() {}"}}},
			}},
			{Code: "module.exports = async function* /* name */ <T /* > */>(value: T) { yield value; };", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "module.exports = async function* /* name */ foo<T /* > */>(value: T) { yield value; };"}}},
			}},
			{Code: "if (ready) module.exports = () => {}; else fallback();", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 32, EndLine: 1, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "while (ready) exports = () => {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 28, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "for (const item of items) exports = () => item;", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 40, EndLine: 1, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "label: exports = () => {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// Additional scope and header boundaries
			{Code: "if (ready) { module.exports = () => {}; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "if (ready) { const foo = () => {};\nmodule.exports = foo; }"}}},
			}},
			{Code: "switch (value) { case 1: exports = () => {}; break; default: exports = async () => {}; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "switch (value) { case 1: const foo = () => {};\nexports = foo; break; default: exports = async () => {}; }"}}},
				{MessageId: "no-anonymous-default-export/error", Message: "The async arrow function should be named.", Line: 1, Column: 81, EndLine: 1, EndColumn: 83, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "switch (value) { case 1: exports = () => {}; break; default: const foo = async () => {};\nexports = foo; }"}}},
			}},
			{Code: "namespace Holder { exports = () => {}; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "namespace Holder { const foo = () => {};\nexports = foo; }"}}},
			}},
			{Code: "if (ready) module.exports = function() {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 29, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "if (ready) module.exports = function foo () {};"}}},
			}},
			{Code: "export default class extends Foo {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo_ extends Foo {}"}}},
			}},
			{Code: "export default class<T> {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo<T> {}"}}},
			}},
			{Code: "export default @decorate(class {}) abstract class {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default @decorate(class {}) abstract class Foo {}"}}},
			}},
			{Code: "export default class {static { Foo; }}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo_ {static { Foo; }}"}}},
			}},
			{Code: "export default () => {}; function other() { var foo; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nexport default foo; function other() { var foo; }"}}},
			}},
			{Code: "const f\\u006fo = 1; export default () => {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 39, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const f\\u006fo = 1; const foo_ = () => {};\nexport default foo_;"}}},
			}},
			{Code: "export default function(foo = foo_) { return foo__; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo___ (foo = foo_) { return foo__; }"}}},
			}},
			{Code: "/* global foo: readonly */ export default () => {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 46, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "/* global foo: readonly */ const foo_ = () => {};\nexport default foo_;"}}},
			}},
			{Code: "/* global foo: off */ export default () => {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 41, EndLine: 1, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "/* global foo: off */ const foo = () => {};\nexport default foo;"}}},
			}},
			{Code: "export default class extends (factory({key: /[{}]/, value: `x${1}`})) {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 70, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo extends (factory({key: /[{}]/, value: `x${1}`})) {}"}}},
			}},
			// AST and scopes
			{Code: "export default (function() {});", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 17, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default (function foo () {});"}}},
			}},
			{Code: "export default ((class {}));", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default ((class Foo {}));"}}},
			}},
			{Code: "(module).exports = (() => {});", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 24, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (() => {});\n(module).exports = foo;"}}},
			}},
			{Code: "(exports) = () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\n(exports) = foo;"}}},
			}},
			{Code: "exports ||= () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nexports ||= foo;"}}},
			}},
			{Code: "module.exports += function() {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "module.exports += function foo () {};"}}},
			}},
			{Code: "function wrapper(module) { module.exports = () => {}; }", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 48, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "function wrapper(module) { const foo = () => {};\nmodule.exports = foo; }"}}},
			}},
			{Code: "export default () => { let foo; return foo; };", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo_ = () => { let foo; return foo; };\nexport default foo_;"}}},
			}},
			{Code: "export default () => { return () => foo; };", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo_ = () => { return () => foo; };\nexport default foo_;"}}},
			}},
			{Code: "export default function () { return foo; }", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo_ () { return foo; }"}}},
			}},
			{Code: "const foo = 1; export default function () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 31, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = 1; export default function foo_ () {}"}}},
			}},
			{Code: "function sibling() { let foo; } export default () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 51, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "function sibling() { let foo; } const foo = () => {};\nexport default foo;"}}},
			}},
			{Code: "export default () => { function sibling(foo) {} };", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo_ = () => { function sibling(foo) {} };\nexport default foo_;"}}},
			}},
			{Code: "export default function (foo = 1) { let foo_; }", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo__ (foo = 1) { let foo_; }"}}},
			}},
			{Code: "export default class { foo() { return foo; } #foo; }", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo { foo() { return foo; } #foo; }"}}},
			}},
			{Code: "export default class { method(Foo) {} }", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo_ { method(Foo) {} }"}}},
			}},
			{Code: "export default class extends (class {}) { [foo]() {} }", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo extends (class {}) { [foo]() {} }"}}},
			}},
			{Code: "export default class /* keep */ extends factory({}) {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 52, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo /* keep */ extends factory({}) {}"}}},
			}},
			{Code: "const emoji = \"😀\"; export default async (foo) => <span>{foo}</span>;", FileName: "/path/to/foo.tsx", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async arrow function should be named.", Line: 1, Column: 48, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const emoji = \"😀\"; const foo_ = async (foo) => <span>{foo}</span>;\nexport default foo_;"}}},
			}},
			{Code: "export default function\t() {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function\t foo () {}"}}},
			}},
			{Code: "export default function/* (comment) */() {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function/* (comment) */ foo () {}"}}},
			}},
			{Code: "export default () => /* keep */ ({}); // after", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => /* keep */ ({});\nexport default foo; // after"}}},
			}},
			{Code: "export default /* before */ ((() => {}) /* inner */);", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 34, EndLine: 1, EndColumn: 36, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = ((() => {}) /* inner */);\nexport default /* before */ foo;"}}},
			}},
			{Code: "module.exports = () => {} /* after */;", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nmodule.exports = foo /* after */;"}}},
			}},
			{Code: "// CR\rexport default () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 2, Column: 19, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "// CR\rconst foo = () => {};\rexport default foo;"}}},
			}},
			{Code: "// LS export default () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 2, Column: 19, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "// LS const foo = () => {}; export default foo;"}}},
			}},
			{Code: "// PS export default () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 2, Column: 19, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "// PS const foo = () => {}; export default foo;"}}},
			}},
			// TypeScript
			{Code: "export default <T>(value: T): T => value;", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 33, EndLine: 1, EndColumn: 35, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = <T>(value: T): T => value;\nexport default foo;"}}},
			}},
			{Code: "export default function<T>(value: T): T { return value; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo<T>(value: T): T { return value; }"}}},
			}},
			{Code: "export default class<T> { value?: T; }", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo<T> { value?: T; }"}}},
			}},
			{Code: "@decorate(class {}) export default class {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 36, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "@decorate(class {}) export default class Foo {}"}}},
			}},
			{Code: "export default @decorate(class {}) class {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 41, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default @decorate(class {}) class Foo {}"}}},
			}},
			{Code: "export default abstract class {}", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 30, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default abstract class Foo {}"}}},
			}},
			{Code: "type foo = string; export default () => {};", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 38, EndLine: 1, EndColumn: 40, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "type foo = string; const foo_ = () => {};\nexport default foo_;"}}},
			}},
			{Code: "export default () => { type foo = string; };", FileName: "/path/to/foo.ts", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo_ = () => { type foo = string; };\nexport default foo_;"}}},
			}},
			// JSDoc wrappers
			{Code: "export default /** @type {Function} */ (() => {});", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 44, EndLine: 1, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (() => {});\nexport default /** @type {Function} */ foo;"}}},
			}},
			{Code: "export default (/** @type {Function} */ (() => {}));", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 45, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (/** @type {Function} */ (() => {}));\nexport default foo;"}}},
			}},
			{Code: "module.exports = /** @type {Function} */ (() => {});", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 46, EndLine: 1, EndColumn: 48, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (() => {});\nmodule.exports = /** @type {Function} */ foo;"}}},
			}},
			// Filename conversion
			{Code: "export default () => {};", FileName: "/path/to/.hidden.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "export default class {}", FileName: "/path/to/.hidden.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/123.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "export default class {}", FileName: "/path/to/123.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/---.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "export default class {}", FileName: "/path/to/---.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/class.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const class_ = () => {};\nexport default class_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/class.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Class {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const module_ = () => {};\nexport default module_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Module {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/arguments.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const arguments_ = () => {};\nexport default arguments_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/arguments.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Arguments {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/await.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const await_ = () => {};\nexport default await_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/await.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Await {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/eval.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const eval_ = () => {};\nexport default eval_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/eval.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Eval {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/undefined.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const undefined_ = () => {};\nexport default undefined_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/undefined.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Undefined {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/global-this.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const globalThis_ = () => {};\nexport default globalThis_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/global-this.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class GlobalThis {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/array.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const array = () => {};\nexport default array;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/array.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Array_ {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/foo-123-bar.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo_123Bar = () => {};\nexport default foo_123Bar;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo-123-bar.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo_123Bar {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/XMLHttp.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const xmlHttp = () => {};\nexport default xmlHttp;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/XMLHttp.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class XmlHttp {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/ßeta.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const ßeta = () => {};\nexport default ßeta;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/ßeta.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class SSeta {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/İstanbul.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const i̇stanbul = () => {};\nexport default i̇stanbul;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/İstanbul.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class İstanbul {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/中文-test.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const 中文Test = () => {};\nexport default 中文Test;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/中文-test.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class 中文Test {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/foo-𐐀.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo𐐀 = () => {};\nexport default foo𐐀;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo-𐐀.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo𐐀 {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/𐐀-test.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const 𐐨Test = () => {};\nexport default 𐐨Test;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/𐐀-test.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class 𐐨Test {}"}}},
			}},
		})
}

func TestNoAnonymousDefaultExportEditDemand(t *testing.T) {
	t.Parallel()
	const source = "export default () => {};"
	helper := rule_tester.NewProgramHelper(fixtures.GetRootDir())
	program, sourceFile, err := helper.CreateTestProgram(source, "foo.js", "tsconfig.json")
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := make(map[rule.EditDemand]rule.RuleDiagnostic, 4)
	for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion, rule.EditDemandAll} {
		var got []rule.RuleDiagnostic
		linter.LintSingleFile(linter.LintSingleFileOptions{
			Program: lintprogram.NewFromCompiler(program), File: sourceFile.FileName(),
			GetRulesForFile: func(*ast.SourceFile) []rule.ConfiguredRule {
				return []rule.ConfiguredRule{{Name: no_anonymous_default_export.NoAnonymousDefaultExportRule.Name, Severity: rule.SeverityError,
					Run: func(ctx rule.RuleContext) rule.RuleListeners {
						return no_anonymous_default_export.NoAnonymousDefaultExportRule.Run(ctx, nil)
					},
				}}
			},
			Consumer: rule.DiagnosticConsumer{Demand: demand, Report: func(d rule.RuleDiagnostic) { got = append(got, d) }},
		})
		if len(got) != 1 {
			t.Fatalf("demand %d: got %d diagnostics", demand, len(got))
		}
		diagnostics[demand] = got[0]
		output, _, fixed := linter.ApplyRuleFixes(source, got)
		if fixed || output != source {
			t.Fatalf("demand %d: unexpected autofix %q", demand, output)
		}
	}
	base := diagnostics[rule.EditDemandNone]
	for demand, d := range diagnostics {
		if d.Range != base.Range || !reflect.DeepEqual(d.Message, base.Message) {
			t.Errorf("demand %d changed diagnostic identity", demand)
		}
		if d.FixesPtr != nil {
			t.Errorf("demand %d produced an autofix", demand)
		}
	}
	if diagnostics[rule.EditDemandNone].Suggestions != nil || diagnostics[rule.EditDemandAutofix].Suggestions != nil {
		t.Fatal("suggestions produced without demand")
	}
	suggestions := diagnostics[rule.EditDemandAll].Suggestions
	if suggestions == nil || len(*suggestions) != 1 || !reflect.DeepEqual(suggestions, diagnostics[rule.EditDemandSuggestion].Suggestions) {
		t.Fatal("inconsistent suggestion artifacts")
	}
	if (*suggestions)[0].Message.Description != "Name it as `foo`." {
		t.Fatal("wrong suggestion message")
	}
	output, _, fixed := linter.ApplyRuleFixes(source, *suggestions)
	if !fixed || output != "const foo = () => {};\nexport default foo;" {
		t.Fatalf("unexpected suggestion output %q", output)
	}
}
