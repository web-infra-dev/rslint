// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-abs.js
import { RuleTester, type TestCase } from '@rslint/rule-tester';

function absCase(
  code: string,
  output: string | null,
  [line, column, endLine, endColumn]: [number, number, number, number],
  filename?: string,
): TestCase {
  return {
    code,
    filename,
    errors: [
      {
        messageId: 'prefer-math-abs',
        message: 'Prefer `Math.abs()` to simplify this expression.',
        line,
        column,
        endLine,
        endColumn,
        suggestions:
          output === null
            ? []
            : [
                {
                  messageId: 'prefer-math-abs/suggestion',
                  desc: 'Replace with `Math.abs()`.',
                  output,
                },
              ],
      },
    ],
  };
}

const project = { parserOptions: { project: './tsconfig.json' } };
const annotatedValid = [
  {
    code: 'function read(): bigint { return source; } let value = read(); value < 0 ? -value : value; function run(value = read()) { return value > limit || value < -limit; }',
    filename: 'file.ts',
  },
  {
    code: '/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value < 0 ? -value : value; }',
    filename: 'file.ts',
  },
  {
    code: '/** @param {bigint} value */ function run(value) { return value < 0 ? -value : value; }',
    filename: 'file.ts',
  },
  {
    code: '/** @param {number | bigint} value */ function run(value) { return 0 < value ? value : -value; }',
    filename: 'file.ts',
  },
  {
    code: '/** @param {bigint} limit */ function run(value, limit) { return -limit >= value || limit <= value; }',
    filename: 'file.ts',
  },
];
const annotatedInvalid = [
  absCase(
    '/** @param {bigint} value */ function run(value: number) { return value < 0 ? -value : value; }',
    '/** @param {bigint} value */ function run(value: number) { return Math.abs(value); }',
    [1, 67, 1, 93],
    'file.ts',
  ),
];

const tester = new RuleTester({
  plugin: 'unicorn',
  fixtureFiles: {
    'tsconfig.json': '{"compilerOptions":{"target":"ESNext","strict":true}}',
  },
});

