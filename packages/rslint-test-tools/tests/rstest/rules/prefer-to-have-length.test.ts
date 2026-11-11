import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { RuleTester } from '@rslint/rule-tester';
import { config, withFixture } from '../config';

const ruleTester = new RuleTester(config);
const message = 'Use `toHaveLength()` instead';

describe('prefer-to-have-length integration contract', () => {
  const ruleName = 'rstest/prefer-to-have-length';
  const run = (code: string, fix = false) =>
    withFixture((root) =>
      lint({
        config: [
          {
            plugins: ['rstest'],
            languageOptions: config.languageOptions,
            rules: { [ruleName]: 'error' },
          },
        ],
        configDirectory: root,
        workingDirectory: root,
        fileContents: { [path.join(root, config.filename!)]: code },
        fix,
      }),
    );

  test('reports exact metadata for every equality matcher in a chain', async () => {
    const result = await run(
      'expect(files.length).toBe(1).and.toStrictEqual(1);',
    );

    expect(
      result.diagnostics.map(({ ruleName, messageId, message, range }) => ({
        ruleName,
        messageId,
        message,
        range,
      })),
    ).toEqual([
      {
        ruleName,
        messageId: 'useToHaveLength',
        message,
        range: {
          start: { line: 1, column: 22 },
          end: { line: 1, column: 26 },
        },
      },
      {
        ruleName,
        messageId: 'useToHaveLength',
        message,
        range: {
          start: { line: 1, column: 34 },
          end: { line: 1, column: 47 },
        },
      },
    ]);
  });

  test('applies only safe fixes and converges', async () => {
    const safeCode = 'expect([1, 2].length).toBe(2);';
    const safe = await run(safeCode, true);
    expect(Object.values(safe.output ?? {})).toEqual([
      'expect([1, 2]).toHaveLength(2);',
    ]);
    expect(safe.diagnostics).toEqual([]);

    for (const unsafeCode of [
      'expect([1, 2].length).toEqual(2);',
      'expect([].length).toBe(-0);',
      'expect(files.length).toBe(1).and.toEqual(1);',
    ]) {
      const unsafe = await run(unsafeCode, true);
      expect(unsafe.output).toBeUndefined();
      expect(unsafe.diagnostics).toHaveLength(
        unsafeCode.includes('.and.') ? 2 : 1,
      );
    }
  });
});

ruleTester.run('prefer-to-have-length', {
  valid: [
    { code: 'expect.hasAssertions()' },
    { code: 'expect(files).toHaveLength(1);' },
    { code: 'expect(users[0]?.permissions?.length).toBe(1);' },
    { code: 'await expect.poll(() => files.length).toBe(1);' },
    { code: 'expect.element(locator).toEqual(1);' },
  ],
  invalid: [
    {
      code: 'expect(files.length).toBe(1);',
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'expect(["file"].length).toEqual(1);',
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 25,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      code: 'expect.soft("file"["length"]).not["toStrictEqual"](0);',
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 35,
          endLine: 1,
          endColumn: 50,
        },
      ],
    },
    {
      code: 'expect([].length).toBe(-0);',
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      code: 'expect(files.length).toBe(1).and.toEqual(1);',
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 22,
          endLine: 1,
          endColumn: 26,
        },
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: "expect(files.length).to.be.a('number').and.toBe(1);",
      errors: [
        {
          messageId: 'useToHaveLength',
          message: 'Use `toHaveLength()` instead',
          line: 1,
          column: 44,
          endLine: 1,
          endColumn: 48,
        },
      ],
    },
  ],
});
