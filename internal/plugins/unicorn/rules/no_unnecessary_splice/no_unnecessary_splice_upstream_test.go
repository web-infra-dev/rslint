// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
package no_unnecessary_splice

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUnnecessarySpliceUpstream(t *testing.T) {
	valid := []rule_tester.ValidTestCase{
		// First snapshot group, including a known non-array receiver.
		{Code: "function f(foo: {splice(start?: number, deleteCount?: number): void}) { foo.splice(); }", FileName: "input.ts"},
		{Code: "array.splice(index, 1);", FileName: "input.js"},
		{Code: "array.splice(index, 2);", FileName: "input.js"},
		{Code: "array.splice(index, 1, element);", FileName: "input.js"},
		{Code: "array.splice(1, 0, element);", FileName: "input.js"},
		{Code: "array.splice(0, 2);", FileName: "input.js"},
		{Code: "array.splice(0, 1, element);", FileName: "input.js"},
		{Code: "array.splice(array.length - 2, 1);", FileName: "input.js"},
		{Code: "array.splice(array.length - 1, 0, element);", FileName: "input.js"},
		{Code: "array.splice(otherArray.length - 1, 1);", FileName: "input.js"},
		{Code: "array.splice(otherArray.length, 0, element);", FileName: "input.js"},
		{Code: "array.splice(1, array.length);", FileName: "input.js"},
		{Code: "array[\"splice\"](0, 1);", FileName: "input.js"},
		{Code: "array[splice](0, 1);", FileName: "input.js"},
		{Code: "array?.splice(0, 1);", FileName: "input.js"},
		{Code: "array.splice?.(0, 1);", FileName: "input.js"},
		{Code: "object?.array.splice(0, 1);", FileName: "input.js"},
		{Code: "getObject?.().array.splice(0, 1);", FileName: "input.js"},
		{Code: "object[foo?.bar].splice(index, 0);", FileName: "input.js"},
		{Code: "object[foo?.bar]().splice(index, 0);", FileName: "input.js"},
		{Code: "array.toSpliced(0, 1);", FileName: "input.js"},
		{Code: "array.splice(- -1, 1);", FileName: "input.js"},
		{Code: "const result = array.splice(0);", FileName: "input.js"},
		{Code: "const result = array.splice(0, array.length);", FileName: "input.js"},
		{Code: "for (const item of array.splice(0)) { item; }", FileName: "input.js"},
		{Code: "for (const item of array.splice(0, array.length)) { item; }", FileName: "input.js"},
		// TypeScript snapshot cases.
		{Code: "array.splice(0 as const, 2);", FileName: "input.ts"},
		{Code: "(object?.array as string[]).splice(index, 0);", FileName: "input.ts"},
		{Code: "((object?.array as {items: string[]}).items).splice(index, 0);", FileName: "input.ts"},
		{Code: "(object?.array satisfies string[]).splice(0);", FileName: "input.ts"},
		{Code: "const result: unknown[] = array.splice(0 as const);", FileName: "input.ts"},
		{Code: "const result: unknown[] = array.splice(0 as const, array.length);", FileName: "input.ts"},
		// Correct documentation examples.
		{Code: "// Remove the no-op call.", FileName: "input.js"},
		{Code: "array.shift();", FileName: "input.js"},
		{Code: "array.unshift(item);", FileName: "input.js"},
		{Code: "array.pop();", FileName: "input.js"},
		{Code: "array.push(item);", FileName: "input.js"},
		{Code: "array.length = 0;", FileName: "input.js"},
	}
	var invalid []rule_tester.InvalidTestCase
	// Each upstream report highlights the six-character method name.
	type upstreamCase struct {
		id     string
		column int
		code   string
		output *string
	}
	addCases := func(filename string, cases []upstreamCase) {
		for _, item := range cases {
			testCase := spliceCase(filename, item.code, item.id, 1, item.column)
			if item.output != nil {
				testCase.Output = []string{*item.output}
			}
			invalid = append(invalid, testCase)
		}
	}
	// JavaScript snapshot cases.
	addCases("input.js", []upstreamCase{
		{"no-op", 7, "array.splice();", new("")},
		// An unknown index could run conversion code or throw.
		{"no-op", 7, "array.splice(index, 0);", nil},
		{"no-op", 7, "array.splice(index, -1);", nil},
		{"no-op", 7, "array.splice(index, +0);", nil},
		{"shift", 7, "array.splice(0, 1);", new("array.shift();")},
		{"shift", 7, "array.splice(+0, +1);", new("array.shift();")},
		{"shift", 7, "array.splice(0, - -1);", new("array.shift();")},
		{"shift", 7, "array.splice(- -0, 1);", new("array.shift();")},
		{"no-op", 7, "array.splice(index, - -0);", nil},
		{"unshift", 7, "array.splice(0, 0, element);", new("array.unshift(element);")},
		{"unshift", 7, "array.splice(0, 0, element, ...moreElements);", new("array.unshift(element, ...moreElements);")},
		{"unshift", 7, "array.splice(0, 0, ...elements);", new("array.unshift(...elements);")},
		{"pop", 7, "array.splice(array.length - 1, 1);", new("array.pop();")},
		{"pop", 7, "array.splice(array.length - +1, +1);", new("array.pop();")},
		{"pop", 14, "object.array.splice(object.array.length - 1, 1);", new("object.array.pop();")},
		{"push", 7, "array.splice(array.length, 0, element);", new("array.push(element);")},
		// An unknown iterator could mutate the receiver during insertion.
		{"push", 7, "array.splice(array.length, 0, element, ...moreElements);", nil},
		{"push", 7, "array.splice(array.length, 0, ...elements);", nil},
		{"push", 14, "object.array.splice(object.array.length, 0, element);", new("object.array.push(element);")},
		{"empty", 7, "array.splice(0);", new("array.length = 0;")},
		{"empty", 14, "object.array.splice(0);", new("object.array.length = 0;")},
		{"empty", 7, "array.splice(0, array.length);", new("array.length = 0;")},
		{"empty", 14, "object.array.splice(0, object.array.length);", new("object.array.length = 0;")},
		{"no-op", 22, "const result = array.splice(index, 0);", nil},
		{"shift", 22, "const result = array.splice(0, 1);", nil},
		{"unshift", 22, "const result = array.splice(0, 0, element);", nil},
		{"pop", 22, "const result = array.splice(array.length - 1, 1);", nil},
		{"push", 22, "const result = array.splice(array.length, 0, element);", nil},
		{"unshift", 7, "array.splice(0, /* comment */ 0, element);", nil},
		{"shift", 7, "array.splice(/* comment */ 0, 1);", nil},
		{"no-op", 7, "array.splice(index, /* comment */ 0);", nil},
		{"no-op", 7, "array.splice(index, 0) /* comment */;", nil},
		{"no-op", 7, "array.splice(index, 0); // comment", nil},
		{"no-op", 7, "array.splice(index, 0); /* comment */", nil},
		{"no-op", 12, "getArray().splice(index, 0);", nil},
		{"no-op", 7, "array.splice(getIndex(), 0);", nil},
		{"no-op", 22, "if (condition) array.splice(index, 0);", nil},
		{"empty", 7, "array.splice(0), f();", new("array.length = 0, f();")},
		{"empty", 7, "array.splice(0, array.length), f();", new("array.length = 0, f();")},
		{"empty", 12, "f(), array.splice(0);", new("f(), array.length = 0;")},
		{"empty", 12, "f(), array.splice(0), g();", new("f(), array.length = 0, g();")},
	})
	// TypeScript snapshot cases.
	addCases("input.ts", []upstreamCase{
		{"shift", 7, "array.splice(0 as const, 1 as const);", new("array.shift();")},
		{"shift", 7, "array.splice(0 satisfies number, 1 satisfies number);", new("array.shift();")},
		{"unshift", 7, "array.splice(<number>0, <number>0, element);", new("array.unshift(element);")},
		{"pop", 7, "array.splice(array.length - 1!, 1!);", new("array.pop();")},
		{"empty", 7, "array.splice(0!, array.length);", new("array.length = 0;")},
		{"push", 28, "(array satisfies string[]).splice((array satisfies string[]).length, 0, element);", new("(array satisfies string[]).push(element);")},
		{"no-op", 33, "function f(foo: number[]) { foo.splice(); }", new("function f(foo: number[]) {  }")},
		{"empty", 7, "array.splice(0, array.length as number);", new("array.length = 0;")},
		{"pop", 7, "array.splice((array.length as number) - 1, 1);", new("array.pop();")},
	})
	// Parenthesized sequence arguments.
	addCases("input.js", []upstreamCase{
		{"unshift", 3, "a.splice(0, 0, (b, c))", new("a.unshift((b, c))")},
		{"push", 3, "a.splice(a.length, 0, (b, c))", new("a.push((b, c))")},
		{"unshift", 3, "a.splice(0, 0, x, (b, c))", new("a.unshift(x, (b, c))")},
		{"unshift", 3, "a.splice(0, 0, x)", new("a.unshift(x)")},
	})
	// Insertion evaluation order.
	addCases("input.js", []upstreamCase{
		{"push", 74, "const array = [1, 2, 3]; function f() { array.push(9); return 7; } array.splice(array.length, 0, f());", nil},
		{"push", 77, "const array = [1, 2, 3]; function f() { array.unshift(9); return 7; } array.splice(array.length, 0, f());", nil},
		{"push", 32, "const array = [1, 2, 3]; array.splice(array.length, 0, foo());", nil},
		{"push", 32, "const array = [1, 2, 3]; array.splice(array.length, 0, 7);", new("const array = [1, 2, 3]; array.push(7);")},
		{"unshift", 32, "const array = [1, 2, 3]; array.splice(0, 0, 7);", new("const array = [1, 2, 3]; array.unshift(7);")},
		{"unshift", 32, "const array = [1, 2, 3]; array.splice(0, 0, foo());", new("const array = [1, 2, 3]; array.unshift(foo());")},
	})
	// Parenthesized callees at statement start.
	addCases("input.js", []upstreamCase{
		{"empty", 35, "(function () { return [1, 2]; }().splice)(0);", new("(function () { return [1, 2]; }()).length = 0;")},
		{"shift", 24, "({array: [1, 2]}.array.splice)(0, 1);", new("({array: [1, 2]}.array).shift();")},
	})
	// Receiver parentheses and comments.
	addCases("input.js", []upstreamCase{
		{"empty", 36, "(function () { return [1, 2]; }()).splice(0);", new("(function () { return [1, 2]; }()).length = 0;")},
		{"empty", 35, "(function () { return [1, 2]; }().splice)(/* keep */ 0);", nil},
	})
	addCases("input.ts", []upstreamCase{
		{"empty", 36, "(function () { return [1, 2]; }()!.splice)(0);", new("(function () { return [1, 2]; }()!).length = 0;")},
	})
	// Parenthesized callee with inserted values.
	addCases("input.js", []upstreamCase{
		{"unshift", 21, "({array: [1]}.array.splice)(0, 0, 2);", new("({array: [1]}.array).unshift(2);")},
	})
	// Upstream runtime cases and the remaining incorrect documentation examples.
	addCases("input.js", []upstreamCase{
		{"empty", 56, "const array = [1, 2]; (function () { return array; }().splice)(0); JSON.stringify(array);", new("const array = [1, 2]; (function () { return array; }()).length = 0; JSON.stringify(array);")},
		{"shift", 38, "const array = [1, 2]; ({array}.array.splice)(0, 1); JSON.stringify(array);", new("const array = [1, 2]; ({array}.array).shift(); JSON.stringify(array);")},
		{"unshift", 7, "array.splice(0, 0, item);", new("array.unshift(item);")},
		{"push", 7, "array.splice(array.length, 0, item);", new("array.push(item);")},
	})
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &NoUnnecessarySpliceRule, valid, invalid)
}

func spliceCase(filename, code, id string, line, column int, output ...string) rule_tester.InvalidTestCase {
	return rule_tester.InvalidTestCase{
		Code: code, FileName: filename, Output: output,
		Errors: []rule_tester.InvalidTestCaseError{{
			MessageId: id, Message: expectedMessages[id], Line: line, Column: column, EndLine: line, EndColumn: column + 6,
		}},
	}
}

var expectedMessages = map[string]string{
	"no-op":   "This `splice()` call does not change the array.",
	"shift":   "Prefer `.shift()` over `.splice()`.",
	"unshift": "Prefer `.unshift()` over `.splice()`.",
	"pop":     "Prefer `.pop()` over `.splice()`.",
	"push":    "Prefer `.push()` over `.splice()`.",
	"empty":   "Prefer setting `.length = 0` over `.splice()` to empty an array.",
}
