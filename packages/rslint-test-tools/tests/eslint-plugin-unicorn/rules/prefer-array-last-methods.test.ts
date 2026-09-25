// Ported from eslint-plugin-unicorn v75.0.0 tests and documentation; see LICENSE.
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const filename = 'src/virtual.js';

const valid = [
  'array.findLast(logic);',
  'array.findLastIndex(logic);',
  'array.lastIndexOf(logic);',
  'array.reduceRight(logic);',
  'array.reverse();',
  'array.toReversed();',
  'array.reverse(logic).find(logic);',
  'array.toReversed(logic).find(logic);',
  'array.reverse?.().find(logic);',
  'array.toReversed?.().find(logic);',
  'array?.reverse().find(logic);',
  'array?.toReversed().find(logic);',
  'array[reverse]().find(logic);',
  'array.reverse()[find](logic);',
  'array.reverse().map(logic);',
  'array.toReversed().map(logic);',
  'const reversed = array.reverse(); reversed.find(logic);',
  'array.reverse().find;',
  'array.toReversed().find;',
].map((code) => ({ code, filename }));

const replacements = new Map([
  ['find', 'findLast'],
  ['findIndex', 'findLastIndex'],
  ['indexOf', 'lastIndexOf'],
  ['reduce', 'reduceRight'],
]);

const invalidCases = [
  ['array.reverse().find(logic);', 'reverse', 'find'],
  ['array.reverse().findIndex(logic);', 'reverse', 'findIndex'],
  ['array.reverse().indexOf(logic);', 'reverse', 'indexOf'],
  ['array.reverse().reduce(logic);', 'reverse', 'reduce'],
  ['array.toReversed().find(logic);', 'toReversed', 'find'],
  ['array.toReversed().findIndex(logic);', 'toReversed', 'findIndex'],
  ['array.toReversed().indexOf(logic);', 'toReversed', 'indexOf'],
  ['array.toReversed().reduce(logic);', 'toReversed', 'reduce'],
  ['array.reverse().find(logic, thisArgument);', 'reverse', 'find'],
  ['array.toReversed().findIndex(logic, thisArgument);', 'toReversed', 'findIndex'],
  ['array.reverse().reduce(logic, initialValue);', 'reverse', 'reduce'],
  ['array.toReversed().reduce(logic, initialValue);', 'toReversed', 'reduce'],
  ['(array.reverse()).find(logic);', 'reverse', 'find'],
  ['(array.toReversed()).find(logic);', 'toReversed', 'find'],
] as const;

const invalid = invalidCases.map(([code, reversingMethod, method]) => {
  const replacement = replacements.get(method)!;
  return {
    code,
    filename,
    errors: [
      {
        messageId: 'prefer-array-last-methods',
        message: `Prefer \`Array#${replacement}()\` over \`Array#${reversingMethod}().${method}()\`.`,
      },
    ],
  };
});

invalid.push(
  {
    code: 'array.reverse(/* comment */).find(logic);',
    filename,
    errors: [
      {
        messageId: 'prefer-array-last-methods',
        message: 'Prefer `Array#findLast()` over `Array#reverse().find()`.',
      },
    ],
  },
  {
    code: 'array.toReversed() /* comment */ .reduce(logic);',
    filename,
    errors: [
      {
        messageId: 'prefer-array-last-methods',
        message:
          'Prefer `Array#reduceRight()` over `Array#toReversed().reduce()`.',
      },
    ],
  },
);

ruleTester.run('prefer-array-last-methods', null as never, { valid, invalid });
