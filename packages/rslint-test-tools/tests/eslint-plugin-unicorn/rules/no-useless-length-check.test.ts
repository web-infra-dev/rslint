// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-useless-length-check.js
import { RuleTester } from '../rule-tester';

const tester = new RuleTester();

// Semicolons: upstream test group 1.
tester.run('no-useless-length-check', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'foo()\narray.length > 0 && (array).some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'foo()\n;(array).some(Boolean)',
    },
    {
      code: 'foo()\narray.length === 0 || (array).every(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'foo()\n;(array).every(Boolean)',
    },
    {
      code: 'const previous = value\narray.length !== 0 && (array.some)(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'const previous = value\n;(array.some)(Boolean)',
    },
    {
      code: 'foo() // keep\narray.length > 0 && (array).some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'foo() // keep\n;(array).some(Boolean)',
    },
    {
      code: 'foo();\narray.length > 0 && (array).some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'foo();\n(array).some(Boolean)',
    },
    {
      code: 'if (ready)\narray.length > 0 && (array).some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'if (ready)\n(array).some(Boolean)',
    },
    {
      code: 'foo()\narray.length > 0 && (array as number[]).some(Boolean)',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'foo()\n;(array as number[]).some(Boolean)',
    },
    {
      code: 'foo()\narray.length > 0 && (array).some(createPredicate())',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'foo()\n;(array).some(createPredicate())',
            },
          ],
        },
      ],
      output: null,
    },
  ],
});

// EditSafety: upstream test group 2.
tester.run('no-useless-length-check', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'array.length > 0 && array.some(createPredicate())',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.some(createPredicate())',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.length > 0 && array.some(predicate)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'array.some(predicate)',
    },
    {
      code: 'array.length === 0 || array.every(value => value > 0)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'array.every(value => value > 0)',
    },
    {
      code: 'array.some(value => value > 0) && array.length > 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 51,
          suggestions: [],
        },
      ],
      output: 'array.some(value => value > 0)',
    },
    {
      // Upstream autofixes; a getter can change the array.
      code: 'array.every(value => value.isValid) || array.length === 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 46,
          endLine: 1,
          endColumn: 58,
          suggestions: [
            {
              messageId: 'remove-length-check',
              output: 'array.every(value => value.isValid)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.some(value => array.pop()) && array.length > 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 43,
          endLine: 1,
          endColumn: 53,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.some(value => array.pop())',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.some(value => { return value > 0; }) && array.length > 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 63,
          suggestions: [],
        },
      ],
      output: 'array.some(value => { return value > 0; })',
    },
    {
      code: 'array.some(predicate) && array.length > 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.some(predicate)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.every(predicate) || array.length === 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 45,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.every(predicate)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.some(predicate) && array.length > 0 && array.some(predicate)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: 'array.some(predicate) && array.some(predicate)',
    },
  ],
});

