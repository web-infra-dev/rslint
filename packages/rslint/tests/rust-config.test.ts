import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, describe, expect, test } from 'rstack/test';
import { admitRustFiles } from '../src/rust/admission.js';
import { resolveRustConfig } from '../src/rust/config.js';

const roots: string[] = [];
function fixture(files: Record<string, string> = {}): string {
  const root = fs.realpathSync(
    fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rust-config-')),
  );
  roots.push(root);
  for (const [name, text] of Object.entries(files)) {
    fs.mkdirSync(path.dirname(path.join(root, name)), { recursive: true });
    fs.writeFileSync(path.join(root, name), text);
  }
  return root;
}

afterEach(() => {
  for (const root of roots.splice(0))
    fs.rmSync(root, { recursive: true, force: true });
});

describe('Rust config selection and evaluation', () => {
  test('evaluates async modules and imported presets without changing their rules or options', async () => {
    const root = fixture({
      'package.json': '{"type":"module"}',
      'config.mjs': `
        export default Promise.resolve((async () => {
          const { preset } = await import('@rslint/core');
          return [preset, { rules: { 'no-debugger': 'warn' } }];
        })());
      `,
      'node_modules/@rslint/core/package.json':
        '{"name":"@rslint/core","type":"module","exports":"./index.js"}',
      'node_modules/@rslint/core/index.js': `
        export const preset = [{
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-unused-vars': ['error', {argsIgnorePattern: '^_'}] },
          languageOptions: { parserOptions: { projectService: true } },
        }];
      `,
    });
    const result = await resolveRustConfig({
      cwd: root,
      overrideConfigFile: './config.mjs',
    });
    expect(result).toEqual({
      cwd: root,
      configDirectory: root,
      configArrayBase: root,
      configEntryDirectories: [root, root],
      config: [
        {
          plugins: ['@typescript-eslint'],
          rules: {
            '@typescript-eslint/no-unused-vars': [
              'error',
              { argsIgnorePattern: '^_' },
            ],
          },
          languageOptions: { parserOptions: { projectService: true } },
        },
        { rules: { 'no-debugger': 'warn' } },
      ],
    });
  });

  test('keeps module and inline origins distinct without adding basePath scopes', async () => {
    const root = fixture({
      'input.ts': '',
      'config/input.ts': '',
      'config/rules.mjs':
        'export default [{files:["input.ts"],rules:{"no-debugger":"error"}}];',
    });
    const cwd = root;
    const resolved = await resolveRustConfig({
      cwd,
      overrideConfigFile: './config/rules.mjs',
      overrideConfig: { files: ['input.ts'], rules: { 'no-empty': 'warn' } },
    });
    expect(resolved.configEntryDirectories).toEqual([
      path.join(root, 'config'),
      cwd,
    ]);
    expect(resolved.config.every((entry) => !('basePath' in entry))).toBe(true);
    const files = admitRustFiles({
      ...resolved,
      patterns: ['input.ts', 'config/input.ts'],
    });
    expect(
      files.map((file) => ({
        path: path.relative(root, file.path),
        rules: file.config.rules,
      })),
    ).toEqual([
      {
        path: path.join('config', 'input.ts'),
        rules: [{ name: 'no-debugger', severity: 2, options: [] }],
      },
      {
        path: 'input.ts',
        rules: [{ name: 'no-empty', severity: 1, options: [] }],
      },
    ]);
  });

  test('explicit authored basePath uses cwd for both module and override entries', async () => {
    const root = fixture({
      'src/a.ts': '',
      'outside.ts': '',
      'config/src/a.ts': '',
      'config/rules.mjs': `export default [{basePath:'src',rules:{'no-debugger':['error']}}];`,
    });
    const cwd = root;
    const resolved = await resolveRustConfig({
      cwd,
      overrideConfigFile: './config/rules.mjs',
      overrideConfig: { basePath: 'src', rules: { 'no-debugger': 'warn' } },
    });
    expect(resolved.configArrayBase).toBe(cwd);
    const files = admitRustFiles({
      ...resolved,
      patterns: ['src/a.ts', 'outside.ts', 'config/src/a.ts'],
    });
    expect(
      files
        .filter((file) => file.config.rules.length > 0)
        .map((file) => ({
          path: path.relative(cwd, file.path),
          rules: file.config.rules,
        })),
    ).toEqual([
      {
        path: path.join('src', 'a.ts'),
        rules: [{ name: 'no-debugger', severity: 1, options: [] }],
      },
    ]);
  });

  test('keeps the winning project declaration origin and the module service boundary', async () => {
    const root = fixture({
      'project/input.ts': '',
      'config/rules.mjs': `export default [{
        languageOptions:{parserOptions:{project:['./module.json']}}
      }];`,
    });
    const cwd = path.join(root, 'project');
    const module = await resolveRustConfig({
      cwd,
      overrideConfigFile: '../config/rules.mjs',
    });
    expect(
      admitRustFiles(module)[0].config.projectPolicy.explicitProject,
    ).toEqual({
      patterns: ['./module.json'],
      baseDirectory: path.join(root, 'config'),
    });
    const inline = await resolveRustConfig({
      cwd,
      overrideConfigFile: '../config/rules.mjs',
      overrideConfig: {
        languageOptions: { parserOptions: { project: ['./inline.json'] } },
      },
    });
    expect(
      admitRustFiles(inline)[0].config.projectPolicy.explicitProject,
    ).toEqual({
      patterns: ['./inline.json'],
      baseDirectory: cwd,
    });
    const service = await resolveRustConfig({
      cwd,
      overrideConfigFile: '../config/rules.mjs',
      overrideConfig: {
        languageOptions: {
          parserOptions: { project: null, projectService: true },
        },
      },
    });
    expect(
      admitRustFiles(service)[0].config.projectPolicy.serviceRootDirectory,
    ).toBe(path.join(root, 'config'));
  });

  test('inline-only mode bypasses a broken local module and snapshots options', async () => {
    const root = fixture({
      'rslint.config.mjs': 'throw new Error("must not execute");',
      'input.ts': '',
    });
    const options = { argsIgnorePattern: '^_' };
    const resolved = await resolveRustConfig({
      cwd: root,
      overrideConfigFile: true,
      overrideConfig: [
        [{ plugins: ['@typescript-eslint'] }],
        { rules: { '@typescript-eslint/no-unused-vars': ['error', options] } },
      ],
    });
    options.argsIgnorePattern = 'changed';
    expect(resolved.configEntryDirectories).toEqual([root, root]);
    expect(resolved.config[1].rules).toEqual({
      '@typescript-eslint/no-unused-vars': [
        'error',
        { argsIgnorePattern: '^_' },
      ],
    });
  });

  test('empty module and empty inline configurations retain admitted source files', async () => {
    // Frozen baseline e0a29f7 retains input.ts with no lint diagnostics in both modes.
    const root = fixture({ 'empty.mjs': 'export default [];', 'input.ts': '' });
    for (const selected of ['empty.mjs', true] as const) {
      const resolved = await resolveRustConfig({
        cwd: root,
        overrideConfigFile: selected,
      });
      expect(resolved.config).toEqual([]);
      expect(resolved.configEntryDirectories).toEqual([]);
      expect(
        admitRustFiles({ ...resolved, patterns: ['input.ts'] }).map((file) => ({
          path: path.relative(root, file.path),
          rules: file.config.rules,
        })),
      ).toEqual([{ path: 'input.ts', rules: [] }]);
    }
  });

  test('normalization preserves the distinction between global and local ignore entries', async () => {
    const root = fixture({ 'input.ts': '' });
    const resolved = await resolveRustConfig({
      cwd: root,
      overrideConfigFile: true,
      overrideConfig: [
        { ignores: ['input.ts'], custom: undefined },
        { rules: { 'no-debugger': 'error' } },
      ],
    });
    expect(resolved.config[0]).toEqual({ ignores: ['input.ts'], settings: {} });
    expect(admitRustFiles(resolved)[0].config.rules).toEqual([
      { name: 'no-debugger', severity: 2, options: [] },
    ]);
  });

  test('reevaluates the top-level ESM module on every request', async () => {
    const root = fixture({
      'rules.mjs': 'export default [{rules:{"no-debugger":"error"}}];',
    });
    const options = { cwd: root, overrideConfigFile: 'rules.mjs' };
    const first = await resolveRustConfig(options);
    fs.writeFileSync(
      path.join(root, 'rules.mjs'),
      'export default [{rules:{"no-debugger":"warn"}}];',
    );
    const second = await resolveRustConfig(options);
    expect(first.config[0].rules).toEqual({ 'no-debugger': 'error' });
    expect(second.config[0].rules).toEqual({ 'no-debugger': 'warn' });
  });

  test('evicts the selected CommonJS module between requests', async () => {
    const root = fixture({
      'rules.cjs': 'module.exports = [{rules:{"no-debugger":"error"}}];',
    });
    const options = { cwd: root, overrideConfigFile: 'rules.cjs' };
    const first = await resolveRustConfig(options);
    fs.writeFileSync(
      path.join(root, 'rules.cjs'),
      'module.exports = [{rules:{"no-debugger":"warn"}}];',
    );
    const second = await resolveRustConfig(options);
    expect(first.config[0].rules).toEqual({ 'no-debugger': 'error' });
    expect(second.config[0].rules).toEqual({ 'no-debugger': 'warn' });
  });

  test('loads native TypeScript config modules with explicit supported extensions', async () => {
    const root = fixture({
      'package.json': '{"type":"module"}',
      'rules.ts': 'const config: object[] = [{}]; export default config;',
      'rules.mts': 'const config: object[] = [{}]; export default config;',
      'rules.cts': 'const config: object[] = [{}]; module.exports = config;',
    });
    for (const extension of ['ts', 'mts', 'cts']) {
      const result = await resolveRustConfig({
        cwd: root,
        overrideConfigFile: `rules.${extension}`,
      });
      expect(result.config).toEqual([{}]);
    }
  });

  test('fails missing explicit files instead of falling back to inline configuration', async () => {
    const root = fixture();
    await expect(
      resolveRustConfig({
        cwd: root,
        overrideConfigFile: 'missing.mjs',
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      }),
    ).rejects.toThrow('ENOENT');
    await expect(
      resolveRustConfig({ cwd: root, overrideConfigFile: 'rslint.json' }),
    ).rejects.toThrow('explicit config path must name a JS/TS config module');
  });

  test('propagates config evaluation failure after exactly one execution', async () => {
    const root = fixture({
      'rules.mjs': `
        import fs from 'node:fs';
        fs.appendFileSync(new URL('./executions.txt',import.meta.url),'x');
        throw new Error('config failure');
      `,
    });
    await expect(
      resolveRustConfig({ cwd: root, overrideConfigFile: 'rules.mjs' }),
    ).rejects.toThrow('config failure');
    expect(fs.readFileSync(path.join(root, 'executions.txt'), 'utf8')).toBe(
      'x',
    );
  });

  test('rejects a module that changes its own source during evaluation', async () => {
    const root = fixture({
      'rules.mjs': `
        import fs from 'node:fs';
        fs.writeFileSync(new URL('./rules.mjs',import.meta.url),'export default [];');
        export default [{}];
      `,
    });
    await expect(
      resolveRustConfig({ cwd: root, overrideConfigFile: 'rules.mjs' }),
    ).rejects.toThrow('config changed while it was being loaded');
  });

  test('rejects invalid flat config and unsupported plugin objects without stripping rules', async () => {
    const root = fixture({
      'invalid.mjs': 'export default {rules:{"no-debugger":"error"}};',
      'plugin.mjs': `export default [{
        plugins:{custom:{rules:{test:{create(){return {};}}}}},
        rules:{'custom/test':'error'}
      }];`,
    });
    await expect(
      resolveRustConfig({ cwd: root, overrideConfigFile: 'invalid.mjs' }),
    ).rejects.toThrow('rslint config must export an array');
    await expect(
      resolveRustConfig({ cwd: root, overrideConfigFile: 'plugin.mjs' }),
    ).rejects.toThrow('does not yet support JavaScript plugin objects');
    await expect(
      resolveRustConfig({
        cwd: root,
        overrideConfigFile: true,
        overrideConfig: [{ plugins: { custom: { rules: {} } } }],
      }),
    ).rejects.toThrow(
      'overrideConfig entry at index 0 uses object-form "plugins"',
    );
  });

  test('refuses automatic discovery instead of silently choosing a config or the no-config API fallback', async () => {
    const root = fixture({
      'rslint.config.mjs': 'throw new Error("must not execute");',
    });
    for (const selected of [undefined, null]) {
      await expect(
        resolveRustConfig({ cwd: root, overrideConfigFile: selected }),
      ).rejects.toThrow(
        'does not yet support automatic configuration discovery',
      );
    }
  });
});
