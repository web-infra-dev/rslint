// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
import { runInNewContext } from 'node:vm';
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { expect, test } from 'rstack/test';

const messages = {
  'no-op': 'This `splice()` call does not change the array.',
  shift: 'Prefer `.shift()` over `.splice()`.',
  unshift: 'Prefer `.unshift()` over `.splice()`.',
  pop: 'Prefer `.pop()` over `.splice()`.',
  push: 'Prefer `.push()` over `.splice()`.',
  empty: 'Prefer setting `.length = 0` over `.splice()` to empty an array.',
} as const;

type SpliceCase = [
  messageId: keyof typeof messages,
  column: number,
  code: string,
  output?: string,
  filename?: string,
];

function invalidCase([
  messageId,
  column,
  code,
  output,
  filename,
]: SpliceCase): TestCase {
  return {
    code,
    filename,
    output,
    errors: [
      {
        messageId,
        message: messages[messageId],
        line: 1,
        column,
        endLine: 1,
        endColumn: column + 6,
      },
    ],
  };
}

const mutationCases: SpliceCase[] = [
  [
    'empty',
    56,
    'const array = [1, 2]; (function () { return array; }().splice)(0); JSON.stringify(array);',
    'const array = [1, 2]; (function () { return array; }()).length = 0; JSON.stringify(array);',
  ],
  [
    'shift',
    38,
    'const array = [1, 2]; ({array}.array.splice)(0, 1); JSON.stringify(array);',
    'const array = [1, 2]; ({array}.array).shift(); JSON.stringify(array);',
  ],
];

