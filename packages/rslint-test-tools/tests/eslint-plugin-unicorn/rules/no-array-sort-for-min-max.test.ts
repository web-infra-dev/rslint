// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
import { runInNewContext } from 'node:vm';
import { RuleTester, type TestCase } from '@rslint/rule-tester';
import { expect, test } from 'rstack/test';

interface MinMaxCase {
  code: string;
  filename?: string;
  method: 'min' | 'max';
  position: [number, number, number, number];
  replacement?: string;
  output?: string;
}

function invalidCase(item: MinMaxCase): TestCase {
  const [line, column, endLine, endColumn] = item.position;
  const extreme = item.method === 'min' ? 'minimum' : 'maximum';
  return {
    code: item.code,
    filename: item.filename,
    errors: [
      {
        messageId: 'no-array-sort-for-min-max/error',
        message: `Use \`Math.${item.method}()\` instead of sorting to get the ${extreme} value.`,
        line,
        column,
        endLine,
        endColumn,
        suggestions:
          item.output === undefined
            ? []
            : [
                {
                  messageId: 'no-array-sort-for-min-max/suggestion',
                  desc: `Replace with \`${item.replacement}\`.`,
                  output: item.output,
                },
              ],
      },
    ],
  };
}

const mutationCases = [
  {
    code: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return(getArray)(/* keep */).sort((a, b) => a - b)[0];}
const result = read();
({result, calls, values});`,
    method: 'min',
    position: [3, 24, 3, 71],
    replacement: 'Math.min(...(getArray)(/* keep */))',
    output: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return Math.min(...(getArray)(/* keep */));}
const result = read();
({result, calls, values});`,
  },
  {
    code: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return(getArray)(/* keep */).toSorted((a, b) => a - b).at(-1);}
const result = read();
({result, calls, values});`,
    method: 'max',
    position: [3, 24, 3, 79],
    replacement: 'Math.max(...(getArray)(/* keep */))',
    output: `let calls = 0;