// Snapshots: upstream test group 3.
tester.run('no-useless-length-check', {} as never, {
  valid: [
    {
      code: 'if (array.length === 0) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length >= 0) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 1) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length != 0) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0 && condition) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { for (const element of otherArray) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { before(); for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { for (const element of array) {} after(); }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { for (const element of array) {} } else { fallback(); }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { for (const index in array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { for (let index = 0; index < array.length; index++) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { while (condition) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array.length > 0) { for await (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array?.length > 0) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array[length] > 0) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'if (array["length"] > 0) { for (const element of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'const array = []; if (array.length > 0) { for (const array of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'const array = []; if (array.length > 0) { for (const {value: array} of array) {} }',
      filename: 'case.js',
    },
    {
      code: 'const object = {array: []}; if (object.array.length > 0) { for (const object of object.array) {} }',
      filename: 'case.js',
    },
    {
      code: 'function iterate(array: Set<number> & {length: number}) { if (array.length > 0) { for (const element of array) {} } }',
      filename: 'case.ts',
    },
    {
      code: 'function f(foo: Set<number>) { return foo.length === 0 || foo.every(Boolean); }',
      filename: 'case.ts',
    },
    {
      code: 'class A { #every() {} foo() { return this.length === 0 || this.#every(Boolean); } }',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 ?? array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 && array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: '(array.length === 0) + (array.every(Boolean))',
      filename: 'case.js',
    },
    {
      code: 'array.length === 1 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === "0" || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0. || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0x0 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length == 0 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: '0 === array.length || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array?.length === 0 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.notLength === 0 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array[length] === 0 || array.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 || array.every?.(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 || array?.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 || array.every',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 || array[every](Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array1.length === 0 || array2.every(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 ?? array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 || array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: '(array.length !== 0) - (array.some(Boolean))',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 1 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== "0" && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0. && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0x0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length <= 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length != 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: '0 !== array.length && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array?.length !== 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.notLength !== 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array[length] !== 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 && array.some?.(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 && array?.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 && array.some',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 && array.notSome(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length !== 0 && array[some](Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array1.length !== 0 && array2.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 ?? array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 || array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: '(array.length > 0) - (array.some(Boolean))',
      filename: 'case.js',
    },
    {
      code: 'array.length > 1 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > "0" && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0. && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0x0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length >= 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: '0 > array.length && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: '0 < array.length && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array?.length > 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.notLength > 0 && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 && array.some?.(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 && array?.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 && array.some',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 && array.notSome(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 && array[some](Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array1.length > 0 && array2.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'if (\n\tfoo &&\n\tarray.length !== 0 &&\n\tbar &&\n\tarray.some(Boolean)\n) {\n\t// ...\n}',
      filename: 'case.js',
    },
    {
      code: '(foo && array.length === 0) || array.every(Boolean) && foo',
      filename: 'case.js',
    },
    {
      code: 'array.length === 0 || (array.every(Boolean) && foo)',
      filename: 'case.js',
    },
    {
      code: '(foo || array.length > 0) && array.some(Boolean)',
      filename: 'case.js',
    },
    {
      code: 'array.length > 0 && (array.some(Boolean) || foo)',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: 'array.length === 0 || array.every(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'array.every(Boolean)',
    },
    {
      code: 'array.length > 0 && array.some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: 'array.some(Boolean)',
    },
    {
      code: 'array.length !== 0 && array.some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'array.some(Boolean)',
    },
    {
      code: '((\n\t((\n\t\t(( array )).length\n\t)) === (( 0 ))\n\t||\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n))',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 3,
          column: 15,
          endLine: 4,
          endColumn: 16,
          suggestions: [],
        },
      ],
      output: '((\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n))',
    },
    {
      code: '((\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n\t||\n\t((\n\t\t(( array )).length\n\t)) === (( 0 ))\n))',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 7,
          column: 15,
          endLine: 8,
          endColumn: 16,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: '((\n\t((\n\t\t(( array )).every(Boolean)\n\t))\n))',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'if ((( array.length > 0 )) && array.some(Boolean));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'if (array.some(Boolean));',
    },
    {
      code: 'if (\n\tarray.length !== 0 &&\n\tarray.some(Boolean) &&\n\tfoo\n) {\n\t// ...\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'if (\n\tarray.some(Boolean) &&\n\tfoo\n) {\n\t// ...\n}',
    },
    {
      code: '(array.length === 0 || array.every(Boolean)) || foo',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: '(array.every(Boolean)) || foo',
    },
    {
      code: 'foo || (array.length === 0 || array.every(Boolean))',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: 'foo || (array.every(Boolean))',
    },
    {
      code: '(array.length > 0 && array.some(Boolean)) && foo',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: '(array.some(Boolean)) && foo',
    },
    {
      code: 'foo && (array.length > 0 && array.some(Boolean))',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
      output: 'foo && (array.some(Boolean))',
    },
    {
      code: 'array.every(Boolean) || array.length === 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 43,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.every(Boolean)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.some(Boolean) && array.length !== 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 42,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.some(Boolean)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'array.some(Boolean) && array.length > 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 40,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.some(Boolean)',
            },
          ],
        },
      ],
      output: null,
    },
    {
      code: 'foo && array.length > 0 && array.some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'foo && array.some(Boolean)',
    },
    {
      code: 'foo || array.length === 0 || array.every(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 14,
          endLine: 1,
          endColumn: 26,
          suggestions: [],
        },
      ],
      output: 'foo || array.every(Boolean)',
    },
    {
      code: '(foo || array.length === 0) || array.every(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 27,
          suggestions: [],
        },
      ],
      output: '(foo) || array.every(Boolean)',
    },
    {
      code: 'array.length === 0 || (array.every(Boolean) || foo)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: '(array.every(Boolean) || foo)',
    },
    {
      code: '(foo && array.length > 0) && array.some(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 25,
          suggestions: [],
        },
      ],
      output: '(foo) && array.some(Boolean)',
    },
    {
      code: 'array.length > 0 && (array.some(Boolean) && foo)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 17,
          suggestions: [],
        },
      ],
      output: '(array.some(Boolean) && foo)',
    },
    {
      code: 'array.every(Boolean) || array.length === 0 || array.every(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
      output: 'array.every(Boolean) || array.every(Boolean)',
    },
    {
      code: 'array.length === 0 || array.every(Boolean) || array.length === 0',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 53,
          endLine: 1,
          endColumn: 65,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output: 'array.length === 0 || array.every(Boolean)',
            },
          ],
        },
      ],
      output: 'array.every(Boolean) || array.length === 0',
    },
    {
      code: 'array1.every(Boolean)\n|| (( array1.length === 0 || array2.length === 0 )) // Both useless\n|| array2.every(Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 26,
          suggestions: [
            {
              messageId: 'remove-length-check',
              desc: 'Remove the length check.',
              output:
                'array1.every(Boolean)\n|| (( array2.length === 0 )) // Both useless\n|| array2.every(Boolean)',
            },
          ],
        },
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 2,
          column: 37,
          endLine: 2,
          endColumn: 49,
          suggestions: [],
        },
      ],
      output:
        'array1.every(Boolean)\n|| (( array1.length === 0 )) // Both useless\n|| array2.every(Boolean)',
    },
    {
      code: "function isUselessLengthCheckNode({node, operator, siblings}) {\n\treturn (\n\t\t(\n\t\t\toperator === '||' &&\n\t\t\tzeroLengthChecks.has(node) &&\n\t\t\tsiblings.length > 0 &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarrayEveryCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t) ||\n\t\t(\n\t\t\toperator === '&&' &&\n\t\t\tnonZeroLengthChecks.has(node) &&\n\t\t\tsiblings.length > 0 &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarraySomeCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t)\n\t);\n}",
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 6,
          column: 13,
          endLine: 6,
          endColumn: 23,
          suggestions: [],
        },
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 15,
          column: 13,
          endLine: 15,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output:
        "function isUselessLengthCheckNode({node, operator, siblings}) {\n\treturn (\n\t\t(\n\t\t\toperator === '||' &&\n\t\t\tzeroLengthChecks.has(node) &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarrayEveryCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t) ||\n\t\t(\n\t\t\toperator === '&&' &&\n\t\t\tnonZeroLengthChecks.has(node) &&\n\t\t\tsiblings.some(condition =>\n\t\t\t\tarraySomeCalls.has(condition) &&\n\t\t\t\tisSameReference(node.left.object, condition.callee.object)\n\t\t\t)\n\t\t)\n\t);\n}",
    },
    {
      code: '(array as any[]).length === 0 || (array as any[]).every(Boolean)',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: '(array as any[]).every(Boolean)',
    },
    {
      code: '(array as any[]).length > 0 && (array as any[]).some(Boolean)',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      output: '(array as any[]).some(Boolean)',
    },
    {
      code: 'array!.length === 0 || array!.every(Boolean)',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 20,
          suggestions: [],
        },
      ],
      output: 'array!.every(Boolean)',
    },
    {
      code: 'array!.length > 0 && array!.some(Boolean)',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 18,
          suggestions: [],
        },
      ],
      output: 'array!.some(Boolean)',
    },
    {
      code: 'function f(array: Int8Array) { return array.length === 0 || array.every(Boolean); }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 57,
          suggestions: [],
        },
      ],
      output: 'function f(array: Int8Array) { return array.every(Boolean); }',
    },
    {
      code: 'if (array.length > 0) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) {\n\tuse(element);\n}',
    },
    {
      code: 'if (array.length !== 0) {\n\tfor (const element of array) {\n\t\t// Do work.\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) {\n\t// Do work.\n}',
    },
    {
      code: 'function iterate(array) {\n\tif (array.length > 0) {\n\t\tfor (const element of array) {\n\t\t\tuse(element);\n\t\t}\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 2,
          column: 12,
          endLine: 2,
          endColumn: 22,
          suggestions: [],
        },
      ],
      output:
        'function iterate(array) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}',
    },
    {
      code: 'if (array.length > 0) for (const element of array) { use(element); }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) { use(element); }',
    },
    {
      code: 'if (array.length > 0) { for (const element of array) use(element); }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) use(element);',
    },
    {
      code: 'if (((array.length > 0))) { for (const element of array) { use(element); } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) { use(element); }',
    },
    {
      code: 'if (object.array.length > 0) { for (const element of object.array) { use(element); } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 18,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      output: 'for (const element of object.array) { use(element); }',
    },
    {
      code: 'if (array.length > 0) { for (const [key, value] of array) { use(key, value); } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const [key, value] of array) { use(key, value); }',
    },
    {
      code: 'if (array.length > 0) { for (const element of array) { /* Keep this comment. */ use(element); } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output:
        'for (const element of array) { /* Keep this comment. */ use(element); }',
    },
    {
      code: 'if (array.length > 0) {\n\tfor (const element of array) {\n\t\t/*\n\t\tKeep this indentation.\n\t\t*/\n\t\tuse(element);\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output:
        'for (const element of array) {\n\t/*\n\tKeep this indentation.\n\t*/\n\tuse(element);\n}',
    },
    {
      code: 'if (array.length > 0)\n\tfor (const element of array) {\n\t\tuse(element);\n\t}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) {\n\tuse(element);\n}',
    },
    {
      code: 'function iterate(array) { if (array.length > 0) { for (var array of array) {} } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 37,
          endLine: 1,
          endColumn: 47,
          suggestions: [],
        },
      ],
      output: 'function iterate(array) { for (var array of array) {} }',
    },
    {
      code: 'function iterate(array) {\n  if (array.length > 0) {\n    for (const element of array) {\n      use(element);\n    }\n  }\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output:
        'function iterate(array) {\n  for (const element of array) {\n    use(element);\n  }\n}',
    },
    {
      code: 'if ((array as any[]).length > 0) { for (const element of (array as any[])) { use(element); } }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: 'for (const element of (array as any[])) { use(element); }',
    },
    {
      code: 'if ((<any[]>array).length > 0) { for (const element of (<any[]>array)) { use(element); } }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 20,
          endLine: 1,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: 'for (const element of (<any[]>array)) { use(element); }',
    },
    {
      code: 'if (array!.length !== 0) { for (const element of array!) { use(element); } }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 24,
          suggestions: [],
        },
      ],
      output: 'for (const element of array!) { use(element); }',
    },
    {
      code: 'if ((array satisfies any[]).length > 0) { for (const element of (array satisfies any[])) { use(element); } }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output:
        'for (const element of (array satisfies any[])) { use(element); }',
    },
    {
      code: 'function iterate(array: Int8Array) { if (array.length > 0) { for (const element of array) { use(element); } } }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 48,
          endLine: 1,
          endColumn: 58,
          suggestions: [],
        },
      ],
      output:
        'function iterate(array: Int8Array) { for (const element of array) { use(element); } }',
    },
    {
      code: 'if (arrays.length > 0) {\n\tfor (const array of arrays) {\n\t\tif (array.length > 0) {\n\t\t\tfor (const element of array) {\n\t\t\t\tuse(element);\n\t\t\t}\n\t\t}\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 22,
          suggestions: [],
        },
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 3,
          column: 13,
          endLine: 3,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output:
        'for (const array of arrays) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}',
    },
  ],
});

