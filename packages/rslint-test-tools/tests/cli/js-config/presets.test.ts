import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { describe, test, expect } from 'rstack/test';
import { normalizeConfig } from '@rslint/core/config-loader';
import { lint } from '@rslint/core/internal';
import {
  type RslintConfigEntry,
  defineConfig,
  globals,
  ts,
  js,
  reactPlugin,
  importPlugin,
  nodePlugin,
  rstestPlugin,
  unicornPlugin,
} from '@rslint/core';
import { createTempDir, cleanupTempDir, runRslint } from './helpers.js';

describe('defineConfig and config presets', () => {
  test('defineConfig should be importable and return input as-is', () => {
    const input = [
      { files: ['**/*.ts'], rules: { 'no-console': 'error' as const } },
    ];
    const result = defineConfig(input);
    expect(result).toBe(input);
  });

  test('globals should be importable from the public root', () => {
    expect(globals.node.process).toBe(false);
    expect(globals.nodeBuiltin.process).toBe(false);
    expect(Object.hasOwn(globals.nodeBuiltin, 'require')).toBe(false);
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
    expect(nodePlugin.configs.recommended).toBeDefined();
    expect(nodePlugin.configs.recommendedModule).toBeDefined();
    expect(nodePlugin.configs.recommendedScript).toBeDefined();
    expect(rstestPlugin).toBeDefined();
    expect(rstestPlugin.configs.recommended).toBeDefined();
    expect(unicornPlugin).toBeDefined();
    expect(unicornPlugin.configs.recommended).toBeDefined();
  });

  test('config presets should be valid config entries', () => {
    for (const plugin of [
      ts,
      js,
      reactPlugin,
      importPlugin,
      nodePlugin,
      rstestPlugin,
      unicornPlugin,
    ]) {
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

  test('import.configs.recommended should declare import plugin and report unresolved imports', async () => {
    const rec = importPlugin.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('eslint-plugin-import');

    const directory = import.meta.dirname;
    const result = await lint({
      config: normalizeConfig([rec]),
      configDirectory: directory,
      workingDirectory: directory,
      fileContents: {
        [path.join(directory, 'import-preset.js')]:
          'import "./missing-import-preset.js";',
      },
    });
    expect(result.fileCount).toBe(1);
    expect(result.diagnostics).toHaveLength(1);
    expect(result.diagnostics[0]).toMatchObject({
      ruleName: 'import/no-unresolved',
      severity: 'error',
      message: "Unable to resolve path to module './missing-import-preset.js'.",
    });
  });

  test('import.configs.recommended reports duplicate exports', async () => {
    const directory = import.meta.dirname;
    const result = await lint({
      config: normalizeConfig([importPlugin.configs.recommended]),
      configDirectory: directory,
      workingDirectory: directory,
      fileContents: {
        [path.join(directory, 'duplicate-exports-preset.ts')]:
          'export const duplicated = 1; export { duplicated };',
      },
    });

    expect(result.fileCount).toBe(1);
    expect(result.diagnostics).toMatchObject([
      {
        ruleName: 'import/export',
        messageId: 'multipleNamed',
        severity: 'error',
      },
      {
        ruleName: 'import/export',
        messageId: 'multipleNamed',
        severity: 'error',
      },
    ]);
  });

  test('import.configs.recommended warns when a default import uses a named export', async () => {
    const directory = import.meta.dirname;
    const result = await lint({
      config: normalizeConfig([importPlugin.configs.recommended]),
      configDirectory: directory,
      workingDirectory: directory,
      fileContents: {
        [path.join(directory, 'named-default-preset.js')]:
          'import foo from "./named-default-preset-dependency.js";',
        [path.join(directory, 'named-default-preset-dependency.js')]:
          'export default 1; export const foo = 2;',
      },
    });

    expect(result.fileCount).toBe(2);
    expect(result.diagnostics).toMatchObject([
      {
        ruleName: 'import/no-named-as-default',
        messageId: 'noNamedAsDefault',
        severity: 'warn',
        message: "Using exported name 'foo' as identifier for default import.",
      },
    ]);
  });

  test('rstestPlugin.configs.recommended should declare rstest plugin and rule', () => {
    const rec = rstestPlugin.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('rstest');
    expect(rec.rules).toEqual({
      'rstest/expect-expect': 'warn',
      'rstest/no-async-mock-factory': 'error',
      'rstest/no-commented-out-tests': 'warn',
      'rstest/no-conditional-expect': 'error',
      'rstest/no-disabled-tests': 'warn',
      'rstest/no-focused-tests': 'error',
      'rstest/no-identical-title': 'error',
      'rstest/no-import-node-test': 'error',
      'rstest/no-interpolation-in-snapshots': 'error',
      'rstest/no-mocks-import': 'error',
      'rstest/no-standalone-expect': 'error',
      'rstest/no-unneeded-async-expect-function': 'error',
      'rstest/prefer-called-exactly-once-with': 'error',
      'rstest/require-local-test-context-for-concurrent-snapshots': 'error',
      'rstest/valid-expect': 'error',
      'rstest/valid-expect-in-promise': 'error',
      'rstest/valid-title': 'error',
    });
  });

  test('unicornPlugin.configs.recommended should declare unicorn plugin and the ported rule', () => {
    const rec = unicornPlugin.configs.recommended;
    expect(rec.plugins).toBeDefined();
    expect(rec.plugins).toContain('unicorn');
    expect(rec.rules?.['unicorn/no-array-fill-with-reference-type']).toBe(
      'error',
    );
    expect(rec.rules?.['unicorn/no-array-reverse']).toBe('error');
    expect(rec.rules?.['unicorn/no-useless-error-capture-stack-trace']).toBe(
      'error',
    );
    expect(rec.rules?.['unicorn/empty-brace-spaces']).toBe('error');
    expect(rec.rules?.['unicorn/no-exports-in-scripts']).toBe('error');
    expect(rec.rules?.['unicorn/no-await-expression-member']).toBe('error');
    expect(rec.rules?.['unicorn/number-literal-case']).toBe('error');
    expect(rec.rules?.['unicorn/prefer-date-now']).toBe('error');
    expect(rec.rules?.['unicorn/require-post-message-target-origin']).toBe(
      'off',
    );
  });

  test('unicornPlugin.configs.recommended disables an earlier require-post-message-target-origin setting', async () => {
    const ruleName = 'unicorn/require-post-message-target-origin';
    const enabled: RslintConfigEntry = {
      plugins: ['unicorn'],
      rules: { [ruleName]: 'error' },
    };
    const preset = unicornPlugin.configs.recommended;
    const directory = import.meta.dirname;
    const options = {
      configDirectory: directory,
      workingDirectory: directory,
      fileContents: {
        [path.join(directory, 'post-message-preset.js')]:
          'window.postMessage(message);',
      },
    };

    // Check a positive control and that an explicit setting after the preset
    // can still enable the rule.
    for (const config of [[enabled], [preset, enabled]]) {
      const result = await lint({
        ...options,
        config: normalizeConfig(config),
      });
      expect(result.fileCount).toBe(1);
      expect(result.diagnostics).toHaveLength(1);
      expect(result.diagnostics[0]).toMatchObject({
        ruleName,
        messageId: 'error',
      });
    }

    const result = await lint({
      ...options,
      config: normalizeConfig([enabled, preset]),
    });
    expect(result.fileCount).toBe(1);
    expect(result.diagnostics).toEqual([]);
  });
});

describe('Node presets', () => {
  interface Diagnostic {
    ruleName: string;
    filePath: string;
    message: string;
  }

  // Temporary projects have no installed dependencies. Resolve the public
  // entry from this workspace before loading it in the CLI subprocess.
  const coreEntry = JSON.stringify(
    pathToFileURL(require.resolve('@rslint/core')).href,
  );
  const nodeConfig = `
    import { defineConfig, nodePlugin } from ${coreEntry};
    export default defineConfig([
      nodePlugin.configs.recommended,
      { rules: { 'no-undef': 'error' } },
    ]);
  `;

  function diagnostics(stdout: string): Diagnostic[] {
    return stdout
      .trim()
      .split('\n')
      .filter(Boolean)
      .map((line) => JSON.parse(line) as Diagnostic);
  }

  test.each(['module', 'commonjs'])(
    'recommended runs every ported upstream rule in a %s package',
    async (type) => {
      const files = {
        'imports.mjs': `import './missing.js'; import 'extra'; import 'dev';`,
        'requires.cjs': `require('./missing.js'); require('extra'); require('dev');`,
        'deprecated.js': 'new Buffer(0);',
        'exports.cjs': 'exports = {};',
        'exit.js': 'process.exit();',
        'syntax.js': 'const value = object?.value;',
        'builtins.js': 'Object.fromEntries([]);',
        'node-builtins.js': `process.getBuiltinModule('fs');`,
        'bin.js': `console.log('ok');`,
      };
      const directory = await createTempDir({
        ...files,
        'package.json': JSON.stringify({
          name: 'node-presets-fixture',
          version: '1.0.0',
          type,
          engines: { node: '>=10.0.0' },
          devDependencies: { dev: '1.0.0' },
          bin: { fixture: 'bin.js' },
        }),
        'node_modules/extra/package.json': '{"name":"extra","main":"index.js"}',
        'node_modules/extra/index.js': '',
        'node_modules/dev/package.json': '{"name":"dev","main":"index.js"}',
        'node_modules/dev/index.js': '',
        'rslint.config.mjs': `
          import { defineConfig, nodePlugin } from ${coreEntry};
          export default defineConfig([nodePlugin.configs.recommended]);
        `,
      });
      try {
        const result = await runRslint(
          ['--format', 'jsonline', ...Object.keys(files)],
          directory,
        );
        expect(result.exitCode, result.stderr).toBe(1);
        expect(result.stdout, result.stderr).not.toBe('');
        const reports = diagnostics(result.stdout);
        const expected = {
          'imports.mjs': [
            'node/no-extraneous-import',
            'node/no-missing-import',
            'node/no-unpublished-import',
          ],
          'requires.cjs': [
            'node/no-extraneous-require',
            'node/no-missing-require',
            'node/no-unpublished-require',
          ],
          'deprecated.js': ['node/no-deprecated-api'],
          'exports.cjs': ['node/no-exports-assign'],
          'exit.js': ['node/no-process-exit'],
          'syntax.js': ['node/no-unsupported-features/es-syntax'],
          'builtins.js': [
            'node/no-unsupported-features/es-builtins',
            'node/no-unsupported-features/es-syntax',
          ],
          'node-builtins.js': ['node/no-unsupported-features/node-builtins'],
          'bin.js': ['node/hashbang'],
        };
        for (const [file, rules] of Object.entries(expected)) {
          expect(
            reports
              .filter((report) => path.basename(report.filePath) === file)
              .map((report) => report.ruleName)
              .sort(),
            file,
          ).toEqual(rules);
        }
        expect(reports).toHaveLength(14);
        // The diagnostic set covers every enabled rule, including nested names.
        expect(
          [...new Set(reports.map((report) => report.ruleName))].sort(),
        ).toEqual(
          Object.keys(nodePlugin.configs.recommendedModule.rules ?? {}).sort(),
        );
      } finally {
        await cleanupTempDir(directory);
      }
    },
  );

  test.each(['module', 'commonjs', undefined])(
    'recommended selects globals and extension overrides with package type %s',
    async (type) => {
      const files = [
        'input.js',
        'input.mjs',
        'input.cjs',
        'nested/.input.mjs',
        'nested/.input.cjs',
      ];
      const directory = await createTempDir({
        'package.json': JSON.stringify({ type }),
        'rslint.config.mjs': nodeConfig,
        ...Object.fromEntries(
          files.map((file) => [file, `require('./missing.js');`]),
        ),
      });
      try {
        const result = await runRslint(
          ['--format', 'jsonline', ...files],
          directory,
        );
        expect(result.exitCode, result.stderr).toBe(1);
        expect(result.stdout, result.stderr).not.toBe('');
        const reports = diagnostics(result.stdout);
        expect(reports).toHaveLength(files.length);
        for (const file of files) {
          const isModule =
            file.endsWith('.mjs') ||
            (file.endsWith('.js') && type === 'module');
          expect(
            reports.find(
              (report) =>
                path.normalize(report.filePath) === path.normalize(file),
            ),
          ).toMatchObject({
            ruleName: isModule ? 'no-undef' : 'node/no-missing-require',
          });
        }
      } finally {
        await cleanupTempDir(directory);
      }
    },
  );

  test.each([
    {
      name: 'ancestor module package',
      parent: '{"type":"module"}',
      child: undefined,
      module: true,
    },
    {
      name: 'nearest package without type',
      parent: '{"type":"module"}',
      child: '{}',
      module: false,
    },
    {
      name: 'malformed package',
      parent: '{"type":"module"}',
      child: '{',
      module: true,
    },
    {
      name: 'non-object package',
      parent: '{"type":"module"}',
      child: 'null',
      module: true,
    },
    {
      name: 'array package',
      parent: '{"type":"module"}',
      child: '[]',
      module: true,
    },
    {
      name: 'no project package',
      parent: undefined,
      child: undefined,
      module: false,
    },
  ])(
    'recommended handles $name from the working directory',
    async ({ parent, child, module }) => {
      const directory = await createTempDir({
        // Stop package lookup at the fixture boundary.
        'package.json': '{"type":"commonjs"}',
        ...(parent === undefined ? {} : { 'project/package.json': parent }),
        ...(child === undefined
          ? {}
          : { 'project/nested/package.json': child }),
        'project/rslint.config.mjs': nodeConfig,
        'project/nested/input.js': `require('./missing.js');`,
      });
      try {
        const result = await runRslint(
          [
            '--config',
            path.join(directory, 'project', 'rslint.config.mjs'),
            '--format',
            'jsonline',
            'input.js',
          ],
          path.join(directory, 'project', 'nested'),
        );
        expect(result.exitCode, result.stderr).toBe(1);
        expect(result.stdout, result.stderr).not.toBe('');
        expect(
          diagnostics(result.stdout).map((report) => report.ruleName),
        ).toEqual([module ? 'no-undef' : 'node/no-missing-require']);
      } finally {
        await cleanupTempDir(directory);
      }
    },
  );

  test.each([
    ['recommendedModule', 'input.cjs'],
    ['recommendedScript', 'input.mjs'],
  ] as const)(
    '%s applies globals and scopes even to %s',
    async (name, filename) => {
      const directory = import.meta.dirname;
      const result = await lint({
        configDirectory: directory,
        workingDirectory: directory,
        config: normalizeConfig(
          defineConfig([
            { languageOptions: { globals: globals.node } },
            nodePlugin.configs[name],
            {
              rules: {
                'no-undef': 'error',
                'no-global-assign': 'error',
                'no-invalid-this': 'error',
              },
            },
          ]),
        ),
        fileContents: {
          [path.join(directory, filename)]: `
          console.log(process, Buffer, Promise, WeakRef, __dirname, __filename, require, module, exports);
          exports = {};
          require = 1;
          function receiver() { return this; }
        `,
        },
      });
      expect(result.fileCount).toBe(1);
      const ruleNames = result.diagnostics
        .map((report) => report.ruleName)
        .sort();
      expect(ruleNames).toEqual(
        name === 'recommendedModule'
          ? [...Array<string>(7).fill('no-undef'), 'no-invalid-this'].sort()
          : ['no-global-assign', 'node/no-exports-assign'],
      );
    },
  );

  test.each([
    ['recommendedModule', 1],
    ['recommendedScript', 2],
  ] as const)(
    '%s preserves its syntax ignores through normalization and native option parsing',
    async (name, count) => {
      const directory = import.meta.dirname;
      const result = await lint({
        configDirectory: directory,
        workingDirectory: directory,
        config: normalizeConfig(
          defineConfig([
            nodePlugin.configs.recommendedModule,
            nodePlugin.configs[name],
            {
              // Both inputs are modules so this isolates the rule's ignores option.
              languageOptions: { sourceType: 'module' },
              settings: { node: { version: '10.0.0' } },
            },
          ]),
        ),
        fileContents: {
          [path.join(directory, 'syntax.mjs')]:
            'export const value = object?.value;',
        },
      });
      expect(result.fileCount).toBe(1);
      expect(result.diagnostics).toHaveLength(count);
      expect(
        result.diagnostics.every(
          (report) =>
            report.ruleName === 'node/no-unsupported-features/es-syntax',
        ),
      ).toBe(true);
      expect(
        result.diagnostics.some((report) =>
          report.message.includes("'modules'"),
        ),
      ).toBe(name === 'recommendedScript');
    },
  );
});