const invalid: SpliceCase[] = [
  // JavaScript snapshot cases.
  ['no-op', 7, 'array.splice();', ''],
  // An unknown index could run conversion code or throw.
  ['no-op', 7, 'array.splice(index, 0);'],
  ['no-op', 7, 'array.splice(index, -1);'],
  ['no-op', 7, 'array.splice(index, +0);'],
  ['shift', 7, 'array.splice(0, 1);', 'array.shift();'],
  ['shift', 7, 'array.splice(+0, +1);', 'array.shift();'],
  ['shift', 7, 'array.splice(0, - -1);', 'array.shift();'],
  ['shift', 7, 'array.splice(- -0, 1);', 'array.shift();'],
  ['no-op', 7, 'array.splice(index, - -0);'],
  ['unshift', 7, 'array.splice(0, 0, element);', 'array.unshift(element);'],
  [
    'unshift',
    7,
    'array.splice(0, 0, element, ...moreElements);',
    'array.unshift(element, ...moreElements);',
  ],
  [
    'unshift',
    7,
    'array.splice(0, 0, ...elements);',
    'array.unshift(...elements);',
  ],
  ['pop', 7, 'array.splice(array.length - 1, 1);', 'array.pop();'],
  ['pop', 7, 'array.splice(array.length - +1, +1);', 'array.pop();'],
  [
    'pop',
    14,
    'object.array.splice(object.array.length - 1, 1);',
    'object.array.pop();',
  ],
  [
    'push',
    7,
    'array.splice(array.length, 0, element);',
    'array.push(element);',
  ],
  // An unknown iterator could mutate the receiver during insertion.
  ['push', 7, 'array.splice(array.length, 0, element, ...moreElements);'],
  ['push', 7, 'array.splice(array.length, 0, ...elements);'],
  [
    'push',
    14,
    'object.array.splice(object.array.length, 0, element);',
    'object.array.push(element);',
  ],
  ['empty', 7, 'array.splice(0);', 'array.length = 0;'],
  ['empty', 14, 'object.array.splice(0);', 'object.array.length = 0;'],
  ['empty', 7, 'array.splice(0, array.length);', 'array.length = 0;'],
  [
    'empty',
    14,
    'object.array.splice(0, object.array.length);',
    'object.array.length = 0;',
  ],
  ['no-op', 22, 'const result = array.splice(index, 0);'],
  ['shift', 22, 'const result = array.splice(0, 1);'],
  ['unshift', 22, 'const result = array.splice(0, 0, element);'],
  ['pop', 22, 'const result = array.splice(array.length - 1, 1);'],
  ['push', 22, 'const result = array.splice(array.length, 0, element);'],
  ['unshift', 7, 'array.splice(0, /* comment */ 0, element);'],
  ['shift', 7, 'array.splice(/* comment */ 0, 1);'],
  ['no-op', 7, 'array.splice(index, /* comment */ 0);'],
  ['no-op', 7, 'array.splice(index, 0) /* comment */;'],
  ['no-op', 7, 'array.splice(index, 0); // comment'],
  ['no-op', 7, 'array.splice(index, 0); /* comment */'],
  ['no-op', 12, 'getArray().splice(index, 0);'],
  ['no-op', 7, 'array.splice(getIndex(), 0);'],
  ['no-op', 22, 'if (condition) array.splice(index, 0);'],
  ['empty', 7, 'array.splice(0), f();', 'array.length = 0, f();'],
  ['empty', 7, 'array.splice(0, array.length), f();', 'array.length = 0, f();'],
  ['empty', 12, 'f(), array.splice(0);', 'f(), array.length = 0;'],
  ['empty', 12, 'f(), array.splice(0), g();', 'f(), array.length = 0, g();'],
  // TypeScript snapshot cases.
  [
    'shift',
    7,
    'array.splice(0 as const, 1 as const);',
    'array.shift();',
    'input.ts',
  ],
  [
    'shift',
    7,
    'array.splice(0 satisfies number, 1 satisfies number);',
    'array.shift();',
    'input.ts',
  ],
  [
    'unshift',
    7,
    'array.splice(<number>0, <number>0, element);',
    'array.unshift(element);',
    'input.ts',
  ],
  [
    'pop',
    7,
    'array.splice(array.length - 1!, 1!);',
    'array.pop();',
    'input.ts',
  ],
  [
    'empty',
    7,
    'array.splice(0!, array.length);',
    'array.length = 0;',
    'input.ts',
  ],
  [
    'push',
    28,
    '(array satisfies string[]).splice((array satisfies string[]).length, 0, element);',
    '(array satisfies string[]).push(element);',
    'input.ts',
  ],
  [
    'no-op',
    33,
    'function f(foo: number[]) { foo.splice(); }',
    'function f(foo: number[]) {  }',
    'input.ts',
  ],
  [
    'empty',
    7,
    'array.splice(0, array.length as number);',
    'array.length = 0;',
    'input.ts',
  ],
  [
    'pop',
    7,
    'array.splice((array.length as number) - 1, 1);',
    'array.pop();',
    'input.ts',
  ],
  // Parenthesized sequence arguments.
  ['unshift', 3, 'a.splice(0, 0, (b, c))', 'a.unshift((b, c))'],
  ['push', 3, 'a.splice(a.length, 0, (b, c))', 'a.push((b, c))'],
  ['unshift', 3, 'a.splice(0, 0, x, (b, c))', 'a.unshift(x, (b, c))'],
  ['unshift', 3, 'a.splice(0, 0, x)', 'a.unshift(x)'],
  // Insertion evaluation order.
  [
    'push',
    74,
    'const array = [1, 2, 3]; function f() { array.push(9); return 7; } array.splice(array.length, 0, f());',
  ],
  [
    'push',
    77,
    'const array = [1, 2, 3]; function f() { array.unshift(9); return 7; } array.splice(array.length, 0, f());',
  ],
  [
    'push',
    32,
    'const array = [1, 2, 3]; array.splice(array.length, 0, foo());',
  ],
  [
    'push',
    32,
    'const array = [1, 2, 3]; array.splice(array.length, 0, 7);',
    'const array = [1, 2, 3]; array.push(7);',
  ],
  [
    'unshift',
    32,
    'const array = [1, 2, 3]; array.splice(0, 0, 7);',
    'const array = [1, 2, 3]; array.unshift(7);',
  ],
  [
    'unshift',
    32,
    'const array = [1, 2, 3]; array.splice(0, 0, foo());',
    'const array = [1, 2, 3]; array.unshift(foo());',
  ],
  // Parenthesized callees at statement start.
  [
    'empty',
    35,
    '(function () { return [1, 2]; }().splice)(0);',
    '(function () { return [1, 2]; }()).length = 0;',
  ],
  [
    'shift',
    24,
    '({array: [1, 2]}.array.splice)(0, 1);',
    '({array: [1, 2]}.array).shift();',
  ],
  // Receiver parentheses and comments.
  [
    'empty',
    36,
    '(function () { return [1, 2]; }()!.splice)(0);',
    '(function () { return [1, 2]; }()!).length = 0;',
    'input.ts',
  ],
  [
    'empty',
    36,
    '(function () { return [1, 2]; }()).splice(0);',
    '(function () { return [1, 2]; }()).length = 0;',
  ],
  ['empty', 35, '(function () { return [1, 2]; }().splice)(/* keep */ 0);'],
  // Parenthesized callee with inserted values.
  [
    'unshift',
    21,
    '({array: [1]}.array.splice)(0, 0, 2);',
    '({array: [1]}.array).unshift(2);',
  ],
  // Remaining incorrect documentation examples.
  ['unshift', 7, 'array.splice(0, 0, item);', 'array.unshift(item);'],
  ['push', 7, 'array.splice(array.length, 0, item);', 'array.push(item);'],
  ...mutationCases,
];

