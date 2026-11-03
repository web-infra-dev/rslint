// Upstream: eslint-plugin-unicorn v79.0.0, including JSDoc, TypeScript and docs.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/no-invalid-character-comparison.js
// Parser-only repeats are deduplicated. Go asserts complete ranges and absent edits;
// this wrapper checks diagnostic counts, message IDs and exact messages.
import { RuleTester } from '../rule-tester';

const tester = new RuleTester();
tester.run('no-invalid-character-comparison', {} as never, {
  valid: [
    // JSDoc / default parser.
    {
      code: '/** @param {?string} text */ function run(text) { return text.charAt(0) === "ab"; }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {Array<string>} values */ function run(values) { return values.at(0) === "ab"; }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {string | Array<string>} value */ function run(value) { return value[0] === "ab"; }',
      filename: 'file.ts',
    },
    {
      code: '/** @param {unknown} text */ function run(text) { return text.charAt(0) === "ab"; }',
      filename: 'file.ts',
    },
    {
      code: '/** @returns {string} */ async function read() { return source; } read().at(0) === "ab";',
      filename: 'file.ts',
    },
    {
      code: '/** @param {string} text */ function run(text) { return text?.[0] === "ab"; }',
      filename: 'file.ts',
    },
    // TypeScript annotation precedence.
    {
      code: '/** @param {string} text */ function run(text: {charAt(): string}) { return text.charAt() === "ab"; }',
      filename: 'file.ts',
    },
    // Snapshot: Single-character comparison value.
    { code: 'str.charAt(4) === "a"', filename: 'case.js' },
    { code: '"a" === str.charAt(4)', filename: 'case.js' },
    { code: 'str.charAt(4) === "\\n"', filename: 'case.js' },
    { code: 'str.charAt(4) !== "\\t"', filename: 'case.js' },
    { code: '"abc".at(0) === "a"', filename: 'case.js' },
    { code: '"abc"[0] === "a"', filename: 'case.js' },
    // Snapshot: Empty string (legitimate out-of-bounds check).
    { code: 'str.charAt(100) === ""', filename: 'case.js' },
    { code: 'str.at(0) === ""', filename: 'case.js' },
    { code: '"abc"[100] === ""', filename: 'case.js' },
    // Snapshot: Not a method call.
    { code: 'charAt(4) === "ab"', filename: 'case.js' },
    { code: 'new charAt(4) === "ab"', filename: 'case.js' },
    // Snapshot: Wrong argument count.
    { code: 'str.charAt(1, 2) === "ab"', filename: 'case.js' },
    // Snapshot: Character access on a receiver that is not provably a string.
    { code: 'str.charAt(0) === "ab"', filename: 'case.js' },
    { code: 'str.charAt() === "ab"', filename: 'case.js' },
    { code: 'tokenizer.charAt(0) === "ab"', filename: 'case.js' },
    { code: 'array.at(0) === "ab"', filename: 'case.js' },
    { code: 'array[0] === "ab"', filename: 'case.js' },
    { code: 'foo.at(0) === "ab"', filename: 'case.js' },
    { code: 'foo[0] === "ab"', filename: 'case.js' },
    // Snapshot: Non-computed access or non-index key.
    { code: '"abc".length === "ab"', filename: 'case.js' },
    { code: '"abc"["length"] === "ab"', filename: 'case.js' },
    { code: 'String(foo)[index] === "ab"', filename: 'case.js' },
    { code: 'String(foo)[property] != "12"', filename: 'case.js' },
    { code: 'String(foo)[i + 1] === "ab"', filename: 'case.js' },
    { code: '"abc"[``] === "ab"', filename: 'case.js' },
    { code: '"abc"[`length`] === "ab"', filename: 'case.js' },
    { code: '"abc"[`00`] === "ab"', filename: 'case.js' },
    { code: '"abc"[`0${index}`] === "ab"', filename: 'case.js' },
    { code: '"abc"[tag`0`] === "ab"', filename: 'case.js' },
    // Snapshot: Dynamic comparison value.
    { code: 'str.charAt(0) === someVariable', filename: 'case.js' },
    // Snapshot: Non-equality operator.
    { code: 'str.charAt(0) < "ab"', filename: 'case.js' },
    { code: 'str.charAt(0) + "ab"', filename: 'case.js' },
    // Snapshot: Optional chaining is intentionally not handled.
    { code: 'str?.charAt(0) === "ab"', filename: 'case.js' },
    { code: '"abc"?.at(0) === "ab"', filename: 'case.js' },
    // Documentation examples.
    { code: "'unicorn'.charAt(0) === 'u';", filename: 'case.js' },
    { code: "'unicorn'.charAt(3) === '\\n';", filename: 'case.js' },
    { code: "'unicorn'.at(0) === 'a';", filename: 'case.js' },
    { code: "'unicorn'[0] === 'a';", filename: 'case.js' },
    { code: "'unicorn'.charAt(100) === '';", filename: 'case.js' },
  ],
  invalid: [
    // JSDoc / default parser.
    {
      code: '/** @param {string} text */ function run(text) { return text.charAt(0) === "😀"; }',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `😀`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '/** @returns {string} */ function read() { return source; } read().at(-1) !== "ab";',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    {
      code: '/** @param {() => string} read */ function run(read) { return read()["0"] == "ab"; }',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '/** @type {string} */ const text = source; const alias = text; alias.charAt() === "ab";',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '/** @param {string} text */ function run(text, cache) { return (cache = text)[0] === "ab"; }',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '/** @param {"left" | "right"} text */ function run(text) { return "ab" !== text.at(1); }',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    // TypeScript annotation precedence.
    {
      code: '/** @param {Array<string>} text */ function run(text: string) { return text.charAt(0) === "ab"; }',
      filename: 'file.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: All equality operators.
    {
      code: '"abc".charAt(4) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".charAt(4) !== "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    {
      code: '"abc".charAt(0) == "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".charAt(0) != "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    // Snapshot: Reversed operand order.
    {
      code: '"ab" === "abc".charAt(4)',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: Parenthesized operands.
    {
      code: '("abc".charAt(0)) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".charAt(0) === ("ab")',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: Typo motivation (`/n` instead of `\n`).
    {
      code: '"abc".charAt(3) === "/n"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `/n`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".charAt(822) !== "foo"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `foo`, so this comparison is always true.',
        },
      ],
    },
    // Snapshot: Omitted positions default to `0`.
    {
      code: '"abc".charAt() === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".at() !== "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    // Snapshot: Emoji is two UTF-16 code units (OXC misses this).
    {
      code: '"abc".charAt(0) === "😀"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `😀`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".charAt(0) === "\\u{1F600}"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `😀`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".charAt(0) === "\\r\\n"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `\r\n`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: Template literal comparison value.
    {
      code: '"abc".charAt(0) === `ab`',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: `charAt` on a provable string receiver.
    {
      code: 'String(foo).charAt(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: 'String.fromCharCode(65).charAt(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '`${foo}`.charAt(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '(foo + "bar").charAt(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: `at` on a provable string receiver.
    {
      code: '"abc".at(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc".at(0) !== "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    {
      code: '"abc".at(-1) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: 'String(foo).at(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: 'String.fromCharCode(65).at(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '`${foo}`.at(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '(foo + "bar").at(0) === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: `[index]` on a provable string receiver.
    {
      code: '"abc"[0] === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc"[0] != "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always true.',
        },
      ],
    },
    {
      code: '"abc"["0"] === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: TypeScript: receiver proven via type annotation.
    {
      code: 'declare const s: string;\nconst x = s.at(0) === "ab";',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: 'declare const s: string;\nconst x = s.charAt(0) === "ab";',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: 'function isAb(s: string) {\n\treturn s["0"] === "ab";\n}',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: TypeScript: receiver proven via `as` assertion.
    {
      code: '(foo as string)[0] === "ab";',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: A single UTF-16 code unit can never equal an astral character, which is two code units.
    {
      code: "'😀'[0] === '😀'",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `😀`, so this comparison is always false.',
        },
      ],
    },
    {
      code: "'😀'.charAt(0) === '😀'",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `😀`, so this comparison is always false.',
        },
      ],
    },
    {
      code: "'a😀'.at(0) === 'a😀'",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `a😀`, so this comparison is always false.',
        },
      ],
    },
    // Snapshot: `String#at()` returns a UTF-16 code unit too, not a whole code point.
    {
      code: "'😀'.at(0) === '😀'",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `😀`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc"[`0`] === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc"[(`1`)] === `ab`',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: '"abc"[`\\u{30}`] === "ab"',
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: 'function isAb(value: string) { return value[`0`] === `ab`; }',
      filename: 'case.ts',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    // Documentation examples.
    {
      code: "'unicorn'.charAt(0) === 'ab';",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: "'unicorn'.charAt(3) === '/n'; // Typo: meant '\\n'",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `/n`, so this comparison is always false.',
        },
      ],
    },
    {
      code: "'unicorn'.at(0) === 'ab';",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
    {
      code: "'unicorn'[0] === 'ab';",
      filename: 'case.js',
      errors: [
        {
          messageId: 'no-invalid-character-comparison',
          message:
            'A single character can never equal the multi-character string `ab`, so this comparison is always false.',
        },
      ],
    },
  ],
});
