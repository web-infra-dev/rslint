// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/require-array-sort-compare.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const valid = (code: string, filename = 'src/virtual.js'): ValidTestCase => ({
  code,
  filename,
});

const invalid = (
  code: string,
  filename = 'src/virtual.js',
): InvalidTestCase => ({
  code,
  filename,
  errors: [
    {
      messageId: 'require-array-sort-compare',
      message: 'Pass a compare function to avoid sorting elements as strings.',
    },
  ],
});

const validCases: ValidTestCase[] = [
  valid('function f(foo: Int8Array) { foo.sort(); }', 'src/virtual.ts'),
  valid('const foo = new Int8Array(); foo.sort();', 'src/virtual.ts'),
  valid('declare function getBytes(): Int8Array; getBytes().sort();', 'src/virtual.ts'),
  valid('array.sort(compareFunction)'),
  valid('array.toSorted(compareFunction)'),
  valid('array.sort((a, b) => a - b)'),
  valid('array.sort((a, b) => a.localeCompare(b))'),
  valid('array.toSorted((a, b) => a - b)'),
  valid('array.toSorted((a, b) => a.localeCompare(b))'),
  valid('array.sort(...[])'),
  valid('array.toSorted(...[])'),
  valid('array.sort?.()'),
  valid('array?.sort?.()'),
  valid('array["sort"]()'),
  valid('array[sort]()'),
  valid('Array.prototype.sort.call(array)'),
  valid('Array.prototype.sort.apply(array)'),
  valid('Array.prototype.toSorted.call(array)'),
  valid('Array.prototype.toSorted.apply(array)'),
  valid('({sort() {}}).sort()'),
  valid('({toSorted() {}}).toSorted()'),
  valid('(() => {}).sort()'),
  valid('(class {}).sort()'),
  valid('new Set().sort()'),
  valid('new Set().toSorted()'),
  valid('const collection = new Set(); collection.sort()'),
  valid('const object = {}; object.sort()'),
  valid('const object = {}; object.toSorted()'),
  valid('const function_ = () => {}; function_.sort()'),
  valid('const array: string = ""; array.sort()', 'src/virtual.ts'),
  valid('declare function getCollection(): {sort(): void}; getCollection().sort();', 'src/virtual.ts'),
];

const invalidCases: InvalidTestCase[] = [
  invalid('array.sort()'),
  invalid('array.toSorted()'),
  invalid('array.sort(undefined)'),
  invalid('array.toSorted(undefined)'),
  invalid('array?.sort()'),
  invalid('array?.toSorted()'),
  invalid('[].sort()'),
  invalid('[].toSorted()'),
  invalid('[3, 2, 1].sort()'),
  invalid('Array.from(iterable).sort()'),
  invalid('Array.of(3, 2, 1).sort()'),
  invalid('new Array(3).sort()'),
  invalid('const array = []; array.sort()'),
  invalid('const array = Array.from(iterable); array.sort()'),
  invalid('const array = Array.of(3, 2, 1); array.toSorted()'),
  invalid('array.sort(/* comment */)'),
  invalid('const array: string[] = []; array.sort()', 'src/virtual.ts'),
  invalid('const array: Array<string> = []; array.toSorted()', 'src/virtual.ts'),
  invalid('(value as string[]).sort()', 'src/virtual.ts'),
  invalid('(<string[]>value).toSorted()', 'src/virtual.ts'),
];

new RuleTester().run('require-array-sort-compare', {} as never, {
  valid: validCases,
  invalid: invalidCases,
});
