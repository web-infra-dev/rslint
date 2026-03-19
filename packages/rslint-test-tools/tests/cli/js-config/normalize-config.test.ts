import { describe, test, expect } from '@rstest/core';
import { normalizeConfig } from '@rslint/core/config-loader';

describe('normalizeConfig', () => {
  test('should accept a valid flat config array', () => {
    const result = normalizeConfig([
      { files: ['**/*.ts'], rules: { 'no-console': 'error' } },
    ]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toEqual(['**/*.ts']);
    expect(result[0].rules).toEqual({ 'no-console': 'error' });
  });

  test('should throw when config is not an array', () => {
    expect(() => normalizeConfig({ rules: {} })).toThrow(
      'rslint config must export an array',
    );
  });

  test('should strip unknown fields', () => {
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

  test('should preserve all known fields', () => {
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

  test('should handle empty array', () => {
    expect(normalizeConfig([])).toEqual([]);
  });

  test('should skip null and non-object entries', () => {
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

  test('should throw when files is a string instead of array', () => {
    expect(() => normalizeConfig([{ files: '**/*.ts', rules: {} }])).toThrow(
      '"files" must be an array',
    );
  });

  test('should throw when ignores is a string instead of array', () => {
    expect(() => normalizeConfig([{ ignores: 'dist/**', rules: {} }])).toThrow(
      '"ignores" must be an array',
    );
  });

  test('should allow omitted files and ignores', () => {
    const result = normalizeConfig([{ rules: { 'no-console': 'error' } }]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toBeUndefined();
    expect(result[0].ignores).toBeUndefined();
  });
});