new RuleTester({ plugin: 'unicorn' }).run('no-unnecessary-splice', {
  valid: [
    // First snapshot group, including a known non-array receiver.
    {
      code: 'function f(foo: {splice(start?: number, deleteCount?: number): void}) { foo.splice(); }',
      filename: 'input.ts',
    },
    'array.splice(index, 1);',
    'array.splice(index, 2);',
    'array.splice(index, 1, element);',
    'array.splice(1, 0, element);',
    'array.splice(0, 2);',
    'array.splice(0, 1, element);',
    'array.splice(array.length - 2, 1);',
    'array.splice(array.length - 1, 0, element);',
    'array.splice(otherArray.length - 1, 1);',
    'array.splice(otherArray.length, 0, element);',
    'array.splice(1, array.length);',
    'array["splice"](0, 1);',
    'array[splice](0, 1);',
    'array?.splice(0, 1);',
    'array.splice?.(0, 1);',
    'object?.array.splice(0, 1);',
    'getObject?.().array.splice(0, 1);',
    'object[foo?.bar].splice(index, 0);',
    'object[foo?.bar]().splice(index, 0);',
    'array.toSpliced(0, 1);',
    'array.splice(- -1, 1);',
    'const result = array.splice(0);',
    'const result = array.splice(0, array.length);',
    'for (const item of array.splice(0)) { item; }',
    'for (const item of array.splice(0, array.length)) { item; }',
    // TypeScript snapshot cases.
    { code: 'array.splice(0 as const, 2);', filename: 'input.ts' },
    {
      code: '(object?.array as string[]).splice(index, 0);',
      filename: 'input.ts',
    },
    {
      code: '((object?.array as {items: string[]}).items).splice(index, 0);',
      filename: 'input.ts',
    },
    {
      code: '(object?.array satisfies string[]).splice(0);',
      filename: 'input.ts',
    },
    {
      code: 'const result: unknown[] = array.splice(0 as const);',
      filename: 'input.ts',
    },
    {
      code: 'const result: unknown[] = array.splice(0 as const, array.length);',
      filename: 'input.ts',
    },
    // Correct documentation examples.
    '// Remove the no-op call.',
    'array.shift();',
    'array.unshift(item);',
    'array.pop();',
    'array.push(item);',
    'array.length = 0;',
  ],
  invalid: invalid.map(invalidCase),
});

test('rewritten parenthesized splice callees retain array mutations', () => {
  for (const [, , code, output] of mutationCases) {
    expect(runInNewContext(output!)).toBe(runInNewContext(code));
  }
});
