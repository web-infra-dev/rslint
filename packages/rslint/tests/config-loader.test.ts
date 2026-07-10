import { describe, test, expect } from '@rstest/core';
import {
  loadConfigFile,
  normalizeConfig,
  collectPluginMeta,
} from '../src/config/config-loader.js';
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

  test('preserves name and strips unknown fields', () => {
    const result = normalizeConfig([
      {
        name: 'my-config',
        files: ['**/*.ts'],
        rules: {},
        unknownField: 123,
      },
    ]);
    expect(result[0].name).toBe('my-config');
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

  test.each([null, undefined, 42, 'string'])(
    'rejects non-object config entry %p',
    (entry) => {
      expect(() => normalizeConfig([entry])).toThrow(/must be an object|null/);
    },
  );

  test('rejects nested config arrays', () => {
    expect(() => normalizeConfig([[{ rules: {} }]])).toThrow(
      /unexpected array/,
    );
  });

  test('throws when files is a string instead of array', () => {
    expect(() => normalizeConfig([{ files: '**/*.ts', rules: {} }])).toThrow(
      '"files" must be an array',
    );
  });

  test('throws when files is an empty array', () => {
    expect(() => normalizeConfig([{ files: [], rules: {} }])).toThrow(
      '"files" must be a non-empty array',
    );
  });

  test('throws when files is null', () => {
    expect(() => normalizeConfig([{ files: null, rules: {} }])).toThrow(
      '"files" must be an array',
    );
  });

  test('throws when files contains invalid selector values', () => {
    expect(() => normalizeConfig([{ files: [123], rules: {} }])).toThrow(
      '"files" must contain only strings or arrays of strings',
    );
  });

  test('preserves ESLint AND selector groups, including an empty group', () => {
    const [entry] = normalizeConfig([
      {
        files: ['**/*.ts', ['**/*.js', '!**/*.test.js'], []],
        rules: {},
      },
    ]);
    expect(entry.files).toEqual(['**/*.ts', ['**/*.js', '!**/*.test.js'], []]);
  });

  test('throws when ignores is a string instead of array', () => {
    expect(() => normalizeConfig([{ ignores: 'dist/**', rules: {} }])).toThrow(
      '"ignores" must be an array',
    );
  });

  test('throws when ignores contains non-string values', () => {
    expect(() => normalizeConfig([{ ignores: [null], rules: {} }])).toThrow(
      '"ignores" must contain only strings',
    );
  });

  test('allows omitted files and ignores', () => {
    const result = normalizeConfig([{ rules: { 'no-console': 'error' } }]);
    expect(result).toHaveLength(1);
    expect(result[0].files).toBeUndefined();
    expect(Object.hasOwn(result[0], 'files')).toBe(false);
    expect(result[0].ignores).toBeUndefined();
    expect(Object.hasOwn(result[0], 'ignores')).toBe(false);
  });

  test('throws when files is explicitly undefined', () => {
    expect(() =>
      normalizeConfig([{ files: undefined, rules: { 'no-console': 'error' } }]),
    ).toThrow('"files" must be an array');
  });

  test('preserves non-global shape when unsupported fields are stripped', () => {
    const [entry] = normalizeConfig([
      { ignores: ['dist/**'], processor: 'example-processor' },
    ]);
    expect(entry).toEqual({ ignores: ['dist/**'], settings: {} });
  });

  test('rejects authored undefined known fields like ESLint v10', () => {
    for (const authoredField of [
      { rules: undefined },
      { languageOptions: undefined },
      { settings: undefined },
      { plugins: undefined },
      { ignores: undefined },
    ]) {
      expect(() =>
        normalizeConfig([{ ignores: ['dist/**'], ...authoredField }]),
      ).toThrow(/must be/);
    }
  });
});

