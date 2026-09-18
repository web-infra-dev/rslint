// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

interface UpstreamCase {
  code: string;
  filename?: string;
  options?: unknown[];
}
function normalizeCase(value: string | UpstreamCase): UpstreamCase {
  return typeof value === 'string'
    ? { code: value, filename: 'src/virtual.js' }
    : { filename: 'src/virtual.js', ...value };
}

const valid: (string | UpstreamCase)[] = [
  {
    code: 'function f(foo: Set<number>) { foo.sort(); }',
    filename: 'src/virtual.ts',
  },
  'sorted =[...array].toSorted()',
  'sorted =array.toSorted()',
  'sorted =[...array].sort',
  'sorted =[...array].sort?.()',
  'array.sort()',
  'array.sort?.()',
  'array?.sort()',
  'if (true) array.sort()',
  'sorted = array.sort(...[])',
  'sorted = array.sort(...[compareFn])',
  'sorted = array.sort(compareFn, extraArgument)',
  'sorted = collection.sort({field: 1})',
  'sorted = query.sort("field")',
  'sorted = query.sort(1)',
  'sorted = query.sort(-1)',
  'sorted = query.sort(+1)',
  'sorted = query.sort(`field`)',
  'sorted = query.sort([criteria])',
  'const docs = collection.find({id}).sort({expireAt: -1}).limit(1).toArray()',
  '[...array].sort({field: 1})',
  {
    code: 'collection.sort({field: 1})',
    options: [{ allowExpressionStatement: false }],
  },
  'const sorted = array.toSorted();',
  'const sorted = [...iterable].toSorted();',
  'const sorted = array.toSorted((a, b) => a - b);',
  'const sortedArray = array.toSorted();',
];

const invalid0: (string | UpstreamCase)[] = [
  'sorted = [...array].sort()',
  'sorted = [...array]?.sort()',
  'sorted = array.sort()',
  'sorted = array?.sort()',
  'sorted = [...array].sort(compareFn)',
  'sorted = [...array]?.sort(compareFn)',
  'sorted = array.sort(compareFn)',
  'sorted = array?.sort(compareFn)',
  { code: 'array.sort()', options: [{ allowExpressionStatement: false }] },
  { code: 'array?.sort()', options: [{ allowExpressionStatement: false }] },
  { code: '[...array].sort()', options: [{ allowExpressionStatement: false }] },
  'sorted = [...(0, array)].sort()',
  'sorted = [...a + b].sort()',
  'const sorted = [...array].sort();',
  'const sorted = [...iterable].sort();',
  'const sorted = [...array].sort((a, b) => a - b);',
  { code: 'array.sort();', options: [{ allowExpressionStatement: false }] },
];

new RuleTester().run('no-array-sort', {} as never, {
  valid: valid.map(normalizeCase),
  invalid: [
    ...invalid0.map((value) => ({
      ...normalizeCase(value),
      errors: [
        {
          messageId: 'error',
          message: 'Use `Array#toSorted()` instead of `Array#sort()`.',
        },
      ],
    })),
  ],
});
