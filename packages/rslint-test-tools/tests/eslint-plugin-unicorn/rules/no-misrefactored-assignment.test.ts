// Ported from eslint-plugin-unicorn v79.0.0 tests and documentation; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-misrefactored-assignment.js
import path from 'node:path';
import { runInNewContext } from 'node:vm';
import { lint } from '@rslint/core/internal';
import {
  RuleTester,
  type ExpectedError,
  type TestCase,
} from '@rslint/rule-tester';
import { expect, test } from 'rstack/test';

function assignmentError(
  target: string,
  [line, column, endLine, endColumn]: [number, number, number, number],
  output?: string,
): ExpectedError {
  return {
    messageId: 'no-misrefactored-assignment/error',
    message: `\`${target}\` is duplicated in the right-hand side, which is likely a refactoring mistake.`,
    line,
    column,
    endLine,
    endColumn,
    suggestions:
      output === undefined
        ? []
        : [
            {
              messageId: 'no-misrefactored-assignment/suggestion',
              desc: `Remove the duplicated \`${target}\`.`,
              output,
            },
          ],
  };
}

function wholeAssignment(
  code: string,
  target: string,
  endLine: number,
  endColumn: number,
  output?: string,
  filename?: string,
): TestCase {
  return {
    code,
    filename,
    errors: [assignmentError(target, [1, 1, endLine, endColumn], output)],
  };
}

new RuleTester({ plugin: 'unicorn' }).run('no-misrefactored-assignment', {
  valid: [
    // No duplication
    'a += b;',
    'a += b + c;',
    'a -= b - c;',
    'a *= b * c;',

    // Operator mismatch
    'a += a - b;',
    'a -= a + b;',
    'a *= a + b;',
    'a /= a * b;',

    // Not a compound assignment we handle
    'a = a + b;',
    'a += a;',

    // Logical assignment operators are intentionally excluded
    'a &&= a && b;',
    'a ||= a || b;',
    'a ??= a ?? b;',

    // Non-commutative operators with the target as the right operand are allowed
    'a -= b - a;',
    'a /= b / a;',
    'a %= b % a;',
    'a **= b ** a;',
    'a <<= b << a;',
    'a >>= b >> a;',
    'a >>>= b >>> a;',

    // Different references
    'a.x += a.y + b;',
    'a[0] += a[1] + b;',
    'foo += bar + baz;',

    // Same property name but a different object
    'a.x += b.x + c;',

    // Computed key with a side effect cannot be proven equal, so it is not matched
    'a[foo()] += a[foo()] + b;',

    // Operator mismatch among bitwise operators
    'a |= a & b;',
    'a <<= a >> b;',

    // The target is not an immediate operand of the right-hand side
    'a += a + b + c;',
    'a += b + a + c;',

    // Documentation examples.
    'a += b;',
    'a *= b;',
    'a += b + c;',
    'a += a - b;',
    'a -= b - a;',
  ],
  invalid: [
    // Comments in the kept operand.
    wholeAssignment(
      'a += a + getValue(/* keep */);',
      'a',
      1,
      30,
      'a += getValue(/* keep */);',
    ),
    wholeAssignment(
      'a *= getValue(/* keep */) * a;',
      'a',
      1,
      30,
      'a *= getValue(/* keep */);',
    ),
    wholeAssignment(
      'a += a + (condition ? /* keep */ first : second);',
      'a',
      1,
      49,
      'a += (condition ? /* keep */ first : second);',
    ),
    wholeAssignment(
      'a += a + values[/* keep */ index];',
      'a',
      1,
      34,
      'a += values[/* keep */ index];',
    ),

    // Comment boundaries around kept and discarded operands.
    wholeAssignment(
      'a += a + (/* keep */ condition ? first : second);',
      'a',
      1,
      49,
      'a += (/* keep */ condition ? first : second);',
    ),
    wholeAssignment(
      'a += a + getValue(\r\n  // keep\r\n);',
      'a',
      3,
      2,
      'a += getValue(\r\n  // keep\r\n);',
    ),
    wholeAssignment(
      'a += a + (/* keep */ first(), second);',
      'a',
      1,
      38,
      'a += (/* keep */ first(), second);',
    ),
    wholeAssignment(
      'a *= a * (left /* keep */ + right);',
      'a',
      1,
      35,
      'a *= left /* keep */ + right;',
    ),
    wholeAssignment(
      'a += a + (value as /* keep */ number);',
      'a',
      1,
      38,
      'a += (value as /* keep */ number);',
      'input.ts',
    ),
    wholeAssignment('a += (/* keep */ a) + value;', 'a', 1, 28),
    wholeAssignment('a += a /* keep */ + value;', 'a', 1, 26),
    wholeAssignment('a += a + (/* keep */ value);', 'a', 1, 28),
    wholeAssignment('a += a + (value /* keep */);', 'a', 1, 28),
    wholeAssignment('a *= a * (/* keep */ left + right);', 'a', 1, 35),

    // Snapshot suite: duplicated left operand for every operator.
    wholeAssignment('a += a + b;', 'a', 1, 11, 'a += b;'),
    wholeAssignment('a -= a - b;', 'a', 1, 11, 'a -= b;'),
    wholeAssignment('a *= a * b;', 'a', 1, 11, 'a *= b;'),
    wholeAssignment('a /= a / b;', 'a', 1, 11, 'a /= b;'),
    wholeAssignment('a %= a % b;', 'a', 1, 11, 'a %= b;'),
    wholeAssignment('a **= a ** b;', 'a', 1, 13, 'a **= b;'),
    wholeAssignment('a &= a & b;', 'a', 1, 11, 'a &= b;'),
    wholeAssignment('a |= a | b;', 'a', 1, 11, 'a |= b;'),
    wholeAssignment('a ^= a ^ b;', 'a', 1, 11, 'a ^= b;'),
    wholeAssignment('a <<= a << b;', 'a', 1, 13, 'a <<= b;'),
    wholeAssignment('a >>= a >> b;', 'a', 1, 13, 'a >>= b;'),
    wholeAssignment('a >>>= a >>> b;', 'a', 1, 15, 'a >>>= b;'),

    // Duplicated right operand for commutative operators.
    wholeAssignment('a += b + a;', 'a', 1, 11, 'a += b;'),
    wholeAssignment('a *= b * a;', 'a', 1, 11, 'a *= b;'),
    wholeAssignment('a &= b & a;', 'a', 1, 11, 'a &= b;'),
    wholeAssignment('a |= b | a;', 'a', 1, 11, 'a |= b;'),
    wholeAssignment('a ^= b ^ a;', 'a', 1, 11, 'a ^= b;'),

    // Member, this, computed and private targets.
    wholeAssignment(
      'foo.bar += foo.bar + 1;',
      'foo.bar',
      1,
      23,
      'foo.bar += 1;',
    ),
    wholeAssignment('this.x += this.x + y;', 'this.x', 1, 21, 'this.x += y;'),
    wholeAssignment('a.b[c] += a.b[c] + d;', 'a.b[c]', 1, 21, 'a.b[c] += d;'),
    wholeAssignment('a.x += a["x"] + b;', 'a.x', 1, 18, 'a.x += b;'),
    {
      code: 'class A { #x = 0; m() { this.#x += this.#x + b; } }',
      errors: [
        assignmentError(
          'this.#x',
          [1, 25, 1, 47],
          'class A { #x = 0; m() { this.#x += b; } }',
        ),
      ],
    },

    // Parenthesized target operand, string concatenation and two matches.
    wholeAssignment('a += (a) + b;', 'a', 1, 13, 'a += b;'),
    wholeAssignment('s += s + "!";', 's', 1, 13, 's += "!";'),
    wholeAssignment('a += a + a;', 'a', 1, 11, 'a += a;'),
    {
      code: 'a += a + (b += b + c);',
      errors: [
        assignmentError('a', [1, 1, 1, 22], 'a += (b += b + c);'),
        assignmentError('b', [1, 11, 1, 21], 'a += a + (b += c);'),
      ],
    },

    // Comments and grouping on the right-hand side.
    wholeAssignment('a += a + /* keep */ b;', 'a', 1, 22),
    wholeAssignment('a += /* lead */ a + b;', 'a', 1, 22, 'a += /* lead */ b;'),
    wholeAssignment('a += a + (b + c);', 'a', 1, 17, 'a += b + c;'),
    wholeAssignment('a += (a + b);', 'a', 1, 13, 'a += (b);'),

    // TypeScript wrappers on the duplicated operand.
    wholeAssignment(
      'a += (a as number) + b;',
      'a',
      1,
      23,
      'a += b;',
      'input.ts',
    ),
    wholeAssignment('a += a! + b;', 'a', 1, 12, 'a += b;', 'input.ts'),
    wholeAssignment('a += (<number>a) + b;', 'a', 1, 21, 'a += b;', 'input.ts'),

    // Grouping of the kept operand.
    wholeAssignment('a += a + (b, c);', 'a', 1, 16, 'a += (b, c);'),
    wholeAssignment('a += a + (b ? c : d);', 'a', 1, 21, 'a += (b ? c : d);'),
    wholeAssignment('a += (b ? c : d) + a;', 'a', 1, 21, 'a += (b ? c : d);'),
    wholeAssignment('a *= a * (b + c);', 'a', 1, 17, 'a *= b + c;'),

    // Documentation examples.
    wholeAssignment('a += a + b;', 'a', 1, 11, 'a += b;'),
    wholeAssignment('a *= b * a;', 'a', 1, 11, 'a *= b;'),
  ],
});