describe('normalizeConfig — community plugins (object-form)', () => {
  type NormalizedPluginEntry = {
    eslintPlugins?: Record<string, { ruleNames: string[] }>;
    plugins?: string[];
  };
  const mockPlugin = {
    meta: { name: 'p' },
    rules: { 'no-foo': {}, 'no-bar': {} },
  };

  test('object-form plugins: strips live objects, emits sorted {ruleNames} meta', () => {
    const [entry] = normalizeConfig([
      {
        files: ['**/*.ts'],
        plugins: { local: mockPlugin },
        rules: { 'local/no-foo': 'error' },
      },
    ]) as NormalizedPluginEntry[];
    expect(entry.eslintPlugins).toEqual({
      local: { ruleNames: ['no-bar', 'no-foo'] },
    });
    // The prefix is merged into the string `plugins` set Go's gate keys off.
    expect(entry.plugins).toContain('local');
    // The live plugin object (carrying `meta`/`create`) must not leak into
    // the serializable payload sent to Go.
    expect(JSON.stringify(entry)).not.toContain('meta');
  });

  test('array-form plugins: native names pass through as the string[] gate, no carrier', () => {
    const [entry] = normalizeConfig([
      {
        files: ['**/*.ts'],
        plugins: ['@typescript-eslint', 'import'],
        rules: {},
      },
    ]) as NormalizedPluginEntry[];
    // Array form is the native-name whitelist: it reaches Go as the plugins
    // string[] and emits NO community-plugin carrier (no live objects).
    expect(entry.plugins).toEqual(['@typescript-eslint', 'import']);
    expect(entry.eslintPlugins).toBeUndefined();
  });

  test('multiple object-form plugin prefixes in one entry', () => {
    // The normal object-form usage mounts more than one community plugin; each
    // prefix's ruleNames + the prefix gate must be collected independently.
    const pluginB = { meta: { name: 'b' }, rules: { 'no-baz': {} } };
    const [entry] = normalizeConfig([
      {
        files: ['**/*.ts'],
        plugins: { local: mockPlugin, other: pluginB },
        rules: { 'local/no-foo': 'error', 'other/no-baz': 'error' },
      },
    ]) as NormalizedPluginEntry[];
    expect(entry.eslintPlugins).toEqual({
      local: { ruleNames: ['no-bar', 'no-foo'] },
      other: { ruleNames: ['no-baz'] },
    });
    expect(entry.plugins).toEqual(['local', 'other']);
  });

  test('throws when a mounted plugin has no rules object', () => {
    expect(() =>
      normalizeConfig([
        { files: ['**/*.ts'], plugins: { bad: { meta: {} } }, rules: {} },
      ]),
    ).toThrow(/must expose a "rules" object/);
  });

  test('throws when an object-form prefix collides with a native plugin name', () => {
    // Asymmetry: a native NAME is legal in the array form (previous test) but
    // illegal as an object-form KEY — native rules always win, so mounting a
    // community plugin under a native prefix would silently shadow it.
    expect(() =>
      normalizeConfig([
        {
          files: ['**/*.ts'],
          plugins: { '@typescript-eslint': mockPlugin },
          rules: {},
        },
      ]),
    ).toThrow(/collides with the built-in plugin/);
  });

  const RESERVED_DECL_ALIASES = [
    'eslint-plugin-import',
    'eslint-plugin-jest',
    'eslint-plugin-jsx-a11y',
    'eslint-plugin-promise',
    'eslint-plugin-react-hooks',
    'eslint-plugin-unicorn',
  ];
  for (const alias of RESERVED_DECL_ALIASES) {
    test(`throws when an object-form prefix is the native decl-name ${alias}`, () => {
      // Each `eslint-plugin-*` alias normalizes to a native prefix in Go, so a
      // community plugin mounted under it would otherwise pass the JS guard but
      // be silently dropped by the Go gate. Reject it loudly.
      expect(() =>
        normalizeConfig([
          { files: ['**/*.ts'], plugins: { [alias]: mockPlugin }, rules: {} },
        ]),
      ).toThrow(/collides with the built-in plugin/);
    });
  }

  test('accepts an eslint-plugin-* key that is NOT a native decl-name', () => {
    // `eslint-plugin-react` has no Go DeclName alias (react is declared bare),
    // so reserving it would wrongly false-reject a legitimate community mount.
    // The asymmetry must be exact: only the 6 aliased names are reserved.
    const [entry] = normalizeConfig([
      {
        files: ['**/*.ts'],
        plugins: { 'eslint-plugin-react': mockPlugin },
        rules: { 'eslint-plugin-react/no-foo': 'error' },
      },
    ]) as NormalizedPluginEntry[];
    expect(entry.plugins).toContain('eslint-plugin-react');
    expect(entry.eslintPlugins).toEqual({
      'eslint-plugin-react': { ruleNames: ['no-bar', 'no-foo'] },
    });
  });

  test('entries with no plugins carry no community-plugin field', () => {
    const [entry] = normalizeConfig([
      { files: ['**/*.ts'], rules: {} },
    ]) as NormalizedPluginEntry[];
    expect(entry.eslintPlugins).toBeUndefined();
    expect(entry.plugins).toBeUndefined();
    expect(Object.hasOwn(entry, 'plugins')).toBe(false);
  });

  test('explicit empty object-form plugins {} is preserved as an empty gate', () => {
    const [entry] = normalizeConfig([
      { files: ['**/*.ts'], plugins: {}, rules: {} },
    ]) as NormalizedPluginEntry[];
    expect(entry.eslintPlugins).toBeUndefined();
    expect(entry.plugins).toEqual([]);
  });

  test('explicit empty array-form plugins [] is preserved on the wire', () => {
    const [entry] = normalizeConfig([
      { files: ['**/*.ts'], plugins: [], rules: {} },
    ]) as NormalizedPluginEntry[];
    expect(entry.eslintPlugins).toBeUndefined();
    expect(entry.plugins).toEqual([]);
    expect(Object.hasOwn(entry, 'plugins')).toBe(true);
  });
});

