import { describe, test, expect } from '@rstest/core';
import { loadConfigFile, normalizeConfig } from '../src/config-loader.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-config-loader-test-'));
}

function cleanup(dir: string): void {
  fs.rmSync(dir, { recursive: true, force: true });
}

describe('loadConfigFile', () => {
  test('loads a .js config file with default export', async () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.js'),
        'export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];',
      );
      const result = await loadConfigFile(path.join(tmp, 'rslint.config.js'));
      expect(Array.isArray(result)).toBe(true);
      expect((result as Array<{ rules: unknown }>)[0].rules).toEqual({
        'no-console': 'error',
      });
    } finally {
      cleanup(tmp);
    }
  });

  test('loads a .mjs config file', async () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.mjs'),
        'export default [{ files: ["**/*.js"], rules: {} }];',
      );
      const result = await loadConfigFile(path.join(tmp, 'rslint.config.mjs'));
      expect(Array.isArray(result)).toBe(true);
    } finally {
      cleanup(tmp);
    }
  });

  test('resolves a thenable (Promise) default export', async () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(
        path.join(tmp, 'rslint.config.js'),
        'export default Promise.resolve([{ files: ["**/*.ts"], rules: { "no-console": "error" } }]);',
      );
      const result = await loadConfigFile(path.join(tmp, 'rslint.config.js'));
      expect(Array.isArray(result)).toBe(true);
      expect((result as Array<{ rules: unknown }>)[0].rules).toEqual({
        'no-console': 'error',
      });
    } finally {
      cleanup(tmp);
    }
  });

  test('throws for an unsupported extension', async () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.yaml'), 'rules: {}');
      await expect(
        loadConfigFile(path.join(tmp, 'rslint.config.yaml')),
      ).rejects.toThrow('Unsupported config file extension');
    } finally {
      cleanup(tmp);
    }
  });
});

describe('normalizeConfig', () => {
  test('accepts a valid flat config array', () => {
    const result = normalizeConfig([
      { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toEqual(['**/*.ts']);
    expect(result[0].rules).toEqual({ 'no-console': 'error' });
  });

  test('throws when config is not an array', () => {
    expect(() => normalizeConfig({ rules: {} })).toThrow(
      'rslint config must export an array',
    );
  });

  test('strips unknown fields', () => {
    const result = normalizeConfig([
      {
        name: 'my-config',
        files: ['**/*.ts'],
        rules: {},
        unknownField: 123,
      },
    ]);
    expect(result[0]).not.toHaveProperty('name');
    expect(result[0]).not.toHaveProperty('unknownField');
  });

  test('preserves all known fields', () => {
    const result = normalizeConfig([
      {
        files: ['**/*.ts'],
        ignores: ['dist/**'],
        languageOptions: {
          parserOptions: { project: ['./tsconfig.json'] },
        },
        rules: { 'no-console': 'error' },
        plugins: ['@typescript-eslint'],
        settings: { key: 'value' },
      },
    ]);
    const entry = result[0];
    expect(entry.files).toEqual(['**/*.ts']);
    expect(entry.ignores).toEqual(['dist/**']);
    expect(entry.rules).toEqual({ 'no-console': 'error' });
    expect(entry.plugins).toEqual(['@typescript-eslint']);
    expect(entry.settings).toEqual({ key: 'value' });
  });

  test('handles empty array', () => {
    expect(normalizeConfig([])).toEqual([]);
  });

  test('skips null and non-object entries', () => {
    const result = normalizeConfig([
      { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
      null,
      undefined,
      42,
      'string',
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].rules).toEqual({ 'no-console': 'error' });
  });

  test('throws when files is a string instead of array', () => {
    expect(() => normalizeConfig([{ files: '**/*.ts', rules: {} }])).toThrow(
      '"files" must be an array',
    );
  });

  test('throws when ignores is a string instead of array', () => {
    expect(() => normalizeConfig([{ ignores: 'dist/**', rules: {} }])).toThrow(
      '"ignores" must be an array',
    );
  });

  test('allows omitted files and ignores', () => {
    const result = normalizeConfig([{ rules: { 'no-console': 'error' } }]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toBeUndefined();
    expect(result[0].ignores).toBeUndefined();
  });
});
