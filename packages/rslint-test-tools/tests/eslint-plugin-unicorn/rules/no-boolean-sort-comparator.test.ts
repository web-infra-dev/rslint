// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-boolean-sort-comparator.js
import { RuleTester } from '../rule-tester';

const js = (code: string) => ({ code, filename: 'src/case.js' });
const ts = (code: string) => ({ code, filename: 'src/case.ts' });

const error = {
  messageId: 'no-boolean-sort-comparator',
  message: 'Do not return a boolean from a sort comparator.',
};

new RuleTester().run('no-boolean-sort-comparator', {} as never, {
  valid: [
    // JSDoc and Reflect cases
    ts('function run(Reflect) { values.sort(Reflect.set); }'),
    ts('values.sort(Reflect.get);'),
    ts('values.toSorted(Object.preventExtensions);'),
    ts(
      '/** @param {Set<object>} values */ function run(values) { values.sort(Reflect.isExtensible); }',
    ),
    ts(
      '/** @returns {number} */ function compare(a, b) { return source; } values.sort(compare);',
    ),
    ts(
      '/** @returns {?boolean} */ function compare(a, b) { return source; } values.sort(compare);',
    ),
    ts(
      '/** @returns {boolean} */ async function compare(a, b) { return source; } values.sort(compare);',
    ),
    ts(
      '/** @returns {boolean} */ let compare = (a, b) => source; compare = replacement; values.sort(compare);',
    ),
    ts(
      '/** @param {Set<number>} values */ function run(values) { values.toSorted((a, b) => a > b); }',
    ),
    // Native return types take precedence over JSDoc
    ts(
      '/** @type {() => boolean} */ const compare = ((): number => source) satisfies Function; values.sort(compare);',
    ),
    // Upstream snapshot cases
    ts('function f(set: Set<number>) { set.sort((a, b) => a > b); }'),
    js('array.sort()'),
    js('array.toSorted()'),
    js('array.sort(compareFunction)'),
    js('array.sort((a, b) => a - b)'),
    js('array.sort((a, b) => b - a)'),
    js('array.sort((a, b) => a.score - b.score)'),
    js('array.sort((a, b) => a.localeCompare(b))'),
    js('array.toSorted((a, b) => a.localeCompare(b))'),
    js('array.sort((a, b) => a > b ? 1 : -1)'),
    js('array.sort((a, b) => a > b ? 1 : a < b ? -1 : 0)'),
    js('array.sort((a, b) => a.foo - b.foo || a.bar - b.bar)'),
    js('array.sort((a, b) => Math.random() - 0.5)'),
    js('array.sort(...comparators)'),
    js('array.sort(...comparators, extraArgument)'),
    js('array[sort]((a, b) => a > b)'),
    js('new Set().sort((a, b) => a > b)'),
    js('({sort() {}}).sort((a, b) => a > b)'),
    js('const collection = new Set(); collection.sort((a, b) => a > b)'),
    js('const object = {}; object.sort((a, b) => a > b)'),
    js('const function_ = () => {}; function_.sort((a, b) => a > b)'),
    js('array.sort(async (a, b) => a > b)'),
    js('array.sort(function * (a, b) { return a > b; })'),
    js('array.sort((a, b) => unknownBoolean)'),
    js('const Boolean = () => 0;\narray.sort(Boolean)'),
    js('const isFinite = () => 0;\narray.sort(isFinite)'),
    js(
      'const Number = {isFinite() { return 0; }};\narray.sort(Number.isFinite)',
    ),
    js('const Number = {isNaN() { return 0; }};\narray.sort(Number["isNaN"])'),
    js('const Array = {isArray() { return 0; }};\narray.sort(Array.isArray)'),
    js('const Object = {is() { return 0; }};\narray.sort(Object.is)'),
    js(
      'array.sort((a, b) => {\n\tif (a > b) {\n\t\treturn true;\n\t}\n\n\treturn false;\n})',
    ),
    js('function compare(a, b) { return a - b; }\narray.sort(compare)'),
    // Documentation examples
    js('// ✅\narray.sort((a, b) => a - b);'),
    js('// ✅\narray.sort((a, b) => a.score - b.score);'),
    js('// ✅\nstrings.toSorted((a, b) => a.localeCompare(b));'),
    js('// ✅ Sorting strings\narray.sort((a, b) => a.localeCompare(b));'),
  ],
  invalid: [
    // JSDoc and Reflect cases
    ts('values.sort(Reflect.defineProperty);'),
    ts('values.sort(Reflect.isExtensible);'),
    ts('values.sort(Reflect.preventExtensions);'),
    ts('values.sort(Reflect.set);'),
    ts('values.sort(Reflect.setPrototypeOf);'),
    ts('values.toSorted(Reflect["isExtensible"]);'),
    ts(
      '/** @returns {boolean} */ function compare(a, b) { return source; } values.sort(compare);',
    ),
    ts(
      '/** @param {(a: number, b: number) => boolean} compare */ function run(compare) { values.toSorted(compare); }',
    ),
    ts(
      '/** @type {{(a: number, b: number): boolean; (): boolean}} */ const compare = source; values.sort(compare);',
    ),
    ts('values.sort(/** @returns {boolean} */ (a, b) => source);'),
    ts(
      '/** @param {Float32Array} values */ function run(values) { values.toSorted((a, b) => a < b); }',
    ),
    // Native return types take precedence over JSDoc
    ts(
      '/** @type {() => number} */ const compare = ((): boolean => source) satisfies Function; values.sort(compare);',
    ),
    // Upstream snapshot cases
    js('array.sort((a, b) => a > b)'),
    js('[].sort((a, b) => a > b)'),
    js('Array.from(iterable).sort((a, b) => a > b)'),
    js('new Array().toSorted((a, b) => a > b)'),
    js('array.sort((a, b) => a >= b)'),
    js('array.sort((a, b) => a < b)'),
    js('array.sort((a, b) => a <= b)'),
    js('array.sort((a, b) => b > a)'),
    js('array.sort((a, b) => b >= a)'),
    js('array.sort((a, b) => b < a)'),
    js('array.sort((a, b) => b <= a)'),
    js('array.sort((a, b) => a.score > b.score)'),
    js('array.sort((a, b) => a.score >= b.score)'),
    js('array.sort((a, b) => a[0] < b[0])'),
    js('array.sort((a, b) => a[i] <= b[i])'),
    js('array.sort((a, b) => a.foo.bar > b.foo.bar)'),
    js('array.sort((a, b) => a > c)'),
    js('array.sort((a, b) => a.foo > b.bar)'),
    js('array.sort((a, b) => a[i] > b[j])'),
    js('array.sort((a, b) => a > b, extraArgument)'),
    js('array["sort"]((a, b) => a > b)'),
    js('array["toSorted"]((a, b) => a > b)'),
    js('const method = "sort";\narray[method]((a, b) => a > b)'),
    js('array.toSorted((a, b) => a > b)'),
    js('array.sort?.((a, b) => a > b)'),
    js('array?.sort((a, b) => a > b)'),
    js('array.sort(function (a, b) { return a > b; })'),
    {
      ...js('array.sort(function (a, a) { return a > a; })'),
      languageOptions: { sourceType: 'script' as const },
    },
    js('array.sort((a, b) => {\n\treturn a > b;\n})'),
    js('array.sort((a, b) => (a) > (b))'),
    js('array.sort((a, b) => {\n\t// Compare\n\treturn a > b;\n})'),
    js('array.sort((a, b) => a === b)'),
    js('strings.toSorted((a, b) => a === b)'),
    js('array.sort((a, b) => a !== b)'),
    js('array.sort((a, b) => true)'),
    js('array.sort((a, b) => false)'),
    js('array.sort((a, b) => Boolean(a - b))'),
    js('array.sort((a, b) => !(a - b))'),
    js('array.sort((a, b) => a > b && c > d)'),
    js('array.sort((a, b) => a > b ? c > d : e > f)'),
    js('array.sort((a, b) => a in b)'),
    js('array.sort((a, b) => a instanceof b)'),
    js('array.sort(Boolean)'),
    js('array.toSorted(Boolean)'),
    js('array.sort(isFinite)'),
    js('array.sort(isNaN)'),
    js('array.sort(Array.isArray)'),
    js('array.sort(ArrayBuffer.isView)'),
    js('array.sort(Atomics.isLockFree)'),
    js('array.sort(Number.isFinite)'),
    js('array.sort(Number.isInteger)'),
    js('array.sort(Number.isNaN)'),
    js('array.sort(Number["isNaN"])'),
    js('array.sort(Number.isSafeInteger)'),
    js('array.sort(Object.hasOwn)'),
    js('array.sort(Object.is)'),
    js('array.sort(Object.isExtensible)'),
    js('array.sort(Object.isFrozen)'),
    js('array.sort(Object.isSealed)'),
    js('array.sort(Reflect.deleteProperty)'),
    js('array.sort(Reflect.has)'),
    {
      ...js('array.sort(URL.canParse)'),
      languageOptions: { globals: { URL: 'readonly' as const } },
    },
    js('const compare = (a, b) => a > b;\narray.sort(compare)'),
    js(
      'const compare = function (a, b) { return a > b; };\narray.sort(compare)',
    ),
    js('function compare(a, b) { return a > b; }\narray.sort(compare)'),
    ts('array.sort((a: number, b: number): boolean => a > b)'),
    ts('array.sort((Boolean as (value: unknown) => boolean))'),
    ts(
      'const compare: (a: number, b: number) => boolean = (a, b) => compareUnknown(a, b);\narray.sort(compare)',
    ),
    ts(
      'type Comparator = (a: number, b: number) => boolean;\narray.sort(((a, b) => a > b) as Comparator)',
    ),
    ts(
      'type Comparator = (a: number, b: number) => boolean;\narray.sort(compare as Comparator)',
    ),
    ts(
      'array.sort(((a: number, b: number) => a > b) as (a: number, b: number) => boolean)',
    ),
    ts('array.sort(compare as (a: number, b: number) => boolean)'),
    ts('array.sort((a: number, b: number): boolean => compare(a, b))'),
    ts(
      'array.sort(function (a: number, b: number): boolean {\n\treturn compare(a, b);\n})',
    ),
    ts('array.sort((a: number, b: number) => (a > b) as boolean)'),
    ts('array.sort((a: number, b: number) => <boolean>(a > b))'),
    ts('array.sort((a: number, b: number) => (a > b) satisfies boolean)'),
    ts('array.sort((a: number, b: number) => (a > b)!)'),
    ts('array.sort(<T extends {score: number}>(a, b) => a.score > b.score)'),
    ts('array.sort((a?, b?) => a > b)'),
    ts(
      'array.sort((a: number, b: number): boolean => {\n\tif (a > b) {\n\t\treturn true;\n\t}\n\n\treturn false;\n})',
    ),
    ts('array.sort((a: unknown, b: unknown) => (a as number) > (b as number))'),
    ts('array.sort(Boolean!)'),
    js('array.sort(([a], [b]) => a > b)'),
    ts('function f(array: Int8Array) { array.sort((a, b) => a > b); }'),
    js('array.sort((a, b) => a.score > b)'),
    ts('array.sort((a, b) => (a as number) > b)'),
    ts('array.sort((compare as (a: number, b: number) => boolean)!)'),
    // Documentation examples
    js('// ❌\narray.sort((a, b) => a > b);'),
    js('// ❌\narray.sort((a, b) => a.score >= b.score);'),
    js('// ❌\nstrings.toSorted((a, b) => a === b);'),
  ].map((testCase) => ({ ...testCase, errors: [error] })),
});
