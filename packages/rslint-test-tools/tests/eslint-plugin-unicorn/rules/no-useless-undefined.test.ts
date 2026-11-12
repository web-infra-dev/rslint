// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-undefined.js
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/snapshots/no-useless-undefined.js.snapshot
import assert from 'node:assert/strict';
import path from 'node:path';
import { lint } from '@rslint/core/internal';
import { RuleTester, type ExpectedError } from '@rslint/rule-tester';
import { test } from 'rstack/test';

function undefinedAt(
  line: number,
  column: number,
  replacement?: string,
): ExpectedError {
  return {
    messageId: 'no-useless-undefined',
    message: 'Do not use useless `undefined`.',
    line,
    column,
    endLine: line,
    endColumn: column + 9,
    ...(replacement === undefined
      ? {}
      : {
          suggestions: [
            {
              messageId: 'no-useless-undefined/suggestion',
              desc: 'Use the indexed access directly.',
              output: replacement,
            },
          ],
        }),
  };
}

new RuleTester({ plugin: 'unicorn' }).run('no-useless-undefined', {
  valid: [
    // JavaScript and index bounds
    'function foo() {return;}',
    'const foo = () => {};',
    'let foo;',
    'var foo;',
    'var a = undefined;',
    'for (const item of items) { var value = undefined; }',
    'const foo = undefined;',
    'foo();',
    'foo(bar,);',
    'foo(undefined, bar);',
    'const {foo} = {};',
    'function foo({bar} = {}) {}',
    'function foo(bar) {}',
    'function* foo() {yield* undefined;}',
    'const foo = index >= 0 ? array[index] : bar;',
    'const foo = index >= 0 ? object.property : undefined;',
    'const foo = index >= 0 ? array?.[index] : undefined;',
    'const foo = index >= 0 ? object?.items[index] : undefined;',
    'const foo = index >= 0 ? getArray()[index] : undefined;',
    'const foo = index >= 0 ? object.items[index] : undefined;',
    'const foo = index >= 0 ? array[index++] : undefined;',
    'const foo = index >= 0 ? array[getIndex()] : undefined;',
    'const foo = index < otherArray.length ? array[index] : undefined;',
    'const foo = index >= 0 ? array[index + 1] : undefined;',
    'const foo = index >= 0 ? array[otherIndex] : undefined;',
    'const foo = index >= 5 ? array[index] : undefined;',
    'const foo = index < array.length ? array[index - 1] : undefined;',
    'const foo = foo?.bar >= 0 ? array[foo.bar] : undefined;',
    'if (Object.is(foo, undefined)){}',
    't.is(foo, undefined)',
    'assert.equal(foo, undefined, message)',
    'assert.notEqual(foo, undefined, message)',
    'assert.strictEqual(foo, undefined, message)',
    'assert.notStrictEqual(foo, undefined, message)',
    'assert.propertyVal(foo, "bar", undefined, message)',
    'assert.notPropertyVal(foo, "bar", undefined, message)',
    'expect(foo).not(undefined)',
    'expect(foo).to.have.property("bar", undefined)',
    'expect(foo).toBe(undefined)',
    'expect(foo).toContain(undefined)',
    'expect(foo).toContainEqual(undefined)',
    'expect(foo).toEqual(undefined)',
    't.same(foo, undefined)',
    't.notSame(foo, undefined)',
    't.strictSame(foo, undefined)',
    't.strictNotSame(foo, undefined)',
    'expect(someFunction).toHaveBeenCalledWith(1, 2, undefined);',
    'set.add(undefined);',
    'map.set(foo, undefined);',
    'array.push(foo, undefined);',
    'array.push(undefined);',
    'array.unshift(foo, undefined);',
    'array.unshift(undefined);',
    'createContext(undefined);',
    'React.createContext(undefined);',
    'useRef(undefined);',
    'React.useRef(undefined);',
    'setState(undefined)',
    'setState?.(undefined)',
    'props.setState(undefined)',
    'props.setState?.(undefined)',
    'array.includes(undefined)',
    'set.has(undefined)',
    'set.delete(undefined)',
    'foo.bind(bar, undefined)',
    'foo.bind(...bar, undefined)',
    'foo.bind(...[], undefined)',
    'foo.bind(...[undefined], undefined)',
    'foo.bind(bar, baz, undefined)',
    'foo?.bind(bar, undefined)',
    {
      code: 'foo(undefined, undefined);',
      options: [{ checkArguments: false }],
    },
    {
      code: 'foo.bind(undefined);',
      options: [{ checkArguments: false }],
    },
    {
      code: 'const foo = () => undefined',
      options: [{ checkArrowFunctionBody: false }],
    },
    'const foo = condition ? array[index] : undefined;',
    'const foo = index >= object?.minimum ? array[index] : undefined;',

    // TypeScript
    {
      code: 'prerenderPaths?.add(entry)',
      filename: 'file.ts',
    },
    {
      code: `function getThing(): string | undefined {
	if (someCondition) {
		return "hello world";
	}

	return undefined;
}`,
      filename: 'file.ts',
    },
    {
      code: `function getThing(): string | undefined {
	if (someCondition) {
		return "hello world";
	} else if (anotherCondition) {
		return undefined;
	}

	return undefined;
}`,
      filename: 'file.ts',
    },
    {
      code: 'const foo = (): undefined => undefined;',
      filename: 'file.ts',
    },
    {
      code: 'const foo = (): string => undefined;',
      filename: 'file.ts',
    },
    {
      code: `function foo(): number | undefined {
	if (bar) {
		return 2;
	}

	return undefined;
}`,
      filename: 'file.ts',
    },
    {
      code: `function foo(): number | void {
	if (bar) {
		return 2;
	}

	return undefined;
}`,
      filename: 'file.ts',
    },
    {
      code: 'function foo(): any {return undefined;}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(): unknown {return undefined;}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(): never {return undefined;}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(): undefined | number {return undefined;}',
      filename: 'file.ts',
    },
    {
      code: 'async function foo(): Promise<void> {return undefined;}',
      filename: 'file.ts',
    },
    {
      code: 'createContext<T>(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'React.createContext<T>(undefined);',
      filename: 'file.ts',
    },
    {
      code: `function test<T extends object | undefined>(argument: T): void {}
test(undefined);`,
      filename: 'file.ts',
    },
    {
      code: 'foo(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'foo(undefined);',
      filename: 'file.tsx',
    },
    {
      code: 'foo(undefined);',
      filename: 'file.mts',
    },
    {
      code: 'foo(undefined);',
      filename: 'file.cts',
    },
    {
      code: 'foo(value, undefined);',
      filename: 'file.ts',
    },
    {
      code: 'foo(undefined, value, undefined);',
      filename: 'file.ts',
    },
    {
      code: 'Promise.resolve(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'Promise.resolve<undefined>(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'Promise.resolve(foo, undefined);',
      filename: 'file.ts',
    },
    {
      code: 'Promise.resolve?.(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'Promise?.resolve(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'Promise["resolve"](undefined);',
      filename: 'file.ts',
    },
    {
      code: 'globalThis.Promise.resolve(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'const resolve = Promise.resolve; resolve(undefined);',
      filename: 'file.ts',
    },
    {
      code: 'NotPromise.resolve(undefined);',
      filename: 'file.ts',
    },

    // Vue single-file components
    {
      code: `<script>
import {ref} from 'vue';

export default {
	setup() {
		return {foo: ref(undefined)};
	}
};
</script>`,
      filename: 'file.vue',
      skip: 'Vue single-file components require vue-eslint-parser.',
    },
    {
      code: `<script setup>
import * as vue from 'vue';
const foo = vue.ref(undefined);
</script>`,
      filename: 'file.vue',
      skip: 'Vue single-file components require vue-eslint-parser.',
    },

    // Resource declarations
    {
      code: 'using resource = undefined;',
      filename: 'file.ts',
    },
    {
      code: 'using resource = undefined, other = undefined;',
      filename: 'file.ts',
    },
    {
      code: 'async function f() { await using resource = undefined; }',
      filename: 'file.ts',
    },
    {
      code: 'for (using resource = undefined; ; ) { break; }',
      filename: 'file.ts',
    },

    // Binding patterns and parentheses
    'let {foo} = undefined;',
    'let [foo] = undefined;',
    'var {foo: bar} = undefined;',
    'let {foo: {bar}} = undefined;',

    // Calls with built-ins
    'array.concat(undefined);',
    'array.concat(other, undefined);',
    'array.splice(start, undefined);',
    'array.splice(start, deleteCount, undefined);',
    'array.toSpliced(start, undefined);',
    'array.lastIndexOf(value, undefined);',
    'array.reduce(fn, undefined);',
    'array.reduceRight(fn, undefined);',
    'String(undefined);',
    'String.fromCharCode(undefined);',
    'String.fromCodePoint(undefined);',
    'Number(undefined);',
    'Math.max(undefined);',
    'Math.min(undefined);',
    'Math.hypot(undefined);',
    'Array(undefined);',
    'Array.of(undefined);',
    'Array.from(undefined);',
    'structuredClone(undefined);',
    'new Date(undefined);',
    'new RegExp(undefined);',
    'new Error(undefined);',

    // https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-undefined.md
    'let foo;',
    'const {foo} = bar;',
    'const noop = () => {};',
    `function foo() {
	return;
}`,
    `function* foo() {
	yield;
}`,
    `function foo(bar) {
}`,
    `function foo({bar}) {
}`,
    'foo();',
    'const foo = array[index];',
    {
      code: 'foo(bar, baz);',
      options: [{ checkArguments: true }],
    },
    {
      code: 'foo(bar, baz, undefined);',
      options: [{ checkArguments: false }],
    },
    {
      code: 'const foo = () => {};',
      options: [{ checkArrowFunctionBody: true }],
    },
    {
      code: 'const foo = () => undefined;',
      options: [{ checkArrowFunctionBody: false }],
    },

    // Empty file
    '',
  ],
  invalid: [
    // Reference contexts
    {
      code: '(index >= 0 ? array[index] : undefined)();',
      errors: [undefinedAt(1, 30)],
    },
    {
      code: '(index >= 0 ? array[index] : undefined)?.();',
      errors: [undefinedAt(1, 30)],
    },
    {
      code: '(index >= 0 ? array[index] : undefined)`template`;',
      errors: [undefinedAt(1, 30)],
    },
    {
      code: 'delete (index >= 0 ? array[index] : undefined);',
      errors: [undefinedAt(1, 37)],
    },
    {
      code: '((index >= 0 ? array[index] : undefined) as Function)();',
      filename: 'file.ts',
      errors: [undefinedAt(1, 31)],
    },

    // Keyword spacing
    {
      code: 'function value(index, array) { return(index) >= 0 ? array[index] : undefined; }',
      errors: [
        undefinedAt(
          1,
          68,
          'function value(index, array) { return array[index]; }',
        ),
      ],
    },
    {
      code: 'for (const value of(index >= 0 ? array[index] : undefined));',
      errors: [undefinedAt(1, 49, 'for (const value of (array[index]));')],
    },
    {
      code: 'function value(index: number, array: number[]) { return(index) >= 0 ? (array[index] as number) : undefined; }',
      filename: 'file.ts',
      errors: [
        undefinedAt(
          1,
          98,
          'function value(index: number, array: number[]) { return array[index] as number; }',
        ),
      ],
    },

    // JavaScript and index bounds
    {
      code: 'function foo() {return undefined;}',
      output: 'function foo() {return;}',
      errors: [undefinedAt(1, 24)],
    },
    {
      code: 'const foo = () => undefined;',
      output: 'const foo = () => {};',
      errors: [undefinedAt(1, 19)],
    },
    {
      code: 'const foo = () => {return undefined;};',
      output: 'const foo = () => {return;};',
      errors: [undefinedAt(1, 27)],
    },
    {
      code: 'function foo() {return       undefined;}',
      output: 'function foo() {return;}',
      errors: [undefinedAt(1, 30)],
    },
    {
      code: 'function foo() {return /* comment */ undefined;}',
      output: 'function foo() {return /* comment */;}',
      errors: [undefinedAt(1, 38)],
    },
    {
      code: 'function* foo() {yield undefined;}',
      output: 'function* foo() {yield;}',
      errors: [undefinedAt(1, 24)],
    },
    {
      code: 'function* foo() {yield                 undefined;}',
      output: 'function* foo() {yield;}',
      errors: [undefinedAt(1, 40)],
    },
    {
      code: 'let a = undefined;',
      output: 'let a;',
      errors: [undefinedAt(1, 9)],
    },
    {
      code: 'let a = undefined, b = 2;',
      output: 'let a, b = 2;',
      errors: [undefinedAt(1, 9)],
    },
    {
      code: 'foo(undefined);',
      output: 'foo();',
      errors: [undefinedAt(1, 5)],
    },
    {
      code: 'foo(undefined, undefined);',
      output: 'foo();',
      errors: [undefinedAt(1, 16)],
    },
    {
      code: 'foo(undefined,);',
      output: 'foo();',
      errors: [undefinedAt(1, 5)],
    },
    {
      code: 'foo(undefined, undefined,);',
      output: 'foo();',
      errors: [undefinedAt(1, 16)],
    },
    {
      code: 'foo(bar, undefined);',
      output: 'foo(bar);',
      errors: [undefinedAt(1, 10)],
    },
    {
      code: 'foo(bar, undefined, undefined);',
      output: 'foo(bar);',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'foo(undefined, bar, undefined);',
      output: 'foo(undefined, bar);',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'foo(bar, undefined,);',
      output: 'foo(bar,);',
      errors: [undefinedAt(1, 10)],
    },
    {
      code: 'foo(undefined, bar, undefined,);',
      output: 'foo(undefined, bar,);',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'foo(bar, undefined, undefined,);',
      output: 'foo(bar,);',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'foo(undefined, bar, undefined, undefined,);',
      output: 'foo(undefined, bar,);',
      errors: [undefinedAt(1, 32)],
    },
    {
      code: 'Promise.resolve(undefined);',
      output: 'Promise.resolve();',
      errors: [undefinedAt(1, 17)],
    },
    {
      code: 'const {foo = undefined} = {};',
      output: 'const {foo} = {};',
      errors: [undefinedAt(1, 14)],
    },
    {
      code: 'const [foo = undefined] = [];',
      output: 'const [foo] = [];',
      errors: [undefinedAt(1, 14)],
    },
    {
      code: 'function foo(bar = undefined) {}',
      output: 'function foo(bar) {}',
      errors: [undefinedAt(1, 20)],
    },
    {
      code: 'function foo({bar = undefined}) {}',
      output: 'function foo({bar}) {}',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'function foo({bar = undefined} = {}) {}',
      output: 'function foo({bar} = {}) {}',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'function foo([bar = undefined]) {}',
      output: 'function foo([bar]) {}',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'function foo([bar = undefined] = []) {}',
      output: 'function foo([bar] = []) {}',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'return undefined;',
      languageOptions: { sourceType: 'script' },
      output: 'return;',
      errors: [undefinedAt(1, 8)],
    },
    {
      code: 'const foo = index >= 0 ? array[index] : undefined;',
      errors: [undefinedAt(1, 41, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index > -1 ? array[index] : undefined;',
      errors: [undefinedAt(1, 41, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index < 0 ? undefined : array[index];',
      errors: [undefinedAt(1, 25, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index >= 0 ? (array[index] as string) : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 53, 'const foo = array[index] as string;')],
    },
    {
      code: 'const foo = index < 0 ? undefined : (array[index] as string);',
      filename: 'file.ts',
      errors: [undefinedAt(1, 25, 'const foo = array[index] as string;')],
    },
    {
      code: 'const foo = -1 < index ? array[index] : undefined;',
      errors: [undefinedAt(1, 41, 'const foo = array[index];')],
    },
    {
      code: 'const foo = 0 > index ? undefined : array[index];',
      errors: [undefinedAt(1, 25, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index <= array.length - 1 ? array[index] : undefined;',
      errors: [undefinedAt(1, 56, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index > array.length - 1 ? undefined : array[index];',
      errors: [undefinedAt(1, 40, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index >= 5 ? array[index - 5] : undefined;',
      errors: [undefinedAt(1, 45, 'const foo = array[index - 5];')],
    },
    {
      code: 'const foo = 0 <= index ? array[index] : undefined;',
      errors: [undefinedAt(1, 41, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index >= array.length ? undefined : array[index];',
      errors: [undefinedAt(1, 37, 'const foo = array[index];')],
    },
    {
      code: 'const foo = array.length <= index ? undefined : array[index];',
      errors: [undefinedAt(1, 37, 'const foo = array[index];')],
    },
    {
      code: 'const foo = array.length > index ? array[index] : undefined;',
      errors: [undefinedAt(1, 51, 'const foo = array[index];')],
    },
    {
      code: 'const foo = 5 <= index ? array[index - 5] : undefined;',
      errors: [undefinedAt(1, 45, 'const foo = array[index - 5];')],
    },
    {
      code: 'const foo = index < 5 ? undefined : array[index - 5];',
      errors: [undefinedAt(1, 25, 'const foo = array[index - 5];')],
    },
    {
      code: 'const foo = index >= 0 ? array[index] /* comment */ : undefined;',
      errors: [undefinedAt(1, 55)],
    },
    {
      code: 'const foo = index /* comment */ >= 0 ? array[index] : undefined;',
      errors: [undefinedAt(1, 55)],
    },
    {
      code: 'const foo = index >= 0 ? array[index] : /* comment */ undefined;',
      errors: [undefinedAt(1, 55)],
    },
    {
      code: `a()
index >= 0 ? [array][index] : undefined;`,
      errors: [
        undefinedAt(
          2,
          31,
          `a()
;[array][index];`,
        ),
      ],
    },

    // TypeScript
    {
      code: 'function shouldBeFlagged(): undefined {return undefined;}',
      filename: 'file.ts',
      output: 'function shouldBeFlagged(): undefined {return;}',
      errors: [undefinedAt(1, 47)],
    },
    {
      code: 'const foo = function (): undefined {return undefined};',
      filename: 'file.ts',
      output: 'const foo = function (): undefined {return};',
      errors: [undefinedAt(1, 44)],
    },
    {
      code: 'function shouldBeFlagged(): void {return undefined;}',
      filename: 'file.ts',
      output: 'function shouldBeFlagged(): void {return;}',
      errors: [undefinedAt(1, 42)],
    },
    {
      code: 'const foo = function (): void {return undefined};',
      filename: 'file.ts',
      output: 'const foo = function (): void {return};',
      errors: [undefinedAt(1, 39)],
    },
    {
      code: 'const foo = (): void => {return undefined};',
      filename: 'file.ts',
      output: 'const foo = (): void => {return};',
      errors: [undefinedAt(1, 33)],
    },
    {
      code: `class A {
	method(): void {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `class A {
	method(): void {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `function outer(): string {
	function inner(): void {
		return undefined;
	}

	return 'foo';
}`,
      filename: 'file.ts',
      output: `function outer(): string {
	function inner(): void {
		return;
	}

	return 'foo';
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `class A {
	method(): undefined {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `class A {
	method(): undefined {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: 'export function foo(): undefined {return undefined}',
      filename: 'file.ts',
      output: 'export function foo(): undefined {return}',
      errors: [undefinedAt(1, 42)],
    },
    {
      code: `const object = {
	method(): undefined {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `const object = {
	method(): undefined {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `const A = class A {
	method(): undefined {
		return undefined
	}
};`,
      filename: 'file.ts',
      output: `const A = class A {
	method(): undefined {
		return
	}
};`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `class A {
	static method(): undefined {
		return undefined
	}
}`,
      filename: 'file.ts',
      output: `class A {
	static method(): undefined {
		return
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `class A {
	get method(): undefined {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `class A {
	get method(): undefined {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `class A {
	static get method(): undefined {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `class A {
	static get method(): undefined {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `class A {
	#method(): undefined {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `class A {
	#method(): undefined {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `class A {
	private method(): undefined {
		return undefined;
	}
}`,
      filename: 'file.ts',
      output: `class A {
	private method(): undefined {
		return;
	}
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: `function foo():undefined {
	function nested() {
		return undefined;
	}

	return nested();
}`,
      filename: 'file.ts',
      output: `function foo():undefined {
	function nested() {
		return;
	}

	return nested();
}`,
      errors: [undefinedAt(3, 10)],
    },
    {
      code: 'const foo = (index as number) >= 0 ? array[index as number] : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 63, 'const foo = array[index as number];')],
    },
    {
      code: 'const foo = index! >= 0 ? array[index!] : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 43, 'const foo = array[index!];')],
    },
    {
      code: 'const foo = (index satisfies number) >= 0 ? array[index satisfies number] : undefined;',
      filename: 'file.ts',
      errors: [
        undefinedAt(1, 77, 'const foo = array[index satisfies number];'),
      ],
    },
    {
      code: 'const foo = <number>index >= 0 ? array[<number>index] : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 57, 'const foo = array[<number>index];')],
    },
    {
      code: 'const foo = index >= (0 as number) ? array[index] : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 53, 'const foo = array[index];')],
    },
    {
      code: 'const foo = index >= 5 ? array[index - (5 as number)] : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 57, 'const foo = array[index - (5 as number)];')],
    },
    {
      code: 'const foo = index <= array.length - (1 as number) ? array[index] : undefined;',
      filename: 'file.ts',
      errors: [undefinedAt(1, 68, 'const foo = array[index];')],
    },

    // JavaScript snapshots
    {
      code: `foo(
	undefined,
	bar,
	undefined,
	undefined,
	undefined,
	undefined,
)`,
      output: `foo(
	undefined,
	bar,
)`,
      errors: [undefinedAt(7, 2)],
    },
    {
      code: 'function foo([bar = undefined] = []) {}',
      output: 'function foo([bar] = []) {}',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'foo(bar, undefined, undefined);',
      output: 'foo(bar);',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'let a = undefined, b = 2;',
      output: 'let a, b = 2;',
      errors: [undefinedAt(1, 9)],
    },
    {
      code: `function foo() {
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
      output: `function foo() {
	return /* */
		/* */

			/* */

			/* */

		/* */
 /* */ ;
}`,
      errors: [undefinedAt(6, 4)],
    },
    {
      code: `function * foo() {
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
      output: `function * foo() {
	yield /* */
		/* */

			/* */

			/* */

		/* */
 /* */ ;
}`,
      errors: [undefinedAt(6, 4)],
    },
    {
      code: `const foo = () => /* */ (
	/* */
	(
		/* */
		undefined
		/* */
	)
	/* */
);`,
      output: `const foo = () => /* */
	/* */

		/* */
 {}
		/* */

	/* */
;`,
      errors: [undefinedAt(5, 3)],
    },
    {
      code: 'foo.bind(undefined)',
      output: 'foo.bind()',
      errors: [undefinedAt(1, 10)],
    },
    {
      code: 'bind(foo, undefined)',
      output: 'bind(foo)',
      errors: [undefinedAt(1, 11)],
    },
    {
      code: 'foo.bind?.(bar, undefined)',
      output: 'foo.bind?.(bar)',
      errors: [undefinedAt(1, 17)],
    },
    {
      code: 'foo[bind](bar, undefined)',
      output: 'foo[bind](bar)',
      errors: [undefinedAt(1, 16)],
    },
    {
      code: 'foo.notBind(bar, undefined)',
      output: 'foo.notBind(bar)',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: `const referencedExports =
	compilation.getDependencyReferencedExports(
		/** @type {Dependency} */ (connection.dependency),
		undefined
	);`,
      output: `const referencedExports =
	compilation.getDependencyReferencedExports(
		/** @type {Dependency} */ (connection.dependency)
	);`,
      errors: [undefinedAt(4, 3)],
    },
    {
      code: 'foo( ((a)), ((undefined)), ((undefined)), )',
      output: 'foo( ((a)), )',
      errors: [undefinedAt(1, 30)],
    },
    {
      code: 'foo( ((undefined)), ((undefined)), )',
      output: 'foo(  )',
      errors: [undefinedAt(1, 23)],
    },

    // Vue single-file components
    {
      code: `<script>
import {nextTick} from 'vue';
const foo = nextTick(undefined);
</script>`,
      filename: 'file.vue',
      skip: 'Vue single-file components require vue-eslint-parser.',
      output: `<script>
import {nextTick} from 'vue';
const foo = nextTick();
</script>`,
      errors: [undefinedAt(3, 22)],
    },

    // TypeScript snapshots
    {
      code: 'function f(foo: Type = undefined) {}',
      filename: 'file.ts',
      output: 'function f(foo?: Type) {}',
      errors: [undefinedAt(1, 24)],
    },
    {
      code: 'function f(foo?: Type = undefined) {}',
      filename: 'file.ts',
      output: 'function f(foo?: Type) {}',
      errors: [undefinedAt(1, 25)],
    },
    {
      // Rslint uses a .ts file for the explicit return type.
      code: 'function f(foo = undefined): string {}',
      filename: 'file.ts',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: 'function f(): string { let bar = undefined; }',
      filename: 'file.ts',
      output: 'function f(): string { let bar; }',
      errors: [undefinedAt(1, 34)],
    },
    {
      code: 'const f = (): string => { let bar = undefined; };',
      filename: 'file.ts',
      output: 'const f = (): string => { let bar; };',
      errors: [undefinedAt(1, 37)],
    },
    {
      code: 'const f = function(foo: Type = undefined) {}',
      filename: 'file.ts',
      output: 'const f = function(foo?: Type) {}',
      errors: [undefinedAt(1, 32)],
    },
    {
      code: 'const f = (foo: Type = undefined) => {}',
      filename: 'file.ts',
      output: 'const f = (foo?: Type) => {}',
      errors: [undefinedAt(1, 24)],
    },
    {
      code: 'const f = {method(foo: Type = undefined){}}',
      filename: 'file.ts',
      output: 'const f = {method(foo?: Type){}}',
      errors: [undefinedAt(1, 31)],
    },
    {
      code: 'const f = class {method(foo: Type = undefined){}}',
      filename: 'file.ts',
      output: 'const f = class {method(foo?: Type){}}',
      errors: [undefinedAt(1, 37)],
    },
    {
      code: 'function f(foo = undefined) {}',
      output: 'function f(foo) {}',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: 'function f(foo = undefined) {}',
      output: 'function f(foo) {}',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: 'function f(foo = undefined) {}',
      filename: 'foo.js',
      output: 'function f(foo) {}',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: 'function f(foo = undefined) {}',
      filename: 'foo.ts',
      errors: [undefinedAt(1, 18)],
    },
    // The foo.MTs snapshot uses an explicit files selector below.
    {
      code: 'function f(foo = undefined) {}',
      filename: 'foo.cts',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: 'function f(foo = undefined) {}',
      filename: 'foo.tsx',
      errors: [undefinedAt(1, 18)],
    },
    {
      code: 'function a({foo} = undefined) {}',
      filename: 'foo.ts',
      errors: [undefinedAt(1, 20)],
    },

    // Comments in initializer edits
    {
      code: 'function f(a /* keep */ = undefined) {}',
      errors: [undefinedAt(1, 27)],
    },
    {
      code: 'let x = /* keep */ undefined;',
      errors: [undefinedAt(1, 20)],
    },
    {
      code: 'const {a /* keep */ = undefined} = object;',
      errors: [undefinedAt(1, 23)],
    },
    {
      code: 'let x /* keep */ = undefined;',
      errors: [undefinedAt(1, 20)],
    },
    {
      code: 'function f(/* keep */ a = undefined) {}',
      output: 'function f(/* keep */ a) {}',
      errors: [undefinedAt(1, 27)],
    },
    {
      code: 'let x = undefined;',
      output: 'let x;',
      errors: [undefinedAt(1, 9)],
    },

    // Resource declarations
    {
      code: 'let resource = undefined;',
      filename: 'file.ts',
      output: 'let resource;',
      errors: [undefinedAt(1, 16)],
    },

    // Binding patterns and parentheses
    {
      code: 'let foo = (undefined);',
      output: 'let foo;',
      errors: [undefinedAt(1, 12)],
    },
    {
      code: 'function foo(bar = (undefined)) {}',
      output: 'function foo(bar) {}',
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'const {bar = (undefined)} = object;',
      output: 'const {bar} = object;',
      errors: [undefinedAt(1, 15)],
    },

    // Calls with built-ins
    {
      code: 'Math.sign(undefined);',
      output: 'Math.sign();',
      errors: [undefinedAt(1, 11)],
    },
    {
      code: 'Math.trunc(undefined);',
      output: 'Math.trunc();',
      errors: [undefinedAt(1, 12)],
    },
    {
      code: 'Math.round(undefined);',
      output: 'Math.round();',
      errors: [undefinedAt(1, 12)],
    },
    {
      code: 'Math.log(undefined);',
      output: 'Math.log();',
      errors: [undefinedAt(1, 10)],
    },
    {
      code: 'parseInt(undefined);',
      output: 'parseInt();',
      errors: [undefinedAt(1, 10)],
    },
    {
      code: 'Number.isNaN(undefined);',
      output: 'Number.isNaN();',
      errors: [undefinedAt(1, 14)],
    },
    {
      code: 'Promise.resolve(undefined);',
      output: 'Promise.resolve();',
      errors: [undefinedAt(1, 17)],
    },
    {
      code: 'Promise.reject(undefined);',
      output: 'Promise.reject();',
      errors: [undefinedAt(1, 16)],
    },
    {
      code: 'Promise.all(undefined);',
      output: 'Promise.all();',
      errors: [undefinedAt(1, 13)],
    },
    {
      code: 'Promise.any(undefined);',
      output: 'Promise.any();',
      errors: [undefinedAt(1, 13)],
    },
    {
      code: 'Promise.race(undefined);',
      output: 'Promise.race();',
      errors: [undefinedAt(1, 14)],
    },
    {
      code: 'Promise.allSettled(undefined);',
      output: 'Promise.allSettled();',
      errors: [undefinedAt(1, 20)],
    },
    {
      code: 'Date(undefined);',
      output: 'Date();',
      errors: [undefinedAt(1, 6)],
    },
    {
      code: 'RegExp(undefined);',
      output: 'RegExp();',
      errors: [undefinedAt(1, 8)],
    },
    {
      code: 'Boolean(undefined);',
      output: 'Boolean();',
      errors: [undefinedAt(1, 9)],
    },
    {
      code: 'Symbol(undefined);',
      output: 'Symbol();',
      errors: [undefinedAt(1, 8)],
    },

    // https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/docs/rules/no-useless-undefined.md
    {
      code: 'let foo = undefined;',
      output: 'let foo;',
      errors: [undefinedAt(1, 11)],
    },
    {
      code: 'const {foo = undefined} = bar;',
      output: 'const {foo} = bar;',
      errors: [undefinedAt(1, 14)],
    },
    {
      code: 'const noop = () => undefined;',
      output: 'const noop = () => {};',
      errors: [undefinedAt(1, 20)],
    },
    {
      code: `function foo() {
	return undefined;
}`,
      output: `function foo() {
	return;
}`,
      errors: [undefinedAt(2, 9)],
    },
    {
      code: `function* foo() {
	yield undefined;
}`,
      output: `function* foo() {
	yield;
}`,
      errors: [undefinedAt(2, 8)],
    },
    {
      code: `function foo(bar = undefined) {
}`,
      output: `function foo(bar) {
}`,
      errors: [undefinedAt(1, 20)],
    },
    {
      code: `function foo({bar = undefined}) {
}`,
      output: `function foo({bar}) {
}`,
      errors: [undefinedAt(1, 21)],
    },
    {
      code: 'foo(undefined);',
      output: 'foo();',
      errors: [undefinedAt(1, 5)],
    },
    {
      code: 'const foo = index >= 0 ? array[index] : undefined;',
      errors: [undefinedAt(1, 41, 'const foo = array[index];')],
    },
    {
      code: 'foo(bar, baz, undefined);',
      options: [{ checkArguments: true }],
      output: 'foo(bar, baz);',
      errors: [undefinedAt(1, 15)],
    },
    {
      code: 'const foo = () => undefined;',
      options: [{ checkArrowFunctionBody: true }],
      output: 'const foo = () => {};',
      errors: [undefinedAt(1, 19)],
    },
  ],
});

test('unicorn/no-useless-undefined with an uppercase TypeScript extension', async () => {
  // The default config file selector covers lowercase extensions.
  const filename = 'foo.MTs';
  const ruleName = 'unicorn/no-useless-undefined';
  const result = await lint({
    configDirectory: import.meta.dirname,
    workingDirectory: import.meta.dirname,
    config: [
      {
        files: [filename],
        plugins: ['unicorn'],
        languageOptions: { parserOptions: { projectService: false } },
        rules: { [ruleName]: 'error' },
      },
    ],
    fileContents: {
      [path.join(import.meta.dirname, filename)]:
        'function f(foo = undefined) {}',
    },
  });

  assert.equal(result.fileCount, 1);
  assert.equal(result.ruleCount, 1);
  assert.deepEqual(result.diagnostics, [
    {
      ruleName,
      severity: 'error',
      filePath: filename,
      messageId: 'no-useless-undefined',
      message: 'Do not use useless `undefined`.',
      range: { start: { line: 1, column: 18 }, end: { line: 1, column: 27 } },
    },
  ]);
});
