// Upstream tests and documentation from eslint-plugin-unicorn v76.0.0 (MIT).
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v76.0.0/test/no-anonymous-default-export.js
package no_anonymous_default_export_test

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_anonymous_default_export"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
	"github.com/web-infra-dev/rslint/internal/testutil/embedfs"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func TestNoAnonymousDefaultExportUpstream(t *testing.T) {
	rule_tester.RunRuleTester(filenameTestRoot(), "filename-tests.json", t, &no_anonymous_default_export.NoAnonymousDefaultExportRule,
		[]rule_tester.ValidTestCase{
			{Code: "export default function named() {}", FileName: "/path/to/case.js"},
			{Code: "export default class named {}", FileName: "/path/to/case.js"},
			{Code: "export default []", FileName: "/path/to/case.js"},
			{Code: "export default {}", FileName: "/path/to/case.js"},
			{Code: "export default 1", FileName: "/path/to/case.js"},
			{Code: "export default false", FileName: "/path/to/case.js"},
			{Code: "export default 0n", FileName: "/path/to/case.js"},
			{Code: "notExports = class {}", FileName: "/path/to/case.js"},
			{Code: "notModule.exports = class {}", FileName: "/path/to/case.js"},
			{Code: "module.notExports = class {}", FileName: "/path/to/case.js"},
			{Code: "module.exports.foo = class {}", FileName: "/path/to/case.js"},
			{Code: "alert(exports = class {})", FileName: "/path/to/case.js"},
			{Code: "foo = module.exports = class {}", FileName: "/path/to/case.js"},
			{Code: "export default class Foo {}", FileName: "/path/to/foo.js"},
			{Code: "export default function foo() {}", FileName: "/path/to/foo.js"},
			{Code: "const foo = () => {};\nexport default foo;", FileName: "/path/to/foo.js"},
			{Code: "module.exports = class Foo {};", FileName: "/path/to/foo.js"},
			{Code: "module.exports = function foo() {};", FileName: "/path/to/foo.js"},
			{Code: "const foo = () => {};\nmodule.exports = foo;", FileName: "/path/to/foo.js"},
		}, []rule_tester.InvalidTestCase{
			// Upstream snapshot cases
			{Code: "export default function () {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function case_ () {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Case {}"}}},
			}},
			{Code: "export default () => {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const case_ = () => {};\nexport default case_;"}}},
			}},
			{Code: "export default function * () {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function * case_ () {}"}}},
			}},
			{Code: "export default async function () {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function case_ () {}"}}},
			}},
			{Code: "export default async function * () {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function * case_ () {}"}}},
			}},
			{Code: "export default async () => {}", FileName: "/path/to/case.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async arrow function should be named.", Line: 1, Column: 25, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const case_ = async () => {};\nexport default case_;"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default class extends class {} {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 38, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo extends class {} {}"}}},
			}},
			{Code: "export default class{}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo{}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo-bar.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class FooBar {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo_bar.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class FooBar {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo+bar.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class FooBar {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo+bar123.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class FooBar123 {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo*.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/[foo].js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/class.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Class {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo.helper.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo.bar.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/foo.test.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default class {}", FileName: "/path/to/.foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "let Foo, Foo_, foo, foo_\nexport default class {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 2, Column: 16, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Foo, Foo_, foo, foo_\nexport default class Foo__ {}"}}},
			}},
			{Code: "let Foo, Foo_, foo, foo_\nexport default (class{})", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 2, Column: 17, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Foo, Foo_, foo, foo_\nexport default (class Foo__{})"}}},
			}},
			{Code: "export default (class extends class {} {})", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 17, EndLine: 1, EndColumn: 39, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default (class Foo extends class {} {})"}}},
			}},
			{Code: "let Exports, Exports_, exports, exports_\nexports = class {}", FileName: "/path/to/exports.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 2, Column: 11, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Exports, Exports_, exports, exports_\nexports = class Exports__ {}"}}},
			}},
			{Code: "module.exports = class {}", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "module.exports = class Module {}"}}},
			}},
			{Code: "module.exports = () => {}", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const module_ = () => {};\nmodule.exports = module_;"}}},
			}},
			{Code: "exports = () => {}", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 14, EndLine: 1, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const module_ = () => {};\nexports = module_;"}}},
			}},
			{Code: "export default function () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo () {}"}}},
			}},
			{Code: "export default function* () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function* foo () {}"}}},
			}},
			{Code: "export default async function* () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function* foo () {}"}}},
			}},
			{Code: "export default async function*() {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 31, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function* foo () {}"}}},
			}},
			{Code: "export default async function *() {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 32, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function * foo () {}"}}},
			}},
			{Code: "export default async function   *   () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 37, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function   *   foo () {}"}}},
			}},
			{Code: "export default async function * /* comment */ () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function * /* comment */ foo () {}"}}},
			}},
			{Code: "export default async function * // comment\n() {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 1, Column: 16, EndLine: 2, EndColumn: 1, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default async function * // comment\n foo () {}"}}},
			}},
			{Code: "let Foo, Foo_, foo, foo_\nexport default async function * () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 2, Column: 16, EndLine: 2, EndColumn: 33, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Foo, Foo_, foo, foo_\nexport default async function * foo__ () {}"}}},
			}},
			{Code: "let Foo, Foo_, foo, foo_\nexport default (async function * () {})", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async generator function should be named.", Line: 2, Column: 17, EndLine: 2, EndColumn: 34, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Foo, Foo_, foo, foo_\nexport default (async function * foo__ () {})"}}},
			}},
			{Code: "let Exports, Exports_, exports, exports_\nexports = function() {}", FileName: "/path/to/exports.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 2, Column: 11, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Exports, Exports_, exports, exports_\nexports = function exports__ () {}"}}},
			}},
			{Code: "module.exports = function() {}", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 26, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "module.exports = function module_ () {}"}}},
			}},
			{Code: "export default () => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nexport default foo;"}}},
			}},
			{Code: "export default async () => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async arrow function should be named.", Line: 1, Column: 25, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = async () => {};\nexport default foo;"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nexport default foo;"}}},
			}},
			{Code: "export default() => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nexport default foo;"}}},
			}},
			{Code: "export default foo => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo_ = foo => {};\nexport default foo_;"}}},
			}},
			{Code: "export default (( () => {} ))", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 22, EndLine: 1, EndColumn: 24, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (( () => {} ));\nexport default foo;"}}},
			}},
			{Code: "/* comment 1 */ export /* comment 2 */ default /* comment 3 */  () => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 68, EndLine: 1, EndColumn: 70, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "/* comment 1 */ const foo = () => {};\nexport /* comment 2 */ default /* comment 3 */  foo;"}}},
			}},
			{Code: "// comment 1\nexport\n// comment 2\ndefault\n// comment 3\n() => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 6, Column: 4, EndLine: 6, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "// comment 1\nconst foo = () => {};\nexport\n// comment 2\ndefault\n// comment 3\nfoo;"}}},
			}},
			{Code: "let Foo, Foo_, foo, foo_\nexport default async () => {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The async arrow function should be named.", Line: 2, Column: 25, EndLine: 2, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Foo, Foo_, foo, foo_\nconst foo__ = async () => {};\nexport default foo__;"}}},
			}},
			{Code: "let Exports, Exports_, exports, exports_\nexports = (( () => {} ))", FileName: "/path/to/exports.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 2, Column: 17, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "let Exports, Exports_, exports, exports_\nconst exports__ = (( () => {} ));\nexports = exports__;"}}},
			}},
			{Code: "// comment 1\nmodule\n// comment 2\n.exports\n// comment 3\n=\n// comment 4\n() => {};", FileName: "/path/to/module.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 8, Column: 4, EndLine: 8, EndColumn: 6, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "// comment 1\nconst module_ = () => {};\nmodule\n// comment 2\n.exports\n// comment 3\n=\n// comment 4\nmodule_;"}}},
			}},
			{Code: "(( exports = (( () => {} )) ))", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (( () => {} ));\n(( exports = foo ));"}}},
			}},
			{Code: "(( module.exports = (( () => {} )) ))", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (( () => {} ));\n(( module.exports = foo ));"}}},
			}},
			{Code: "(( exports = (( () => {} )) ));", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 20, EndLine: 1, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (( () => {} ));\n(( exports = foo ));"}}},
			}},
			{Code: "(( module.exports = (( () => {} )) ));", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 27, EndLine: 1, EndColumn: 29, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = (( () => {} ));\n(( module.exports = foo ));"}}},
			}},
			// Upstream line endings
			{Code: "export default () => {\r\n\tbar();\r\n};\r\n", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {\r\n\tbar();\r\n};\r\nexport default foo;\r\n"}}},
			}},
			// Documentation examples
			{Code: "export default class {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default class Foo {}"}}},
			}},
			{Code: "export default function () {}", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 16, EndLine: 1, EndColumn: 25, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "export default function foo () {}"}}},
			}},
			{Code: "export default () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 19, EndLine: 1, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nexport default foo;"}}},
			}},
			{Code: "module.exports = class {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The class should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "module.exports = class Foo {};"}}},
			}},
			{Code: "module.exports = function () {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The function should be named.", Line: 1, Column: 18, EndLine: 1, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "module.exports = function foo () {};"}}},
			}},
			{Code: "module.exports = () => {};", FileName: "/path/to/foo.js", Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "no-anonymous-default-export/error", Message: "The arrow function should be named.", Line: 1, Column: 21, EndLine: 1, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "no-anonymous-default-export/suggestion", Output: "const foo = () => {};\nmodule.exports = foo;"}}},
			}},
		})
}

func filenameTestRoot() embedfs.Root {
	root := fixtures.GetRootDir()
	// TypeScript's default wildcard excludes dotfiles. Explicit inclusion keeps
	// upstream's no-suggestion case in the normal rule tester.
	root.FS = utils.NewOverlayVFS(root.FS, map[string]string{
		tspath.ResolvePath(root.Dir, "filename-tests.json"): `{"extends":"./tsconfig.json","include":["**/*","/path/to/.foo.js","/path/to/.hidden.js"]}`,
	})
	return root
}
