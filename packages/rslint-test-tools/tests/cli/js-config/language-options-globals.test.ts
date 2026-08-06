import { describe, test, expect } from '@rstest/core';
import {
  runRslint,
  createTempDir,
  cleanupTempDir,
  TS_CONFIG,
} from './helpers.js';

interface Diagnostic {
  ruleName: string;
  filePath: string;
  message: string;
}

/**
 * Lint a single file whose config declares `globals`, and return the rules
 * that fired. `globals` is authored exactly as in an ESLint flat config, so
 * this also covers the boolean shape the `globals` npm package exports.
 */
async function lintWithGlobals(
  code: string,
  globals: Record<string, unknown>,
): Promise<string[]> {
  const config = `export default [
    {
      files: ['**/*.ts'],
      languageOptions: {
        globals: ${JSON.stringify(globals)},
        parserOptions: { projectService: false, project: ['./tsconfig.json'] },
      },
      rules: { 'no-global-assign': 'error' },
    },
  ];`;
  const tempDir = await createTempDir({
    'rslint.config.mjs': config,
    'tsconfig.json': TS_CONFIG,
    'index.ts': code,
  });
  try {
    const result = await runRslint(['--format', 'jsonline'], tempDir);
    return result.stdout
      .trim()
      .split('\n')
      .filter((line) => line.trim())
      .map((line) => (JSON.parse(line) as Diagnostic).ruleName);
  } finally {
    await cleanupTempDir(tempDir);
  }
}

describe('languageOptions.globals access levels', () => {
  test('writable lifts the built-in readonly setting', async () => {
    expect(
      await lintWithGlobals('Object = 1;', { Object: 'writable' }),
    ).toEqual([]);
  });

  test('readonly keeps a built-in reported', async () => {
    expect(
      await lintWithGlobals('Object = 1;', { Object: 'readonly' }),
    ).toEqual(['no-global-assign']);
  });

  test('off removes the built-in entirely', async () => {
    expect(await lintWithGlobals('Object = 1;', { Object: 'off' })).toEqual([]);
  });

  test('a readonly project global is reported like a built-in', async () => {
    expect(
      await lintWithGlobals('myGlobal = 1;', { myGlobal: 'readonly' }),
    ).toEqual(['no-global-assign']);
  });

  test('a writable project global is assignable', async () => {
    expect(
      await lintWithGlobals('myGlobal = 1;', { myGlobal: 'writable' }),
    ).toEqual([]);
  });

  // The `globals` package spells the two declared levels as booleans.
  test('boolean true is writable', async () => {
    expect(await lintWithGlobals('Object = 1;', { Object: true })).toEqual([]);
  });

  test('boolean false is readonly, not off', async () => {
    expect(await lintWithGlobals('myGlobal = 1;', { myGlobal: false })).toEqual(
      ['no-global-assign'],
    );
  });

  test('an inline comment overrides the config setting', async () => {
    expect(
      await lintWithGlobals('/* global Object: writable */\nObject = 1;', {
        Object: 'readonly',
      }),
    ).toEqual([]);
  });
});
