import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

const valid = (code: string) => ({ code, filename: 'file.mjs' });

const invalid = (code: string, output: string) => ({
  code,
  filename: 'file.mjs',
  errors: [
    {
      message:
        'Unexpected bitwise XOR operator `^`. Did you mean the exponentiation operator `**`?',
      suggestions: [
        {
          messageId: 'no-xor-as-exponentiation/suggestion',
          desc: 'Replace `^` with `**`.',
          output,
        },
      ],
    },
  ],
});

ruleTester.run('no-xor-as-exponentiation', null as never, {
  valid: [
    // Already correct exponentiation.
    valid('2 ** 32'),

    // Non-decimal literals are likely intentional bitwise XOR.
    valid('0xFF ^ 8'),
    valid('2 ^ 0x10'),
    valid('0b100 ^ 2'),
    valid('0o20 ^ 2'),
    valid('2 ^ 0o20'),

    // Non-literal operands.
    valid('a ^ b'),
    valid('x ^ 2'),
    valid('2 ^ y'),
    valid('flags ^ MASK'),

    // Floats.
    valid('2.5 ^ 3'),
    valid('2 ^ 3.5'),
    valid('2e3 ^ 2'),

    // BigInt.
    valid('2n ^ 32n'),

    // Unary operand is not a literal.
    valid('2 ^ -3'),
    valid('-2 ^ 3'),
    valid('2 ^ +3'),

    // Other operators.
    valid('2 | 8'),
    valid('2 & 8'),
    valid('2 << 8'),
    valid('2 ** 8'),
  ],
  invalid: [
    // Plain decimal-integer pairs.
    invalid('2 ^ 32', '2 ** 32'),
    invalid('3 ^ 3', '3 ** 3'),
    invalid('10 ^ 6', '10 ** 6'),
    invalid('0 ^ 0', '0 ** 0'),
    invalid('2 ^ 8', '2 ** 8'),

    // Surrounding whitespace preserved by the suggestion.
    invalid('2  ^  8', '2  **  8'),

    // Inside an expression.
    invalid('const x = 2 ^ 8;', 'const x = 2 ** 8;'),
    invalid('foo(2 ^ 8)', 'foo(2 ** 8)'),

    // Numeric separators are still decimal integers.
    invalid('10 ^ 1_000', '10 ** 1_000'),

    // Nested: only the inner literal pair fires.
    invalid('2 ^ 8 ^ 2', '2 ^ 8 ** 2'),

    // Comment between operands must be preserved by the suggestion.
    invalid('2 /* comment */ ^ 8', '2 /* comment */ ** 8'),
  ],
});
