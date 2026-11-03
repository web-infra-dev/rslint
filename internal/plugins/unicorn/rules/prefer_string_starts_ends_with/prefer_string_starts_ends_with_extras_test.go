package prefer_string_starts_ends_with_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_string_starts_ends_with"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

// Additional AST, Unicode, mutation and edit-safety cases. Unsafe upstream
// edits have conservative expectations that preserve runtime behavior.
func TestPreferStringStartsEndsWithExtras(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_string_starts_ends_with.PreferStringStartsEndsWithRule,
		[]rule_tester.ValidTestCase{
			{Code: "\"abc\".indexOf(\"a\") === +0", FileName: "file.ts"},
			{Code: "\"abc\".indexOf(\"a\") === 0n", FileName: "file.ts"},
			{Code: "\"abc\".slice(0, +1) === \"a\"", FileName: "file.ts"},
			{Code: "\"abc\".slice(0, 1_0) === \"a\"", FileName: "file.ts"},
			{Code: "const needle=\"a\"; \"abc\".slice(0, needle[\"length\"]) === needle", FileName: "file.ts"},
			{Code: "const needle=\"a\"; \"abc\".slice(0, needle?.length) === needle", FileName: "file.ts"},
			{Code: "const needle=\"c\"; \"abc\".slice(-needle[\"length\"]) === needle", FileName: "file.ts"},
			{Code: "(/^a/?.test)(value)", FileName: "file.ts"},
			{Code: "/^a/.test?.(value)", FileName: "file.ts"},
			{Code: "/^a/?.test(value)", FileName: "file.ts"},
			{Code: "/^a/[\"test\"](value)", FileName: "file.ts"},
			{Code: "/^a/.test(...value)", FileName: "file.ts"},
			{Code: "/^a/.test(value, 1)", FileName: "file.ts"},
			{Code: "/^a/.test()", FileName: "file.ts"},
			{Code: "new RegExp(\"^a\").test(value)", FileName: "file.ts"},
			{Code: "/^a/.test(foo as number)", FileName: "file.ts"},
			{Code: "/** @type {unknown} */ let text = \"abc\"; text.indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "const text = {a:\"abc\"}; text.a.indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "let text = \"abc\"; text = \"xyz\"; text.indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "class C { #indexOf() { return 0; } m() { return this.#indexOf(\"a\") === 0; } }", FileName: "file.ts"},
			{Code: "const X = <A.B />;", FileName: "file.tsx"},
			{Code: "const read = async (): string => value; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "function read() { return \"abc\" }; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "const read = (): string => value; const alias = read; alias().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "const read: () => string = () => value; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "const read = <T>(): string => value; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "function* read(): string { return value }; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "function read(): string { return value }; read = replacement; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "let read = (): string => value; read = replacement; read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			{Code: "const read = (() => value) satisfies (() => string); read().indexOf(\"a\") === 0;", FileName: "file.ts"},
			// Optional bindings, JSDoc types, and source-only inference boundaries.
			{Code: "function f(value?: string) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "function f(value?: \"abc\") { return value.indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "/** @param {string} [value] */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "/** @param {'abc'} [value] */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "/** @param {string|string} [value] */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "/** @param {!string} [value] */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "/** @type {(()=>string) & (()=>number)} */ let read; read().indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "declare const read: () => string; read().indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "let value = \"abc\"; (flag ? value : \"x\").indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "let value = \"abc\"; (value || \"x\").indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "let value = \"abc\"; const alias = value; alias.indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "let value = \"abc\"; value!.indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "type Text = string; (unknown satisfies Text).indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "/** @type {unknown} */ const value = \"abc\"; (true ? value : \"x\").indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "/** @type {unknown} */ const value = \"abc\"; (value || \"x\").indexOf(\"a\") === 0;", FileName: "review.ts"},
			{Code: "/** @param {()=>string} [read] */ function f(read) { return read().indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "/** @param {()=>string} read */ function f(read = (): string => value) { return read().indexOf(\"a\") === 0; }", FileName: "review.ts"},
			{Code: "function f(value: `a${string}`) { return value.indexOf(\"a\") === 0; }", FileName: "file.ts"},
			// Same-named local bindings are not evidence of built-in string results.
			{Code: "function String() { return [\"a\"]; } String().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function Date() { return [\"a\"]; } Date().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function encodeURI() { return [\"a\"]; } encodeURI().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function encodeURIComponent() { return [\"a\"]; } encodeURIComponent().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function decodeURI() { return [\"a\"]; } decodeURI().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function decodeURIComponent() { return [\"a\"]; } decodeURIComponent().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function atob() { return [\"a\"]; } atob().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function btoa() { return [\"a\"]; } btoa().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function String(): string[] { return [\"a\"]; } String().slice(0, 1) === \"a\";", FileName: "safety.ts"},
			{Code: "function f(String) { return String().indexOf(\"a\") === 0; }", FileName: "safety.ts"},
			{Code: "import { String } from \"custom\"; String().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "const { String } = custom; String().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "const String = { fromCharCode() { return [\"a\"]; }, fromCodePoint() { return [\"a\"]; }, raw() { return [\"a\"]; } }; String.fromCharCode(65).indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "const String = { fromCodePoint() { return [\"a\"]; } }; String.fromCodePoint(65).indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "const String = { raw() { return [\"a\"]; } }; String.raw({raw: [\"a\"]}).indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "const String = { raw() { return [\"a\"]; } }; String.raw`a`.indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "class Date { toISOString() { return [\"a\"]; } } new Date().toISOString().indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "class RegExp { get source() { return [\"a\"]; } } new RegExp().source.indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "function RegExp() { return {source: [\"a\"]}; } RegExp().source.indexOf(\"a\") === 0;", FileName: "safety.ts"},
			{Code: "type Date = { toISOString(): string[] }; function f(value: Date) { return value.toISOString().indexOf(\"a\") === 0; }", FileName: "safety.ts"},
			{Code: "type RegExp = { source: string[] }; function f(value: RegExp) { return value.source.indexOf(\"a\") === 0; }", FileName: "safety.ts"},
			{Code: "namespace Types { export type Date = { toISOString(): string[] }; } function f(value: Types.Date) { return value.toISOString().indexOf(\"a\") === 0; }", FileName: "safety.ts"},
		},
		[]rule_tester.InvalidTestCase{
			{Code: "\"abc\".slice(0, -0) === \"\"", FileName: "file.ts",
				Output: []string{"\"abc\".startsWith(\"\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 26},
				},
			},
			{Code: "\"abc\".slice(0, 0) === \"\"", FileName: "file.ts",
				Output: []string{"\"abc\".startsWith(\"\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{Code: "\"😀x\".slice(0, 2) === \"😀\"", FileName: "file.ts",
				Output: []string{"\"😀x\".startsWith(\"😀\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 27},
				},
			},
			{Code: "\"x😀\".slice(-2) === \"😀\"", FileName: "file.ts",
				Output: []string{"\"x😀\".endsWith(\"😀\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ends-with-slice", Message: "Prefer `String#endsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25},
				},
			},
			{Code: "\"abc\".slice(0, 0x1) === \"a\"", FileName: "file.ts",
				Output: []string{"\"abc\".startsWith(\"a\")"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 28},
				},
			},
			{Code: "(/^a/.test)(value)", FileName: "file.ts",
				Output: []string{"(value.startsWith)('a')"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useStringCasting", Output: "(String(value).startsWith)('a')"},
							{MessageId: "useOptionalChaining", Output: "(value?.startsWith)('a')"},
							{MessageId: "useNullishCoalescing", Output: "((value ?? '').startsWith)('a')"},
						},
					},
				},
			},
			{Code: "/^/.test(value)", FileName: "file.ts",
				Output: []string{"value.startsWith('')"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useStringCasting", Output: "String(value).startsWith('')"},
							{MessageId: "useOptionalChaining", Output: "value?.startsWith('')"},
							{MessageId: "useNullishCoalescing", Output: "(value ?? '').startsWith('')"},
						},
					},
				},
			},
			{Code: "/$/.test(value)", FileName: "file.ts",
				Output: []string{"value.endsWith('')"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-ends-with", Message: "Prefer `String#endsWith()` over a regex with `$`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 16,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useStringCasting", Output: "String(value).endsWith('')"},
							{MessageId: "useOptionalChaining", Output: "value?.endsWith('')"},
							{MessageId: "useNullishCoalescing", Output: "(value ?? '').endsWith('')"},
						},
					},
				},
			},
			{Code: "/^a/gs.test(value)", FileName: "file.ts",
				Output: []string{"value.startsWith('a')"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 19,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useStringCasting", Output: "String(value).startsWith('a')"},
							{MessageId: "useOptionalChaining", Output: "value?.startsWith('a')"},
							{MessageId: "useNullishCoalescing", Output: "(value ?? '').startsWith('a')"},
						},
					},
				},
			},
			{Code: "/^a/.test(foo as string)", FileName: "file.ts",
				Output: []string{"(foo as string).startsWith('a')"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 25,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useStringCasting", Output: "String(foo as string).startsWith('a')"},
							{MessageId: "useOptionalChaining", Output: "(foo as string)?.startsWith('a')"},
							{MessageId: "useNullishCoalescing", Output: "(foo as string ?? '').startsWith('a')"},
						},
					},
				},
			},
			{Code: "/** @type {string} */ const text = \"abc\"; text.indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"/** @type {string} */ const text = \"abc\"; text.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 66},
				},
			},
			{Code: "const text: string | null = \"abc\"; text.indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"const text: string | null = \"abc\"; text.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 36, EndLine: 1, EndColumn: 59},
				},
			},
			{Code: "const text: unknown = \"abc\"; text.indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"const text: unknown = \"abc\"; text.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 30, EndLine: 1, EndColumn: 53},
				},
			},
			{Code: "({a:\"abc\"}.a).indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"({a:\"abc\"}.a).startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 33},
				},
			},
			{Code: "const text = \"abc\"; (text.trim()).indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"const text = \"abc\"; (text.trim()).startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 21, EndLine: 1, EndColumn: 53},
				},
			},
			{Code: "const text = [\"a\",\"b\"]; text.join(\"\").indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"const text = [\"a\",\"b\"]; text.join(\"\").startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 25, EndLine: 1, EndColumn: 57},
				},
			},
			{Code: "function f(text: string[]) { return text.join(\"\").indexOf(\"a\") === 0; }", FileName: "file.ts",
				Output: []string{"function f(text: string[]) { return text.join(\"\").startsWith(\"a\"); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 69},
				},
			},
			{Code: "/abc/.source.indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"/abc/.source.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 32},
				},
			},
			{Code: "const re = /abc/; re.source.indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"const re = /abc/; re.source.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 47},
				},
			},
			{Code: "new Date().toISOString().indexOf(\"2\") === 0;", FileName: "file.ts",
				Output: []string{"new Date().toISOString().startsWith(\"2\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 44},
				},
			},
			{Code: "const date = new Date(); date.toISOString().indexOf(\"2\") === 0;", FileName: "file.ts",
				Output: []string{"const date = new Date(); date.toISOString().startsWith(\"2\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 26, EndLine: 1, EndColumn: 63},
				},
			},
			{Code: "/** @type {Date} */ let date; date.toISOString().indexOf(\"2\") === 0;", FileName: "file.ts",
				Output: []string{"/** @type {Date} */ let date; date.toISOString().startsWith(\"2\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 31, EndLine: 1, EndColumn: 68},
				},
			},
			{Code: "let text = \"abc\"; text.indexOf(\"a\") === 0;", FileName: "file.ts",
				Output: []string{"let text = \"abc\"; text.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 19, EndLine: 1, EndColumn: 42},
				},
			},
			{Code: "const text=\"abc\"; const search=/a/; text.indexOf(search) === 0;", FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 63},
				},
			},
			{Code: "const text=\"abc\"; const search=\"a\"; text./* keep */indexOf(search) === 0;", FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 73},
				},
			},
			{Code: "const text=\"abc\"; /* 前😀 */\ntext.slice(0,1) === \"a\";", FileName: "file.ts",
				Output: []string{"const text=\"abc\"; /* 前😀 */\ntext.startsWith(\"a\");"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 2, Column: 1, EndLine: 2, EndColumn: 24},
				},
			},
			{Code: "const X = <div>{/^a/.test(value)}</div>;", FileName: "file.tsx",
				Output: []string{"const X = <div>{value.startsWith('a')}</div>;"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 17, EndLine: 1, EndColumn: 33,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "useStringCasting", Output: "const X = <div>{String(value).startsWith('a')}</div>;"},
							{MessageId: "useOptionalChaining", Output: "const X = <div>{value?.startsWith('a')}</div>;"},
							{MessageId: "useNullishCoalescing", Output: "const X = <div>{(value ?? '').startsWith('a')}</div>;"},
						},
					},
				},
			},
			{Code: "\"a\" === String(read()).slice(0,1)", FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 34,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-string-starts-ends-with/slice-suggestion", Output: "String(read()).startsWith(\"a\")"},
						},
					},
				},
			},
			{Code: "\"a\" === (calls++,\"abc\").slice(0,1)", FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 35,
						Suggestions: []rule_tester.InvalidTestCaseSuggestion{
							{MessageId: "prefer-string-starts-ends-with/slice-suggestion", Output: "(calls++,\"abc\").startsWith(\"a\")"},
						},
					},
				},
			},
			{Code: "\"a\" === String(read()).slice(0, /* keep */1)", FileName: "file.ts",
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 1, EndLine: 1, EndColumn: 45},
				},
			},
			{Code: "/** @type {string} */ const text = \"abc\"; /** @type {string} */ const needle=\"a\"; text.slice(0,needle.length)===needle;", FileName: "file.ts",
				Output: []string{"/** @type {string} */ const text = \"abc\"; /** @type {string} */ const needle=\"a\"; text.startsWith(needle);"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 83, EndLine: 1, EndColumn: 119},
				},
			},
			{Code: "function f(text: string, needle:string) { return text.slice(0, needle.length)===needle; }", FileName: "file.ts",
				Output: []string{"function f(text: string, needle:string) { return text.startsWith(needle); }"},
				Errors: []rule_tester.InvalidTestCaseError{
					{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 50, EndLine: 1, EndColumn: 87},
				},
			},
			// Contextual keywords can also be ordinary expression identifiers.
			{Code: `/^a/.test(async)`, FileName: "file.ts", Output: []string{`async.startsWith('a')`},
				Errors: []rule_tester.InvalidTestCaseError{{
					MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17,
					Suggestions: []rule_tester.InvalidTestCaseSuggestion{
						{MessageId: "useStringCasting", Output: `String(async).startsWith('a')`},
						{MessageId: "useOptionalChaining", Output: `async?.startsWith('a')`},
						{MessageId: "useNullishCoalescing", Output: `(async ?? '').startsWith('a')`},
					},
				}},
			},
			// Literal regex bodies use JavaScript string escaping in fixes.
			{Code: "/^'/.test(\"abc\")", FileName: "file.js", Output: []string{"\"abc\".startsWith('\\'')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}},
			{Code: "/^'\"/.test(\"abc\")", FileName: "file.js", Output: []string{"\"abc\".startsWith('\\'\"')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			{Code: "/^\t/.test(\"abc\")", FileName: "file.js", Output: []string{"\"abc\".startsWith('\\t')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}},
			{Code: "/^\u0000/.test(\"abc\")", FileName: "file.js", Output: []string{"\"abc\".startsWith('\\u{0}')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}},
			{Code: "/^\u001f/.test(\"abc\")", FileName: "file.js", Output: []string{"\"abc\".startsWith('\\u{1f}')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 17}}},
			{Code: "/^😀/.test(\"abc\")", FileName: "file.js", Output: []string{"\"abc\".startsWith('😀')"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 18}}},
			// Function-valued bindings use authored return annotations, without inferring bodies.
			{Code: "const read = (): string => value; read().indexOf(\"a\") === 0;", FileName: "file.ts", Output: []string{"const read = (): string => value; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 60}}},
			{Code: "const read = function (): string { return value }; read().indexOf(\"a\") === 0;", FileName: "file.ts", Output: []string{"const read = function (): string { return value }; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 52, EndLine: 1, EndColumn: 77}}},
			{Code: "/** @returns {string} */ const read = () => value; read().indexOf(\"a\") === 0;", FileName: "file.ts", Output: []string{"/** @returns {string} */ const read = () => value; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 52, EndLine: 1, EndColumn: 77}}},
			{Code: "function read(): string { return value }; read().indexOf(\"a\") === 0;", FileName: "file.ts", Output: []string{"function read(): string { return value }; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 68}}},
			{Code: "let read = (): string => value; read().indexOf(\"a\") === 0;", FileName: "file.ts", Output: []string{"let read = (): string => value; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 33, EndLine: 1, EndColumn: 58}}},
			{Code: "const read = (() : string => value) satisfies (() => string); read().indexOf(\"a\") === 0;", FileName: "file.ts", Output: []string{"const read = (() : string => value) satisfies (() => string); read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 63, EndLine: 1, EndColumn: 88}}},
			// JSDoc callable results and preserved comment-range boundaries.
			{Code: "/** @param {'abc'} value */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts", Output: []string{"/** @param {'abc'} value */ function f(value) { return value.startsWith(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 56, EndLine: 1, EndColumn: 80}}},
			{Code: "/** @type {'abc'} */ const value = \"abc\"; value.indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"/** @type {'abc'} */ const value = \"abc\"; value.startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 67}}},
			{Code: "/** @param {string|string} value */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts", Output: []string{"/** @param {string|string} value */ function f(value) { return value.startsWith(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 64, EndLine: 1, EndColumn: 88}}},
			{Code: "/** @param {!string} value */ function f(value) { return value.indexOf(\"a\") === 0; }", FileName: "review.ts", Output: []string{"/** @param {!string} value */ function f(value) { return value.startsWith(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 58, EndLine: 1, EndColumn: 82}}},
			{Code: "/** @type {(()=>string) | (()=>string)} */ let read; read().indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"/** @type {(()=>string) | (()=>string)} */ let read; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 54, EndLine: 1, EndColumn: 79}}},
			{Code: "/** @type {{(): string; (): string}} */ let read; read().indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"/** @type {{(): string; (): string}} */ let read; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 51, EndLine: 1, EndColumn: 76}}},
			{Code: "/** @type {()=>!string} */ let read; read().indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"/** @type {()=>!string} */ let read; read().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 38, EndLine: 1, EndColumn: 63}}},
			{Code: "(RegExp(\"abc\").source).indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"(RegExp(\"abc\").source).startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 42}}},
			{Code: "declare const read: () => string; (read() as string).indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"declare const read: () => string; (read() as string).startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 72}}},
			{Code: "/* before */ \"abc\".indexOf(\"a\") === 0; /* after */", FileName: "review.ts", Output: []string{"/* before */ \"abc\".startsWith(\"a\"); /* after */"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 38}}},
			{Code: "/* before */ \"abc\".indexOf(\"a\") /* removed */ === 0; /* after */", FileName: "review.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 14, EndLine: 1, EndColumn: 52}}},
			{Code: "let value = \"abc\"; value.trim().indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"let value = \"abc\"; value.trim().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 51}}},
			{Code: "let value = \"abc\"; (value + unknown).indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"let value = \"abc\"; (value + unknown).startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 20, EndLine: 1, EndColumn: 56}}},
			{Code: "(unknown satisfies string).indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"(unknown satisfies string).startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 1, EndLine: 1, EndColumn: 46}}},
			{Code: "/** @type {string} */ const value = \"abc\"; (value || \"x\").indexOf(\"a\") === 0;", FileName: "review.ts", Output: []string{"/** @type {string} */ const value = \"abc\"; (value || \"x\").startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 44, EndLine: 1, EndColumn: 77}}},
			{Code: "function f(value?: string) { return /^a/.test(value); }", FileName: "review.ts", Output: []string{"function f(value?: string) { return value.startsWith('a'); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useStringCasting", Output: "function f(value?: string) { return String(value).startsWith('a'); }"}, {MessageId: "useOptionalChaining", Output: "function f(value?: string) { return value?.startsWith('a'); }"}, {MessageId: "useNullishCoalescing", Output: "function f(value?: string) { return (value ?? '').startsWith('a'); }"}}}}},
			{Code: "/** @param {string} [value] */ function f(value) { return /^a/.test(value); }", FileName: "review.ts", Output: []string{"/** @param {string} [value] */ function f(value) { return value.startsWith('a'); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 59, EndLine: 1, EndColumn: 75, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useStringCasting", Output: "/** @param {string} [value] */ function f(value) { return String(value).startsWith('a'); }"}, {MessageId: "useOptionalChaining", Output: "/** @param {string} [value] */ function f(value) { return value?.startsWith('a'); }"}, {MessageId: "useNullishCoalescing", Output: "/** @param {string} [value] */ function f(value) { return (value ?? '').startsWith('a'); }"}}}}},
			// tsgo exposes a static template literal type as a string-like literal.
			{Code: "function f(value: `abc`) { return value.indexOf(\"a\") === 0; }", FileName: "file.ts", Output: []string{"function f(value: `abc`) { return value.startsWith(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 35, EndLine: 1, EndColumn: 59}}},
			{Code: "function f(value: \"abc\" | `def`) { return value.indexOf(\"a\") === 0; }", FileName: "file.ts", Output: []string{"function f(value: \"abc\" | `def`) { return value.startsWith(\"a\"); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 43, EndLine: 1, EndColumn: 67}}},
			// Preserve getter evaluation counts and the order of captured values.
			{Code: "let reads = 0; const box = { get prefix() { return ++reads === 1 ? \"a\" : \"b\"; } }; const result = \"ab\".slice(0, (box.prefix as string).length) === (box.prefix as string); ({result, reads});", FileName: "safety.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 99, EndLine: 1, EndColumn: 170}}},
			{Code: "function f(text: string, box: {prefix: string}) { return text.slice(0, (box.prefix as string).length) === (box.prefix as string); }", FileName: "safety.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 58, EndLine: 1, EndColumn: 129}}},
			{Code: "function f(text: string, box: {prefix: string}, key: string) { return text.slice(0, (box[key] as string).length) === (box[key] as string); }", FileName: "safety.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 71, EndLine: 1, EndColumn: 138}}},
			{Code: "let prefix: string = \"a\"; const box = { get text() { prefix = \"b\"; return \"ab\"; } }; const result = prefix === (box.text as string).slice(0, prefix.length); result;", FileName: "safety.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 101, EndLine: 1, EndColumn: 156, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "prefer-string-starts-ends-with/slice-suggestion", Output: "let prefix: string = \"a\"; const box = { get text() { prefix = \"b\"; return \"ab\"; } }; const result = (box.text as string).startsWith(prefix); result;"}}}}},
			{Code: "function String(): string { return \"abc\"; } String().indexOf(\"a\") === 0;", FileName: "safety.ts", Output: []string{"function String(): string { return \"abc\"; } String().startsWith(\"a\");"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.", Line: 1, Column: 45, EndLine: 1, EndColumn: 72}}},
			{Code: "function f(text: string, prefix: string) { return text.slice(0, prefix.length) === prefix; }", FileName: "safety.ts", Output: []string{"function f(text: string, prefix: string) { return text.startsWith(prefix); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with-slice", Message: "Prefer `String#startsWith()` over comparing a string slice.", Line: 1, Column: 51, EndLine: 1, EndColumn: 90}}},
			{Code: "function f() { const suffix = \"c\"; return \"abc\".slice(-suffix.length) === suffix; }", FileName: "safety.ts", Output: []string{"function f() { const suffix = \"c\"; return \"abc\".endsWith(suffix); }"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-ends-with-slice", Message: "Prefer `String#endsWith()` over comparing a string slice.", Line: 1, Column: 43, EndLine: 1, EndColumn: 81}}},
			{Code: "function String() { return [\"a\"]; } /^a/.test(value);", FileName: "safety.ts", Output: []string{"function String() { return [\"a\"]; } value.startsWith('a');"}, Errors: []rule_tester.InvalidTestCaseError{{MessageId: "prefer-starts-with", Message: "Prefer `String#startsWith()` over a regex with `^`.", Line: 1, Column: 37, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "useOptionalChaining", Output: "function String() { return [\"a\"]; } value?.startsWith('a');"}, {MessageId: "useNullishCoalescing", Output: "function String() { return [\"a\"]; } (value ?? '').startsWith('a');"}}}}},
		},
	)
}

func TestPreferStringStartsEndsWithEditDemand(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		code, fixed string
		suggestions []string
	}{
		{`/^a/.test(value)`, `value.startsWith('a')`, []string{
			`String(value).startsWith('a')`, `value?.startsWith('a')`, `(value ?? '').startsWith('a')`,
		}},
		{`"abc".indexOf("a") === 0`, `"abc".startsWith("a")`, nil},
		{`"a" === String(read()).slice(0, 1)`, "", []string{`String(read()).startsWith("a")`}},
		{`"ab".slice(0, (box.prefix as string).length) === (box.prefix as string)`, "", nil},
		{`"a" === (box.text as string).slice(0, 1)`, "", []string{`(box.text as string).startsWith("a")`}},
	} {
		t.Run(test.code, func(t *testing.T) {
			t.Parallel()
			source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/edit-demand.ts", Path: "/edit-demand.ts"}, test.code, core.ScriptKindTS)
			run := func(demand rule.EditDemand) rule.RuleDiagnostic {
				t.Helper()
				var diagnostics []rule.RuleDiagnostic
				comments := rule.NewCommentStore(source)
				ctx := rule.RuleContext{SourceFile: source, Comments: comments, DisableManager: rule.NewDisableManager(source, comments)}.WithDiagnosticConsumer(
					prefer_string_starts_ends_with.PreferStringStartsEndsWithRule.Name, rule.SeverityError,
					rule.DiagnosticConsumer{Demand: demand, Report: func(diagnostic rule.RuleDiagnostic) { diagnostics = append(diagnostics, diagnostic) }},
				)
				listeners := prefer_string_starts_ends_with.PreferStringStartsEndsWithRule.Run(ctx, nil)
				var visit ast.Visitor
				visit = func(node *ast.Node) bool {
					if listener := listeners[node.Kind]; listener != nil {
						listener(node)
					}
					return node.ForEachChild(visit)
				}
				source.AsNode().ForEachChild(visit)
				if len(diagnostics) != 1 {
					t.Fatalf("demand %d: got %d diagnostics, want 1", demand, len(diagnostics))
				}
				return diagnostics[0]
			}
			all := run(rule.EditDemandAll)
			for _, demand := range []rule.EditDemand{rule.EditDemandNone, rule.EditDemandAutofix, rule.EditDemandSuggestion} {
				diagnostic := run(demand)
				metadata, want := diagnostic, all
				metadata.FixesPtr, want.FixesPtr = nil, nil
				metadata.Suggestions, want.Suggestions = nil, nil
				if !reflect.DeepEqual(metadata, want) {
					t.Errorf("demand %d changed diagnostic metadata", demand)
				}
				if demand&rule.EditDemandAutofix == 0 {
					if diagnostic.FixesPtr != nil {
						t.Errorf("demand %d unexpectedly built autofixes", demand)
					}
				} else if !reflect.DeepEqual(diagnostic.FixesPtr, all.FixesPtr) {
					t.Errorf("demand %d changed autofixes", demand)
				}
				if demand&rule.EditDemandSuggestion == 0 {
					if diagnostic.Suggestions != nil && len(*diagnostic.Suggestions) != 0 {
						t.Errorf("demand %d unexpectedly built suggestions", demand)
					}
				} else if !reflect.DeepEqual(diagnostic.Suggestions, all.Suggestions) {
					t.Errorf("demand %d changed suggestions", demand)
				}
			}
			output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{all})
			if fixed != (test.fixed != "") || (fixed && output != test.fixed) {
				t.Errorf("autofix = %q, fixed = %v; want %q", output, fixed, test.fixed)
			}
			var suggestions []rule.RuleSuggestion
			if all.Suggestions != nil {
				suggestions = *all.Suggestions
			}
			if len(suggestions) != len(test.suggestions) {
				t.Fatalf("got %d suggestions, want %d", len(suggestions), len(test.suggestions))
			}
			for i, suggestion := range suggestions {
				wantID, wantDescription := "prefer-string-starts-ends-with/slice-suggestion", "Replace with `String#startsWith()`."
				if len(test.suggestions) == 3 {
					wantID = []string{"useStringCasting", "useOptionalChaining", "useNullishCoalescing"}[i]
					wantDescription = []string{
						"Convert to string `String(…).startsWith()`.",
						"Use optional chaining `…?.startsWith()`.",
						"Use nullish coalescing `(… ?? '').startsWith()`.",
					}[i]
				}
				if suggestion.Message.Id != wantID || suggestion.Message.Description != wantDescription || !reflect.DeepEqual(suggestion.Message.Data, map[string]string{"method": "startsWith"}) {
					t.Errorf("suggestion %d message = %+v", i, suggestion.Message)
				}
				diagnostic := all
				diagnostic.FixesPtr = &suggestion.FixesArr
				output, _, fixed := linter.ApplyRuleFixes(test.code, []rule.RuleDiagnostic{diagnostic})
				if !fixed || output != test.suggestions[i] {
					t.Errorf("suggestion %d = %q, want %q", i, output, test.suggestions[i])
				}
			}
		})
	}
}

func TestPreferStringStartsEndsWithRepeatedAliases(t *testing.T) {
	// Repeated references must share their proof instead of expanding a
	// conditional alias tree exponentially.
	var source strings.Builder
	source.WriteString(`const value0 = "abc";`)
	for i := 1; i <= 40; i++ {
		fmt.Fprintf(&source, "const value%d = flag ? value%d : value%d;", i, i-1, i-1)
	}
	prefix := source.String()
	code := prefix + `value40.indexOf("a") === 0;`
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t,
		&prefer_string_starts_ends_with.PreferStringStartsEndsWithRule, nil,
		[]rule_tester.InvalidTestCase{{
			Code: code, FileName: "file.js", Output: []string{prefix + `value40.startsWith("a");`},
			Errors: []rule_tester.InvalidTestCaseError{{
				MessageId: "prefer-starts-with-indexOf", Message: "Prefer `String#startsWith()` over `String#indexOf() === 0`.",
				Line: 1, Column: len(prefix) + 1, EndLine: 1, EndColumn: len(code),
			}},
		}})
}