tester.run('prefer-math-abs', {
  valid: [
    // Parentheses and comments that cannot be kept.
    'const value = 1n; (/* keep */ value) < 0 ? -value : value;',

    // BigInt initializers and default values from annotated calls.
    '/** @returns {bigint} */ function read() { return source; } function run(value = read()) { return value < 0 ? -value : value; }',
    '/** @returns {bigint | number} */ function read() { return source; } function run(value = read()) { return value > limit || value < -limit; }',
    '/** @param {() => bigint} read */ function run(read, value = read()) { return value < 0 ? -value : value; }',
    '/** @returns {bigint} */ function read() { return source; } let value = read(); value < 0 ? -value : value;',
    '/** @returns {bigint | number} */ function read() { return source; } var value = read(); value > limit || value < -limit;',
    '/** @returns {bigint} */ function read() { return source; } let value = -read(); value < 0 ? -value : value;',
    '/** @returns {bigint} */ function read() { return source; } function run(value, limit = read()) { return value > limit || value < -limit; }',

    // JSDoc declarations, unions, and operations.
    '/** @param {bigint} input */ function run(input) { const value = input + 1n; return value < 0 ? -value : value; }',
    '/** @type {bigint} */ let value; value < 0 ? -value : value;',
    '/** @param {number | bigint} value */ function run(value) { return 0 < value ? value : -value; }',
    '/** @param {?bigint} value */ function run(value) { return value >= 0 ? value : -value; }',
    '/** @param {bigint} [value] */ function run(value) { return value < 0 ? -value : value; }',
    '/** @param {bigint} value */ function run(value, limit) { return value > limit || value < -limit; }',
    '/** @param {bigint} limit */ function run(value, limit) { return -limit >= value || limit <= value; }',
    '/** @returns {bigint} */ function read() { return source; } const value = read(); value < 0 ? -value : value;',
    '/** @param {bigint} input */ function run(input, limit) { const value = ~input; return value > limit || value < -limit; }',

    // TypeScript declarations and JSDoc, with and without a project.
    ...annotatedValid,
    ...annotatedValid.map((item) => ({ ...item, languageOptions: project })),

    // Runtime BigInt examples.
    '/** @returns {bigint} */ function read() { return -5n; } function run(value = read()) { return value < 0 ? -value : value; } run();',
    '/** @returns {bigint} */ function read() { return -5n; } let value = read(); value < 0 ? -value : value;',
    '/** @param {bigint} value */ function run(value) { return value < 0 ? -value : value; } run(-5n);',
    '/** @param {bigint} value */ function run(value, limit) { return value > limit || value < -limit; } run(-5n, 2);',

    // JavaScript snapshot cases.
    'const absolute = Math.abs(value);',
    'value < 0 ? value : -value',
    'value > 0 ? -value : value',
    'value < 1 ? -value : value',
    'value < 0 ? -otherValue : value',
    'value < 0 ? -value : otherValue',
    'value < 0n ? -value : value',
    'value < BigInt(0) ? -value : value',
    'const value = 1n; value < 0 ? -value : value;',
    'const value = BigInt(1); value < 0 ? -value : value;',
    'function foo(value = 1n) { return value < 0 ? -value : value; }',
    'function foo(value = BigInt(1)) { return value < 0 ? -value : value; }',
    'value > limit || value < limit',
    'value > limit || otherValue < -limit',
    'value > limit || value <= -limit',
    'value > limit || value < -otherLimit',
    'value > 0n || value < -0n',
    'value > BigInt(0) || value < -BigInt(0)',
    'const value = 1n; const limit = 2n; value > limit || value < -limit;',
    '(value + 1) < 0 ? -(value + 1) : value + 1',
    'Math.abs(value) > limit',
    'Math.abs(value) >= limit',
    'value > limit && value < -limit',

    // TypeScript snapshot cases.
    {
      code: 'function foo(value: bigint) {\n\treturn value < 0 ? -value : value;\n}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value) {\n\treturn (value as bigint) < 0 ? -(value as bigint) : (value as bigint);\n}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value) {\n\treturn (<bigint>value) < 0 ? -(<bigint>value) : (<bigint>value);\n}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value: bigint, limit: bigint) {\n\treturn value > limit || value < -limit;\n}',
      filename: 'file.ts',
    },
    {
      code: 'const value: bigint = 1n;\nvalue < 0 ? -value : value;',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value: number | bigint) {\n\treturn value < 0 ? -value : value;\n}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value) {\n\treturn (value as number | bigint) < 0 ? -(value as number | bigint) : (value as number | bigint);\n}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value: bigint & {}) {\n\treturn value < 0 ? -value : value;\n}',
      filename: 'file.ts',
    },
    {
      code: 'const value: 1n = 1n;\nvalue < 0 ? -value : value;',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value) {\n\treturn (value satisfies bigint) < 0 ? -(value satisfies bigint) : (value satisfies bigint);\n}',
      filename: 'file.ts',
    },
    {
      code: 'function foo(value) {\n\treturn (value as bigint)! < 0 ? -(value as bigint)! : (value as bigint)!;\n}',
      filename: 'file.ts',
    },

    // Upstream documentation examples.
    'const absolute = Math.abs(value);',
    'if (Math.abs(number) > MAXIMUM) {\n}',
  ],
  invalid: [
    // Comments kept in the selected operands.
    absCase(
      '(/* keep */ value) < 0 ? -value : value;',
      'Math.abs((/* keep */ value));',
      [1, 1, 1, 40],
    ),
    absCase(
      'value > (/* keep */ limit) || value < -limit;',
      'Math.abs(value) > (/* keep */ limit);',
      [1, 1, 1, 45],
    ),
    absCase(
      'value < -limit || (/* keep */ value) > limit;',
      'Math.abs((/* keep */ value)) > limit;',
      [1, 1, 1, 45],
    ),
    absCase(
      '(value as /* keep */ number) < 0 ? -value : value;',
      'Math.abs((value as /* keep */ number));',
      [1, 1, 1, 50],
      'file.ts',
    ),

    // Parentheses and comments that cannot be kept.
    absCase(
      '((value /* keep */)) >= 0 ? value : -value;',
      'Math.abs(((value /* keep */)));',
      [1, 1, 1, 43],
    ),
    absCase(
      'value > (limit // keep\r\n) || value < -limit;',
      'Math.abs(value) > (limit // keep\r\n);',
      [1, 1, 2, 20],
    ),
    absCase(
      'limit as /* keep */ number < value || value < -(limit as number);',
      'Math.abs(value) > (limit as /* keep */ number);',
      [1, 1, 1, 65],
      'file.ts',
    ),
    absCase('value /* keep */ < 0 ? -value : value;', null, [1, 1, 1, 38]),
    absCase('value < 0 ? -(/* keep */ value) : value;', null, [1, 1, 1, 40]),
    absCase('value < 0 ? -value : (/* keep */ value);', null, [1, 1, 1, 40]),
    absCase(
      'value > limit || (/* keep */ value) < -limit;',
      null,
      [1, 1, 1, 45],
    ),
    absCase(
      'value > limit || value < -(/* keep */ limit);',
      null,
      [1, 1, 1, 45],
    ),
    absCase(
      '(/* keep */ object.value) < 0 ? -object.value : object.value;',
      null,
      [1, 1, 1, 61],
    ),
    absCase(
      'value > (/* keep */ object.limit) || value < -object.limit;',
      null,
      [1, 1, 1, 59],
    ),

    // Suggestion-only rewrite for coercion and signed zero.
    absCase(
      '(/* keep */ value) < 0 ? -value : value;',
      'Math.abs((/* keep */ value));',
      [1, 1, 1, 40],
    ),

    // BigInt initializers and default values from annotated calls.
    absCase(
      '/** @returns {number} */ function read() { return source; } function run(value = read()) { return value < 0 ? -value : value; }',
      '/** @returns {number} */ function read() { return source; } function run(value = read()) { return Math.abs(value); }',
      [1, 99, 1, 125],
    ),
    absCase(
      '/** @returns {number} */ function read() { return source; } let value = read(); value > limit || value < -limit;',
      '/** @returns {number} */ function read() { return source; } let value = read(); Math.abs(value) > limit;',
      [1, 81, 1, 112],
    ),

    // JSDoc declarations, unions, and operations.
    absCase(
      '/** @param {number} value */ function run(value) { return value < 0 ? -value : value; }',
      '/** @param {number} value */ function run(value) { return Math.abs(value); }',
      [1, 59, 1, 85],
    ),
    absCase(
      '/** @param {number} value */ function run(value, limit) { return value > limit || value < -limit; }',
      '/** @param {number} value */ function run(value, limit) { return Math.abs(value) > limit; }',
      [1, 66, 1, 97],
    ),

    // TypeScript declarations and JSDoc, with and without a project.
    ...annotatedInvalid,
    ...annotatedInvalid.map((item) => ({ ...item, languageOptions: project })),

    // TypeScript assertions used as the range limit.
    absCase(
      'limit as number < value || value < -(limit as number)',
      'Math.abs(value) > (limit as number)',
      [1, 1, 1, 54],
      'file.ts',
    ),
    absCase(
      'value <= -(limit as number) || limit as number <= value',
      'Math.abs(value) >= (limit as number)',
      [1, 1, 1, 56],
      'file.ts',
    ),
    absCase(
      '(limit as number) < value || value < -(limit as number)',
      'Math.abs(value) > (limit as number)',
      [1, 1, 1, 56],
      'file.ts',
    ),
    absCase(
      'limit satisfies number < value || value < -(limit satisfies number)',
      'Math.abs(value) > (limit satisfies number)',
      [1, 1, 1, 68],
      'file.ts',
    ),
    absCase(
      'value <= -(limit satisfies number) || limit satisfies number <= value',
      'Math.abs(value) >= (limit satisfies number)',
      [1, 1, 1, 70],
      'file.ts',
    ),
    absCase(
      '(limit satisfies number) < value || value < -(limit satisfies number)',
      'Math.abs(value) > (limit satisfies number)',
      [1, 1, 1, 70],
      'file.ts',
    ),

    // JavaScript snapshot cases.
    absCase('value < 0 ? -value : value', 'Math.abs(value)', [1, 1, 1, 27]),
    absCase('value <= 0 ? -value : value', 'Math.abs(value)', [1, 1, 1, 28]),
    absCase('value > 0 ? value : -value', 'Math.abs(value)', [1, 1, 1, 27]),
    absCase('value >= 0 ? value : -value', 'Math.abs(value)', [1, 1, 1, 28]),
    absCase('value < 0 ? 0 - value : value', 'Math.abs(value)', [1, 1, 1, 30]),
    absCase('value <= 0 ? 0 - value : value', 'Math.abs(value)', [1, 1, 1, 31]),
    absCase('value < 0 ? value * -1 : value', 'Math.abs(value)', [1, 1, 1, 31]),
    absCase('value < 0 ? -1 * value : value', 'Math.abs(value)', [1, 1, 1, 31]),
    absCase(
      'object.value < 0 ? -object.value : object.value',
      null,
      [1, 1, 1, 48],
    ),
    absCase('0 > value ? -value : value', 'Math.abs(value)', [1, 1, 1, 27]),
    absCase('0 >= value ? -value : value', 'Math.abs(value)', [1, 1, 1, 28]),
    absCase('0 < value ? value : -value', 'Math.abs(value)', [1, 1, 1, 27]),
    absCase('0 <= value ? value : -value', 'Math.abs(value)', [1, 1, 1, 28]),
    absCase(
      'function foo() {return value < 0 ? -value : value;}',
      'function foo() {return Math.abs(value);}',
      [1, 24, 1, 50],
    ),
    absCase(
      'function foo() {return(value < 0 ? -value : value);}',
      'function foo() {return (Math.abs(value));}',
      [1, 24, 1, 50],
    ),
    absCase(
      'const absolute = value > limit || value < -limit;',
      'const absolute = Math.abs(value) > limit;',
      [1, 18, 1, 49],
    ),
    absCase(
      'const absolute = value >= limit || value <= -limit;',
      'const absolute = Math.abs(value) >= limit;',
      [1, 18, 1, 51],
    ),
    absCase(
      'const absolute = value < -limit || value > limit;',
      'const absolute = Math.abs(value) > limit;',
      [1, 18, 1, 49],
    ),
    absCase(
      'const absolute = value <= -limit || value >= limit;',
      'const absolute = Math.abs(value) >= limit;',
      [1, 18, 1, 51],
    ),
    absCase(
      'const absolute = limit < value || -limit > value;',
      'const absolute = Math.abs(value) > limit;',
      [1, 18, 1, 49],
    ),
    absCase(
      'const absolute = limit <= value || -limit >= value;',
      'const absolute = Math.abs(value) >= limit;',
      [1, 18, 1, 51],
    ),
    absCase(
      'function foo() {return value > limit || value < -limit;}',
      'function foo() {return Math.abs(value) > limit;}',
      [1, 24, 1, 55],
    ),
    absCase(
      'function foo() {return(value > limit || value < -limit);}',
      'function foo() {return (Math.abs(value) > limit);}',
      [1, 24, 1, 55],
    ),
    absCase(
      'const absolute = object.value > limit || object.value < -limit;',
      null,
      [1, 18, 1, 63],
    ),
    absCase(
      'const absolute = value > object.limit || value < -object.limit;',
      null,
      [1, 18, 1, 63],
    ),
    absCase('value < 0 ? -/* comment */value : value', null, [1, 1, 1, 40]),
    absCase(
      'value > limit || value < -/* comment */limit',
      null,
      [1, 1, 1, 45],
    ),
    absCase(
      'if (number > POSITIVE_CONSTANT || number < -POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}',
      'if (Math.abs(number) > POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}',
      [1, 5, 1, 62],
    ),
    absCase(
      'if (number >= POSITIVE_CONSTANT || number <= -POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}',
      'if (Math.abs(number) >= POSITIVE_CONSTANT) {\n\tconsole.log(number);\n}',
      [1, 5, 1, 64],
    ),
    absCase(
      'import {value} from "./value.js"; value < 0 ? -value : value;',
      'import {value} from "./value.js"; Math.abs(value);',
      [1, 35, 1, 61],
    ),

    // TypeScript snapshot cases.
    absCase(
      'function foo(value: number) {\n\treturn value < 0 ? -value : value;\n}',
      'function foo(value: number) {\n\treturn Math.abs(value);\n}',
      [2, 9, 2, 35],
      'file.ts',
    ),
    absCase(
      'function foo(value) {\n\treturn (value satisfies number) < 0 ? -(value satisfies number) : (value satisfies number);\n}',
      'function foo(value) {\n\treturn Math.abs((value satisfies number));\n}',
      [2, 9, 2, 92],
      'file.ts',
    ),

    // Suggestion-only ternaries.
    absCase(
      'const a = b < 0 ? -b : b;',
      'const a = Math.abs(b);',
      [1, 11, 1, 25],
    ),
    absCase(
      'const a = b <= 0 ? -b : b;',
      'const a = Math.abs(b);',
      [1, 11, 1, 26],
    ),
    absCase(
      'const a = b > 0 ? b : -b;',
      'const a = Math.abs(b);',
      [1, 11, 1, 25],
    ),
    absCase(
      'const a = b >= 0 ? b : -b;',
      'const a = Math.abs(b);',
      [1, 11, 1, 26],
    ),

    // Upstream documentation examples.
    absCase(
      'const absolute = value < 0 ? -value : value;',
      'const absolute = Math.abs(value);',
      [1, 18, 1, 44],
    ),
    absCase(
      'const absolute = value <= 0 ? 0 - value : value;',
      'const absolute = Math.abs(value);',
      [1, 18, 1, 48],
    ),
    absCase(
      'const absolute = value > 0 ? value : -value;',
      'const absolute = Math.abs(value);',
      [1, 18, 1, 44],
    ),
    absCase(
      'if (number > MAXIMUM || number < -MAXIMUM) {\n}',
      'if (Math.abs(number) > MAXIMUM) {\n}',
      [1, 5, 1, 42],
    ),
  ],
});
