package no_useless_undefined_test

// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-undefined.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/snapshots/no-useless-undefined.js.snapshot

import (
	"testing"

	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/fixtures"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/rules/no_useless_undefined"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/rule_tester"
)

func TestNoUselessUndefinedUpstream(t *testing.T) {
	rule_tester.RunRuleTester(fixtures.GetRootDir(), "tsconfig.json", t, &no_useless_undefined.NoUselessUndefinedRule, []rule_tester.ValidTestCase{
		// JavaScript and index bounds
		{Code: "function foo() {return;}", FileName: "file.js"},
		{Code: "const foo = () => {};", FileName: "file.js"},
		{Code: "let foo;", FileName: "file.js"},
		{Code: "var foo;", FileName: "file.js"},
		{Code: "var a = undefined;", FileName: "file.js"},
		{Code: "for (const item of items) { var value = undefined; }", FileName: "file.js"},
		{Code: "const foo = undefined;", FileName: "file.js"},
		{Code: "foo();", FileName: "file.js"},
		{Code: "foo(bar,);", FileName: "file.js"},
		{Code: "foo(undefined, bar);", FileName: "file.js"},
		{Code: "const {foo} = {};", FileName: "file.js"},
		{Code: "function foo({bar} = {}) {}", FileName: "file.js"},
		{Code: "function foo(bar) {}", FileName: "file.js"},
		{Code: "function* foo() {yield* undefined;}", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? array[index] : bar;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? object.property : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? array?.[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? object?.items[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? getArray()[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? object.items[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? array[index++] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? array[getIndex()] : undefined;", FileName: "file.js"},
		{Code: "const foo = index < otherArray.length ? array[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? array[index + 1] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 0 ? array[otherIndex] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= 5 ? array[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index < array.length ? array[index - 1] : undefined;", FileName: "file.js"},
		{Code: "const foo = foo?.bar >= 0 ? array[foo.bar] : undefined;", FileName: "file.js"},
		{Code: "if (Object.is(foo, undefined)){}", FileName: "file.js"},
		{Code: "t.is(foo, undefined)", FileName: "file.js"},
		{Code: "assert.equal(foo, undefined, message)", FileName: "file.js"},
		{Code: "assert.notEqual(foo, undefined, message)", FileName: "file.js"},
		{Code: "assert.strictEqual(foo, undefined, message)", FileName: "file.js"},
		{Code: "assert.notStrictEqual(foo, undefined, message)", FileName: "file.js"},
		{Code: "assert.propertyVal(foo, \"bar\", undefined, message)", FileName: "file.js"},
		{Code: "assert.notPropertyVal(foo, \"bar\", undefined, message)", FileName: "file.js"},
		{Code: "expect(foo).not(undefined)", FileName: "file.js"},
		{Code: "expect(foo).to.have.property(\"bar\", undefined)", FileName: "file.js"},
		{Code: "expect(foo).toBe(undefined)", FileName: "file.js"},
		{Code: "expect(foo).toContain(undefined)", FileName: "file.js"},
		{Code: "expect(foo).toContainEqual(undefined)", FileName: "file.js"},
		{Code: "expect(foo).toEqual(undefined)", FileName: "file.js"},
		{Code: "t.same(foo, undefined)", FileName: "file.js"},
		{Code: "t.notSame(foo, undefined)", FileName: "file.js"},
		{Code: "t.strictSame(foo, undefined)", FileName: "file.js"},
		{Code: "t.strictNotSame(foo, undefined)", FileName: "file.js"},
		{Code: "expect(someFunction).toHaveBeenCalledWith(1, 2, undefined);", FileName: "file.js"},
		{Code: "set.add(undefined);", FileName: "file.js"},
		{Code: "map.set(foo, undefined);", FileName: "file.js"},
		{Code: "array.push(foo, undefined);", FileName: "file.js"},
		{Code: "array.push(undefined);", FileName: "file.js"},
		{Code: "array.unshift(foo, undefined);", FileName: "file.js"},
		{Code: "array.unshift(undefined);", FileName: "file.js"},
		{Code: "createContext(undefined);", FileName: "file.js"},
		{Code: "React.createContext(undefined);", FileName: "file.js"},
		{Code: "useRef(undefined);", FileName: "file.js"},
		{Code: "React.useRef(undefined);", FileName: "file.js"},
		{Code: "setState(undefined)", FileName: "file.js"},
		{Code: "setState?.(undefined)", FileName: "file.js"},
		{Code: "props.setState(undefined)", FileName: "file.js"},
		{Code: "props.setState?.(undefined)", FileName: "file.js"},
		{Code: "array.includes(undefined)", FileName: "file.js"},
		{Code: "set.has(undefined)", FileName: "file.js"},
		{Code: "set.delete(undefined)", FileName: "file.js"},
		{Code: "foo.bind(bar, undefined)", FileName: "file.js"},
		{Code: "foo.bind(...bar, undefined)", FileName: "file.js"},
		{Code: "foo.bind(...[], undefined)", FileName: "file.js"},
		{Code: "foo.bind(...[undefined], undefined)", FileName: "file.js"},
		{Code: "foo.bind(bar, baz, undefined)", FileName: "file.js"},
		{Code: "foo?.bind(bar, undefined)", FileName: "file.js"},
		{Code: "foo(undefined, undefined);", FileName: "file.js", Options: map[string]any{"checkArguments": false}},
		{Code: "foo.bind(undefined);", FileName: "file.js", Options: map[string]any{"checkArguments": false}},
		{Code: "const foo = () => undefined", FileName: "file.js", Options: map[string]any{"checkArrowFunctionBody": false}},
		{Code: "const foo = condition ? array[index] : undefined;", FileName: "file.js"},
		{Code: "const foo = index >= object?.minimum ? array[index] : undefined;", FileName: "file.js"},

		// TypeScript
		{Code: "prerenderPaths?.add(entry)", FileName: "file.ts"},
		{Code: `function getThing(): string | undefined {
	if (someCondition) {
		return "hello world";
	}

	return undefined;
}`, FileName: "file.ts"},
		{Code: `function getThing(): string | undefined {
	if (someCondition) {
		return "hello world";
	} else if (anotherCondition) {
		return undefined;
	}

	return undefined;
}`, FileName: "file.ts"},
		{Code: "const foo = (): undefined => undefined;", FileName: "file.ts"},
		{Code: "const foo = (): string => undefined;", FileName: "file.ts"},
		{Code: `function foo(): number | undefined {
	if (bar) {
		return 2;
	}

	return undefined;
}`, FileName: "file.ts"},
		{Code: `function foo(): number | void {
	if (bar) {
		return 2;
	}

	return undefined;
}`, FileName: "file.ts"},
		{Code: "function foo(): any {return undefined;}", FileName: "file.ts"},
		{Code: "function foo(): unknown {return undefined;}", FileName: "file.ts"},
		{Code: "function foo(): never {return undefined;}", FileName: "file.ts"},
		{Code: "function foo(): undefined | number {return undefined;}", FileName: "file.ts"},
		{Code: "async function foo(): Promise<void> {return undefined;}", FileName: "file.ts"},
		{Code: "createContext<T>(undefined);", FileName: "file.ts"},
		{Code: "React.createContext<T>(undefined);", FileName: "file.ts"},
		{Code: `function test<T extends object | undefined>(argument: T): void {}
test(undefined);`, FileName: "file.ts"},
		{Code: "foo(undefined);", FileName: "file.ts"},
		{Code: "foo(undefined);", FileName: "file.tsx"},
		{Code: "foo(undefined);", FileName: "file.mts"},
		{Code: "foo(undefined);", FileName: "file.cts"},
		{Code: "foo(value, undefined);", FileName: "file.ts"},
		{Code: "foo(undefined, value, undefined);", FileName: "file.ts"},
		{Code: "Promise.resolve(undefined);", FileName: "file.ts"},
		{Code: "Promise.resolve<undefined>(undefined);", FileName: "file.ts"},
		{Code: "Promise.resolve(foo, undefined);", FileName: "file.ts"},
		{Code: "Promise.resolve?.(undefined);", FileName: "file.ts"},
		{Code: "Promise?.resolve(undefined);", FileName: "file.ts"},
		{Code: "Promise[\"resolve\"](undefined);", FileName: "file.ts"},
		{Code: "globalThis.Promise.resolve(undefined);", FileName: "file.ts"},
		{Code: "const resolve = Promise.resolve; resolve(undefined);", FileName: "file.ts"},
		{Code: "NotPromise.resolve(undefined);", FileName: "file.ts"},

		// Vue single-file components require vue-eslint-parser.
		{Code: `<script>
import {ref} from 'vue';

export default {
	setup() {
		return {foo: ref(undefined)};
	}
};
</script>`, FileName: "file.vue", Skip: true},
		{Code: `<script setup>
import * as vue from 'vue';
const foo = vue.ref(undefined);
</script>`, FileName: "file.vue", Skip: true},

		// Resource declarations
		{Code: "using resource = undefined;", FileName: "file.ts"},
		{Code: "using resource = undefined, other = undefined;", FileName: "file.ts"},
		{Code: "async function f() { await using resource = undefined; }", FileName: "file.ts"},
		{Code: "for (using resource = undefined; ; ) { break; }", FileName: "file.ts"},

		// Binding patterns and parentheses
		{Code: "let {foo} = undefined;", FileName: "file.js"},
		{Code: "let [foo] = undefined;", FileName: "file.js"},
		{Code: "var {foo: bar} = undefined;", FileName: "file.js"},
		{Code: "let {foo: {bar}} = undefined;", FileName: "file.js"},

		// Calls with built-ins
		{Code: "array.concat(undefined);", FileName: "file.js"},
		{Code: "array.concat(other, undefined);", FileName: "file.js"},
		{Code: "array.splice(start, undefined);", FileName: "file.js"},
		{Code: "array.splice(start, deleteCount, undefined);", FileName: "file.js"},
		{Code: "array.toSpliced(start, undefined);", FileName: "file.js"},
		{Code: "array.lastIndexOf(value, undefined);", FileName: "file.js"},
		{Code: "array.reduce(fn, undefined);", FileName: "file.js"},
		{Code: "array.reduceRight(fn, undefined);", FileName: "file.js"},
		{Code: "String(undefined);", FileName: "file.js"},
		{Code: "String.fromCharCode(undefined);", FileName: "file.js"},
		{Code: "String.fromCodePoint(undefined);", FileName: "file.js"},
		{Code: "Number(undefined);", FileName: "file.js"},
		{Code: "Math.max(undefined);", FileName: "file.js"},
		{Code: "Math.min(undefined);", FileName: "file.js"},
		{Code: "Math.hypot(undefined);", FileName: "file.js"},
		{Code: "Array(undefined);", FileName: "file.js"},
		{Code: "Array.of(undefined);", FileName: "file.js"},
		{Code: "Array.from(undefined);", FileName: "file.js"},
		{Code: "structuredClone(undefined);", FileName: "file.js"},
		{Code: "new Date(undefined);", FileName: "file.js"},
		{Code: "new RegExp(undefined);", FileName: "file.js"},
		{Code: "new Error(undefined);", FileName: "file.js"},

		// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-undefined.md
		{Code: "let foo;", FileName: "file.js"},
		{Code: "const {foo} = bar;", FileName: "file.js"},
		{Code: "const noop = () => {};", FileName: "file.js"},
		{Code: `function foo() {
	return;
}`, FileName: "file.js"},
		{Code: `function* foo() {
	yield;
}`, FileName: "file.js"},
		{Code: `function foo(bar) {
}`, FileName: "file.js"},
		{Code: `function foo({bar}) {
}`, FileName: "file.js"},
		{Code: "foo();", FileName: "file.js"},
		{Code: "const foo = array[index];", FileName: "file.js"},
		{Code: "foo(bar, baz);", FileName: "file.js", Options: map[string]any{"checkArguments": true}},
		{Code: "foo(bar, baz, undefined);", FileName: "file.js", Options: map[string]any{"checkArguments": false}},
		{Code: "const foo = () => {};", FileName: "file.js", Options: map[string]any{"checkArrowFunctionBody": true}},
		{Code: "const foo = () => undefined;", FileName: "file.js", Options: map[string]any{"checkArrowFunctionBody": false}},

		// Empty file
		{Code: "", FileName: "file.js"},
	}, []rule_tester.InvalidTestCase{
		// Reference contexts
		{Code: "(index >= 0 ? array[index] : undefined)();", FileName: "file.js", Errors: upstreamErrorAt(1, 30)},
		{Code: "(index >= 0 ? array[index] : undefined)?.();", FileName: "file.js", Errors: upstreamErrorAt(1, 30)},
		{Code: "(index >= 0 ? array[index] : undefined)`template`;", FileName: "file.js", Errors: upstreamErrorAt(1, 30)},
		{Code: "delete (index >= 0 ? array[index] : undefined);", FileName: "file.js", Errors: upstreamErrorAt(1, 37)},
		{Code: "((index >= 0 ? array[index] : undefined) as Function)();", FileName: "file.ts", Errors: upstreamErrorAt(1, 31)},

		// Keyword spacing
		{
			Code:     "function value(index, array) { return(index) >= 0 ? array[index] : undefined; }",
			FileName: "file.js",
			Errors:   upstreamErrorAt(1, 68, "function value(index, array) { return array[index]; }"),
		},
		{Code: "for (const value of(index >= 0 ? array[index] : undefined));", FileName: "file.js", Errors: upstreamErrorAt(1, 49, "for (const value of (array[index]));")},
		{
			Code:     "function value(index: number, array: number[]) { return(index) >= 0 ? (array[index] as number) : undefined; }",
			FileName: "file.ts",
			Errors:   upstreamErrorAt(1, 98, "function value(index: number, array: number[]) { return array[index] as number; }"),
		},

		// JavaScript and index bounds
		{Code: "function foo() {return undefined;}", FileName: "file.js", Output: []string{"function foo() {return;}"}, Errors: upstreamErrorAt(1, 24)},
		{Code: "const foo = () => undefined;", FileName: "file.js", Output: []string{"const foo = () => {};"}, Errors: upstreamErrorAt(1, 19)},
		{Code: "const foo = () => {return undefined;};", FileName: "file.js", Output: []string{"const foo = () => {return;};"}, Errors: upstreamErrorAt(1, 27)},
		{Code: "function foo() {return       undefined;}", FileName: "file.js", Output: []string{"function foo() {return;}"}, Errors: upstreamErrorAt(1, 30)},
		{
			Code:     "function foo() {return /* comment */ undefined;}",
			FileName: "file.js",
			Output:   []string{"function foo() {return /* comment */;}"},
			Errors:   upstreamErrorAt(1, 38),
		},
		{Code: "function* foo() {yield undefined;}", FileName: "file.js", Output: []string{"function* foo() {yield;}"}, Errors: upstreamErrorAt(1, 24)},
		{Code: "function* foo() {yield                 undefined;}", FileName: "file.js", Output: []string{"function* foo() {yield;}"}, Errors: upstreamErrorAt(1, 40)},
		{Code: "let a = undefined;", FileName: "file.js", Output: []string{"let a;"}, Errors: upstreamErrorAt(1, 9)},
		{Code: "let a = undefined, b = 2;", FileName: "file.js", Output: []string{"let a, b = 2;"}, Errors: upstreamErrorAt(1, 9)},
		{Code: "foo(undefined);", FileName: "file.js", Output: []string{"foo();"}, Errors: upstreamErrorAt(1, 5)},
		{Code: "foo(undefined, undefined);", FileName: "file.js", Output: []string{"foo(undefined);", "foo();"}, Errors: upstreamErrorAt(1, 16)},
		{Code: "foo(undefined,);", FileName: "file.js", Output: []string{"foo();"}, Errors: upstreamErrorAt(1, 5)},
		{Code: "foo(undefined, undefined,);", FileName: "file.js", Output: []string{"foo(undefined,);", "foo();"}, Errors: upstreamErrorAt(1, 16)},
		{Code: "foo(bar, undefined);", FileName: "file.js", Output: []string{"foo(bar);"}, Errors: upstreamErrorAt(1, 10)},
		{Code: "foo(bar, undefined, undefined);", FileName: "file.js", Output: []string{"foo(bar, undefined);", "foo(bar);"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "foo(undefined, bar, undefined);", FileName: "file.js", Output: []string{"foo(undefined, bar);"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "foo(bar, undefined,);", FileName: "file.js", Output: []string{"foo(bar,);"}, Errors: upstreamErrorAt(1, 10)},
		{Code: "foo(undefined, bar, undefined,);", FileName: "file.js", Output: []string{"foo(undefined, bar,);"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "foo(bar, undefined, undefined,);", FileName: "file.js", Output: []string{"foo(bar, undefined,);", "foo(bar,);"}, Errors: upstreamErrorAt(1, 21)},
		{
			Code:     "foo(undefined, bar, undefined, undefined,);",
			FileName: "file.js",
			Output:   []string{"foo(undefined, bar, undefined,);", "foo(undefined, bar,);"},
			Errors:   upstreamErrorAt(1, 32),
		},
		{Code: "Promise.resolve(undefined);", FileName: "file.js", Output: []string{"Promise.resolve();"}, Errors: upstreamErrorAt(1, 17)},
		{Code: "const {foo = undefined} = {};", FileName: "file.js", Output: []string{"const {foo} = {};"}, Errors: upstreamErrorAt(1, 14)},
		{Code: "const [foo = undefined] = [];", FileName: "file.js", Output: []string{"const [foo] = [];"}, Errors: upstreamErrorAt(1, 14)},
		{Code: "function foo(bar = undefined) {}", FileName: "file.js", Output: []string{"function foo(bar) {}"}, Errors: upstreamErrorAt(1, 20)},
		{Code: "function foo({bar = undefined}) {}", FileName: "file.js", Output: []string{"function foo({bar}) {}"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "function foo({bar = undefined} = {}) {}", FileName: "file.js", Output: []string{"function foo({bar} = {}) {}"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "function foo([bar = undefined]) {}", FileName: "file.js", Output: []string{"function foo([bar]) {}"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "function foo([bar = undefined] = []) {}", FileName: "file.js", Output: []string{"function foo([bar] = []) {}"}, Errors: upstreamErrorAt(1, 21)},
		{
			Code:            "return undefined;",
			FileName:        "file.js",
			LanguageOptions: rule.LanguageOptions{SourceType: "script"},
			Output:          []string{"return;"},
			Errors:          upstreamErrorAt(1, 8),
		},
		{Code: "const foo = index >= 0 ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 41, "const foo = array[index];")},
		{Code: "const foo = index > -1 ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 41, "const foo = array[index];")},
		{Code: "const foo = index < 0 ? undefined : array[index];", FileName: "file.js", Errors: upstreamErrorAt(1, 25, "const foo = array[index];")},
		{Code: "const foo = index >= 0 ? (array[index] as string) : undefined;", FileName: "file.ts", Errors: upstreamErrorAt(1, 53, "const foo = array[index] as string;")},
		{Code: "const foo = index < 0 ? undefined : (array[index] as string);", FileName: "file.ts", Errors: upstreamErrorAt(1, 25, "const foo = array[index] as string;")},
		{Code: "const foo = -1 < index ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 41, "const foo = array[index];")},
		{Code: "const foo = 0 > index ? undefined : array[index];", FileName: "file.js", Errors: upstreamErrorAt(1, 25, "const foo = array[index];")},
		{Code: "const foo = index <= array.length - 1 ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 56, "const foo = array[index];")},
		{Code: "const foo = index > array.length - 1 ? undefined : array[index];", FileName: "file.js", Errors: upstreamErrorAt(1, 40, "const foo = array[index];")},
		{Code: "const foo = index >= 5 ? array[index - 5] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 45, "const foo = array[index - 5];")},
		{Code: "const foo = 0 <= index ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 41, "const foo = array[index];")},
		{Code: "const foo = index >= array.length ? undefined : array[index];", FileName: "file.js", Errors: upstreamErrorAt(1, 37, "const foo = array[index];")},
		{Code: "const foo = array.length <= index ? undefined : array[index];", FileName: "file.js", Errors: upstreamErrorAt(1, 37, "const foo = array[index];")},
		{Code: "const foo = array.length > index ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 51, "const foo = array[index];")},
		{Code: "const foo = 5 <= index ? array[index - 5] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 45, "const foo = array[index - 5];")},
		{Code: "const foo = index < 5 ? undefined : array[index - 5];", FileName: "file.js", Errors: upstreamErrorAt(1, 25, "const foo = array[index - 5];")},
		{Code: "const foo = index >= 0 ? array[index] /* comment */ : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 55)},
		{Code: "const foo = index /* comment */ >= 0 ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 55)},
		{Code: "const foo = index >= 0 ? array[index] : /* comment */ undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 55)},
		{Code: `a()
index >= 0 ? [array][index] : undefined;`, FileName: "file.js", Errors: upstreamErrorAt(2, 31, `a()
;[array][index];`)},

		// TypeScript
		{
			Code:     "function shouldBeFlagged(): undefined {return undefined;}",
			FileName: "file.ts",
			Output:   []string{"function shouldBeFlagged(): undefined {return;}"},
			Errors:   upstreamErrorAt(1, 47),
		},
		{
			Code:     "const foo = function (): undefined {return undefined};",
			FileName: "file.ts",
			Output:   []string{"const foo = function (): undefined {return};"},
			Errors:   upstreamErrorAt(1, 44),
		},
		{
			Code:     "function shouldBeFlagged(): void {return undefined;}",
			FileName: "file.ts",
			Output:   []string{"function shouldBeFlagged(): void {return;}"},
			Errors:   upstreamErrorAt(1, 42),
		},
		{
			Code:     "const foo = function (): void {return undefined};",
			FileName: "file.ts",
			Output:   []string{"const foo = function (): void {return};"},
			Errors:   upstreamErrorAt(1, 39),
		},
		{Code: "const foo = (): void => {return undefined};", FileName: "file.ts", Output: []string{"const foo = (): void => {return};"}, Errors: upstreamErrorAt(1, 33)},
		{
			Code: `class A {
	method(): void {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	method(): void {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `function outer(): string {
	function inner(): void {
		return undefined;
	}

	return 'foo';
}`,
			FileName: "file.ts",
			Output: []string{`function outer(): string {
	function inner(): void {
		return;
	}

	return 'foo';
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `class A {
	method(): undefined {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	method(): undefined {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code:     "export function foo(): undefined {return undefined}",
			FileName: "file.ts",
			Output:   []string{"export function foo(): undefined {return}"},
			Errors:   upstreamErrorAt(1, 42),
		},
		{
			Code: `const object = {
	method(): undefined {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`const object = {
	method(): undefined {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `const A = class A {
	method(): undefined {
		return undefined
	}
};`,
			FileName: "file.ts",
			Output: []string{`const A = class A {
	method(): undefined {
		return
	}
};`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `class A {
	static method(): undefined {
		return undefined
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	static method(): undefined {
		return
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `class A {
	get method(): undefined {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	get method(): undefined {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `class A {
	static get method(): undefined {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	static get method(): undefined {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `class A {
	#method(): undefined {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	#method(): undefined {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `class A {
	private method(): undefined {
		return undefined;
	}
}`,
			FileName: "file.ts",
			Output: []string{`class A {
	private method(): undefined {
		return;
	}
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code: `function foo():undefined {
	function nested() {
		return undefined;
	}

	return nested();
}`,
			FileName: "file.ts",
			Output: []string{`function foo():undefined {
	function nested() {
		return;
	}

	return nested();
}`},
			Errors: upstreamErrorAt(3, 10),
		},
		{
			Code:     "const foo = (index as number) >= 0 ? array[index as number] : undefined;",
			FileName: "file.ts",
			Errors:   upstreamErrorAt(1, 63, "const foo = array[index as number];"),
		},
		{Code: "const foo = index! >= 0 ? array[index!] : undefined;", FileName: "file.ts", Errors: upstreamErrorAt(1, 43, "const foo = array[index!];")},
		{
			Code:     "const foo = (index satisfies number) >= 0 ? array[index satisfies number] : undefined;",
			FileName: "file.ts",
			Errors:   upstreamErrorAt(1, 77, "const foo = array[index satisfies number];"),
		},
		{Code: "const foo = <number>index >= 0 ? array[<number>index] : undefined;", FileName: "file.ts", Errors: upstreamErrorAt(1, 57, "const foo = array[<number>index];")},
		{Code: "const foo = index >= (0 as number) ? array[index] : undefined;", FileName: "file.ts", Errors: upstreamErrorAt(1, 53, "const foo = array[index];")},
		{
			Code:     "const foo = index >= 5 ? array[index - (5 as number)] : undefined;",
			FileName: "file.ts",
			Errors:   upstreamErrorAt(1, 57, "const foo = array[index - (5 as number)];"),
		},
		{
			Code:     "const foo = index <= array.length - (1 as number) ? array[index] : undefined;",
			FileName: "file.ts",
			Errors:   upstreamErrorAt(1, 68, "const foo = array[index];"),
		},

		// JavaScript snapshots
		{
			Code: `foo(
	undefined,
	bar,
	undefined,
	undefined,
	undefined,
	undefined,
)`,
			FileName: "file.js",
			Output: []string{`foo(
	undefined,
	bar,
	undefined,
	undefined,
	undefined,
)`, `foo(
	undefined,
	bar,
	undefined,
	undefined,
)`, `foo(
	undefined,
	bar,
	undefined,
)`, `foo(
	undefined,
	bar,
)`},
			Errors: upstreamErrorAt(7, 2),
		},
		{Code: "function foo([bar = undefined] = []) {}", FileName: "file.js", Output: []string{"function foo([bar] = []) {}"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "foo(bar, undefined, undefined);", FileName: "file.js", Output: []string{"foo(bar, undefined);", "foo(bar);"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "let a = undefined, b = 2;", FileName: "file.js", Output: []string{"let a, b = 2;"}, Errors: upstreamErrorAt(1, 9)},
		{
			Code: `function foo() {
	return /* */ (
		/* */
		(
			/* */
			undefined
			/* */
		)
		/* */
	) /* */ ;
}`,
			FileName: "file.js",
			Output: []string{`function foo() {
	return /* */
		/* */

			/* */

			/* */

		/* */
 /* */ ;
}`},
			Errors: upstreamErrorAt(6, 4),
		},
		{
			Code: `function * foo() {
	yield /* */ (
		/* */
		(
			/* */
			undefined
			/* */
		)
		/* */
	) /* */ ;
}`,
			FileName: "file.js",
			Output: []string{`function * foo() {
	yield /* */
		/* */

			/* */

			/* */

		/* */
 /* */ ;
}`},
			Errors: upstreamErrorAt(6, 4),
		},
		{
			Code: `const foo = () => /* */ (
	/* */
	(
		/* */
		undefined
		/* */
	)
	/* */
);`,
			FileName: "file.js",
			Output: []string{`const foo = () => /* */
	/* */

		/* */
 {}
		/* */

	/* */
;`},
			Errors: upstreamErrorAt(5, 3),
		},
		{Code: "foo.bind(undefined)", FileName: "file.js", Output: []string{"foo.bind()"}, Errors: upstreamErrorAt(1, 10)},
		{Code: "bind(foo, undefined)", FileName: "file.js", Output: []string{"bind(foo)"}, Errors: upstreamErrorAt(1, 11)},
		{Code: "foo.bind?.(bar, undefined)", FileName: "file.js", Output: []string{"foo.bind?.(bar)"}, Errors: upstreamErrorAt(1, 17)},
		{Code: "foo[bind](bar, undefined)", FileName: "file.js", Output: []string{"foo[bind](bar)"}, Errors: upstreamErrorAt(1, 16)},
		{Code: "foo.notBind(bar, undefined)", FileName: "file.js", Output: []string{"foo.notBind(bar)"}, Errors: upstreamErrorAt(1, 18)},
		{
			Code: `const referencedExports =
	compilation.getDependencyReferencedExports(
		/** @type {Dependency} */ (connection.dependency),
		undefined
	);`,
			FileName: "file.js",
			Output: []string{`const referencedExports =
	compilation.getDependencyReferencedExports(
		/** @type {Dependency} */ (connection.dependency)
	);`},
			Errors: upstreamErrorAt(4, 3),
		},
		{
			Code:     "foo( ((a)), ((undefined)), ((undefined)), )",
			FileName: "file.js",
			Output:   []string{"foo( ((a)), ((undefined)), )", "foo( ((a)), )"},
			Errors:   upstreamErrorAt(1, 30),
		},
		{Code: "foo( ((undefined)), ((undefined)), )", FileName: "file.js", Output: []string{"foo( ((undefined)), )", "foo(  )"}, Errors: upstreamErrorAt(1, 23)},

		// Vue single-file components require vue-eslint-parser.
		{
			Code: `<script>
import {nextTick} from 'vue';
const foo = nextTick(undefined);
</script>`,
			FileName: "file.vue",
			Skip:     true,
			Output: []string{`<script>
import {nextTick} from 'vue';
const foo = nextTick();
</script>`},
			Errors: upstreamErrorAt(3, 22),
		},

		// TypeScript snapshots
		{Code: "function f(foo: Type = undefined) {}", FileName: "file.ts", Output: []string{"function f(foo?: Type) {}"}, Errors: upstreamErrorAt(1, 24)},
		{Code: "function f(foo?: Type = undefined) {}", FileName: "file.ts", Output: []string{"function f(foo?: Type) {}"}, Errors: upstreamErrorAt(1, 25)},
		// This snapshot needs a .ts file for the return type. Keep the parameter's inferred type.
		{Code: "function f(foo = undefined): string {}", FileName: "file.ts", Errors: upstreamErrorAt(1, 18)},
		{Code: "function f(): string { let bar = undefined; }", FileName: "file.ts", Output: []string{"function f(): string { let bar; }"}, Errors: upstreamErrorAt(1, 34)},
		{
			Code:     "const f = (): string => { let bar = undefined; };",
			FileName: "file.ts",
			Output:   []string{"const f = (): string => { let bar; };"},
			Errors:   upstreamErrorAt(1, 37),
		},
		{Code: "const f = function(foo: Type = undefined) {}", FileName: "file.ts", Output: []string{"const f = function(foo?: Type) {}"}, Errors: upstreamErrorAt(1, 32)},
		{Code: "const f = (foo: Type = undefined) => {}", FileName: "file.ts", Output: []string{"const f = (foo?: Type) => {}"}, Errors: upstreamErrorAt(1, 24)},
		{Code: "const f = {method(foo: Type = undefined){}}", FileName: "file.ts", Output: []string{"const f = {method(foo?: Type){}}"}, Errors: upstreamErrorAt(1, 31)},
		{
			Code:     "const f = class {method(foo: Type = undefined){}}",
			FileName: "file.ts",
			Output:   []string{"const f = class {method(foo?: Type){}}"},
			Errors:   upstreamErrorAt(1, 37),
		},
		{Code: "function f(foo = undefined) {}", FileName: "file.js", Output: []string{"function f(foo) {}"}, Errors: upstreamErrorAt(1, 18)},
		{Code: "function f(foo = undefined) {}", FileName: "file.js", Output: []string{"function f(foo) {}"}, Errors: upstreamErrorAt(1, 18)},
		{Code: "function f(foo = undefined) {}", FileName: "foo.js", Output: []string{"function f(foo) {}"}, Errors: upstreamErrorAt(1, 18)},
		// Unlike Unicorn, keep defaults that supply a parameter's inferred type.
		{Code: "function f(foo = undefined) {}", FileName: "foo.ts", Errors: upstreamErrorAt(1, 18)},
		// The foo.MTs snapshot is exercised via an explicit source root below.
		{Code: "function f(foo = undefined) {}", FileName: "foo.cts", Errors: upstreamErrorAt(1, 18)},
		{Code: "function f(foo = undefined) {}", FileName: "foo.tsx", Errors: upstreamErrorAt(1, 18)},
		// TypeScript rejects the optional binding pattern in Unicorn's fix.
		{Code: "function a({foo} = undefined) {}", FileName: "foo.ts", Errors: upstreamErrorAt(1, 20)},

		// Comments in initializer edits
		{Code: "function f(a /* keep */ = undefined) {}", FileName: "file.js", Errors: upstreamErrorAt(1, 27)},
		{Code: "let x = /* keep */ undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 20)},
		{Code: "const {a /* keep */ = undefined} = object;", FileName: "file.js", Errors: upstreamErrorAt(1, 23)},
		{Code: "let x /* keep */ = undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 20)},
		{Code: "function f(/* keep */ a = undefined) {}", FileName: "file.js", Output: []string{"function f(/* keep */ a) {}"}, Errors: upstreamErrorAt(1, 27)},
		{Code: "let x = undefined;", FileName: "file.js", Output: []string{"let x;"}, Errors: upstreamErrorAt(1, 9)},

		// Resource declarations
		{Code: "let resource = undefined;", FileName: "file.ts", Output: []string{"let resource;"}, Errors: upstreamErrorAt(1, 16)},

		// Binding patterns and parentheses
		{Code: "let foo = (undefined);", FileName: "file.js", Output: []string{"let foo;"}, Errors: upstreamErrorAt(1, 12)},
		{Code: "function foo(bar = (undefined)) {}", FileName: "file.js", Output: []string{"function foo(bar) {}"}, Errors: upstreamErrorAt(1, 21)},
		{Code: "const {bar = (undefined)} = object;", FileName: "file.js", Output: []string{"const {bar} = object;"}, Errors: upstreamErrorAt(1, 15)},

		// Calls with built-ins
		{Code: "Math.sign(undefined);", FileName: "file.js", Output: []string{"Math.sign();"}, Errors: upstreamErrorAt(1, 11)},
		{Code: "Math.trunc(undefined);", FileName: "file.js", Output: []string{"Math.trunc();"}, Errors: upstreamErrorAt(1, 12)},
		{Code: "Math.round(undefined);", FileName: "file.js", Output: []string{"Math.round();"}, Errors: upstreamErrorAt(1, 12)},
		{Code: "Math.log(undefined);", FileName: "file.js", Output: []string{"Math.log();"}, Errors: upstreamErrorAt(1, 10)},
		{Code: "parseInt(undefined);", FileName: "file.js", Output: []string{"parseInt();"}, Errors: upstreamErrorAt(1, 10)},
		{Code: "Number.isNaN(undefined);", FileName: "file.js", Output: []string{"Number.isNaN();"}, Errors: upstreamErrorAt(1, 14)},
		{Code: "Promise.resolve(undefined);", FileName: "file.js", Output: []string{"Promise.resolve();"}, Errors: upstreamErrorAt(1, 17)},
		{Code: "Promise.reject(undefined);", FileName: "file.js", Output: []string{"Promise.reject();"}, Errors: upstreamErrorAt(1, 16)},
		{Code: "Promise.all(undefined);", FileName: "file.js", Output: []string{"Promise.all();"}, Errors: upstreamErrorAt(1, 13)},
		{Code: "Promise.any(undefined);", FileName: "file.js", Output: []string{"Promise.any();"}, Errors: upstreamErrorAt(1, 13)},
		{Code: "Promise.race(undefined);", FileName: "file.js", Output: []string{"Promise.race();"}, Errors: upstreamErrorAt(1, 14)},
		{Code: "Promise.allSettled(undefined);", FileName: "file.js", Output: []string{"Promise.allSettled();"}, Errors: upstreamErrorAt(1, 20)},
		{Code: "Date(undefined);", FileName: "file.js", Output: []string{"Date();"}, Errors: upstreamErrorAt(1, 6)},
		{Code: "RegExp(undefined);", FileName: "file.js", Output: []string{"RegExp();"}, Errors: upstreamErrorAt(1, 8)},
		{Code: "Boolean(undefined);", FileName: "file.js", Output: []string{"Boolean();"}, Errors: upstreamErrorAt(1, 9)},
		{Code: "Symbol(undefined);", FileName: "file.js", Output: []string{"Symbol();"}, Errors: upstreamErrorAt(1, 8)},

		// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-undefined.md
		{Code: "let foo = undefined;", FileName: "file.js", Output: []string{"let foo;"}, Errors: upstreamErrorAt(1, 11)},
		{Code: "const {foo = undefined} = bar;", FileName: "file.js", Output: []string{"const {foo} = bar;"}, Errors: upstreamErrorAt(1, 14)},
		{Code: "const noop = () => undefined;", FileName: "file.js", Output: []string{"const noop = () => {};"}, Errors: upstreamErrorAt(1, 20)},
		{Code: `function foo() {
	return undefined;
}`, FileName: "file.js", Output: []string{`function foo() {
	return;
}`}, Errors: upstreamErrorAt(2, 9)},
		{Code: `function* foo() {
	yield undefined;
}`, FileName: "file.js", Output: []string{`function* foo() {
	yield;
}`}, Errors: upstreamErrorAt(2, 8)},
		{Code: `function foo(bar = undefined) {
}`, FileName: "file.js", Output: []string{`function foo(bar) {
}`}, Errors: upstreamErrorAt(1, 20)},
		{Code: `function foo({bar = undefined}) {
}`, FileName: "file.js", Output: []string{`function foo({bar}) {
}`}, Errors: upstreamErrorAt(1, 21)},
		{Code: "foo(undefined);", FileName: "file.js", Output: []string{"foo();"}, Errors: upstreamErrorAt(1, 5)},
		{Code: "const foo = index >= 0 ? array[index] : undefined;", FileName: "file.js", Errors: upstreamErrorAt(1, 41, "const foo = array[index];")},
		{
			Code:     "foo(bar, baz, undefined);",
			FileName: "file.js",
			Options:  map[string]any{"checkArguments": true},
			Output:   []string{"foo(bar, baz);"},
			Errors:   upstreamErrorAt(1, 15),
		},
		{
			Code:     "const foo = () => undefined;",
			FileName: "file.js",
			Options:  map[string]any{"checkArrowFunctionBody": true},
			Output:   []string{"const foo = () => {};"},
			Errors:   upstreamErrorAt(1, 19),
		},
	})
}

func TestNoUselessUndefinedUpstreamUppercaseExtension(t *testing.T) {
	t.Parallel()
	// The compiler fixture's tsconfig glob excludes uppercase extensions.
	const code = "function f(foo = undefined) {}"
	program, file := undefinedTestProgram(t, "foo.MTs", code)
	diagnostics := lintUndefined(program, file, rule.EditDemandAll)
	if len(diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %v", diagnostics)
	}
	diagnostic := diagnostics[0]
	if diagnostic.Message.Id != "no-useless-undefined" || diagnostic.Message.Description != "Do not use useless `undefined`." ||
		diagnostic.Range.Pos() != 17 || diagnostic.Range.End() != 26 || len(diagnostic.Fixes()) != 0 || diagnostic.Suggestions != nil {
		t.Fatalf("unexpected diagnostic %+v", diagnostic)
	}
}

func upstreamErrorAt(line, column int, replacement ...string) []rule_tester.InvalidTestCaseError {
	diagnostic := rule_tester.InvalidTestCaseError{
		MessageId: "no-useless-undefined",
		Message:   "Do not use useless `undefined`.",
		Line:      line, Column: column, EndLine: line, EndColumn: column + len("undefined"),
	}
	if len(replacement) > 0 {
		diagnostic.Suggestions = []rule_tester.InvalidTestCaseSuggestion{{
			MessageId: "no-useless-undefined/suggestion",
			Output:    replacement[0],
		}}
	}
	return []rule_tester.InvalidTestCaseError{diagnostic}
}