describe('collectPluginMeta', () => {
  const mockPlugin = {
    meta: { name: 'p' },
    rules: { 'no-foo': {}, 'no-bar': {} },
  };

  test('aggregates entries + descriptors, only for plugin-mounting configs', () => {
    const { eslintPluginEntries, pluginConfigs } = collectPluginMeta([
      {
        configPath: '/proj/rslint.config.mjs',
        configDirectory: '/proj',
        entries: normalizeConfig([
          {
            files: ['**/*.ts'],
            plugins: { local: mockPlugin },
            rules: {},
          },
        ]),
      },
      {
        configPath: '/proj/sub/rslint.config.mjs',
        configDirectory: '/proj/sub',
        entries: normalizeConfig([{ files: ['**/*.ts'], rules: {} }]),
      },
    ]);
    expect(eslintPluginEntries).toEqual([
      { prefix: 'local', ruleNames: ['no-bar', 'no-foo'] },
    ]);
    // Only the plugin-mounting config gets a worker-pool descriptor; the plain
    // config stays zero-overhead (no worker spun up for it).
    expect(pluginConfigs).toEqual([
      { configPath: '/proj/rslint.config.mjs', configDirectory: '/proj' },
    ]);
  });

  test('ruleNames are merged across configs sharing a prefix (per-config-unique rules still register)', () => {
    // Go registers ONE global placeholder set per prefix but the worker routes
    // per-config; a rule mounted only by /b's `local` (zzz) must still register,
    // or Go never dispatches it for files under /b (silent false-green).
    const { eslintPluginEntries } = collectPluginMeta([
      {
        configPath: '/a/rslint.config.mjs',
        configDirectory: '/a',
        entries: normalizeConfig([
          { plugins: { local: mockPlugin }, rules: {} },
        ]),
      },
      {
        configPath: '/b/rslint.config.mjs',
        configDirectory: '/b',
        entries: normalizeConfig([
          { plugins: { local: { rules: { zzz: {} } } }, rules: {} },
        ]),
      },
    ]);
    expect(eslintPluginEntries).toEqual([
      { prefix: 'local', ruleNames: ['no-bar', 'no-foo', 'zzz'] },
    ]);
  });
});
