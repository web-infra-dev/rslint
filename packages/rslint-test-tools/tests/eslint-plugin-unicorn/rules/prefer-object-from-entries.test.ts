// Expected reduce fixes also remove sparse-array holes.
// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-object-from-entries.js
import { RuleTester } from '../rule-tester';

new RuleTester().run('prefer-object-from-entries', {} as never, {
  valid: [
    // Upstream typed-array runtime regressions (lint and applied output).
    {
      code: 'JSON.stringify(new Uint8Array([1, 2]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(new Uint8Array([1, 2]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(new Float64Array([1, 2]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(new Float64Array([1, 2]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(new BigInt64Array([1n, 2n]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(new BigInt64Array([1n, 2n]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(Uint16Array.from([1, 2]).reduce((object, value) => ({...object, [value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(Uint16Array.from([1, 2]).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(Uint8Array.of(1, 2).reduce((object, value) => ({...object, [value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'JSON.stringify(Uint8Array.of(1, 2).reduce((object, value) => Object.assign(object, {[value]: Number(value)}), {}));',
      filename: 'src/virtual.js',
    },
    // test.snapshot: reduce and for-of.
    {
      code: 'new Uint8Array([1, 2]).reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'const values = new Int16Array([1, 2]); values.reduce((object, value) => Object.assign(object, {[value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint16Array.from([1, 2]).reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'Uint8Array.of(1, 2).reduce((object, value) => Object.assign(object, {[value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'const values = Float32Array.of(...input); values.reduce((object, value) => Object.assign(object, {[value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'const values = new Uint8Array(input); const alias = values; alias.reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: '((new Uint8Array(values))).reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.create(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.create(null),);',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), (( Object.create(null) )),);',
      filename: 'src/virtual.js',
    },
    {
      code: 'let object = Object.create(null);\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, __proto__: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {__proto__: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, "__proto__": value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: "pairs.reduce(object => Object.assign(object, {'__proto__': value}), {});",
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), {}, extraArgument);',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce({}, object => ({...object, key}));',
      filename: 'src/virtual.js',
    },
    {
      code: 'reduce(object => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'new reduce(object => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce?.(object => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs?.reduce(object => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.notReduce(object => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), {notEmpty});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), []);',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(...[(object => ({...object, key}))], {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), ...[{}]);',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.create());',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.create(null, extraArgument));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.create?.(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object?.create(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), window.Object.create(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.notCreate(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), NotObject.create(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), object.create(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), object.CREATE(null));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), Object.create("null"));',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(callback, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(callback, Object.create(null))',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(async function * () {}, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce()',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(callback, {}, extraArgument)',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce?.(callback, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs?.reduce(callback, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.notReduce(callback, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs[reduce](callback, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(...callback, {})',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(function(object) {Object.assign(object, {key})}, {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key} + 1), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce((object = {}) => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce((object) => ({...NotSameObject, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key, anotherKey}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({keyFirst, ...object}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(async object => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(async object => await {...object, key}, {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce((...object) => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(({object}) => ({...object, key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, ...key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.assign(NotSameObject, {key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {...key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.assign?.(object, {key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object?.assign(object, {key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.notAssign(object, {key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => NotObject.assign(object, {key}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, object}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key: Object.keys(object)}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce((object, [key, value = object]) => ({...object, [key]: value}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce((object) => Object.assign(object, {object}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, key: function () { return object; }}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, method() {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {async * method() {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, async method() {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, * method() {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, async * method() {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, get key() {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'pairs.reduce(object => ({...object, set key(v) {}}), {});',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = doSomething(value);\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n\tcount++;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfoo();\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {}, otherObject = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'async function foo() {\n\tconst object = {};\n\tfor await (const [key, value] of pairs) {\n\t\tobject[key] = value;\n\t}\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] += value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject.key = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\totherObject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tthis.object[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[value] = key;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (var [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, value] of object) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [object, value] of pairs) {\n\tobject[object] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key = defaultKey, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = {};\nfor (const [key, ...value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'let object;\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
    },
    {
      code: 'const flattened = arrayOfObjects.reduce((flattened, next) => Object.assign(flattened, next), {});',
      filename: 'src/virtual.js',
    },
    // test.typescript.
    {
      code: 'function convert(values: Uint8Array) { return values.reduce((object, value) => ({...object, [value]: value}), {}); }',
      filename: 'src/virtual.ts',
    },
    {
      code: '(values as Int32Array).reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.ts',
    },
    {
      code: '(<Float32Array>values).reduce((object, value) => Object.assign(object, {[value]: value}), {});',
      filename: 'src/virtual.ts',
    },
    {
      code: 'function convert(values: Uint8Array) { return values!.reduce((object, value) => ({...object, [value]: value}), {}); }',
      filename: 'src/virtual.ts',
    },
    {
      code: '(new Uint8Array(values) satisfies Uint8Array).reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.ts',
    },
    {
      code: 'type Bytes = Uint8Array; function convert(values: Bytes) { return values.reduce((object, value) => ({...object, [value]: value}), {}); }',
      filename: 'src/virtual.ts',
    },
    // test: typed parser services.
    {
      code: 'declare function getValues(): Uint8Array; getValues().subarray(0).reduce((object, value) => ({...object, [value]: value}), {});',
      filename: 'src/virtual.ts',
    },
    // test.snapshot: functions.
    {
      code: 'underscore.fromPairs(pairs)',
      filename: 'src/virtual.js',
    },
    {
      code: '_.fromPairs',
      filename: 'src/virtual.js',
    },
    {
      code: '_.fromPairs()',
      filename: 'src/virtual.js',
    },
    {
      code: 'new _.fromPairs(pairs)',
      filename: 'src/virtual.js',
    },
    {
      code: '_.fromPairs(...[pairs])',
      filename: 'src/virtual.js',
    },
    {
      code: '_?.fromPairs(pairs)',
      filename: 'src/virtual.js',
    },
    {
      code: '_.foo(pairs)',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['foo'],
        },
      ],
    },
    {
      code: 'foo(pairs)',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['utils.object.foo'],
        },
      ],
    },
    {
      code: 'object.foo(pairs)',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['utils.object.foo'],
        },
      ],
    },
    // Documentation examples.
    {
      code: 'const object = Object.fromEntries(pairs);',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = pairs.reduce(\n\t(object, [key, value]) => ({...object, [key]: value}),\n\tObject.create(null)\n);',
      filename: 'src/virtual.js',
    },
    {
      code: 'const object = Object.fromEntries(pairs);',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['getObjectFromKeyValue', 'utils.fromPairs'],
        },
      ],
    },
  ],
  invalid: [
    // Upstream typed-array runtime regressions (lint and applied output).
    {
      code: 'JSON.stringify([1, 2].reduce((object, value) => ({...object, [value]: value}), {}));',
      filename: 'src/virtual.js',
      output:
        'JSON.stringify(Object.fromEntries([1, 2].flatMap((value) => [[value, value]])));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 29,
          suggestions: [],
        },
      ],
    },
    {
      code: 'JSON.stringify(new Array(1, 2).reduce((object, value) => ({...object, [value]: value}), {}));',
      filename: 'src/virtual.js',
      output:
        'JSON.stringify(Object.fromEntries(new Array(1, 2).flatMap((value) => [[value, value]])));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
    },
    {
      code: 'JSON.stringify(Array.from([1, 2]).reduce((object, value) => ({...object, [value]: value}), {}));',
      filename: 'src/virtual.js',
      output:
        'JSON.stringify(Object.fromEntries(Array.from([1, 2]).flatMap((value) => [[value, value]])));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 41,
          suggestions: [],
        },
      ],
    },
    {
      code: 'JSON.stringify(Array.of(1, 2).reduce((object, value) => ({...object, [value]: value}), {}));',
      filename: 'src/virtual.js',
      output:
        'JSON.stringify(Object.fromEntries(Array.of(1, 2).flatMap((value) => [[value, value]])));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 37,
          suggestions: [],
        },
      ],
    },
    // test.snapshot: reduce and for-of.
    {
      code: 'pairs.reduce(object => ({...object, key}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['key', key]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, /* keep this */ key}), {});',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), {},);',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['key', key]],));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, key,}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['key', key]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), (( {} )));',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['key', key]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce( (( object => ({...object, key}) )) , {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap( (( () => [['key', key]] )) ));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce( (( (object) => ({...object, key}) )) , {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap( (( () => [['key', key]] )) ));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce( (( (object,) => ({...object, key}) )) , {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap( (( () => [['key', key]] )) ));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [((key))] : ((value))}), {});',
      filename: 'src/virtual.js',
      output:
        'Object.fromEntries(pairs.flatMap(() => [[((key)), ((value))]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: '((\n\t(( pairs ))\n\t.reduce(\n\t\t((\n\t\t\t(object,) => ((\n\t\t\t\t((\n\t\t\t\t\tObject\n\t\t\t\t)).assign(\n\t\t\t\t\t((\n\t\t\t\t\t\tobject\n\t\t\t\t\t)),\n\t\t\t\t\t(({\n\t\t\t\t\t\t[ ((key)) ] : ((value)),\n\t\t\t\t\t}))\n\t\t\t\t)\n\t\t\t))\n\t\t)),\n\t\t({})\n\t)\n));',
      filename: 'src/virtual.js',
      output:
        '((\n\tObject.fromEntries((( pairs ))\n\t.flatMap(\n\t\t((\n\t\t\t() => \n\t\t\t\t[[((key)), ((value))]]\n\t\t\t\n\t\t))\n\t))\n));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 9,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, 0: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[0, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, true: value}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['true', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, 0n: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[0n, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, undefined: value}), {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap(() => [['undefined', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, null: value}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['null', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, var: value}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['var', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, for: value}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['for', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, default: value}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['default', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, string: value}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['string', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, "string": value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [["string", value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [0]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[0, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [true]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[true, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [0n]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[0n, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [undefined]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[undefined, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [null]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[null, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, ["for"]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [["for", value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, [string]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[string, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, ["string"]: value}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [["string", value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {key}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['key', key]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {key,}), {});',
      filename: 'src/virtual.js',
      output: "Object.fromEntries(pairs.flatMap(() => [['key', key]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => Object.assign(object, {[key]: value,}), {});',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs.flatMap(() => [[key, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce((object, element, index, array) => ({...object, key}), {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap((element, index, array) => [['key', key]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce((object, [key, value], index, array,) => ({...object, [key]: value + index + array.length}), {});',
      filename: 'src/virtual.js',
      output:
        'Object.fromEntries(pairs.flatMap(([key, value], index, array,) => [[key, value + index + array.length]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, key: function (object) { return object; }}), {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap(() => [['key', function (object) { return object; }]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, method: async () => {}}), {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap(() => [['method', async () => {}]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, method: async function * (){}}), {});',
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap(() => [['method', async function * (){}]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = ({});\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: 'const object = (Object.fromEntries(pairs));',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {\n\t// Comment\n};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 4,
          column: 1,
          endLine: 6,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\n// Comment\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 3,
          column: 1,
          endLine: 5,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\t// Comment\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 5,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n} // Comment',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of ((pairs))) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(((pairs)));',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[(key)] = (value);\n}',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (let [key, value] of pairs) {\n\tobject[key] = value;\n}\nconsole.log(object);',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);\nconsole.log(object);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n[1].forEach(foo);',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs)\n;\n[1].forEach(foo);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n(foo)();',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs)\n;\n(foo)();',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n`foo`;',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs)\n;\n`foo`;',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n+foo;',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs)\n;\n+foo;',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n/foo/.test(bar);',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs)\n;\n/foo/.test(bar);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {}\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\nfoo();',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs)\n;\nfoo();',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}\n[1].forEach(foo);',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);\n[1].forEach(foo);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    // test.typescript.
    {
      code: 'array.reduce<Record<string, Data & {b?: string}>>((result, entry) => ({...result, [entry.id]: entry.data}), {});',
      filename: 'src/virtual.ts',
      output: null,
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    // test.snapshot: functions.
    {
      code: '_.fromPairs(pairs)',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs)',
      errors: [
        {
          messageId: 'function',
          message: 'Prefer `Object.fromEntries()` over `_.fromPairs()`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 12,
          suggestions: [],
        },
      ],
    },
    {
      code: 'lodash.fromPairs(pairs)',
      filename: 'src/virtual.js',
      output: 'Object.fromEntries(pairs)',
      errors: [
        {
          messageId: 'function',
          message: 'Prefer `Object.fromEntries()` over `lodash.fromPairs()`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
    },
    {
      code: 'myFromPairsFunction(pairs)',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['myFromPairsFunction'],
        },
      ],
      output: 'Object.fromEntries(pairs)',
      errors: [
        {
          messageId: 'function',
          message:
            'Prefer `Object.fromEntries()` over `myFromPairsFunction()`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
    },
    {
      code: 'utils.object.foo(pairs)',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['utils.object.foo'],
        },
      ],
      output: 'Object.fromEntries(pairs)',
      errors: [
        {
          messageId: 'function',
          message: 'Prefer `Object.fromEntries()` over `utils.object.foo()`.',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
    },
    // test: prototype properties and comments.
    {
      code: "pairs.reduce(object => ({...object, ['__proto__']: value}), {});",
      filename: 'src/virtual.js',
      output:
        "Object.fromEntries(pairs.flatMap(() => [['__proto__', value]]));",
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, key}),/* keep */ {});',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'pairs.reduce(object => ({...object, key}), /* keep */ {});',
      filename: 'src/virtual.js',
      output: null,
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 13,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = pairs.reduce((object, [key, value]) => ({...object, [key]: value}), {});',
      filename: 'src/virtual.js',
      output:
        'const object = Object.fromEntries(pairs.flatMap(([key, value]) => [[key, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = pairs.reduce((object, element, index, array) => ({...object, [key]: value}), {});',
      filename: 'src/virtual.js',
      output:
        'const object = Object.fromEntries(pairs.flatMap((element, index, array) => [[key, value]]));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    // test.snapshot: unbraced loop.
    {
      code: 'const object = {};\nfor (const [key, value] of pairs)\n\tobject[key] = value;',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 3,
          endColumn: 22,
          suggestions: [],
        },
      ],
    },
    // Documentation examples.
    {
      code: 'const object = pairs.reduce(\n\t(object, [key, value]) => ({...object, [key]: value}),\n\t{}\n);',
      filename: 'src/virtual.js',
      output:
        'const object = Object.fromEntries(pairs.flatMap(\n\t([key, value]) => [[key, value]]\n));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = pairs.reduce(\n\t(object, [key, value]) => Object.assign(object, {[key]: value}),\n\t{}\n);',
      filename: 'src/virtual.js',
      output:
        'const object = Object.fromEntries(pairs.flatMap(\n\t([key, value]) => [[key, value]]\n));',
      errors: [
        {
          messageId: 'reduce',
          message: 'Prefer `Object.fromEntries()` over `Array#reduce()`.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = {};\nfor (const [key, value] of pairs) {\n\tobject[key] = value;\n}',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);',
      errors: [
        {
          messageId: 'loop',
          message: 'Prefer `Object.fromEntries()` over a `for-of` loop.',
          line: 2,
          column: 1,
          endLine: 4,
          endColumn: 2,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = _.fromPairs(pairs);',
      filename: 'src/virtual.js',
      output: 'const object = Object.fromEntries(pairs);',
      errors: [
        {
          messageId: 'function',
          message: 'Prefer `Object.fromEntries()` over `_.fromPairs()`.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
    },
    {
      code: 'const object = utils.fromPairs(pairs);',
      filename: 'src/virtual.js',
      options: [
        {
          functions: ['utils.fromPairs'],
        },
      ],
      output: 'const object = Object.fromEntries(pairs);',
      errors: [
        {
          messageId: 'function',
          message: 'Prefer `Object.fromEntries()` over `utils.fromPairs()`.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
    },
  ],
});