// LoopFixSafety: upstream test group 4.
tester.run('no-useless-length-check', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'if (array.length > 0) { /* Keep this comment. */ for (const element of array) { use(element); } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'if (array.length /* Keep this comment. */ > 0) { for (const element of array) { use(element); } }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 46,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'if (array.length > 0) {\n\tfor (const element of array) {\n\t\tconst value = `\n\t\t\tKeep this indentation.\n\t\t`;\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'if (array.length > 0) {\n\tfor (const element of array) {\n\t\t/*\nKeep this column.\n\t\t*/\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output:
        'for (const element of array) {\n\t/*\nKeep this column.\n\t*/\n}',
    },
    {
      code: 'if (array.length > 0) {\n\tfor (const element of array) {\nuse(element);\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) {\nuse(element);\n}',
    },
    {
      code: 'if (array.length > 0) { for (const element of array) {\n\tuse(element);\n} }',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'before(); if (array.length > 0) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 21,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'if (array.length > 0) {\r\n\tfor (const element of array) {\r\n\r\n\t\tuse(element);\r\n\t}\r\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: 'for (const element of array) {\r\n\r\n\tuse(element);\r\n}',
    },
    {
      code: 'if (array.length > 0) {\r\n\tfor (const element of array) {\r\n\r\n\t\tuse(element);\n\t\tuseAgain(element);\r\n\t}\r\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output:
        'for (const element of array) {\r\n\r\n\tuse(element);\n\tuseAgain(element);\r\n}',
    },
  ],
});