const getArray = () => {calls++; return values;};
function read() {return Math.max(...(getArray)(/* keep */));}
const result = read();
({result, calls, values});`,
  },
] satisfies MinMaxCase[];

const constructorCases = [
  {
    code: 'new ([3, 1, 2].toSorted((a, b) => a - b))[0].constructor("42");',
    method: 'min',
    position: [1, 5, 1, 45],
    replacement: 'Math.min(...[3, 1, 2])',
    output: 'new (Math.min(...[3, 1, 2])).constructor("42");',
  },
  {
    code: 'new (([3, 1, 2].toSorted((a, b) => a - b))[0]).constructor("42");',
    method: 'min',
    position: [1, 6, 1, 46],
    replacement: 'Math.min(...[3, 1, 2])',
    output: 'new (Math.min(...[3, 1, 2])).constructor("42");',
  },
] satisfies MinMaxCase[];

const invalid: MinMaxCase[] = [
  // Keyword spacing.
  {
    code: 'function f() {return(getArray)(/* keep */).toSorted((a, b) => a - b).at(-1);}',
    method: 'max',
    position: [1, 21, 1, 76],
    replacement: 'Math.max(...(getArray)(/* keep */))',
    output: 'function f() {return Math.max(...(getArray)(/* keep */));}',
  },
  {
    code: 'function f() {throw(getArray)(/* keep */).toSorted((a, b) => a - b).at(-1);}',
    method: 'max',
    position: [1, 20, 1, 75],
    replacement: 'Math.max(...(getArray)(/* keep */))',
    output: 'function f() {throw Math.max(...(getArray)(/* keep */));}',
  },
  {
    code: 'if (flag) {}else(array.sort((a, b) => a - b))[0];',
    method: 'min',
    position: [1, 17, 1, 49],
    replacement: 'Math.min(...array)',
    output: 'if (flag) {}else Math.min(...array);',
  },
  {
    code: 'do(array.sort((a, b) => a - b))[0];while (flag);',
    method: 'min',
    position: [1, 3, 1, 35],
    replacement: 'Math.min(...array)',
    output: 'do Math.min(...array);while (flag);',
  },
  // Comments in the receiver.
  {
    code: '[3, /* keep */ 1, 2].sort((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 1, 1, 46],
    replacement: 'Math.min(...[3, /* keep */ 1, 2])',
    output: 'Math.min(...[3, /* keep */ 1, 2]);',
  },
  {
    code: 'getArray(/* keep */).toSorted((a, b) => a - b).at(-1);',
    method: 'max',
    position: [1, 1, 1, 54],
    replacement: 'Math.max(...getArray(/* keep */))',
    output: 'Math.max(...getArray(/* keep */));',
  },
  // Comments and TypeScript.
  {
    code: '[3, // keep\r\n1, 2].toSorted((a, b) => b - a)[0];',
    method: 'max',
    position: [1, 1, 2, 35],
    replacement: 'Math.max(...[3, // keep\r\n1, 2])',
    output: 'Math.max(...[3, // keep\r\n1, 2]);',
  },
  {
    code: '(array as /* keep */ number[]).sort((a, b) => a - b).at(0);',
    filename: 'input.ts',
    method: 'min',
    position: [1, 1, 1, 59],
    replacement: 'Math.min(...(array as /* keep */ number[]))',
    output: 'Math.min(...(array as /* keep */ number[]));',
  },
  {
    code: '(/* discard */ array).sort((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 1, 1, 47],
  },
  {
    code: 'array.sort((a, b) => a /* discard */ - b)[0];',
    method: 'min',
    position: [1, 1, 1, 45],
  },
  {
    code: 'array.sort((a, b) => a - b)[/* discard */ 0];',
    method: 'min',
    position: [1, 1, 1, 45],
  },
  {
    code: 'array./* discard */toSorted((a, b) => a - b).at(-1);',
    method: 'max',
    position: [1, 1, 1, 52],
  },
  {
    code: 'getArray(/* keep */).sort((a, b) => a - b).at(/* discard */ 0);',
    method: 'min',
    position: [1, 1, 1, 63],
  },
  // Comment-preserving suggestions with mutation and empty arrays.
  ...mutationCases,
  // Upstream snapshot cases.
  {
    code: 'const minimum = array.sort((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 17, 1, 47],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const minimum = array.toSorted((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 17, 1, 51],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const minimum = arrays[key?.name].toSorted((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 17, 1, 63],
    replacement: 'Math.min(...arrays[key?.name])',
    output: 'const minimum = Math.min(...arrays[key?.name]);',
  },
  {
    code: 'const minimum = array.sort((a, b) => a - b).at(0);',
    method: 'min',
    position: [1, 17, 1, 50],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const minimum = array.toSorted((a, b) => a - b).at(0);',
    method: 'min',
    position: [1, 17, 1, 54],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const maximum = array.sort((a, b) => a - b).at(-1);',
    method: 'max',
    position: [1, 17, 1, 51],
    replacement: 'Math.max(...array)',
    output: 'const maximum = Math.max(...array);',
  },
  {
    code: 'const maximum = array.toSorted((a, b) => a - b).at(-1);',
    method: 'max',
    position: [1, 17, 1, 55],
    replacement: 'Math.max(...array)',
    output: 'const maximum = Math.max(...array);',
  },
  {
    code: 'const maximum = array.sort((a, b) => b - a)[0];',
    method: 'max',
    position: [1, 17, 1, 47],
    replacement: 'Math.max(...array)',
    output: 'const maximum = Math.max(...array);',
  },
  {
    code: 'const maximum = array.toSorted((a, b) => b - a)[0];',
    method: 'max',
    position: [1, 17, 1, 51],
    replacement: 'Math.max(...array)',
    output: 'const maximum = Math.max(...array);',
  },
  {
    code: 'const minimum = array.sort((a, b) => b - a).at(-1);',
    method: 'min',
    position: [1, 17, 1, 51],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const minimum = array.toSorted((a, b) => b - a).at(-1);',
    method: 'min',
    position: [1, 17, 1, 55],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: `const minimum = array.sort(function (a, b) {
	return a - b;
})[0];`,
    method: 'min',
    position: [1, 17, 3, 6],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: `const maximum = array.toSorted((a, b) => {
	return a - b;
}).at(-1);`,
    method: 'max',
    position: [1, 17, 3, 10],
    replacement: 'Math.max(...array)',
    output: 'const maximum = Math.max(...array);',
  },
  {
    code: 'const minimum = (array ?? fallback).toSorted((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 17, 1, 65],
    replacement: 'Math.min(...(array ?? fallback))',
    output: 'const minimum = Math.min(...(array ?? fallback));',
  },
  {
    code: `const minimum = array.toSorted((a, b) => {
	// Compare numerically.
	return a - b;
})[0];`,
    method: 'min',
    position: [1, 17, 4, 6],
  },
  {
    code: `class ArraySubclass extends Array {
	getMinimum() {
		return super.sort((a, b) => a - b)[0];
	}
}`,
    method: 'min',
    position: [3, 10, 3, 40],
  },
  {
    code: 'const minimum = array.sort((a: number, b: number) => a - b)[0];',
    filename: 'input.ts',
    method: 'min',
    position: [1, 17, 1, 63],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const minimum = array.toSorted((a, b) => (a - b) as number)[0];',
    filename: 'input.ts',
    method: 'min',
    position: [1, 17, 1, 63],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'const minimum = array.toSorted((a, b) => a - b satisfies number)[0];',
    filename: 'input.ts',
    method: 'min',
    position: [1, 17, 1, 68],
    replacement: 'Math.min(...array)',
    output: 'const minimum = Math.min(...array);',
  },
  {
    code: 'function f(array: Int8Array) { return array.toSorted((a, b) => a - b)[0]; }',
    filename: 'input.ts',
    method: 'min',
    position: [1, 39, 1, 73],
    replacement: 'Math.min(...array)',
    output: 'function f(array: Int8Array) { return Math.min(...array); }',
  },
  {
    code: 'for (const value of array.sort((a, b) => a - b)[0]) {}',
    method: 'min',
    position: [1, 21, 1, 51],
    replacement: 'Math.min(...array)',
    output: 'for (const value of Math.min(...array)) {}',
  },
  // Constructor callee.
  ...constructorCases,
  {
    code: 'const minimum = [3, 1, 2].toSorted((a, b) => a - b)[0];',
    method: 'min',
    position: [1, 17, 1, 55],
    replacement: 'Math.min(...[3, 1, 2])',
    output: 'const minimum = Math.min(...[3, 1, 2]);',
  },
  // TypeScript constructor callee.
  {
    code: 'new ([3, 1, 2].toSorted((a, b) => a - b))[0]!.constructor("42");',
    filename: 'input.ts',
    method: 'min',
    position: [1, 5, 1, 45],
    replacement: 'Math.min(...[3, 1, 2])',
    output: 'new (Math.min(...[3, 1, 2]))!.constructor("42");',
  },
];

new RuleTester({ plugin: 'unicorn' }).run('no-array-sort-for-min-max', {
  valid: [
    // Upstream snapshot cases.
    {
      code: 'function f(set: Set<number>) { return set.toSorted((a, b) => a - b)[0]; }',
      filename: 'input.ts',
    },
    'const minimum = Math.min(...array);',
    'const minimum = array.reduce((minimum, value) => Math.min(minimum, value));',
    'const minimum = array.sort()[0];',
    'const minimum = array.toSorted()[0];',
    'const first = array.sort((a, b) => a.localeCompare(b))[0];',
    'const first = array.sort(compare)[0];',
    'const first = array.sort((a, b) => a - b || a.id - b.id)[0];',
    'const first = array.sort((a, b) => a.value - b.value)[0];',
    'const first = array.sort((a, b) => a - b)[1];',
    'const first = array.sort((a, b) => a - b).at(1);',
    'const first = array.sort((a, b) => a - b).at(index);',
    'const first = array.sort?.((a, b) => a - b)[0];',
    'const first = array?.sort((a, b) => a - b)[0];',
    'const first = array.sort((a, b) => a - b)?.[0];',
    'const first = object?.array.toSorted((a, b) => a - b)[0];',
    'const first = object?.array.toSorted((a, b) => a - b).at(0);',
    'const first = object?.array.sort((a, b) => a - b)[0];',
    'const first = object?.getArray().toSorted((a, b) => a - b)[0];',
    'const first = object?.getArray().toSorted((a, b) => a - b).at(0);',
    'const first = array["sort"]((a, b) => a - b)[0];',
    'const first = array.sort((a, b) => { return a - b; return 0; })[0];',
    'const first = array.sort((a, b) => {})[0];',
    'const first = array.sort(async (a, b) => a - b)[0];',
    'const first = array.sort(function * (a, b) { return a - b; })[0];',
    'const first = ({sort() {}}).sort((a, b) => a - b)[0];',
    'const first = new Set().sort((a, b) => a - b)[0];',
    'array.sort((a, b) => a - b)[0] = value;',
    'array.sort((a, b) => a - b)[0]++;',
    'delete array.sort((a, b) => a - b)[0];',
    'delete array.sort((a, b) => a - b).at(0);',
    'for (array.sort((a, b) => a - b)[0] of values) {}',
    'for (array.sort((a, b) => a - b)[0] in object) {}',
    {
      code: 'const first = array.sort(function (a, a) { return a - a; })[0];',
      languageOptions: { sourceType: 'script' },
    },
    {
      code: 'const first = object?.array!.toSorted((a: number, b: number) => a - b)[0];',
      filename: 'input.ts',
    },
    {
      code: 'const first = object?.getArray()!.toSorted((a: number, b: number) => a - b)[0];',
      filename: 'input.ts',
    },
    {
      code: 'const first = array.sort((a: number, b: number) => a - b).at(1);',
      filename: 'input.ts',
    },
    'new BigInt64Array().sort((a, b) => a - b)[0];',
    'new BigUint64Array().sort((a, b) => a - b).at(-1);',
    'BigInt64Array.from([]).toSorted((a, b) => a - b)[0];',
    'BigUint64Array.of(1n).toSorted((a, b) => a - b).at(-1);',
    {
      code: 'function f(values: BigInt64Array) { return values.sort((a, b) => a - b)[0]; }',
      filename: 'input.ts',
    },
    {
      code: 'function f(values: BigUint64Array) { return values.toSorted((a, b) => a - b).at(-1); }',
      filename: 'input.ts',
    },
    {
      code: 'function f(values: BigInt64Array | Int8Array) { return values.toSorted((a, b) => a - b)[0]; }',
      filename: 'input.ts',
    },
    // Additional examples from the upstream documentation.
    'const maximum = Math.max(...array);',
    'const sorted = array.toSorted((a, b) => a - b);',
    'const first = array.toSorted((a, b) => a.value - b.value)[0];',
  ],
  invalid: invalid.map(invalidCase),
});

interface ExecutionResult {
  calls: number;
  result: number | undefined;
  values: number[];
}

// The integration cases above assert the exact suggestion outputs used here.
test('the suggested code evaluates the receiver once and leaves the array in its original order', () => {
  for (const { code, method, output } of mutationCases) {
    for (const values of [[3, 1, 2], []]) {
      const original: ExecutionResult = runInNewContext(code, {
        values: [...values],
      });
      const suggested: ExecutionResult = runInNewContext(output, {
        values: [...values],
      });
      expect(original.calls).toBe(1);
      expect(suggested.calls).toBe(1);
      expect(suggested.result).toBe(Math[method](...values));
      expect(original.result).toBe(
        values.length > 0 ? suggested.result : undefined,
      );
      expect([...suggested.values]).toEqual(values);
      expect([...original.values]).toEqual(
        method === 'min' ? values.toSorted((a, b) => a - b) : values,
      );
    }
  }
});

test('the suggested code preserves the number constructor', () => {
  for (const { code, output } of constructorCases) {
    const original: unknown = runInNewContext(code);
    const suggested: unknown = runInNewContext(output);
    expect(typeof suggested).toBe('object');
    expect(Number(suggested)).toBe(Number(original));
    expect(Number(suggested)).toBe(42);
  }
});
