// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-string-raw.js
package prefer_string_raw_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_raw"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func preferStringRawWithoutTypes() rule.Rule {
	r := prefer_string_raw.PreferStringRawRule
	// RuleTester supplies a compiler. Hide it for syntax-only cases.
	run := r.Run
	r.Run = func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ctx.TypeChecker = nil
		return run(ctx, options)
	}
	return r
}

func TestPreferStringRawUpstreamSyntax(t *testing.T) {
	r := preferStringRawWithoutTypes()
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &r,
		[]rule_tester.ValidTestCase{
			// TypeScript literal and enum contexts.
			{Code: "type Path = `a\\\\b`;", FileName: "case.ts"},
			{Code: "enum Paths {File = `a\\\\b`}", FileName: "case.ts"},
			// String literals under const assertions.
			{Code: "const path = 'a\\\\b' as const;", FileName: "file.ts"},
			{Code: "const path = <const>'a\\\\b';", FileName: "file.ts"},
			// Template literals under const assertions.
			{Code: "const path = `a\\\\b` as const;", FileName: "file.ts"},
			{Code: "const path = <const>`a\\\\b`;", FileName: "file.ts"},
			// Snapshot cases for string literals.
			{Code: "a = '\\''", FileName: "case.js"},
			{Code: "a = \"\\\\a \\\n\tb\"", FileName: "case.js"},
			{Code: "a = 'a\\\\b\\u{51}c'", FileName: "case.js"},
			{Code: "a = \"a\\\\b`\"", FileName: "case.js"},
			{Code: "a = \"a\\\\b${foo}\"", FileName: "case.js"},
			{Code: "a = \"a\\\\b\\tc\"", FileName: "case.js"},
			{Code: "a = \"a\\\\b\\nc\"", FileName: "case.js"},
			{Code: "a = '\\\\'", FileName: "case.js"},
			{Code: "a = 'a\\\\b\\\"'", FileName: "case.js"},
			// Snapshot cases for template literals.
			{Code: "a = `a`", FileName: "case.js"},
			{Code: "a = `${foo}`", FileName: "case.js"},
			{Code: "a = `a${100}b`", FileName: "case.js"},
			{Code: "a = `a\\t${foo.bar}b\\\\c`", FileName: "case.js"},
			{Code: "a = `${foo}\\\\a${bar}\\``", FileName: "case.js"},
			{Code: "a = `a\\${`", FileName: "case.js"},
			{Code: "a = `a$\\{`", FileName: "case.js"},
			{Code: "a = `${a}\\'${b}\\\\`", FileName: "case.js"},
			{Code: "a = `\\\"a\\\\b`", FileName: "case.js"},
			{Code: "a = `\\\\a${foo}b\\\\${foo}`", FileName: "case.js"},
			{Code: "a = String.raw`a\\\\b`", FileName: "case.js"},
			{Code: "a = `\\\\a \\\nb`", FileName: "case.js"},
			// Snapshot cases for unnecessary String.raw.
			{Code: "a = foo`ab`", FileName: "case.js"},
			{Code: "a = foo().bar`ab`", FileName: "case.js"},
			{Code: "a = foo.bar()`ab`", FileName: "case.js"},
			{Code: "a = String[\"raw\"]`ab`", FileName: "case.js"},
			{Code: "a = foo.raw`ab`", FileName: "case.js"},
			{Code: "a = String.foo`ab`", FileName: "case.js"},
			{Code: "a = String.raw`a\\b`", FileName: "case.js"},
			{Code: "a = String.raw`a\\b${foo}cd`", FileName: "case.js"},
			{Code: "a = String.raw`ab${foo}c\\nd`", FileName: "case.js"},
			{Code: "a = String.raw`a\n\tb\\c\n\tde`", FileName: "case.js"},
			{Code: "a = String.raw\n\t// Comment\n\t`ab\\c`", FileName: "case.js"},
			// Shadowing, comments and yield.
			{Code: "const String = {raw() {}};\na = String.raw`abc`", FileName: "case.js"},
			// Snapshot cases for restricted syntax and Jest inline snapshots.
			{Code: "'a\\\\b';", FileName: "case.ts"},
			{Code: "import 'a\\\\b';", FileName: "case.ts"},
			{Code: "export {} from 'a\\\\b';", FileName: "case.ts"},
			{Code: "export * from 'a\\\\b';", FileName: "case.ts"},
			{Code: "import 'm' with {'a\\\\b': 'v'};", FileName: "case.ts"},
			{Code: "export {} from 'm' with {'a\\\\b': 'v'};", FileName: "case.ts"},
			{Code: "import 'm' with {k: 'a\\\\b'};", FileName: "case.ts"},
			{Code: "export {} from 'm' with {k: 'a\\\\b'};", FileName: "case.ts"},
			{Code: "import {'a\\\\b' as s} from 'm';", FileName: "case.ts"},
			{Code: "export {'a\\\\b' as s} from 'm';", FileName: "case.ts"},
			{Code: "export {s as 'a\\\\b'} from 'm';", FileName: "case.ts"},
			{Code: "export * as 'a\\\\b' from 'm';", FileName: "case.ts"},
			{Code: "<Component attribute='a\\\\b' />", FileName: "case.tsx"},
			{Code: "enum E {'a\\\\b' = 1}", FileName: "case.ts"},
			{Code: "enum E {K = 'a\\\\b'}", FileName: "case.ts"},
			{Code: "module 'a\\\\b' {}", FileName: "case.ts"},
			{Code: "import type T = require('a\\\\b');", FileName: "case.ts"},
			{Code: "type T = 'a\\\\b';", FileName: "case.ts"},
			{Code: "type T = import('a\\\\b');", FileName: "case.ts"},
			{Code: "({ 'a\\\\b': 1 })", FileName: "case.ts"},
			{Code: "class C { 'a\\\\b' = 1 }", FileName: "case.ts"},
			{Code: "class C { 'a\\\\b'(){} }", FileName: "case.ts"},
			{Code: "class C { accessor 'a\\\\b' = 1 }", FileName: "case.ts"},
			{Code: "abstract class C { abstract 'a\\\\b' }", FileName: "case.ts"},
			{Code: "abstract class C { abstract ['a\\\\b'] }", FileName: "case.ts"},
			{Code: "abstract class C { abstract 'a\\\\b'() }", FileName: "case.ts"},
			{Code: "abstract class C { abstract ['a\\\\b']() }", FileName: "case.ts"},
			{Code: "abstract class C { abstract accessor 'a\\\\b' }", FileName: "case.ts"},
			{Code: "abstract class C { abstract accessor ['a\\\\b'] }", FileName: "case.ts"},
			{Code: "interface I { 'a\\\\b' }", FileName: "case.ts"},
			{Code: "interface I { ['a\\\\b'] }", FileName: "case.ts"},
			{Code: "interface I { 'a\\\\b'(): void }", FileName: "case.ts"},
			{Code: "interface I { ['a\\\\b'](): void }", FileName: "case.ts"},
			{Code: "expect(foo).toMatchInlineSnapshot('a\\\\b')", FileName: "case.ts"},
			{Code: "expect(foo).toMatchInlineSnapshot(`a\\\\b`)", FileName: "case.ts"},
			// Documentation examples.
			{Code: "const file = String.raw`C:\\windows\\style\\path\\to\\file.js`;", FileName: "case.js"},
			{Code: "const regexp = new RegExp(String.raw`foo\\.bar`);", FileName: "case.js"},
			{Code: "const file = String.raw`C:\\windows\\temp\\my-app-${process.pid}.log`;", FileName: "case.js"},
			{Code: "const noBackslash = 'foobar'", FileName: "case.js"},
			// Shared upstream empty-file smoke test.
			{Code: "", FileName: "case.js"},
		},
		[]rule_tester.InvalidTestCase{
			// TypeScript literal and enum contexts.
			{Code: "type Separator = 'a\\\\b' | 'c\\\\d'; const separator: Separator = 'a\\\\b';", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 64, 1, 70, "type Separator = 'a\\\\b' | 'c\\\\d'; const separator: Separator = String.raw`a\\b`;")}},
			{Code: "const options = {path: 'C:\\\\Users'} as const; const exact: 'C:\\\\Users' = options.path;", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 24, 1, 35, "const options = {path: String.raw`C:\\Users`} as const; const exact: 'C:\\\\Users' = options.path;")}},
			// String literals with and without type information.
			{Code: "const path = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b`;")}},
			{Code: "const path: 'a\\\\b' = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 22, 1, 28, "const path: 'a\\\\b' = String.raw`a\\b`;")}},
			{Code: "type PathText = 'a\\\\b' | undefined; const path: PathText = 'a\\\\b'; const exact: 'a\\\\b' = path;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 60, 1, 66, "type PathText = 'a\\\\b' | undefined; const path: PathText = String.raw`a\\b`; const exact: 'a\\\\b' = path;")}},
			{Code: "declare function use(path: 'a\\\\b'): void; use('a\\\\b');", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 47, 1, 53, "declare function use(path: 'a\\\\b'): void; use(String.raw`a\\b`);")}},
			{Code: "const options = {path: 'a\\\\b'} as const;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 24, 1, 30, "const options = {path: String.raw`a\\b`} as const;")}},
			{Code: "const options = ['a\\\\b'] as const;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 18, 1, 24, "const options = [String.raw`a\\b`] as const;")}},
			{Code: "const options = {nested: {path: 'a\\\\b'}} as const;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 33, 1, 39, "const options = {nested: {path: String.raw`a\\b`}} as const;")}},
			{Code: "class Paths {readonly path = 'a\\\\b';}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 30, 1, 36, "class Paths {readonly path = String.raw`a\\b`;}")}},
			{Code: "const path = 'a\\\\b' satisfies string;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b` satisfies string;")}},
			{Code: "const path = 'a\\\\b' satisfies 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b` satisfies 'a\\\\b';")}},
			{Code: "declare function use<const Type>(value: Type): Type; const path = use('a\\\\b');", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 71, 1, 77, "declare function use<const Type>(value: Type): Type; const path = use(String.raw`a\\b`);")}},
			{Code: "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use('a\\\\b');", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 103, 1, 109, "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use(String.raw`a\\b`);")}},
			{Code: "const path = (/* keep */ 'a\\\\b');", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 26, 1, 32, "const path = (/* keep */ String.raw`a\\b`);")}},
			{Code: "function path(): string {return 'a\\\\b';}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 33, 1, 39, "function path(): string {return String.raw`a\\b`;}")}},
			{Code: "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = 'a\\\\b';\nif (getTag(path) === 'string') {}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(3, 27, 3, 33, "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = String.raw`a\\b`;\nif (getTag(path) === 'string') {}")}},
			// Widened string bindings.
			{Code: "const path: string = 'a\\\\b';", FileName: "file.ts", Output: []string{"const path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 22, 1, 28)}},
			{Code: "let path: string = 'a\\\\b';", FileName: "file.ts", Output: []string{"let path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 20, 1, 26)}},
			{Code: "let path = 'a\\\\b';", FileName: "file.ts", Output: []string{"let path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "var path = 'a\\\\b';", FileName: "file.ts", Output: []string{"var path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "const path: string = (/* keep */ 'a\\\\b');", FileName: "file.ts", Output: []string{"const path: string = (/* keep */ String.raw`a\\b`);"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 34, 1, 40)}},
			// Nullable and aliased string bindings.
			{Code: "type PathText = string; const path: PathText = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 48, 1, 54, "type PathText = string; const path: PathText = String.raw`a\\b`;")}},
			{Code: "const path: string | undefined = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 34, 1, 40, "const path: string | undefined = String.raw`a\\b`;")}},
			{Code: "let path: string | null = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 27, 1, 33, "let path: string | null = String.raw`a\\b`;")}},
			{Code: "type PathText = string | undefined; const path: PathText = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 60, 1, 66, "type PathText = string | undefined; const path: PathText = String.raw`a\\b`;")}},
			// Template literals with and without type information.
			{Code: "const path = `a\\\\b`;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b`;")}},
			{Code: "const path: 'a\\\\b' = `a\\\\b`;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 22, 1, 28, "const path: 'a\\\\b' = String.raw`a\\b`;")}},
			{Code: "type PathText = 'a\\\\b' | undefined; const path: PathText = `a\\\\b`; const exact: 'a\\\\b' = path;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 60, 1, 66, "type PathText = 'a\\\\b' | undefined; const path: PathText = String.raw`a\\b`; const exact: 'a\\\\b' = path;")}},
			{Code: "declare function use(path: 'a\\\\b'): void; use(`a\\\\b`);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 47, 1, 53, "declare function use(path: 'a\\\\b'): void; use(String.raw`a\\b`);")}},
			{Code: "const options = {path: `a\\\\b`} as const;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 24, 1, 30, "const options = {path: String.raw`a\\b`} as const;")}},
			{Code: "const options = [`a\\\\b`] as const;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 18, 1, 24, "const options = [String.raw`a\\b`] as const;")}},
			{Code: "const options = {nested: {path: `a\\\\b`}} as const;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 33, 1, 39, "const options = {nested: {path: String.raw`a\\b`}} as const;")}},
			{Code: "class Paths {readonly path = `a\\\\b`;}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 30, 1, 36, "class Paths {readonly path = String.raw`a\\b`;}")}},
			{Code: "const path = `a\\\\b` satisfies string;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b` satisfies string;")}},
			{Code: "const path = `a\\\\b` satisfies 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b` satisfies 'a\\\\b';")}},
			{Code: "declare function use<const Type>(value: Type): Type; const path = use(`a\\\\b`);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 71, 1, 77, "declare function use<const Type>(value: Type): Type; const path = use(String.raw`a\\b`);")}},
			{Code: "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use(`a\\\\b`);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 103, 1, 109, "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use(String.raw`a\\b`);")}},
			{Code: "const path = (/* keep */ `a\\\\b`);", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 26, 1, 32, "const path = (/* keep */ String.raw`a\\b`);")}},
			{Code: "function path(): string {return `a\\\\b`;}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 33, 1, 39, "function path(): string {return String.raw`a\\b`;}")}},
			{Code: "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = `a\\\\b`;\nif (getTag(path) === 'string') {}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(3, 27, 3, 33, "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = String.raw`a\\b`;\nif (getTag(path) === 'string') {}")}},
			// Widened template bindings.
			{Code: "const path: string = `a\\\\b`;", FileName: "file.ts", Output: []string{"const path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 22, 1, 28)}},
			{Code: "let path: string = `a\\\\b`;", FileName: "file.ts", Output: []string{"let path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 20, 1, 26)}},
			{Code: "let path = `a\\\\b`;", FileName: "file.ts", Output: []string{"let path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "var path = `a\\\\b`;", FileName: "file.ts", Output: []string{"var path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "const path: string = (/* keep */ `a\\\\b`);", FileName: "file.ts", Output: []string{"const path: string = (/* keep */ String.raw`a\\b`);"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 34, 1, 40)}},
			// Nullable and aliased template bindings.
			{Code: "type PathText = string; const path: PathText = `a\\\\b`;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 48, 1, 54, "type PathText = string; const path: PathText = String.raw`a\\b`;")}},
			{Code: "const path: string | undefined = `a\\\\b`;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 34, 1, 40, "const path: string | undefined = String.raw`a\\b`;")}},
			{Code: "let path: string | null = `a\\\\b`;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 27, 1, 33, "let path: string | null = String.raw`a\\b`;")}},
			{Code: "type PathText = string | undefined; const path: PathText = `a\\\\b`;", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 60, 1, 66, "type PathText = string | undefined; const path: PathText = String.raw`a\\b`;")}},
			// An interpolated template in a return statement.
			{Code: "declare const suffix: string; function path(): string {return`a\\\\b${suffix}`;}", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 62, 1, 77, "declare const suffix: string; function path(): string {return String.raw`a\\b${suffix}`;}")}},
			// Type-sensitive templates and TypeScript extensions.
			{Code: "const path = 'a\\\\b';", FileName: "file.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b`;")}},
			{Code: "const path = 'a\\\\b';", FileName: "file.mts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b`;")}},
			{Code: "const path = 'a\\\\b';", FileName: "file.cts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b`;")}},
			{Code: "const path = 'a\\\\b';", FileName: "file.tsx", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 14, 1, 20, "const path = String.raw`a\\b`;")}},
			// Upstream uses the TypeScript parser on file.js. The native parser selects JavaScript behavior.
			{Code: "const path = 'a\\\\b';", FileName: "file.js", Output: []string{"const path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 14, 1, 20)}},
			{Code: "const path = String.raw`abc`;", FileName: "file.ts", Output: []string{"const path = 'abc';"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 14, 1, 24)}},
			// Snapshot cases for string literals.
			{Code: "TEST_STRING = 'a\\\\b';", FileName: "case.js", Output: []string{"TEST_STRING = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 15, 1, 21)}},
			{Code: "function a() {return'a\\\\b'}", FileName: "case.js", Output: []string{"function a() {return String.raw`a\\b`}"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 21, 1, 27)}},
			{Code: "const foo = \"foo \\\\x46\";", FileName: "case.js", Output: []string{"const foo = String.raw`foo \\x46`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 13, 1, 24)}},
			{Code: "a = 'a\\\\b\\''", FileName: "case.js", Output: []string{"a = String.raw`a\\b'`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 13)}},
			{Code: "a = \"a\\\\b\\\"\"", FileName: "case.js", Output: []string{"a = String.raw`a\\b\"`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 13)}},
			// Snapshot cases for template literals.
			{Code: "a = `a\\\\b`", FileName: "case.js", Output: []string{"a = String.raw`a\\b`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 11)}},
			{Code: "function a() {return`a\\\\b${foo}cd`}", FileName: "case.js", Output: []string{"function a() {return String.raw`a\\b${foo}cd`}"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 21, 1, 35)}},
			{Code: "a = {[`a\\\\b${foo}cd${foo.bar}e\\\\f`]: b}", FileName: "case.js", Output: []string{"a = {[String.raw`a\\b${foo}cd${foo.bar}e\\f`]: b}"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 7, 1, 35)}},
			{Code: "a = `a${foo}${foo.bar}b\\\\c`", FileName: "case.js", Output: []string{"a = String.raw`a${foo}${foo.bar}b\\c`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 28)}},
			{Code: "a = `a\\\\b${\"c\\\\d\"}e`", FileName: "case.js", Output: []string{"a = String.raw`a\\b${String.raw`c\\d`}e`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 21), preferRawError(1, 12, 1, 18)}},
			{Code: "a = `\\\\a\nb`", FileName: "case.js", Output: []string{"a = String.raw`\\a\nb`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 2, 3)}},
			{Code: "a = `\\\\a${foo}\nb${bar}c\nd\\\\\\\\e`", FileName: "case.js", Output: []string{"a = String.raw`\\a${foo}\nb${bar}c\nd\\\\e`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 3, 8)}},
			{Code: "a = `a\\\\b${ foo /* bar */}c\\\\d`", FileName: "case.js", Output: []string{"a = String.raw`a\\b${ foo /* bar */}c\\d`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 32)}},
			{Code: "a = `a\\\\b${ foo + bar }`", FileName: "case.js", Output: []string{"a = String.raw`a\\b${ foo + bar }`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 25)}},
			{Code: "a = `${ foo .bar }a\\\\b`", FileName: "case.js", Output: []string{"a = String.raw`${ foo .bar }a\\b`"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 5, 1, 24)}},
			// Snapshot cases for unnecessary String.raw.
			{Code: "a = String.raw`abc`", FileName: "case.js", Output: []string{"a = 'abc'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab${foo}cd`", FileName: "case.js", Output: []string{"a = `ab${foo}cd`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab\"c`", FileName: "case.js", Output: []string{"a = 'ab\"c'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab'c`", FileName: "case.js", Output: []string{"a = 'ab\\'c'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab'\"c`", FileName: "case.js", Output: []string{"a = 'ab\\'\"c'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab\r\nc`", FileName: "case.js", Output: []string{"a = `ab\r\nc`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`a\n\tbc\n\tde`", FileName: "case.js", Output: []string{"a = `a\n\tbc\n\tde`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`\na${foo}b\n${bar}cd`", FileName: "case.js", Output: []string{"a = `\na${foo}b\n${bar}cd`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a\nString.raw`abc`", FileName: "case.js", Output: []string{"a\n'abc'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 1, 2, 11)}},
			{Code: "a\nString.raw`a${b}`", FileName: "case.js", Output: []string{"a\n;`a${b}`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 1, 2, 11)}},
			{Code: "a = String.raw\n// Comment\n`ab\nc`", FileName: "case.js", Output: []string{"a = \n// Comment\n`ab\nc`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw  /* comment */  `ab\nc`", FileName: "case.js", Output: []string{"a =   /* comment */  `ab\nc`"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "function foo() {\n\treturn String.raw`abc`\n}", FileName: "case.js", Output: []string{"function foo() {\n\treturn 'abc'\n}"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 9, 2, 19)}},
			{Code: "function foo() {\n\treturn String.raw\n\t\t\t// comment\n\t\t\t`abc`\n}", FileName: "case.js", Output: []string{"function foo() {\n\treturn ( \n\t\t\t// comment\n\t\t\t'abc')\n}"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 9, 2, 19)}},
			{Code: "function foo() {\n\treturn (String.raw\n\t\t\t// already parenthesized\n\t\t\t`abc`)\n}", FileName: "case.js", Output: []string{"function foo() {\n\treturn (\n\t\t\t// already parenthesized\n\t\t\t'abc')\n}"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 10, 2, 20)}},
			{Code: "function foo() {\n\treturn (String.raw) // remove parenthesis\n\t\t\t`abc`\n}", FileName: "case.js", Output: []string{"function foo() {\n\treturn (  // remove parenthesis\n\t\t\t'abc')\n}"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 10, 2, 20)}},
			{Code: "String.raw`use strict`", FileName: "case.js", Output: []string{"'use strict'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 1, 1, 11)}},
			// Shadowing, comments and yield.
			{Code: "String /* keep */ .raw`abc`;", FileName: "case.js", Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 1, 1, 23)}},
			{Code: "function* foo() {\n\tyield String.raw\n\t\t`abc`;\n}", FileName: "case.js", Output: []string{"function* foo() {\n\tyield (\n\t\t'abc');\n}"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 8, 2, 18)}},
			{Code: "function* foo() {\n\tyield String.raw\n\t\t`a${value}`;\n}", FileName: "case.js", Output: []string{"function* foo() {\n\tyield (\n\t\t`a${value}`);\n}"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(2, 8, 2, 18)}},
			// Snapshot cases for restricted syntax and Jest inline snapshots.
			{Code: "({ ['a\\\\b']: 1 })", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 5, 1, 11, "({ [String.raw`a\\b`]: 1 })")}},
			{Code: "class C { ['a\\\\b'] = 1 }", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 12, 1, 18, "class C { [String.raw`a\\b`] = 1 }")}},
			{Code: "class C { ['a\\\\b'](){} }", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 12, 1, 18, "class C { [String.raw`a\\b`](){} }")}},
			{Code: "class C { accessor ['a\\\\b'] = 1 }", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 21, 1, 27, "class C { accessor [String.raw`a\\b`] = 1 }")}},
			{Code: "expect('a\\\\b').toMatchInlineSnapshot(\"\")", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 8, 1, 14, "expect(String.raw`a\\b`).toMatchInlineSnapshot(\"\")")}},
			{Code: "expect(`a\\\\b`).toMatchInlineSnapshot(``)", FileName: "case.ts", Errors: []rule_tester.InvalidTestCaseError{suggestRaw(1, 8, 1, 14, "expect(String.raw`a\\b`).toMatchInlineSnapshot(``)")}},
			// Line terminators.
			{Code: "a = String.raw`ab\rd`;", FileName: "case.js", Output: []string{"a = `ab\rd`;"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab\nd`;", FileName: "case.js", Output: []string{"a = `ab\nd`;"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`ab\u2028d`;", FileName: "case.js", Output: []string{"a = `ab\u2028d`;"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			{Code: "a = String.raw`abc`;", FileName: "case.js", Output: []string{"a = 'abc';"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 5, 1, 15)}},
			// Removing the tag after return, throw or yield.
			{Code: "function read(ready) { return String.raw\n`x` && ready; }", FileName: "case.js", Output: []string{"function read(ready) { return ( \n'x' && ready); }"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 31, 1, 41)}},
			{Code: "function read(ready) { throw String.raw\n`x` && ready; }", FileName: "case.js", Output: []string{"function read(ready) { throw ( \n'x' && ready); }"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 30, 1, 40)}},
			{Code: "function* read(ready) { yield String.raw\n`x` && ready; }", FileName: "case.js", Output: []string{"function* read(ready) { yield (\n'x') && ready; }"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 31, 1, 41)}},
			{Code: "function read(ready) { return String.raw`x` && ready; }", FileName: "case.js", Output: []string{"function read(ready) { return 'x' && ready; }"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 31, 1, 41)}},
			{Code: "function read(ready) { return (String.raw\n`x` && ready); }", FileName: "case.js", Output: []string{"function read(ready) { return (\n'x' && ready); }"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 32, 1, 42)}},
			{Code: "function read(ready) { return String.raw /* keep */\n`x` && ready; }", FileName: "case.js", Output: []string{"function read(ready) { return (  /* keep */\n'x' && ready); }"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 31, 1, 41)}},
			// Documentation examples.
			{Code: "const file = \"C:\\\\windows\\\\style\\\\path\\\\to\\\\file.js\";", FileName: "case.js", Output: []string{"const file = String.raw`C:\\windows\\style\\path\\to\\file.js`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 14, 1, 53)}},
			{Code: "const regexp = new RegExp('foo\\\\.bar');", FileName: "case.js", Output: []string{"const regexp = new RegExp(String.raw`foo\\.bar`);"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 27, 1, 38)}},
			{Code: "const file = `C:\\\\windows\\\\temp\\\\my-app-${process.pid}.log`;", FileName: "case.js", Output: []string{"const file = String.raw`C:\\windows\\temp\\my-app-${process.pid}.log`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 14, 1, 60)}},
			{Code: "const noBackslash = String.raw`foobar`", FileName: "case.js", Output: []string{"const noBackslash = 'foobar'"}, Errors: []rule_tester.InvalidTestCaseError{unnecessaryRawError(1, 21, 1, 31)}},
		},
	)
}

func TestPreferStringRawUpstreamTyped(t *testing.T) {
	r := prefer_string_raw.PreferStringRawRule
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &r,
		[]rule_tester.ValidTestCase{
			// String literals with and without type information.
			{Code: "const path = 'a\\\\b';", FileName: "file.ts"},
			{Code: "const path: 'a\\\\b' = 'a\\\\b';", FileName: "file.ts"},
			{Code: "type PathText = 'a\\\\b' | undefined; const path: PathText = 'a\\\\b'; const exact: 'a\\\\b' = path;", FileName: "file.ts"},
			{Code: "declare function use(path: 'a\\\\b'): void; use('a\\\\b');", FileName: "file.ts"},
			{Code: "const options = {path: 'a\\\\b'} as const;", FileName: "file.ts"},
			{Code: "const options = ['a\\\\b'] as const;", FileName: "file.ts"},
			{Code: "const options = {nested: {path: 'a\\\\b'}} as const;", FileName: "file.ts"},
			{Code: "class Paths {readonly path = 'a\\\\b';}", FileName: "file.ts"},
			{Code: "const path = 'a\\\\b' satisfies string;", FileName: "file.ts"},
			{Code: "const path = 'a\\\\b' satisfies 'a\\\\b';", FileName: "file.ts"},
			{Code: "declare function use<const Type>(value: Type): Type; const path = use('a\\\\b');", FileName: "file.ts"},
			{Code: "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use('a\\\\b');", FileName: "file.ts"},
			{Code: "const path = (/* keep */ 'a\\\\b');", FileName: "file.ts"},
			{Code: "function path(): string {return 'a\\\\b';}", FileName: "file.ts"},
			{Code: "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = 'a\\\\b';\nif (getTag(path) === 'string') {}", FileName: "file.ts"},
			// String literals under const assertions.
			{Code: "const path = 'a\\\\b' as const;", FileName: "file.ts"},
			{Code: "const path = <const>'a\\\\b';", FileName: "file.ts"},
			// Template literals with and without type information.
			{Code: "const path = `a\\\\b`;", FileName: "file.ts"},
			{Code: "const path: 'a\\\\b' = `a\\\\b`;", FileName: "file.ts"},
			{Code: "type PathText = 'a\\\\b' | undefined; const path: PathText = `a\\\\b`; const exact: 'a\\\\b' = path;", FileName: "file.ts"},
			{Code: "declare function use(path: 'a\\\\b'): void; use(`a\\\\b`);", FileName: "file.ts"},
			{Code: "const options = {path: `a\\\\b`} as const;", FileName: "file.ts"},
			{Code: "const options = [`a\\\\b`] as const;", FileName: "file.ts"},
			{Code: "const options = {nested: {path: `a\\\\b`}} as const;", FileName: "file.ts"},
			{Code: "class Paths {readonly path = `a\\\\b`;}", FileName: "file.ts"},
			{Code: "const path = `a\\\\b` satisfies string;", FileName: "file.ts"},
			{Code: "const path = `a\\\\b` satisfies 'a\\\\b';", FileName: "file.ts"},
			{Code: "declare function use<const Type>(value: Type): Type; const path = use(`a\\\\b`);", FileName: "file.ts"},
			{Code: "declare function use(value: 'a\\\\b'): 1; declare function use(value: string): 2; const result: 1 = use(`a\\\\b`);", FileName: "file.ts"},
			{Code: "const path = (/* keep */ `a\\\\b`);", FileName: "file.ts"},
			{Code: "function path(): string {return `a\\\\b`;}", FileName: "file.ts"},
			{Code: "type Tag<Value> = Value extends string ? 'string' : 'other';\ndeclare function getTag<Value>(value: Value): Tag<Value>;\nconst path: {} | 'a\\\\b' = `a\\\\b`;\nif (getTag(path) === 'string') {}", FileName: "file.ts"},
			// Template literals under const assertions.
			{Code: "const path = `a\\\\b` as const;", FileName: "file.ts"},
			{Code: "const path = <const>`a\\\\b`;", FileName: "file.ts"},
			// Type-sensitive templates and TypeScript extensions.
			{Code: "declare const suffix: string; declare function use<const Type>(value: Type): Type; const result = use({path: `a\\\\b${suffix}`});", FileName: "file.ts"},
			{Code: "declare const suffix: string; const path: `a\\\\b${string}` = `a\\\\b${suffix}`;", FileName: "file.ts"},
		},
		[]rule_tester.InvalidTestCase{
			// Widened string bindings.
			{Code: "const path: string = 'a\\\\b';", FileName: "file.ts", Output: []string{"const path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 22, 1, 28)}},
			{Code: "let path: string = 'a\\\\b';", FileName: "file.ts", Output: []string{"let path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 20, 1, 26)}},
			{Code: "let path = 'a\\\\b';", FileName: "file.ts", Output: []string{"let path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "var path = 'a\\\\b';", FileName: "file.ts", Output: []string{"var path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "const path: string = (/* keep */ 'a\\\\b');", FileName: "file.ts", Output: []string{"const path: string = (/* keep */ String.raw`a\\b`);"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 34, 1, 40)}},
			// Nullable and aliased string bindings.
			{Code: "type PathText = string; const path: PathText = 'a\\\\b';", FileName: "file.ts", Output: []string{"type PathText = string; const path: PathText = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 48, 1, 54)}},
			{Code: "const path: string | undefined = 'a\\\\b';", FileName: "file.ts", Output: []string{"const path: string | undefined = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 34, 1, 40)}},
			{Code: "let path: string | null = 'a\\\\b';", FileName: "file.ts", Output: []string{"let path: string | null = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 27, 1, 33)}},
			{Code: "type PathText = string | undefined; const path: PathText = 'a\\\\b';", FileName: "file.ts", Output: []string{"type PathText = string | undefined; const path: PathText = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 60, 1, 66)}},
			// Widened template bindings.
			{Code: "const path: string = `a\\\\b`;", FileName: "file.ts", Output: []string{"const path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 22, 1, 28)}},
			{Code: "let path: string = `a\\\\b`;", FileName: "file.ts", Output: []string{"let path: string = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 20, 1, 26)}},
			{Code: "let path = `a\\\\b`;", FileName: "file.ts", Output: []string{"let path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "var path = `a\\\\b`;", FileName: "file.ts", Output: []string{"var path = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 12, 1, 18)}},
			{Code: "const path: string = (/* keep */ `a\\\\b`);", FileName: "file.ts", Output: []string{"const path: string = (/* keep */ String.raw`a\\b`);"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 34, 1, 40)}},
			// Nullable and aliased template bindings.
			{Code: "type PathText = string; const path: PathText = `a\\\\b`;", FileName: "file.ts", Output: []string{"type PathText = string; const path: PathText = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 48, 1, 54)}},
			{Code: "const path: string | undefined = `a\\\\b`;", FileName: "file.ts", Output: []string{"const path: string | undefined = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 34, 1, 40)}},
			{Code: "let path: string | null = `a\\\\b`;", FileName: "file.ts", Output: []string{"let path: string | null = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 27, 1, 33)}},
			{Code: "type PathText = string | undefined; const path: PathText = `a\\\\b`;", FileName: "file.ts", Output: []string{"type PathText = string | undefined; const path: PathText = String.raw`a\\b`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 60, 1, 66)}},
			// An interpolated template in a return statement.
			{Code: "declare const suffix: string; function path(): string {return`a\\\\b${suffix}`;}", FileName: "file.ts", Output: []string{"declare const suffix: string; function path(): string {return String.raw`a\\b${suffix}`;}"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 62, 1, 77)}},
			// Type-sensitive templates and TypeScript extensions.
			{Code: "declare const suffix: string; const path = `a\\\\b${suffix}`;", FileName: "file.ts", Output: []string{"declare const suffix: string; const path = String.raw`a\\b${suffix}`;"}, Errors: []rule_tester.InvalidTestCaseError{preferRawError(1, 44, 1, 59)}},
		},
	)
}

func preferRawError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "prefer-string-raw",
		Message:   "`String.raw` should be used to avoid escaping `\\`.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func unnecessaryRawError(line, column, endLine, endColumn int) rule_tester.InvalidTestCaseError {
	return rule_tester.InvalidTestCaseError{
		MessageId: "unnecessary-string-raw",
		Message:   "Using `String.raw` is unnecessary as the string does not contain any `\\`.",
		Line:      line, Column: column, EndLine: endLine, EndColumn: endColumn,
	}
}

func suggestRaw(line, column, endLine, endColumn int, output string) rule_tester.InvalidTestCaseError {
	diagnostic := preferRawError(line, column, endLine, endColumn)
	diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{
		{MessageId: "prefer-string-raw/suggestion", Output: output},
	}
	return diagnostic
}
