// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-return-array-push.js
import assert from 'node:assert/strict';
import { runInNewContext } from 'node:vm';
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { test } from 'rstack/test';

const iifeReturn =
  'const array = []; function run() { return function () { return array; }().push(1); } run(); array.join();';
const iifeSuggestion =
  'const array = []; function run() { (function () { return array; }().push(1)); return; } run(); array.join();';

function invalid(
  code: string,
  method: 'push' | 'unshift',
  line: number,
  column: number,
  suggested?: string,
  filename?: string,
): TestCase {
  return {
    code,
    filename,
    errors: [
      {
        messageId: 'no-return-array-push/error',
        message: `Do not use the return value of \`.${method}(…)\`.`,
        line,
        column,
        endLine: line,
        endColumn: column + method.length,
        suggestions: suggested
          ? [
              {
                messageId: 'no-return-array-push/suggestion',
                desc: `Separate the \`${method}()\` call from \`return\`.`,
                output: suggested,
              },
            ]
          : [],
      },
    ],
  };
}

function withTypeInformation<T extends { code: string }>(testCase: T) {
  return {
    ...testCase,
    filename: 'file.ts',
    languageOptions: { parserOptions: { projectService: true } },
  };
}

const tester = new RuleTester({
  plugin: 'unicorn',
  fixtureFiles: {
    'tsconfig.json': '{"compilerOptions":{"strict":true},"files":["file.ts"]}',
  },
});