// SpaceIndentation: upstream test group 5.
tester.run('no-useless-length-check', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'function run() {\n  if (array.length > 0) {\n    for (const element of array) {\n      use({\n        element,\n      });\n    }\n  }\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 2,
          column: 13,
          endLine: 2,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output:
        'function run() {\n  for (const element of array) {\n    use({\n      element,\n    });\n  }\n}',
    },
  ],
});

// Comments: upstream test group 6.
tester.run('no-useless-length-check', {} as never, {
  valid: [],
  invalid: [
    {
      code: 'if (array.length !== 0 && /* keep */ array.some(Boolean) && foo) {}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'if (array.length !== 0 && array.some(Boolean) && foo) {}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: 'if (array.some(Boolean) && foo) {}',
    },
    {
      code: 'if (array.some(Boolean) /* keep */ && array.length !== 0) {}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 45,
          endLine: 1,
          endColumn: 57,
          suggestions: [],
        },
      ],
      output: null,
    },
    {
      code: 'if (/* keep */ array.length !== 0 && array.some(Boolean)) {}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 34,
          suggestions: [],
        },
      ],
      output: 'if (/* keep */ array.some(Boolean)) {}',
    },
  ],
});

// Syntax: upstream test group 7.
tester.run('no-useless-length-check', {} as never, {
  valid: [],
  invalid: [
    {
      code: '(array satisfies number[]).length > 0 && (array satisfies number[]).some(value => value > 0)',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 28,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
      output: '(array satisfies number[]).some(value => value > 0)',
    },
    {
      code: '(<number[]>array).every(value => value > 0) || (<number[]>array).length === 0',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 66,
          endLine: 1,
          endColumn: 78,
          suggestions: [],
        },
      ],
      output: '(<number[]>array).every(value => value > 0)',
    },
    {
      code: 'const element = <div>{array.length > 0 && array.some(value => value > 0)}</div>;',
      filename: 'case.tsx',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 1,
          column: 29,
          endLine: 1,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output: 'const element = <div>{array.some(value => value > 0)}</div>;',
    },
    {
      code: 'array.length === 0 || array.every(/* keep */ Boolean)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 19,
          suggestions: [],
        },
      ],
      output: 'array.every(/* keep */ Boolean)',
    },
  ],
});

