import path from 'node:path';

import { lint } from '@rslint/core/internal';

import { buildConfigForSettings } from '../../src/util/load-test-config';
import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const message = 'Use `toHaveLength()` instead';
const error = (column: number, endColumn: number) => ({
  messageId: 'useToHaveLength',
  message,
  line: 1,
  column,
  endLine: 1,
  endColumn,
});

describe('prefer-to-have-length integration contract', () => {
  const ruleName = 'rstest/prefer-to-have-length';
  const filename = path.resolve(
    import.meta.dirname,
    '../src/prefer-to-have-length.ts',
  );
  const configFile = path.resolve(import.meta.dirname, '../rslint.config.mjs');

  const run = async (code: string, fix = false) => {
    const { config, configDirectory } = await buildConfigForSettings(
      configFile,
      undefined,
    );
    return lint({
      config: [
        ...config,
        {
          rules: {
            [ruleName]: 'error',
          },
        },
      ],
      configDirectory,
      workingDirectory: process.cwd(),
      fileContents: { [filename]: code },
      fix,
    });
  };

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

ruleTester.run('prefer-to-have-length', {} as never, {
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
      output: null,
      errors: [error(22, 26)],
    },
    {
      code: 'expect(["file"].length).toEqual(1);',
      output: null,
      errors: [error(25, 32)],
    },
    {
      code: 'expect.soft("file"["length"]).not["toStrictEqual"](0);',
      output: null,
      errors: [error(35, 50)],
    },
    {
      code: 'expect([].length).toBe(-0);',
      output: null,
      errors: [error(19, 23)],
    },
    {
      code: 'expect(files.length).toBe(1).and.toEqual(1);',
      output: null,
      errors: [error(22, 26), error(34, 41)],
    },
    {
      code: "expect(files.length).to.be.a('number').and.toBe(1);",
      output: null,
      errors: [error(44, 48)],
    },
  ],
});