tester.run('no-return-array-push', {
  valid: [
    iifeSuggestion,
    // JSDoc annotations.
    {
      code: '/** @param {Map<string, number>} values */ function run(values) { return values.push(value); }',
      filename: 'file.ts',
    },
    {
      code: '/** @returns {Set<number>} */ function read() { return source; } function run() { return read().unshift(value); }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {() => string} read */ function run(read) { return read().push(value); }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {?Array<number>} stream */ function run(stream) { return stream.push(value); }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {Array<number>} values */ function run(values) { return values.push(value).then(consume); }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {Array<number>} values */ function run(values) { return void values.push(value); }',
      filename: 'file.ts',
    },
    // TypeScript annotations take precedence over JSDoc.
    {
      code: '/** @param {Array<number>} values */ function run(values: Set<number>) { return values.push(value); }',
      filename: 'file.ts',
    },
    // JavaScript.
    `function foo() {
	array.push(value);
	return;
}`,
    `function foo() {
	array.unshift(value);
	return;
}`,
    `function foo() {
	return array.length + 1;
}`,
    `function foo() {
	return array.push();
}`,
    `function foo() {
	return array.unshift();
}`,
    `function foo() {
	return stream.push(chunk);
}`,
    `function foo() {
	return this.push(chunk);
}`,
    `function foo() {
	return this.stream.push(chunk);
}`,
    `function foo() {
	return process.stdin.push(chunk);
}`,
    `function foo() {
	return process.stdout.push(chunk);
}`,
    `function foo() {
	return process.stderr.push(chunk);
}`,
    `function foo() {
	return router.push(to);
}`,
    `function foo() {
	return this.router.push(to);
}`,
    `function foo() {
	return this.$router.push(to);
}`,
    `function foo() {
	return stream?.push(chunk);
}`,
    `function foo() {
	return this?.push(chunk);
}`,
    `function foo() {
	return process.stdin?.push(chunk);
}`,
    `function foo() {
	return array['push'](value);
}`,
    `function foo() {
	return array['unshift'](value);
}`,
    `function foo() {
	return array[push](value);
}`,
    `function foo() {
	return array[unshift](value);
}`,
    `function foo() {
	return push(value);
}`,
    `function foo() {
	return unshift(value);
}`,
    `function foo() {
	array?.push(value);
}`,
    `function foo() {
	array?.unshift(value);
}`,
    'void array.push(value);',
    'void array.unshift(value);',
    'void array?.push(value);',
    'void array?.unshift(value);',
    'const length = void array.push(value);',
    'const foo = value => void array.push(value);',
    `function foo() {
	return void array.push(value);
}`,
    `function foo() {
	return router.push(to).catch(() => {});
}`,
    `function foo() {
	return router.push(to).then(onFulfilled);
}`,
    'const promise = router.push(to).finally(cleanup);',
    `function foo() {
	return array.push(value).toString();
}`,
    `function foo() {
	return array.unshift(value).foo;
}`,
    `function foo() {
	return router.push(to)?.catch(() => {});
}`,
    'array.push(value)[0];',
    // TypeScript syntax.
    { code: 'array.push(value) as number;', filename: 'file.ts' },
    { code: 'array.unshift(value) as number;', filename: 'file.ts' },
    { code: '<number>array.push(value);', filename: 'file.ts' },
    { code: '<number>array.unshift(value);', filename: 'file.ts' },
    { code: 'array.push(value)!;', filename: 'file.ts' },
    { code: 'array.unshift(value)!;', filename: 'file.ts' },
    { code: 'array.push(value) satisfies number;', filename: 'file.ts' },
    { code: 'array.unshift(value) satisfies number;', filename: 'file.ts' },
    { code: 'void (array.push(value) as number);', filename: 'file.ts' },
    { code: 'void (array.unshift(value) as number);', filename: 'file.ts' },
    {
      code: 'const onClick = () => router.push(`/${region}/customers/${String(result.data)}` as Route);',
      filename: 'file.ts',
    },
    {
      code: `function foo() {
	return (array.push(value) as Foo).bar;
}`,
      filename: 'file.ts',
    },
    {
      code: `interface Router {
	push(to: string): Promise<void>;
}

function foo(router: Router) {
	return router.push(to);
}`,
      filename: 'file.ts',
    },
    {
      code: `class Queue {
	unshift(value: unknown) {
		return value;
	}
}

function foo(queue: Queue) {
	return queue.unshift(value);
}`,
      filename: 'file.ts',
    },
    // Type information.
    ...[
      'declare const router: {push(to: string): Promise<void>}; function foo() { return router.push(to); }',
      'function foo(navigation: {push(to: string): Promise<void>}) { return navigation.push(to); }',
      'declare const queue: {unshift(value: unknown): number}; function foo() { return queue.unshift(value); }',
    ].map((code) => withTypeInformation({ code })),
    // Documentation examples.
    `function add(item) {
	items.push(item);
	return;
}`,
    `const add = item => {
	items.push(item);
};`,
    `items.unshift(item);
const length = items.length;`,
    `function getNextLength(item) {
	items.push(item);
	return items.length;
}`,
    'void items.push(item);',
    'target.push(...source);',
  ],
  invalid: [
    // JSDoc annotations.
    invalid(
      '/** @param {Array<number>} stream */ function run(stream) { return stream.push(value); }',
      'push',
      1,
      75,
      '/** @param {Array<number>} stream */ function run(stream) { stream.push(value); return; }',
      'file.ts',
    ),
    invalid(
      '/** @returns {ReadonlyArray<number>} */ function read() { return source; } function run() { return read().unshift(value); }',
      'unshift',
      1,
      107,
      '/** @returns {ReadonlyArray<number>} */ function read() { return source; } function run() { read().unshift(value); return; }',
      'file.ts',
    ),
    invalid(
      '/** @param {?Array<number>} values */ function run(values) { return values.push(value); }',
      'push',
      1,
      76,
      '/** @param {?Array<number>} values */ function run(values) { values.push(value); return; }',
      'file.ts',
    ),
    invalid(
      '/** @param {Uint8Array} values */ function run(values) { return values.push(value); }',
      'push',
      1,
      72,
      '/** @param {Uint8Array} values */ function run(values) { values.push(value); return; }',
      'file.ts',
    ),
    invalid(
      '/** @param {Array<number>} values */ const run = values => values.push(value);',
      'push',
      1,
      67,
      undefined,
      'file.ts',
    ),
    invalid(
      '/** @param {Array<number>} values */ function run(values) { return /* keep */ values.unshift(value); }',
      'unshift',
      1,
      86,
      undefined,
      'file.ts',
    ),
    // TypeScript annotations take precedence over JSDoc.
    invalid(
      '/** @param {Map<string, number>} stream */ function run(stream: number[]) { return stream.push(value); }',
      'push',
      1,
      91,
      '/** @param {Map<string, number>} stream */ function run(stream: number[]) { stream.push(value); return; }',
      'file.ts',
    ),
    // JavaScript.
    invalid('const length = array.push(value);', 'push', 1, 22),
    invalid('const length = routerItems.push(value);', 'push', 1, 28),
    invalid('const length = array.unshift(value);', 'unshift', 1, 22),
    invalid('const length = stream.unshift(chunk);', 'unshift', 1, 23),
    invalid('console.log(array.push(value));', 'push', 1, 19),
    invalid('console.log(array.unshift(value));', 'unshift', 1, 19),
    invalid('const length = await array.push(value);', 'push', 1, 28),
    invalid('const length = await array.unshift(value);', 'unshift', 1, 28),
    invalid('async function foo() { await array.push(value); }', 'push', 1, 36),
    invalid(
      'async function foo() { await array.unshift(value); }',
      'unshift',
      1,
      36,
    ),
    invalid('(array.push(value), sideEffect());', 'push', 1, 8),
    invalid('(array.unshift(value), sideEffect());', 'unshift', 1, 8),
    invalid('(sideEffect(), array.push(value));', 'push', 1, 22),
    invalid('(sideEffect(), array.unshift(value));', 'unshift', 1, 22),
    invalid('condition && array.push(value);', 'push', 1, 20),
    invalid('condition && array.unshift(value);', 'unshift', 1, 20),
    invalid('array.push(value) && sideEffect();', 'push', 1, 7),
    invalid('array.unshift(value) && sideEffect();', 'unshift', 1, 7),
    invalid('condition || array.push(value);', 'push', 1, 20),
    invalid('condition || array.unshift(value);', 'unshift', 1, 20),
    invalid('array.push(value) || sideEffect();', 'push', 1, 7),
    invalid('array.unshift(value) || sideEffect();', 'unshift', 1, 7),
    invalid('condition ?? array.push(value);', 'push', 1, 20),
    invalid('condition ?? array.unshift(value);', 'unshift', 1, 20),
    invalid('array.push(value) ?? sideEffect();', 'push', 1, 7),
    invalid('array.unshift(value) ?? sideEffect();', 'unshift', 1, 7),
    invalid('condition ? array.push(value) : sideEffect();', 'push', 1, 19),
    invalid(
      'condition ? array.unshift(value) : sideEffect();',
      'unshift',
      1,
      19,
    ),
    invalid('condition ? sideEffect() : array.push(value);', 'push', 1, 34),
    invalid(
      'condition ? sideEffect() : array.unshift(value);',
      'unshift',
      1,
      34,
    ),
    invalid('array.push(value) ? sideEffect() : other();', 'push', 1, 7),
    invalid('array.unshift(value) ? sideEffect() : other();', 'unshift', 1, 7),
    invalid('for (array.push(value); condition; ) {}', 'push', 1, 12),
    invalid('for (; array.push(value); ) {}', 'push', 1, 14),
    invalid('for (; condition; array.unshift(value)) {}', 'unshift', 1, 25),
    invalid('for (; array.unshift(value); ) {}', 'unshift', 1, 14),
    invalid(
      `function foo() {
	return array.push(value);
}`,
      'push',
      2,
      15,
      `function foo() {
	array.push(value); return;
}`,
    ),
    invalid(
      `function foo() {
	return array.unshift(value);
}`,
      'unshift',
      2,
      15,
      `function foo() {
	array.unshift(value); return;
}`,
    ),
    invalid(
      `function foo() {
	return (array.push(value));
}`,
      'push',
      2,
      16,
      `function foo() {
	array.push(value); return;
}`,
    ),
    invalid(
      `function foo() {
	return (array.unshift(value));
}`,
      'unshift',
      2,
      16,
      `function foo() {
	array.unshift(value); return;
}`,
    ),
    invalid(
      `function foo() {
	return array?.push(value);
}`,
      'push',
      2,
      16,
    ),
    invalid(
      `function foo() {
	return array?.unshift(value);
}`,
      'unshift',
      2,
      16,
    ),
    invalid(
      `function foo() {
	return array.push?.(value);
}`,
      'push',
      2,
      15,
    ),
    invalid(
      `function foo() {
	return array.unshift?.(value);
}`,
      'unshift',
      2,
      15,
    ),
    invalid(
      `async function foo() {
	return await array.push(value);
}`,
      'push',
      2,
      21,
    ),
    invalid(
      `async function foo() {
	return await array.unshift(value);
}`,
      'unshift',
      2,
      21,
    ),
    invalid('const foo = value => array.push(value);', 'push', 1, 28),
    invalid('const foo = value => array.unshift(value);', 'unshift', 1, 28),
    invalid(
      'const foo = async value => array.push(await value);',
      'push',
      1,
      34,
    ),
    invalid(
      'const foo = async value => array.unshift(await value);',
      'unshift',
      1,
      34,
    ),
    invalid(
      `function foo() {
	return /* comment */ array.push(value);
}`,
      'push',
      2,
      29,
    ),
    invalid(
      `function foo() {
	return /* comment */ array.unshift(value);
}`,
      'unshift',
      2,
      29,
    ),
    invalid(
      `function foo() {
	return array.push(/* comment */ value);
}`,
      'push',
      2,
      15,
      `function foo() {
	array.push(/* comment */ value); return;
}`,
    ),
    invalid(
      `function foo() {
	return array.unshift(/* comment */ value);
}`,
      'unshift',
      2,
      15,
      `function foo() {
	array.unshift(/* comment */ value); return;
}`,
    ),
    invalid(
      `function foo() {
	if (condition) return array.push(value);
}`,
      'push',
      2,
      30,
    ),
    invalid(
      `function foo() {
	if (condition) return array.unshift(value);
}`,
      'unshift',
      2,
      30,
    ),
    invalid(
      `function foo() {
	return condition && array.push(value);
}`,
      'push',
      2,
      28,
    ),
    invalid(
      `function foo() {
	return condition && array.unshift(value);
}`,
      'unshift',
      2,
      28,
    ),
    invalid(
      `function foo() {
	return condition ? array.push(value) : value;
}`,
      'push',
      2,
      27,
    ),
    invalid(
      `function foo() {
	return condition ? array.unshift(value) : value;
}`,
      'unshift',
      2,
      27,
    ),
    invalid(
      `function foo() {
	return condition ? value : array.push(value);
}`,
      'push',
      2,
      35,
    ),
    invalid(
      `function foo() {
	return condition ? value : array.unshift(value);
}`,
      'unshift',
      2,
      35,
    ),
    invalid(
      `function foo() {
	return (sideEffect(), array.push(value));
}`,
      'push',
      2,
      30,
    ),
    invalid(
      `function foo() {
	return (sideEffect(), array.unshift(value));
}`,
      'unshift',
      2,
      30,
    ),
    invalid(
      `function foo() {
	foo()
	return [array].push(value);
}`,
      'push',
      3,
      17,
      `function foo() {
	foo()
	;[array].push(value); return;
}`,
    ),
    invalid(
      `function foo() {
	foo()
	return [array].unshift(value);
}`,
      'unshift',
      3,
      17,
      `function foo() {
	foo()
	;[array].unshift(value); return;
}`,
    ),
    // TypeScript syntax.
    invalid(
      'const foo = (value: string) => array.push(value);',
      'push',
      1,
      38,
      undefined,
      'file.ts',
    ),
    invalid(
      'const foo = (value: string) => array.unshift(value);',
      'unshift',
      1,
      38,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return array.push(value) as number;
}`,
      'push',
      2,
      15,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return array.unshift(value) as number;
}`,
      'unshift',
      2,
      15,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return <number>array.push(value);
}`,
      'push',
      2,
      23,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return <number>array.unshift(value);
}`,
      'unshift',
      2,
      23,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return array.push(value)!;
}`,
      'push',
      2,
      15,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return array.unshift(value)!;
}`,
      'unshift',
      2,
      15,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return array.push(value) satisfies number;
}`,
      'push',
      2,
      15,
      undefined,
      'file.ts',
    ),
    invalid(
      `function foo() {
	return array.unshift(value) satisfies number;
}`,
      'unshift',
      2,
      15,
      undefined,
      'file.ts',
    ),
    invalid(
      'const foo = (value: string) => array.push(value) satisfies number;',
      'push',
      1,
      38,
      undefined,
      'file.ts',
    ),
    invalid(
      'const foo = (value: string) => array.unshift(value) satisfies number;',
      'unshift',
      1,
      38,
      undefined,
      'file.ts',
    ),
    // Type information.
    ...[
      invalid(
        'declare const array: number[]; function foo() { return array.push(value); }',
        'push',
        1,
        62,
        'declare const array: number[]; function foo() { array.push(value); return; }',
      ),
      invalid(
        'function foo(stream: string[]) { return stream.push(value); }',
        'push',
        1,
        48,
        'function foo(stream: string[]) { stream.push(value); return; }',
      ),
      invalid(
        'function foo(this: string[]) { return this.push(value); }',
        'push',
        1,
        44,
        'function foo(this: string[]) { this.push(value); return; }',
      ),
      invalid(
        'function foo(router: string[]) { return router.push(value); }',
        'push',
        1,
        48,
        'function foo(router: string[]) { router.push(value); return; }',
      ),
      invalid(
        'function foo(this: {router: string[]}) { return this.router.push(value); }',
        'push',
        1,
        61,
        'function foo(this: {router: string[]}) { this.router.push(value); return; }',
      ),
      invalid(
        'function foo(this: {$router: string[]}) { return this.$router.push(value); }',
        'push',
        1,
        63,
        'function foo(this: {$router: string[]}) { this.$router.push(value); return; }',
      ),
      invalid(
        'function foo(this: {stream: string[]}) { return this.stream.push(value); }',
        'push',
        1,
        61,
        'function foo(this: {stream: string[]}) { this.stream.push(value); return; }',
      ),
      invalid(
        'declare const process: {stdin: string[]}; function foo() { return process.stdin.push(value); }',
        'push',
        1,
        81,
        'declare const process: {stdin: string[]}; function foo() { process.stdin.push(value); return; }',
      ),
      invalid(
        'declare const process: {stdout: string[]}; function foo() { return process.stdout.push(value); }',
        'push',
        1,
        83,
        'declare const process: {stdout: string[]}; function foo() { process.stdout.push(value); return; }',
      ),
      invalid(
        'declare const process: {stderr: string[]}; function foo() { return process.stderr.push(value); }',
        'push',
        1,
        83,
        'declare const process: {stderr: string[]}; function foo() { process.stderr.push(value); return; }',
      ),
    ].map(withTypeInformation),
    // IIFE receivers.
    invalid(
      'function run() { return function () { return []; }().push(1); }',
      'push',
      1,
      54,
      'function run() { (function () { return []; }().push(1)); return; }',
    ),
    invalid(
      'function run() { return (function () { return []; }().push(1)); }',
      'push',
      1,
      55,
      'function run() { (function () { return []; }().push(1)); return; }',
    ),
    invalid(
      'function run() { return function () { return []; }().unshift(1); }',
      'unshift',
      1,
      54,
      'function run() { (function () { return []; }().unshift(1)); return; }',
    ),
    invalid(
      'function run() { return (function () { return []; }().unshift(1)); }',
      'unshift',
      1,
      55,
      'function run() { (function () { return []; }().unshift(1)); return; }',
    ),
    // IIFE comments and semicolons.
    invalid(
      `function run() { setup()
return function () { return []; }().push(1); }`,
      'push',
      2,
      37,
      `function run() { setup()
;(function () { return []; }().push(1)); return; }`,
    ),
    invalid(
      'function run() { return function () { return []; }().push(/* keep */ 1); }',
      'push',
      1,
      54,
      'function run() { (function () { return []; }().push(/* keep */ 1)); return; }',
    ),
    invalid(
      'function run() { return function () { return []; }()!.push(1); }',
      'push',
      1,
      55,
      'function run() { (function () { return []; }()!.push(1)); return; }',
      'file.ts',
    ),
    // IIFE suggestion preserves the side effect.
    invalid(iifeReturn, 'push', 1, 75, iifeSuggestion),
    // Documentation examples.
    invalid(
      `function add(item) {
	return items.push(item);
}`,
      'push',
      2,
      15,
      `function add(item) {
	items.push(item); return;
}`,
    ),
    invalid('const add = item => items.push(item);', 'push', 1, 27),
    invalid('const length = items.unshift(item);', 'unshift', 1, 22),
    invalid('source.forEach(item => target.push(item));', 'push', 1, 31),
  ],
});

test('separating return preserves the IIFE receiver and mutation', () => {
  assert.equal(runInNewContext(iifeSuggestion), runInNewContext(iifeReturn));
});
