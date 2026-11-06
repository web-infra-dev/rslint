// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-string-slice.js
import { RuleTester } from '../rule-tester';

const errorsSubstr = [
  {
    messageId: 'substr',
    message: 'Prefer `String#slice()` over `String#substr()`.',
  },
];
const errorsSubstring = [
  {
    messageId: 'substring',
    message: 'Prefer `String#slice()` over `String#substring()`.',
  },
];

new RuleTester().run('prefer-string-slice', {} as never, {
  valid: [
    // Basic cases.
    {
      code: 'const substr = foo.substr',
      filename: 'file.js',
    },
    {
      code: 'const substring = foo.substring',
      filename: 'file.js',
    },
    {
      code: 'foo.slice()',
      filename: 'file.js',
    },
    {
      code: 'foo.slice(0)',
      filename: 'file.js',
    },
    {
      code: 'foo.slice(1, 2)',
      filename: 'file.js',
    },
    {
      code: 'foo?.slice(1, 2)',
      filename: 'file.js',
    },
    {
      code: 'foo?.slice?.(1, 2)',
      filename: 'file.js',
    },
    {
      code: 'foo?.bar.baz.slice(1, 2)',
      filename: 'file.js',
    },
    {
      code: 'foo.slice(-3, -2)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(0, 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(1, 2)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(2, 1)',
      filename: 'file.js',
    },
    {
      code: '"foo".substring(1, 2)',
      filename: 'file.js',
    },
    {
      code: '"foo".substring(2, 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(index, index + 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(index, 1 + index)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(index - 1, index)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(index + 1, index)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(1 + index, index)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(index, /* comment */ index + 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring(index, index + /* comment */ 1)',
      filename: 'file.js',
    },
    {
      code: 'foo?.substring(index, index + 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring?.(index, index + 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring((( index )), (( index )) + 1)',
      filename: 'file.js',
    },
    {
      code: 'foo.substring((( index )) - 1, (( index )))',
      filename: 'file.js',
    },
    {
      code: 'function foo(collection: {substring(start: number): string}) { collection.substring(1); }',
      filename: 'file.ts',
    },
  ],
  invalid: [
    // Basic cases.
    {
      code: 'const modes = new Set(["foo"]); modes.clear(); foo.substr(modes.size ? 0 : 1, length);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'const modes = new Set(["foo"]); modes.clear(); foo.substr((modes.size ? 0 : 1) as number, length);',
      filename: 'file.ts',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'const modes = new Set(["foo"]); modes.clear(); foo.substr(+(modes.size ? 0 : 1), length);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'const modes = new Set(["foo"]); modes.clear(); foo.substr((modes.size && 0) || index, length);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'const object = {value: true}; Object.defineProperty(object, "value", {get() { return 0; }}); foo.substr(object.value ? 0 : index, length);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr()',
      filename: 'file.js',
      output: 'foo.slice()',
      errors: errorsSubstr,
    },
    {
      code: 'foo?.substr()',
      filename: 'file.js',
      output: 'foo?.slice()',
      errors: errorsSubstr,
    },
    {
      code: 'foo.bar?.substring()',
      filename: 'file.js',
      output: 'foo.bar?.slice()',
      errors: errorsSubstring,
    },
    {
      code: 'foo?.[0]?.substring()',
      filename: 'file.js',
      output: 'foo?.[0]?.slice()',
      errors: errorsSubstring,
    },
    {
      code: 'foo.bar.substr?.()',
      filename: 'file.js',
      output: 'foo.bar.slice?.()',
      errors: errorsSubstr,
    },
    {
      code: 'foo.bar?.substring?.()',
      filename: 'file.js',
      output: 'foo.bar?.slice?.()',
      errors: errorsSubstring,
    },
    {
      code: 'foo.bar?.baz?.substr()',
      filename: 'file.js',
      output: 'foo.bar?.baz?.slice()',
      errors: errorsSubstr,
    },
    {
      code: 'foo.bar?.baz.substring()',
      filename: 'file.js',
      output: 'foo.bar?.baz.slice()',
      errors: errorsSubstring,
    },
    {
      code: 'foo.bar.baz?.substr()',
      filename: 'file.js',
      output: 'foo.bar.baz?.slice()',
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr()',
      filename: 'file.js',
      output: '"foo".slice()',
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(bar.length, Math.min(baz, 100))',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(1, "abc".length)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr("1", 2)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'const length = 123;\n"foo".substr(1, length)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'const length = 123;\n"foo".substr(0, length)',
      filename: 'file.js',
      output:
        'const length = 123;\n"foo".slice(0, Math.max(0, length === undefined ? Infinity : length))',
      errors: errorsSubstr,
    },
    {
      code: 'const length = 123;\n"foo".substr(\'0\', length)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(0, -1)',
      filename: 'file.js',
      output: '"foo".slice(0, 0)',
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(0, "foo".length)',
      filename: 'file.js',
      output: '"foo".slice(0, "foo".length)',
      errors: errorsSubstr,
    },
    {
      code: 'const length = 123;\n"foo".substr(1, length - 4)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(1, length)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: "const uri = 'foo';\n((uri || '')).substr(1)",
      filename: 'file.js',
      output: "const uri = 'foo';\n((uri || '')).slice(1)",
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(start)',
      filename: 'file.js',
      output: 'foo.slice(start)',
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(1)',
      filename: 'file.js',
      output: '"foo".slice(1)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(start, length)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(1, 2)',
      filename: 'file.js',
      output: '"foo".slice(1, 3)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(1, 0)',
      filename: 'file.js',
      output: 'foo.slice(1, 1)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(1, 2, 3)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: '"Sample".substr(0, "Sample".lastIndexOf("/"))',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, foo.length - 2)',
      filename: 'file.js',
      output: 'foo.slice(0, -2)',
      errors: errorsSubstr,
    },
    {
      code: 'foo?.substr(0, foo.length - 2)',
      filename: 'file.js',
      output: 'foo?.slice(0, -2)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, (foo.length - 2))',
      filename: 'file.js',
      output: 'foo.slice(0, -2)',
      errors: errorsSubstr,
    },
    {
      code: '"foo".substr(0, "foo".length - 2)',
      filename: 'file.js',
      output: '"foo".slice(0, -2)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.bar.substr(0, foo.bar.length - 2)',
      filename: 'file.js',
      output:
        'foo.bar.slice(0, Math.max(0, foo.bar.length - 2 === undefined ? Infinity : foo.bar.length - 2))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, foo.length - count)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, foo.length - count === undefined ? Infinity : foo.length - count))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, foo.length - 0)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, foo.length - 0 === undefined ? Infinity : foo.length - 0))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, foo.length - 2.5)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, foo.length - 2.5 === undefined ? Infinity : foo.length - 2.5))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, foo.length - -1)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, foo.length - -1 === undefined ? Infinity : foo.length - -1))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, bar.length - 2)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, bar.length - 2 === undefined ? Infinity : bar.length - 2))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, foo.length - /* comment */ 2)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, foo.length - /* comment */ 2 === undefined ? Infinity : foo.length - /* comment */ 2))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, /* comment */ foo.length - 2)',
      filename: 'file.js',
      output:
        'foo.slice(0, /* comment */ Math.max(0, foo.length - 2 === undefined ? Infinity : foo.length - 2))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, (/* comment */ foo.length - 2))',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (/* comment */ foo.length - 2) === undefined ? Infinity : (/* comment */ foo.length - 2)))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, (/* comment */ (foo.length - 2)))',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (/* comment */ (foo.length - 2)) === undefined ? Infinity : (/* comment */ (foo.length - 2))))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, /* comment */ (foo.length - 2))',
      filename: 'file.js',
      output:
        'foo.slice(0, /* comment */ Math.max(0, (foo.length - 2) === undefined ? Infinity : (foo.length - 2)))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring()',
      filename: 'file.js',
      output: 'foo.slice()',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring()',
      filename: 'file.js',
      output: '"foo".slice()',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(1)',
      filename: 'file.js',
      output: '"foo".slice(1)',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(2, 4)',
      filename: 'file.js',
      output: '"foo".slice(2, 4)',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(4, 2)',
      filename: 'file.js',
      output: '"foo".slice(2, 4)',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(-1, -5)',
      filename: 'file.js',
      output: '"foo".slice(0, 0)',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(-1, 2)',
      filename: 'file.js',
      output: '"foo".slice(0, 2)',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(length)',
      filename: 'file.js',
      output: '"foo".slice(Math.max(0, length))',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring("fo".length)',
      filename: 'file.js',
      output: '"foo".slice("fo".length)',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(0, length)',
      filename: 'file.js',
      output:
        '"foo".slice(0, Math.max(0, length === undefined ? Infinity : length))',
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(length, 0)',
      filename: 'file.js',
      output: '"foo".slice(0, Math.max(0, length))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(start)',
      filename: 'file.js',
      output: 'foo.slice(Math.max(0, start))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(start, end)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: '"foo".substring(1, 3)',
      filename: 'file.js',
      output: '"foo".slice(1, 3)',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(1, 2, 3)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    // TypeScript assertions.
    {
      code: 'function foo() {\n\treturn (bar as string).substr(3);\n}',
      filename: 'file.ts',
      output: 'function foo() {\n\treturn (bar as string).slice(3);\n}',
      errors: errorsSubstr,
    },
    {
      code: 'function foo() {\n\treturn ((bar as string)).substring(3);\n}',
      filename: 'file.ts',
      output: 'function foo() {\n\treturn ((bar as string)).slice(3);\n}',
      errors: errorsSubstring,
    },
    {
      code: "const value: 'foo' = 'foo';\nvalue.substring(1);",
      filename: 'file.ts',
      output: "const value: 'foo' = 'foo';\nvalue.slice(1);",
      errors: errorsSubstring,
    },
    // Snapshot cases.
    {
      code: '/* 1 */ (( /* 2 */ 0 /* 3 */, /* 4 */ foo /* 5 */ )) /* 6 */\n\t. /* 7 */ substring /* 8 */ (\n\t\t/* 9 */ (( /* 10 */ bar /* 11 */ )) /* 12 */,\n\t\t/* 13 */ (( /* 14 */ 0 /* 15 */ )) /* 16 */,\n\t\t/* 17 */\n\t)\n/* 18 */',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, ...bar)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(...bar)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, (100, 1))',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (100, 1) === undefined ? Infinity : (100, 1)))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, 1, extraArgument)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr((0, bar.length), (0, baz.length))',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring((10, 1), 0)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, (10, 1)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, (10, 1))',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (10, 1) === undefined ? Infinity : (10, 1)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, await 1)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring((10, bar))',
      filename: 'file.js',
      output: 'foo.slice(Math.max(0, (10, bar)))',
      errors: errorsSubstring,
    },
    {
      code: 'const string = "::";\nconst output = string.substr(-2, 2);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0.5, 0.5)',
      filename: 'file.js',
      output: 'foo.slice(0.5, 0)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(1.5, 2.5)',
      filename: 'file.js',
      output: 'foo.slice(1.5, 3)',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(-1.5, 2.5)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring(1, -n)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(1, -Infinity)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(1, -foo.length)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring((/* keep */ 5), 2)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring((/* keep */ -1))',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring((/* keep */ -1), 3)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(1, (/* keep */ 2))',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, (/* keep */ -1))',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring(0, (/* keep */ 2))',
      filename: 'file.js',
      output: 'foo.slice(0, (/* keep */ 2))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, undefined)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, undefined === undefined ? Infinity : undefined))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring(0, undefined)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, undefined === undefined ? Infinity : undefined))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, null)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, null === undefined ? Infinity : null))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, NaN)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, NaN === undefined ? Infinity : NaN))',
      errors: errorsSubstring,
    },
    // Operator precedence and side effects.
    {
      code: 'foo.substring(0, a || b)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (a || b) === undefined ? Infinity : (a || b)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, a && b)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (a && b) === undefined ? Infinity : (a && b)))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring(0, () => a || b)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (() => a || b) === undefined ? Infinity : (() => a || b)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, () => {})',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (() => {}) === undefined ? Infinity : (() => {})))',
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring(0, a ? b : c)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (a ? b : c) === undefined ? Infinity : (a ? b : c)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, a = b)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'async function f() { foo.substring(0, await a || b); }',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'function * f() { foo.substring(0, yield a || b); }',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, (a || b))',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (a || b) === undefined ? Infinity : (a || b)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, a ?? b)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, (a ?? b) === undefined ? Infinity : (a ?? b)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, a + b)',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, a + b === undefined ? Infinity : a + b))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, a.b)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, a.b === undefined ? Infinity : a.b))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, -a)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, -a === undefined ? Infinity : -a))',
      errors: errorsSubstring,
    },
    // Swapped arguments.
    {
      code: 'foo.substring(a, 0)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, a))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(a || b, 0)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, a || b))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(a, 0.0)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, a))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(bar(), 0)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, bar()))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring((a, b), 0)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, (a, b)))',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring((a /* comment */), 0)',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, a)',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, a === undefined ? Infinity : a))',
      errors: errorsSubstring,
    },
    // a `null` end is not rewritten to the whole string.
    {
      code: 'foo.substring(0, null);',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, null === undefined ? Infinity : null));',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, null);',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, null === undefined ? Infinity : null));',
      errors: errorsSubstr,
    },
    {
      code: 'const n = null; foo.substring(0, n);',
      filename: 'file.js',
      output:
        'const n = null; foo.slice(0, Math.max(0, n === undefined ? Infinity : n));',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, undefined);',
      filename: 'file.js',
      output:
        'foo.slice(0, Math.max(0, undefined === undefined ? Infinity : undefined));',
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, n);',
      filename: 'file.js',
      output: 'foo.slice(0, Math.max(0, n === undefined ? Infinity : n));',
      errors: errorsSubstring,
    },
    // an end with a side effect is not rewritten.
    {
      code: 'foo.substring(0, getEnd());',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, end());',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substring(0, a++);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, getEnd());',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substr(0, tag`end`);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    {
      code: 'foo.substring(0, tag`end`);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstring,
    },
    {
      code: 'foo.substr(0, [...ends][0]);',
      filename: 'file.js',
      output: null,
      errors: errorsSubstr,
    },
    // Static template receivers.
    {
      code: '`abc`.substr(0, `abc`.length - 1);',
      filename: 'file.js',
      output: '`abc`.slice(0, -1);',
      errors: errorsSubstr,
    },
    {
      code: '`abc`.substr(0, "abc".length - 1);',
      filename: 'file.js',
      output: '`abc`.slice(0, -1);',
      errors: errorsSubstr,
    },
    {
      code: '`ab\\u{63}`.substr(0, `abc`.length - 1);',
      filename: 'file.js',
      output: '`ab\\u{63}`.slice(0, -1);',
      errors: errorsSubstr,
    },
    // Documentation examples.
    {
      code: "const str = 'hello world';\n\n// ❌\nstr.substr(0, 5);\n\n// ✅\nstr.slice(0, 5);\n// → 'hello'\n",
      filename: 'file.js',
      output:
        "const str = 'hello world';\n\n// ❌\nstr.slice(0, 5);\n\n// ✅\nstr.slice(0, 5);\n// → 'hello'\n",
      errors: errorsSubstr,
    },
    {
      code: "const str = 'hello world';\n\n// ❌\nstr.substring(6, 11);\n\n// ✅\nstr.slice(6, 11);\n// → 'world'\n",
      filename: 'file.js',
      output:
        "const str = 'hello world';\n\n// ❌\nstr.slice(6, 11);\n\n// ✅\nstr.slice(6, 11);\n// → 'world'\n",
      errors: errorsSubstring,
    },
    {
      code: "const str = 'hello world';\n\n// ❌ substring swaps arguments when start > end\nstr.substring(5, 0);\n// → 'hello' (confusing!)\n\n// ✅ slice handles it intuitively\nstr.slice(5, 0);\n// → '' (empty, as expected)\n\n// ✅ slice handles negative indices\nstr.slice(-5);\n// → 'world' (last 5 characters)\n",
      filename: 'file.js',
      output:
        "const str = 'hello world';\n\n// ❌ substring swaps arguments when start > end\nstr.slice(0, 5);\n// → 'hello' (confusing!)\n\n// ✅ slice handles it intuitively\nstr.slice(5, 0);\n// → '' (empty, as expected)\n\n// ✅ slice handles negative indices\nstr.slice(-5);\n// → 'world' (last 5 characters)\n",
      errors: errorsSubstring,
    },
  ],
});