test('keeps the call and its comment when the suggestion is applied', async () => {
  const code =
    'let value = 2; let evaluations = 0; function getValue() { evaluations++; return 3; } value += value + getValue(/* keep */); JSON.stringify({value, evaluations});';
  const root = import.meta.dirname;
  const ruleName = 'unicorn/no-misrefactored-assignment';
  const result = await lint({
    workingDirectory: root,
    configDirectory: root,
    config: [
      {
        plugins: ['unicorn'],
        languageOptions: { parserOptions: { projectService: false } },
        rules: { [ruleName]: 'error' },
      },
    ],
    fileContents: { [path.join(root, 'input.js')]: code },
    fix: true,
  });
  expect(result.fileCount).toBe(1);
  expect(result.ruleCount).toBe(1);
  expect(result.output).toBeUndefined();
  expect(result.fixableErrorCount).toBe(0);
  expect(result.diagnostics).toHaveLength(1);
  const [diagnostic] = result.diagnostics;
  expect(diagnostic.ruleName).toBe(ruleName);
  expect(diagnostic.messageId).toBe('no-misrefactored-assignment/error');
  expect(diagnostic.fixes).toBeUndefined();
  const suggestions = diagnostic.suggestions ?? [];
  expect(suggestions).toHaveLength(1);
  const [suggestion] = suggestions;
  expect(suggestion.messageId).toBe('no-misrefactored-assignment/suggestion');
  const edits = suggestion.fixes ?? [];
  expect(edits).toHaveLength(1);
  const [{ startPos, endPos, text }] = edits;
  const suggested = code.slice(0, startPos) + text + code.slice(endPos);
  expect(suggested).toContain('/* keep */');
  expect(runInNewContext(code)).toBe('{"value":7,"evaluations":1}');
  expect(runInNewContext(suggested)).toBe('{"value":5,"evaluations":1}');
});
