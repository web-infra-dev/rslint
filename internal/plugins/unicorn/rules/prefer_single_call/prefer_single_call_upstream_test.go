// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v77.0.0/test/prefer-single-call.js
// Includes every upstream test and runnable documentation example. Expected edits
// and complete diagnostic ranges were checked with ESLint 10.9.0 and Unicorn 77.0.0.
package prefer_single_call_test

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/prefer_single_call"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestPreferSingleCallUpstream(t *testing.T) {
	t.Run("Array push", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "function f(foo: {push(value: number): void}) { foo.push(1); foo.push(2); }", FileName: "case.ts"},
			{Code: "function makeSink() { return {push(value: number) {}}; } const sink = makeSink(); sink.push(1); sink.push(2);", FileName: "case.ts"},
			{Code: "foo.forEach(fn);\nfoo.forEach(fn);", FileName: "case.js"},
			{Code: "foo.push(1);", FileName: "case.js"},
			{Code: "foo.push(1);\nfoo.unshift(2);", FileName: "case.js"},
			{Code: "foo.push(1);; // <- there is an \"EmptyStatement\" between\nfoo.push(2);", FileName: "case.js"},
			{Code: "foo.push(1);\nbar.push(2);", FileName: "case.js"},
			{Code: "foo.push(1);push(2)", FileName: "case.js"},
			{Code: "push(1);foo.push(2)", FileName: "case.js"},
			{Code: "new foo.push(1);foo.push(2)", FileName: "case.js"},
			{Code: "foo.push(1);new foo.push(2)", FileName: "case.js"},
			{Code: "foo[push](1);foo.push(2)", FileName: "case.js"},
			{Code: "foo.push(1);foo[push](2)", FileName: "case.js"},
			{Code: "foo.push(foo.push(1));", FileName: "case.js"},
			{Code: "const length = foo.push(1);\nfoo.push(2);", FileName: "case.js"},
			{Code: "foo.push(1);\nconst length = foo.push(2);", FileName: "case.js"},
			{Code: "foo().push(1);\nfoo().push(2);", FileName: "case.js"},
			{Code: "foo().bar.push(1);\nfoo().bar.push(2);", FileName: "case.js"},
			{Code: "const stream = new Readable();\nstream.push('one string');\nstream.push('another string');", FileName: "case.js"},
			{Code: "class FooReadable extends Readable {\n\tpushAndEnd(chunk) {\n\t\tthis.push(chunk);\n\t\tthis.push(null);\n\t}\n}", FileName: "case.js"},
			{Code: "class Foo {\n\tpushAndEnd(chunk) {\n\t\tthis.stream.push(chunk);\n\t\tthis.stream.push(null);\n\t}\n}", FileName: "case.js"},
			{Code: "process.stdin.push(chunk);\nprocess.stdin.push(null);", FileName: "case.js"},
			{Code: "process.stdout.push(chunk);\nprocess.stdout.push(null);", FileName: "case.js"},
			{Code: "process.stderr.push(chunk);\nprocess.stderr.push(null);", FileName: "case.js"},
			{Code: "foo.push(1);\nfoo.push(2);\nfoo.bar.push(1);\nfoo.bar.push(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"foo.push", "foo.bar.push"}}}},
			{Code: "for (const _ of []) foo.push(bar);", FileName: "case.js"},
			{Code: "function bar() {}\nfoo.push(bindEvents);", FileName: "case.js"},
			{Code: "foo.push?.(1);\nfoo.push?.(2);", FileName: "case.js"},
			{Code: "foo.push(1);\nfoo.push?.(2);", FileName: "case.js"},
			{Code: "foo.push?.(1);\nfoo.push(2);", FileName: "case.js"},
			{Code: "const foo = new Foo(); foo.push(1); foo.push(2);", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "const object = {value: 0};\nObject.defineProperty(object, 'value', {get() { return foo.length; }});\nfoo.push(1);\nfoo.push(object.value);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 4, Column: 5, EndLine: 4, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const object = {value: 0};\nObject.defineProperty(object, 'value', {get() { return foo.length; }});\nfoo.push(1, object.value);"}}},
			}},
			{Code: "foo.push(1);\nfoo.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2);"}}},
			}},
			{Code: "(foo.push)(1);\n(foo.push)(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(foo.push)(1, 2);"}}},
			}},
			{Code: "foo.bar.push(1);\nfoo.bar.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 9, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.bar.push(1, 2);"}}},
			}},
			{Code: "foo.push(1);\n(foo).push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 7, EndLine: 2, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2);"}}},
			}},
			{Code: "foo.push();\nfoo.push();", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push();"}}},
			}},
			{Code: "foo.push(1);\nfoo.push();", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1);"}}},
			}},
			{Code: "foo.push();\nfoo.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(2);"}}},
			}},
			{Code: "foo.push(1, 2);\nfoo.push((3), (4));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2, (3), (4));"}}},
			}},
			{Code: "foo.push(1, 2,);\nfoo.push(3, 4);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2, 3, 4);"}}},
			}},
			{Code: "foo.push(1, 2);\nfoo.push(3, 4,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2, 3, 4,);"}}},
			}},
			{Code: "foo.push(1, 2,);\nfoo.push(3, 4,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2, 3, 4,);"}}},
			}},
			{Code: "foo.push(1, 2, ...a,);\nfoo.push(...b,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2, ...a, ...b,);"}}},
			}},
			{Code: "foo.push(bar());\nfoo.push(1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(bar(), 1);"}}},
			}},
			{Code: "foo.push(1);\nfoo.push(bar());", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, bar());"}}},
			}},
			{Code: "foo.push(1,);\nfoo.push(2,);\nfoo.push(3,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2,);\nfoo.push(3,);"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 5, EndLine: 3, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1,);\nfoo.push(2, 3,);"}}},
			}},
			{Code: "if (a) {\n\tfoo.push(1);\n\tfoo.push(2);\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 6, EndLine: 3, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (a) {\n\tfoo.push(1, 2);\n}"}}},
			}},
			{Code: "switch (a) {\n\tdefault:\n\t\tfoo.push(1);\n\t\tfoo.push(2);\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 4, Column: 7, EndLine: 4, EndColumn: 11, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "switch (a) {\n\tdefault:\n\t\tfoo.push(1, 2);\n}"}}},
			}},
			{Code: "function a() {\n\tfoo.push(1);\n\tfoo.push(2);\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 6, EndLine: 3, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function a() {\n\tfoo.push(1, 2);\n}"}}},
			}},
			{Code: "foo.push(1)\nfoo.push(2)\n;[foo].forEach(bar)", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2);[foo].forEach(bar)"}}},
			}},
			{Code: "foo.bar.push(1);\n(foo)['bar'].push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 14, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.bar.push(1, 2);"}}},
			}},
			{Code: "foo.push(1);\nfoo.push(2);\nstream.push(1);\nstream.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2);\nstream.push(1);\nstream.push(2);"}}},
			}},
			{Code: "foo.bar.push(1);\nfoo.bar.push(2);\nfoo.push(1);\nfoo.push(2);\nbar.foo.push(1);\nbar.foo.push(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"foo", "foo.bar"}}}, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 9, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.bar.push(1, 2);\nfoo.push(1);\nfoo.push(2);\nbar.foo.push(1);\nbar.foo.push(2);"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 4, Column: 5, EndLine: 4, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.bar.push(1);\nfoo.bar.push(2);\nfoo.push(1, 2);\nbar.foo.push(1);\nbar.foo.push(2);"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 6, Column: 9, EndLine: 6, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.bar.push(1);\nfoo.bar.push(2);\nfoo.push(1);\nfoo.push(2);\nbar.foo.push(1, 2);"}}},
			}},
			{Code: "foo.push(1);\nfoo?.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.push(1, 2);"}}},
			}},
			{Code: "foo?.push(1);\nfoo.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo?.push(1, 2);"}}},
			}},
			{Code: "foo?.push(1);\nfoo?.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo?.push(1, 2);"}}},
			}},
			{Code: "foo?.bar.push(1);\nfoo?.bar.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 10, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo?.bar.push(1, 2);"}}},
			}},
			{Code: "(foo as any[]).push(1);\n(foo as any[]).push(2);", FileName: "case.ts", Output: []string{"(foo as any[]).push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 16, EndLine: 2, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo!.push(1);\nfoo!.push(2);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo!.push(1, 2);"}}},
			}},
			{Code: "function f(foo: number[]) { foo.push(1); foo.push(2); }", FileName: "case.ts", Output: []string{"function f(foo: number[]) { foo.push(1, 2); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 46, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			// rslint always supplies TypeScript type information; the identical type-aware upstream case above/below covers this input.
			{Code: "const array = [0].map(value => value); array.push(1); array.push(2);", FileName: "case.ts", Skip: true, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 61, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const array = [0].map(value => value); array.push(1, 2);"}}},
			}},
			{Code: "const array = [0].map(value => value); array.push(1); array.push(2);", FileName: "case.ts", Output: []string{"const array = [0].map(value => value); array.push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 61, EndLine: 1, EndColumn: 65, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "declare const receiver: number[] | {push(value: number): void}; receiver.push(1); receiver.push(2);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 92, EndLine: 1, EndColumn: 96, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "declare const receiver: number[] | {push(value: number): void}; receiver.push(1, 2);"}}},
			}},
			// rslint always supplies TypeScript type information; the identical type-aware upstream case above/below covers this input.
			{Code: "function makeSink() { return {push(value: number) {}}; } const sink = makeSink(); sink.push(1); sink.push(2);", FileName: "case.ts", Skip: true, Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 102, EndLine: 1, EndColumn: 106, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function makeSink() { return {push(value: number) {}}; } const sink = makeSink(); sink.push(1, 2);"}}},
			}},
			{Code: "const container = {data: {entries: {push(value) { console.log(value); }}}}; container.data.entries.push(1); container.data.entries.push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 132, EndLine: 1, EndColumn: 136, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const container = {data: {entries: {push(value) { console.log(value); }}}}; container.data.entries.push(1, 2);"}}},
			}},
			{Code: "const values = []; values.push(1); values.push(2);", FileName: "case.js", Output: []string{"const values = []; values.push(1, 2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 43, EndLine: 1, EndColumn: 47, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Argument evaluation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const array = []; array.push(1); array.push(array.length);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 40, EndLine: 1, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const array = []; array.push(1, array.length);"}}},
			}},
			{Code: "function f(array: unknown[], value: unknown) { array.push(1); array.push(value); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 69, EndLine: 1, EndColumn: 73, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(array: unknown[], value: unknown) { array.push(1, value); }"}}},
			}},
			{Code: "function f(array: unknown[], value: unknown) { array.push(value); array.push(2); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 73, EndLine: 1, EndColumn: 77, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(array: unknown[], value: unknown) { array.push(value, 2); }"}}},
			}},
			{Code: "const array = [1]; array.push(2); array.push(...array);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 41, EndLine: 1, EndColumn: 45, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const array = [1]; array.push(2, ...array);"}}},
			}},
			{Code: "function f(array: unknown[]) { array.push(array = []); array.push(2); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 62, EndLine: 1, EndColumn: 66, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(array: unknown[]) { array.push(array = [], 2); }"}}},
			}},
			{Code: "function f(array: unknown[]) { const values = { *[Symbol.iterator]() { array = []; yield 1; } }; array.push(...values); array.push(2); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 127, EndLine: 1, EndColumn: 131, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(array: unknown[]) { const values = { *[Symbol.iterator]() { array = []; yield 1; } }; array.push(...values, 2); }"}}},
			}},
			{Code: "function f(array: unknown[]) { const values = { *[Symbol.iterator]() { array = []; yield 1; } }; array.push([...values]); array.push(2); }", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 1, Column: 129, EndLine: 1, EndColumn: 133, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function f(array: unknown[]) { const values = { *[Symbol.iterator]() { array = []; yield 1; } }; array.push([...values], 2); }"}}},
			}},
		})
	})
	t.Run("Array unshift", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "function f(foo: {unshift(value: number): void}) { foo.unshift(1); foo.unshift(2); }", FileName: "case.ts"},
			{Code: "foo.unshift(1);", FileName: "case.js"},
			{Code: "foo.push(1);\nfoo.unshift(2);", FileName: "case.js"},
			{Code: "foo.unshift(1);\nfoo.push(2);", FileName: "case.js"},
			{Code: "foo.unshift(1);; // <- there is an \"EmptyStatement\" between\nfoo.unshift(2);", FileName: "case.js"},
			{Code: "foo.unshift(1);\nbar.unshift(2);", FileName: "case.js"},
			{Code: "foo.unshift(1);unshift(2)", FileName: "case.js"},
			{Code: "unshift(1);foo.unshift(2)", FileName: "case.js"},
			{Code: "new foo.unshift(1);foo.unshift(2)", FileName: "case.js"},
			{Code: "foo.unshift(1);new foo.unshift(2)", FileName: "case.js"},
			{Code: "foo[unshift](1);foo.unshift(2)", FileName: "case.js"},
			{Code: "foo.unshift(1);foo[unshift](2)", FileName: "case.js"},
			{Code: "foo.unshift(foo.unshift(1));", FileName: "case.js"},
			{Code: "const length = foo.unshift(1);\nfoo.unshift(2);", FileName: "case.js"},
			{Code: "foo.unshift(1);\nconst length = foo.unshift(2);", FileName: "case.js"},
			{Code: "foo().unshift(1);\nfoo().unshift(2);", FileName: "case.js"},
			{Code: "foo().bar.unshift(1);\nfoo().bar.unshift(2);", FileName: "case.js"},
			{Code: "const stream = new Readable();\nstream.unshift('one string');\nstream.unshift('another string');", FileName: "case.js"},
			{Code: "class FooReadable extends Readable {\n\tunshiftAndEnd(chunk) {\n\t\tthis.unshift(chunk);\n\t\tthis.unshift(null);\n\t}\n}", FileName: "case.js"},
			{Code: "class Foo {\n\tunshiftAndEnd(chunk) {\n\t\tthis.stream.unshift(chunk);\n\t\tthis.stream.unshift(null);\n\t}\n}", FileName: "case.js"},
			{Code: "process.stdin.unshift(chunk);\nprocess.stdin.unshift(null);", FileName: "case.js"},
			{Code: "process.stdout.unshift(chunk);\nprocess.stdout.unshift(null);", FileName: "case.js"},
			{Code: "process.stderr.unshift(chunk);\nprocess.stderr.unshift(null);", FileName: "case.js"},
			{Code: "foo.unshift(1);\nfoo.unshift(2);\nfoo.bar.unshift(1);\nfoo.bar.unshift(2);", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"foo.unshift", "foo.bar.unshift"}}}},
			{Code: "for (const _ of []) foo.unshift(bar);", FileName: "case.js"},
			{Code: "function bar() {}\nfoo.unshift(bindEvents);", FileName: "case.js"},
			{Code: "foo.unshift?.(1);\nfoo.unshift?.(2);", FileName: "case.js"},
			{Code: "foo.unshift(1);\nfoo.unshift?.(2);", FileName: "case.js"},
			{Code: "foo.unshift?.(1);\nfoo.unshift(2);", FileName: "case.js"},
			{Code: "const foo = new Foo(); foo.unshift(1); foo.unshift(2);", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "foo.unshift(1);\nfoo.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2, 1);"}}},
			}},
			{Code: "(foo.unshift)(1);\n(foo.unshift)(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(foo.unshift)(2, 1);"}}},
			}},
			{Code: "foo.bar.unshift(1);\nfoo.bar.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 9, EndLine: 2, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.bar.unshift(2, 1);"}}},
			}},
			{Code: "foo.unshift(1);\n(foo).unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 7, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(foo).unshift(2, 1);"}}},
			}},
			{Code: "bar()\nfoo.unshift(1);\n(foo).unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 7, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "bar()\n;(foo).unshift(2, 1);"}}},
			}},
			{Code: "foo.unshift();\nfoo.unshift();", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift();"}}},
			}},
			{Code: "foo.unshift(1);\nfoo.unshift();", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(1);"}}},
			}},
			{Code: "foo.unshift();\nfoo.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2);"}}},
			}},
			{Code: "foo.unshift(1, 2);\nfoo.unshift((3), (4));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift((3), (4), 1, 2);"}}},
			}},
			{Code: "foo.unshift(1, 2,);\nfoo.unshift(3, 4);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(3, 4, 1, 2,);"}}},
			}},
			{Code: "foo.unshift(1, 2);\nfoo.unshift(3, 4,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(3, 4, 1, 2);"}}},
			}},
			{Code: "foo.unshift(1, 2,);\nfoo.unshift(3, 4,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(3, 4, 1, 2,);"}}},
			}},
			{Code: "foo.unshift(1, 2, ...a,);\nfoo.unshift(...b,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(...b, 1, 2, ...a,);"}}},
			}},
			{Code: "foo.unshift(bar());\nfoo.unshift(1);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(1, bar());"}}},
			}},
			{Code: "foo.unshift(1);\nfoo.unshift(bar());", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(bar(), 1);"}}},
			}},
			{Code: "foo.unshift(x);\nfoo.unshift(foo.length);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(foo.length, x);"}}},
			}},
			{Code: "foo.unshift(1);\n// Keep this comment\nfoo.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 5, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.unshift(1,);\nfoo.unshift(2,);\nfoo.unshift(3,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2, 1,);\nfoo.unshift(3,);"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 5, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(1,);\nfoo.unshift(3, 2,);"}}},
			}},
			{Code: "if (a) {\n\tfoo.unshift(1);\n\tfoo.unshift(2);\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 6, EndLine: 3, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "if (a) {\n\tfoo.unshift(2, 1);\n}"}}},
			}},
			{Code: "switch (a) {\n\tdefault:\n\t\tfoo.unshift(1);\n\t\tfoo.unshift(2);\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 4, Column: 7, EndLine: 4, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "switch (a) {\n\tdefault:\n\t\tfoo.unshift(2, 1);\n}"}}},
			}},
			{Code: "function a() {\n\tfoo.unshift(1);\n\tfoo.unshift(2);\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 6, EndLine: 3, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "function a() {\n\tfoo.unshift(2, 1);\n}"}}},
			}},
			{Code: "foo.unshift(1)\nfoo.unshift(2)\n;[foo].forEach(bar)", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2, 1)\n;[foo].forEach(bar)"}}},
			}},
			{Code: "foo.bar.unshift(1);\n(foo)['bar'].unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 14, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "(foo)['bar'].unshift(2, 1);"}}},
			}},
			{Code: "foo.unshift(1);\nfoo?.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2, 1);"}}},
			}},
			{Code: "foo.unshift(1);\nfoo?.unshift(2,);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2, 1);"}}},
			}},
			{Code: "foo?.unshift(1);\nfoo.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 5, EndLine: 2, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.unshift(2, 1);"}}},
			}},
			{Code: "foo?.unshift(1);\nfoo?.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo?.unshift(2, 1);"}}},
			}},
			{Code: "foo?.bar.unshift(1);\nfoo?.bar.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 10, EndLine: 2, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo?.bar.unshift(2, 1);"}}},
			}},
			{Code: "(foo as any[]).unshift(1);\n(foo as any[]).unshift(2);", FileName: "case.ts", Output: []string{"(foo as any[]).unshift(2, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 16, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo!.unshift(1);\nfoo!.unshift(2);", FileName: "case.ts", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 13, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo!.unshift(2, 1);"}}},
			}},
			{Code: "function f(foo: number[]) { foo.unshift(1); foo.unshift(2); }", FileName: "case.ts", Output: []string{"function f(foo: number[]) { foo.unshift(2, 1); }"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 49, EndLine: 1, EndColumn: 56, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const container = {data: {entries: {unshift(value) { console.log(value); }}}}; container.data.entries.unshift(1); container.data.entries.unshift(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 138, EndLine: 1, EndColumn: 145, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const container = {data: {entries: {unshift(value) { console.log(value); }}}}; container.data.entries.unshift(2, 1);"}}},
			}},
			{Code: "const values = []; values.unshift(1); values.unshift(2);", FileName: "case.js", Output: []string{"const values = []; values.unshift(2, 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 46, EndLine: 1, EndColumn: 53, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "const array = []; array.unshift(1); array.unshift(array.length);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 1, Column: 43, EndLine: 1, EndColumn: 50, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const array = []; array.unshift(array.length, 1);"}}},
			}},
		})
	})
	t.Run("classList", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "foo.classList.toggle('foo');\nfoo.classList.toggle('bar');", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nfoo.classList.remove(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");; // <- there is an \"EmptyStatement\" between\nfoo.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nbar.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");add(\"bar\")", FileName: "case.js"},
			{Code: "add(\"foo\");foo.classList(\"bar\")", FileName: "case.js"},
			{Code: "new foo.classList.add(\"foo\");foo.classList.add(\"bar\")", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");new foo.classList.add(\"bar\")", FileName: "case.js"},
			{Code: "foo.classList[add](\"foo\");foo.classList.add(\"bar\")", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");foo.classList[add](\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(foo.classList.add(\"foo\"));", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nfoo[classList].add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nclassList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\n(new foo.classList).add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nfoo.classList.add?.(\"bar\");", FileName: "case.js"},
			{Code: "foo.notClassList.add(\"foo\");\nfoo.notClassList.add(\"bar\");", FileName: "case.js"},
			{Code: "classList.add(\"foo\");\nclassList.add(\"bar\");", FileName: "case.js"},
			{Code: "const _ = foo.classList.add(\"foo\");\nfoo.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nconst _ = foo.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo().classList.add(\"foo\");\nfoo().classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo().bar.classList.add(\"foo\");\nfoo().bar.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList?.add(\"foo\");\nfoo.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add(\"foo\");\nfoo.classList?.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.add?.(\"foo\");\nfoo.classList.add(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList?.remove(\"foo\");\nfoo.classList.remove(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.remove(\"foo\");\nfoo.classList?.remove(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.remove?.(\"foo\");\nfoo.classList.remove(\"bar\");", FileName: "case.js"},
			{Code: "foo.classList.remove(\"foo\");\nfoo.classList.remove?.(\"bar\");", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "foo.classList.add(\"foo\");\nfoo.classList.add(\"bar\");", FileName: "case.js", Output: []string{"foo.classList.add(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.remove(\"foo\");\nfoo.classList.remove(\"bar\");", FileName: "case.js", Output: []string{"foo.classList.remove(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.remove()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 21, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "(foo.classList.add)(\"foo\");\n(foo.classList.add)(\"bar\");", FileName: "case.js", Output: []string{"(foo.classList.add)(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 16, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.bar.classList.add(\"foo\");\nfoo.bar.classList.add(\"bar\");", FileName: "case.js", Output: []string{"foo.bar.classList.add(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 19, EndLine: 2, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(\"foo\");\n(foo).classList.add(\"bar\");", FileName: "case.js", Output: []string{"foo.classList.add(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 17, EndLine: 2, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add();\nfoo.classList.add();", FileName: "case.js", Output: []string{"foo.classList.add();"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(\"foo\");\nfoo.classList.add();", FileName: "case.js", Output: []string{"foo.classList.add(\"foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add();\nfoo.classList.add(2);", FileName: "case.js", Output: []string{"foo.classList.add(2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(a, b);\nfoo.classList.add((c), (d));", FileName: "case.js", Output: []string{"foo.classList.add(a, b, (c), (d));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add.push(a, b,);\nfoo.classList.add.push(c, d);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 19, EndLine: 2, EndColumn: 23, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.classList.add.push(a, b, c, d);"}}},
			}},
			{Code: "foo.classList.add(a, b);\nfoo.classList.add(c, d,);", FileName: "case.js", Output: []string{"foo.classList.add(a, b, c, d,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(a, b,);\nfoo.classList.add(c, d,);", FileName: "case.js", Output: []string{"foo.classList.add(a, b, c, d,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(a, b, ...c,);\nfoo.classList.add(...d,);", FileName: "case.js", Output: []string{"foo.classList.add(a, b, ...c, ...d,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(bar());\nfoo.classList.add(\"foo\");", FileName: "case.js", Output: []string{"foo.classList.add(bar(), \"foo\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(a);\nfoo.classList.add(bar());", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "foo.classList.add(a, bar());"}}},
			}},
			{Code: "foo.classList.add(a,);\nfoo.classList.add(b,);\nfoo.classList.add(c,);", FileName: "case.js", Output: []string{"foo.classList.add(a, b,);\nfoo.classList.add(c,);", "foo.classList.add(a, b, c,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 3, Column: 15, EndLine: 3, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (a) {\n\tfoo.classList.add(a);\n\tfoo.classList.add(b);\n}", FileName: "case.js", Output: []string{"if (a) {\n\tfoo.classList.add(a, b);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 3, Column: 16, EndLine: 3, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "switch (a) {\n\tdefault:\n\t\tfoo.classList.add(a);\n\t\tfoo.classList.add(b);\n}", FileName: "case.js", Output: []string{"switch (a) {\n\tdefault:\n\t\tfoo.classList.add(a, b);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 4, Column: 17, EndLine: 4, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function _() {\n\tfoo.classList.add(a);\n\tfoo.classList.add(b);\n}", FileName: "case.js", Output: []string{"function _() {\n\tfoo.classList.add(a, b);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 3, Column: 16, EndLine: 3, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(a)\nfoo.classList.add(b)\n;[foo].forEach(bar)", FileName: "case.js", Output: []string{"foo.classList.add(a, b);[foo].forEach(bar)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.bar.classList.add(a);\n(foo)['bar'].classList.add(b);", FileName: "case.js", Output: []string{"foo.bar.classList.add(a, b);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 24, EndLine: 2, EndColumn: 27, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo?.classList.add(\"foo\");\nfoo.classList.add(\"bar\");", FileName: "case.js", Output: []string{"foo?.classList.add(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(\"foo\");\nfoo?.classList.add(\"bar\");", FileName: "case.js", Output: []string{"foo.classList.add(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 16, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo?.classList.add(\"foo\");\nfoo?.classList.add(\"bar\");", FileName: "case.js", Output: []string{"foo?.classList.add(\"foo\", \"bar\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 16, EndLine: 2, EndColumn: 19, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("importScripts", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "importScripts('foo.js');\nnotImportScripts('bar.js');", FileName: "case.js"},
			{Code: "importScripts(\"foo.js\");", FileName: "case.js"},
			{Code: "importScripts(\"foo.js\");; // <- there is an \"EmptyStatement\" between\nimportScripts(\"bar.js\");", FileName: "case.js"},
			{Code: "new importScripts(\"foo.js\");importScripts(\"bar.js\")", FileName: "case.js"},
			{Code: "importScripts(\"foo.js\");new importScripts(\"bar.js\")", FileName: "case.js"},
			{Code: "const _ = importScripts(\"foo.js\");\nimportScripts(\"bar.js\");", FileName: "case.js"},
			{Code: "importScripts(\"foo.js\");\nconst _ = importScripts(\"bar.js\");", FileName: "case.js"},
			{Code: "importScripts(\"foo.js\");\nimportScripts(\"bar.js\");", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"importScripts"}}}},
		}, []rule_tester.InvalidTestCase{
			{Code: "importScripts(\"foo.js\");\nimportScripts(\"bar.js\");", FileName: "case.js", Output: []string{"importScripts(\"foo.js\", \"bar.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "(importScripts)(\"foo.js\");\n(importScripts)(\"bar.js\");", FileName: "case.js", Output: []string{"(importScripts)(\"foo.js\", \"bar.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 2, EndLine: 2, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts();\nimportScripts();", FileName: "case.js", Output: []string{"importScripts();"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(\"foo.js\");\nimportScripts();", FileName: "case.js", Output: []string{"importScripts(\"foo.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts();\nimportScripts(2);", FileName: "case.js", Output: []string{"importScripts(2);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a, b);\nimportScripts((c), (d));", FileName: "case.js", Output: []string{"importScripts(a, b, (c), (d));"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a, b,);\nimportScripts(c, d);", FileName: "case.js", Output: []string{"importScripts(a, b, c, d);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a, b);\nimportScripts(c, d,);", FileName: "case.js", Output: []string{"importScripts(a, b, c, d,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a, b,);\nimportScripts(c, d,);", FileName: "case.js", Output: []string{"importScripts(a, b, c, d,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "foo.classList.add(a, b, ...c,);\nfoo.classList.add(...d,);", FileName: "case.js", Output: []string{"foo.classList.add(a, b, ...c, ...d,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 2, Column: 15, EndLine: 2, EndColumn: 18, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(bar());\nimportScripts(\"foo.js\");", FileName: "case.js", Output: []string{"importScripts(bar(), \"foo.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a);\nimportScripts(bar());", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "importScripts(a, bar());"}}},
			}},
			{Code: "importScripts(a,);\nimportScripts(b,);\nimportScripts(c,);", FileName: "case.js", Output: []string{"importScripts(a, b,);\nimportScripts(c,);", "importScripts(a, b, c,);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "if (a) {\n\timportScripts(a);\n\timportScripts(b);\n}", FileName: "case.js", Output: []string{"if (a) {\n\timportScripts(a, b);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 3, Column: 2, EndLine: 3, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "switch (a) {\n\tdefault:\n\t\timportScripts(a);\n\t\timportScripts(b);\n}", FileName: "case.js", Output: []string{"switch (a) {\n\tdefault:\n\t\timportScripts(a, b);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 4, Column: 3, EndLine: 4, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "function _() {\n\timportScripts(a);\n\timportScripts(b);\n}", FileName: "case.js", Output: []string{"function _() {\n\timportScripts(a, b);\n}"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 3, Column: 2, EndLine: 3, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(a)\nimportScripts(b)\n;[foo].forEach(bar)", FileName: "case.js", Output: []string{"importScripts(a, b);[foo].forEach(bar)"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts?.(\"foo.js\");\nimportScripts(\"bar.js\");", FileName: "case.js", Output: []string{"importScripts?.(\"foo.js\", \"bar.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts(\"foo.js\");\nimportScripts?.(\"bar.js\");", FileName: "case.js", Output: []string{"importScripts(\"foo.js\", \"bar.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "importScripts?.(\"foo.js\");\nimportScripts?.(\"bar.js\");", FileName: "case.js", Output: []string{"importScripts?.(\"foo.js\", \"bar.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 2, Column: 1, EndLine: 2, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Reference identity", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "'1'.someMagicPropertyReturnsAnArray.push(1);\n(1).someMagicPropertyReturnsAnArray.push(2);\n\n/a/i.someMagicPropertyReturnsAnArray.push(1);\n/b/g.someMagicPropertyReturnsAnArray.push(2);\n\n1n.someMagicPropertyReturnsAnArray.push(1);\n2n.someMagicPropertyReturnsAnArray.push(2);\n\n(true).someMagicPropertyReturnsAnArray.push(1);\n(false).someMagicPropertyReturnsAnArray.push(2);", FileName: "case.js"},
		}, []rule_tester.InvalidTestCase{
			{Code: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 4, Column: 10, EndLine: 4, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1, 2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 7, Column: 11, EndLine: 7, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1, 2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 10, Column: 12, EndLine: 10, EndColumn: 16, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1, 1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 13, Column: 16, EndLine: 13, EndColumn: 20, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1, 1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 16, Column: 13, EndLine: 16, EndColumn: 17, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1, 1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 19, Column: 39, EndLine: 19, EndColumn: 43, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1, 2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 22, Column: 40, EndLine: 22, EndColumn: 44, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1, 2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 25, Column: 38, EndLine: 25, EndColumn: 42, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1, 2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1);\n\t\t(true).someMagicPropertyReturnsAnArray.push(2);\n\t}\n}"}}},
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 28, Column: 42, EndLine: 28, EndColumn: 46, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "class A extends B {\n\tfoo() {\n\t\tthis.x.push(1);\n\t\tthis.x.push(2);\n\n\t\tsuper.x.push(1);\n\t\tsuper.x.push(2);\n\n\t\t((a?.x).y).push(1);\n\t\t(a.x?.y).push(1);\n\n\t\t((a?.x.y).z).push(1);\n\t\t((a.x?.y).z).push(1);\n\n\t\ta[null].push(1);\n\t\ta['null'].push(1);\n\n\t\t'1'.someMagicPropertyReturnsAnArray.push(1);\n\t\t'1'.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(1);\n\t\t/a/i.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t1n.someMagicPropertyReturnsAnArray.push(1);\n\t\t1n.someMagicPropertyReturnsAnArray.push(2);\n\n\t\t(true).someMagicPropertyReturnsAnArray.push(1, 2);\n\t}\n}"}}},
			}},
			{Code: "a[x].push(1);\na[x].push(2);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 2, Column: 6, EndLine: 2, EndColumn: 10, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "a[x].push(1, 2);"}}},
			}},
		})
	})
	t.Run("Receiver reads", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{}, []rule_tester.InvalidTestCase{
			{Code: "const result = [];\nresult.push(\"a\");\nresult.push(result.length);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 8, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = [];\nresult.push(\"a\", result.length);"}}},
			}},
			{Code: "const result = [];\nresult.push(\"a\");\nresult.push(String(result));", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 8, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = [];\nresult.push(\"a\", String(result));"}}},
			}},
			{Code: "const result = [];\nresult.unshift(\"a\");\nresult.unshift(result.length);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 8, EndLine: 3, EndColumn: 15, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "const result = [];\nresult.unshift(result.length, \"a\");"}}},
			}},
			{Code: "const result = [];\nresult.push(\"a\");\nresult.push(1);", FileName: "case.js", Output: []string{"const result = [];\nresult.push(\"a\", 1);"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 8, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
	t.Run("Documentation", func(t *testing.T) {
		rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &prefer_single_call.PreferSingleCallRule, []rule_tester.ValidTestCase{
			{Code: "// ✅\nfoo.push(1, 2, 3);", FileName: "case.js"},
			{Code: "// ✅\nfoo.unshift(2, 3, 1);", FileName: "case.js"},
			{Code: "// ✅\nelement.classList.add('foo', 'bar', 'baz');", FileName: "case.js"},
			{Code: "// ✅\nimportScripts(\n\t\"https://example.com/foo.js\",\n\t\"https://example.com/bar.js\",\n);", FileName: "case.js"},
			{Code: "/* eslint unicorn/prefer-single-call: [\"error\", {\"ignore\": [\"readable.push\"]}] */\nimport {Readable} from 'node:stream';\n\nconst readable = new Readable();\nreadable.push('one');\nreadable.push('another');\nreadable.push(null);\n", FileName: "case.js", Options: []any{map[string]any{"ignore": []any{"readable.push"}}}},
		}, []rule_tester.InvalidTestCase{
			{Code: "// ❌\nfoo.push(1);\nfoo.push(2, 3);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#push()` multiple times.", Line: 3, Column: 5, EndLine: 3, EndColumn: 9, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "// ❌\nfoo.push(1, 2, 3);"}}},
			}},
			{Code: "// ❌\nfoo.unshift(1);\nfoo.unshift(2, 3);", FileName: "case.js", Output: []string{}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Array#unshift()` multiple times.", Line: 3, Column: 5, EndLine: 3, EndColumn: 12, Suggestions: []rule_tester.InvalidTestCaseSuggestion{{MessageId: "suggestion", Output: "// ❌\nfoo.unshift(2, 3, 1);"}}},
			}},
			{Code: "// ❌\nelement.classList.add('foo');\nelement.classList.add('bar', 'baz');", FileName: "case.js", Output: []string{"// ❌\nelement.classList.add('foo', 'bar', 'baz');"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `Element#classList.add()` multiple times.", Line: 3, Column: 19, EndLine: 3, EndColumn: 22, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
			{Code: "// ❌\nimportScripts(\"https://example.com/foo.js\");\nimportScripts(\"https://example.com/bar.js\");", FileName: "case.js", Output: []string{"// ❌\nimportScripts(\"https://example.com/foo.js\", \"https://example.com/bar.js\");"}, Errors: []rule_tester.InvalidTestCaseError{
				{MessageId: "error/array-push", Message: "Do not call `importScripts()` multiple times.", Line: 3, Column: 1, EndLine: 3, EndColumn: 14, Suggestions: []rule_tester.InvalidTestCaseSuggestion{}},
			}},
		})
	})
}
