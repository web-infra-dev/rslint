import { describe, test, expect } from '@rstest/core';
import { normalizeConfig } from '@rslint/core/config-loader';
import {
  defineConfig,
  ts,
  js,
  reactPlugin,
  importPlugin,
  rstestPlugin,
} from '@rslint/core';

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
    expect(rstestPlugin).toBeDefined();
    expect(rstestPlugin.configs.recommended).toBeDefined();
  });

  test('config presets should be valid config entries', () => {
    for (const plugin of [ts, js, reactPlugin, importPlugin, rstestPlugin]) {
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
    expect(normalized.length).toBe(ts.configs.recommended.length + 1);
    const lastEntry = normalized[normalized.length - 1];
    expect(lastEntry.rules).toEqual({
      '@typescript-eslint/no-explicit-any': 'off',
    });
  });

  test('ts.configs.recommended should declare @typescript-eslint plugin', () => {
    const plugins = ts.configs.recommended.flatMap((entry) =>
      Array.isArray(entry.plugins) ? entry.plugins : [],
    );
    expect(plugins).toContain('@typescript-eslint');
  });

  const tsPresetNames = [
    'recommended',
    'recommendedTypeChecked',
    'strict',
    'strictTypeChecked',
    'stylistic',
    'stylisticTypeChecked',
  ] as const;

  test.each(tsPresetNames)(
    'ts.configs.%s layers the base entry and the eslint-recommended override',
    (name) => {
      const entries = ts.configs[name];
      expect(Array.isArray(entries)).toBe(true);

      const plugins = entries.flatMap((entry) =>
        Array.isArray(entry.plugins) ? entry.plugins : [],
      );
      expect(plugins).toContain('@typescript-eslint');

      // The eslint-recommended layer is what turns off the core rules
      // TypeScript already reports on.
      const overrides = entries.find((entry) =>
        entry.files?.includes('**/*.mts'),
      );
      expect(overrides?.rules?.['no-undef']).toBe('off');

      // The preset's own rules always land in the last entry.
      const rules = entries[entries.length - 1].rules ?? {};
      expect(Object.keys(rules).length).toBeGreaterThan(0);
      for (const ruleName of Object.keys(rules)) {
        // A preset only ever enables its own plugin's rules; core rules
        // appear solely as 'off' to make room for their TS-aware counterpart.
        if (!ruleName.startsWith('@typescript-eslint/')) {
          expect(rules[ruleName]).toBe('off');
        }
      }
    },
  );

  test.each([
    ['recommended', 'recommendedTypeChecked'],
    ['strict', 'strictTypeChecked'],
    ['stylistic', 'stylisticTypeChecked'],
    ['recommended', 'strict'],
    ['recommendedTypeChecked', 'strictTypeChecked'],
  ] as const)(
    'ts.configs.%s is contained in ts.configs.%s',
    (subset, superset) => {
      const enabled = (name: (typeof tsPresetNames)[number]) => {
        const entries = ts.configs[name];
        const rules = entries[entries.length - 1].rules ?? {};
        return Object.keys(rules).filter((key) => rules[key] !== 'off');
      };
      expect(enabled(superset)).toEqual(
        expect.arrayContaining(enabled(subset)),
      );
    },
  );

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

  test('rstestPlugin.configs.recommended should declare rstest plugin and rule', () => {
    const rec = rstestPlugin.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('rstest');
    expect(rec.rules).toEqual({
      'rstest/no-commented-out-tests': 'warn',
      'rstest/no-conditional-expect': 'error',
      'rstest/no-disabled-tests': 'warn',
      'rstest/no-focused-tests': 'error',
      'rstest/no-identical-title': 'error',
      'rstest/no-interpolation-in-snapshots': 'error',
      'rstest/no-mocks-import': 'error',
      'rstest/valid-title': 'error',
    });
  });
});