// Documentation: upstream documentation examples.
tester.run('no-useless-length-check', {} as never, {
  valid: [
    {
      code: '// ✅\nif (array.every(Boolean));',
      filename: 'case.js',
    },
    {
      code: '// ✅\nif (array.some(Boolean));',
      filename: 'case.js',
    },
    {
      code: '// ✅\nif (array.some(Boolean));',
      filename: 'case.js',
    },
    {
      code: '// ✅\nfor (const element of array) {\n\tuse(element);\n}',
      filename: 'case.js',
    },
    {
      code: '// ✅\nconst isAllTruthyOrEmpty = array.every(Boolean);',
      filename: 'case.js',
    },
    {
      code: '// ✅\nif (array.length === 0 || anotherCheck() || array.every(Boolean));',
      filename: 'case.js',
    },
    {
      code: '// ✅\nconst isNonEmptyAllTruthyArray = array.length > 0 && array.every(Boolean);',
      filename: 'case.js',
    },
    {
      code: '// ✅\nconst isEmptyArrayOrAllTruthy = array.length === 0 || array.some(Boolean);',
      filename: 'case.js',
    },
  ],
  invalid: [
    {
      code: '// ❌\nif (array.length === 0 || array.every(Boolean));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: '// ❌\nif (array.every(Boolean));',
    },
    {
      code: '// ❌\nif (array.length !== 0 && array.some(Boolean));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 23,
          suggestions: [],
        },
      ],
      output: '// ❌\nif (array.some(Boolean));',
    },
    {
      code: '// ❌\nif (array.length > 0 && array.some(Boolean));',
      filename: 'case.js',
      errors: [
        {
          messageId: 'non-zero',
          message:
            'The non-empty check is useless as `Array#some()` returns `false` for an empty array.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: '// ❌\nif (array.some(Boolean));',
    },
    {
      code: '// ❌\nif (array.length > 0) {\n\tfor (const element of array) {\n\t\tuse(element);\n\t}\n}',
      filename: 'case.js',
      errors: [
        {
          messageId: 'for-of',
          message:
            'The non-empty check is useless as `for…of` does not iterate over an empty array.',
          line: 2,
          column: 11,
          endLine: 2,
          endColumn: 21,
          suggestions: [],
        },
      ],
      output: '// ❌\nfor (const element of array) {\n\tuse(element);\n}',
    },
    {
      code: '// ❌\nconst isAllTruthyOrEmpty = array.length === 0 || array.every(Boolean);',
      filename: 'case.js',
      errors: [
        {
          messageId: 'zero',
          message:
            'The empty check is useless as `Array#every()` returns `true` for an empty array.',
          line: 2,
          column: 34,
          endLine: 2,
          endColumn: 46,
          suggestions: [],
        },
      ],
      output: '// ❌\nconst isAllTruthyOrEmpty = array.every(Boolean);',
    },
  ],
});
