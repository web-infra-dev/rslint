import { describe, test, expect } from '@rstest/core';
import { normalizeConfig } from '@rslint/core/config-loader';
import { defineConfig, ts, js, reactPlugin, importPlugin } from '@rslint/core';

describe('defineConfig and config presets', () => {
  test('defineConfig should be importable and return input as-is', () => {
    const input = [
      { files: ['**/*.ts'], rules: { 'no-console': 'error' as const } },
    ];
    const result = defineConfig(input);
    expect(result).toBe(input);
  });

  test('config presets should be importable', () => {
    expect(ts).toBeDefined();
    expect(ts.configs.recommended).toBeDefined();
    expect(js).toBeDefined();
    expect(js.configs.recommended).toBeDefined();
    expect(reactPlugin).toBeDefined();
    expect(reactPlugin.configs.recommended).toBeDefined();
    expect(importPlugin).toBeDefined();
    expect(importPlugin.configs.recommended).toBeDefined();
  });

  test('config presets should be valid config entries', () => {
    for (const plugin of [ts, js, reactPlugin, importPlugin]) {
      const rec = plugin.configs.recommended;
      expect(typeof rec).toBe('object');
      expect(rec).not.toBeNull();
    }
  });

  test('defineConfig with preset should work with normalizeConfig', () => {
    const config = defineConfig([
      ts.configs.recommended,
      { rules: { '@typescript-eslint/no-explicit-any': 'off' } },
    ]);
    const normalized = normalizeConfig(config);
    expect(normalized.length).toBe(2);
    const lastEntry = normalized[normalized.length - 1];
    expect(lastEntry.rules).toEqual({
      '@typescript-eslint/no-explicit-any': 'off',
    });
  });

  test('ts.configs.recommended should declare @typescript-eslint plugin', () => {
    const rec = ts.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('@typescript-eslint');
  });

  test('react.configs.recommended should declare react plugin', () => {
    const rec = reactPlugin.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('react');
  });

  test('import.configs.recommended should declare import plugin', () => {
    const rec = importPlugin.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('eslint-plugin-import');
  });
});
