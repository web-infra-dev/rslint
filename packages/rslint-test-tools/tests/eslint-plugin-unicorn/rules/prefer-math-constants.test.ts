// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation. See LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-math-constants.js
import { RuleTester, type TestCase } from '@rslint/rule-tester';

function approximation(
  code: string,
  literal: string,
  constant: string,
  filename = 'file.js',
): TestCase {
  const start = code.indexOf(literal);
  const end = start + literal.length;
  return {
    code,
    filename,
    errors: [
      {
        messageId: 'prefer-math-constants/error',
        message: `Prefer \`${constant}\` over the approximate value \`${literal}\`.`,
        line: 1,
        column: start + 1,
        endLine: 1,
        endColumn: end + 1,
        suggestions: [
          {
            messageId: 'prefer-math-constants/suggestion',
            desc: `Replace \`${literal}\` with \`${constant}\`.`,
            output: code.slice(0, start) + constant + code.slice(end),
          },
        ],
      },
    ],
  };
}

const tester = new RuleTester({ plugin: 'unicorn' });
tester.run('prefer-math-constants', {
  valid: [
    // Already using the constant.
    'Math.PI',
    'const tau = Math.PI * 2;',
    'Math.E',
    // Too few fraction digits.
    '3.1',
    '2.71',
    '1.41',
    '0.69',
    '2.30',
    '1.44',
    '0.43',
    '0.70',
    // Different integer part, unrelated digits, extra precision or different rounding.
    '4.14',
    '5.718',
    '1.23456',
    '0.12345',
    '3.1415926535897932',
    '1.4143',
    // Literal types, including negative ones.
    { code: 'type Tau = 3.14;', filename: 'file.ts' },
    { code: 'type Tau = -3.14;', filename: 'file.ts' },
    { code: 'let value: -3.14;', filename: 'file.ts' },
    { code: 'const value = foo as -3.14;', filename: 'file.ts' },
    { code: 'function f(value: -3.14) {}', filename: 'file.ts' },
    { code: 'const x = y as 3.14;', filename: 'file.ts' },
    // Other numeric spellings.
    '3',
    '0x1F',
    '0b101',
    '3n',
    '3.14e2',
    '1_000',
    // Non-computed property keys.
    'const object = {3.14: 1};',
    'class Foo {3.14 = 1}',
    'class Foo {3.14() {}}',
    { code: 'interface Foo {3.14: number}', filename: 'file.ts' },
    // Documentation examples.
    'const area = Math.PI * radius ** 2;',
    'const growth = amount * Math.E;',
    'const diagonal = side * Math.SQRT2;',
  ],
  invalid: [
    // PI.
    approximation('3.14', '3.14', 'Math.PI'),
    approximation('const area = 3.14 * radius ** 2;', '3.14', 'Math.PI'),
    approximation('3.141', '3.141', 'Math.PI'),
    approximation('3.1416', '3.1416', 'Math.PI'),
    approximation('3.141592653589793', '3.141592653589793', 'Math.PI'),
    // E, including rounding.
    approximation('2.718', '2.718', 'Math.E'),
    approximation('const growth = amount * 2.718;', '2.718', 'Math.E'),
    approximation('2.718281828459045', '2.718281828459045', 'Math.E'),
    approximation('2.7183', '2.7183', 'Math.E'),
    // LN2, LN10, LOG2E, LOG10E, SQRT2 and SQRT1_2.
    approximation('0.6931', '0.6931', 'Math.LN2'),
    approximation('0.6931471805599453', '0.6931471805599453', 'Math.LN2'),
    approximation('2.3025', '2.3025', 'Math.LN10'),
    approximation('2.3026', '2.3026', 'Math.LN10'),
    approximation('1.4426', '1.4426', 'Math.LOG2E'),
    approximation('1.4427', '1.4427', 'Math.LOG2E'),
    approximation('0.4342', '0.4342', 'Math.LOG10E'),
    approximation('1.4142', '1.4142', 'Math.SQRT2'),
    approximation('0.7071', '0.7071', 'Math.SQRT1_2'),
    // Negative, leading-dot, separators, and nested expressions.
    approximation('-3.14', '3.14', 'Math.PI'),
    approximation('.6931', '.6931', 'Math.LN2'),
    approximation('3.141_592', '3.141_592', 'Math.PI'),
    approximation('1 / 3.14', '3.14', 'Math.PI'),
    approximation('foo(3.14)', '3.14', 'Math.PI'),
    approximation('[2.718]', '2.718', 'Math.E'),
    // Computed keys, JSX and comments.
    approximation('const object = {[3.14]: 1};', '3.14', 'Math.PI'),
    approximation('class Foo {[3.14] = 1}', '3.14', 'Math.PI'),
    approximation('object[3.14]', '3.14', 'Math.PI'),
    approximation('<div>{3.14}</div>', '3.14', 'Math.PI', 'file.jsx'),
    approximation('foo(/* before */ 3.14 /* after */)', '3.14', 'Math.PI'),
    // TypeScript runtime expressions.
    approximation('enum Foo {a = 3.14}', '3.14', 'Math.PI', 'file.ts'),
    approximation(
      'const x = 3.14 satisfies number;',
      '3.14',
      'Math.PI',
      'file.ts',
    ),
    approximation('const x = 3.14 as number;', '3.14', 'Math.PI', 'file.ts'),
    approximation('const x = <number>3.14;', '3.14', 'Math.PI', 'file.ts'),
    approximation(
      'const value = -3.14 as number;',
      '3.14',
      'Math.PI',
      'file.ts',
    ),
    // Documentation example.
    approximation('const diagonal = side * 1.4142;', '1.4142', 'Math.SQRT2'),
  ],
});
