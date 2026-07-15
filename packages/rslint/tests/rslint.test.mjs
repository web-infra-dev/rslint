import { Rslint } from '@rslint/core';
import { lint } from '@rslint/core/internal';
import { describe, test, expect, rs } from '@rstest/core';
import { ConfigModuleHost } from '@rslint/core/config-loader';
import path from 'node:path';
import os from 'node:os';
import { pathToFileURL } from 'node:url';
import {
  writeFile,
  rm,
  mkdtemp,
  mkdir,
  readFile,
  cp,
  symlink,
} from 'node:fs/promises';

const fixturesDir = path.resolve(import.meta.dirname, '../fixtures');
const eslintPluginFixturesDir = path.resolve(
  import.meta.dirname,
  'eslint-plugin/fixtures',
);

// A self-contained config (overrideConfigFile:true → no discovery, only
// overrideConfig). array-type is non-type-aware so it runs even on a gap file.
const arrayTypeConfig = [
  {
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    rules: { '@typescript-eslint/array-type': 'error' },
    plugins: ['@typescript-eslint'],
  },
];

describe('Rslint class', () => {
  test('published declarations expose async disposal under ES2022 consumer libs', async () => {
    const distDir = path.resolve(import.meta.dirname, '../dist');
    const packageJson = JSON.parse(
      await readFile(
        path.resolve(import.meta.dirname, '../package.json'),
        'utf8',
      ),
    );
    const distDts = await readFile(path.join(distDir, 'index.d.ts'), 'utf8');
    expect(distDts).toContain('reference lib="esnext.disposable"');
    expect(distDts).toContain('[Symbol.asyncDispose](): Promise<void>');

    const ts = await import('typescript');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-dts-consumer-'));
    try {
      const packageRoot = path.join(tmp, 'node_modules', '@rslint', 'core');
      await mkdir(packageRoot, { recursive: true });
      await writeFile(
        path.join(packageRoot, 'package.json'),
        JSON.stringify({
          name: packageJson.name,
          type: packageJson.type,
          exports: {
            '.': packageJson.exports['.'],
          },
        }),
      );
      await cp(distDir, path.join(packageRoot, 'dist'), { recursive: true });
      const entry = path.join(tmp, 'index.ts');
      await writeFile(
        entry,
        [
          "import { Rslint } from '@rslint/core';",
          'const rslint: Rslint = new Rslint();',
          'rslint.close();',
          '',
        ].join('\n'),
      );

      const options = {
        target: ts.ScriptTarget.ES2022,
        module: ts.ModuleKind.NodeNext,
        moduleResolution: ts.ModuleResolutionKind.NodeNext,
        lib: ['lib.es2022.d.ts'],
        noEmit: true,
        strict: true,
        skipLibCheck: false,
      };
      const program = ts.createProgram([entry], options);
      const diagnostics = ts.getPreEmitDiagnostics(program);
      const rendered = diagnostics
        .map((d) => ts.flattenDiagnosticMessageText(d.messageText, '\n'))
        .join('\n');
      expect(rendered).not.toContain('asyncDispose');
      expect(diagnostics).toEqual([]);
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText returns ESLint-shaped LintResult[]', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];', {
        filePath: 'gap-rslint.ts',
      });
      expect(results).toHaveLength(1);
      const r = results[0];
      // filePath is absolute (ESLint contract; outputFixes guards on it).
      expect(path.isAbsolute(r.filePath)).toBe(true);
      expect(r.filePath.endsWith('gap-rslint.ts')).toBe(true);
      expect(r.errorCount).toBe(1);
      expect(r.warningCount).toBe(0);
      expect(r.messages).toHaveLength(1);
      const m = r.messages[0];
      expect(m.ruleId).toBe('@typescript-eslint/array-type');
      expect(m.severity).toBe(2); // error → 2
      expect(m.line).toBe(1);
      expect(m.column).toBe(8);
      expect(m.endLine).toBe(1);
      expect(m.endColumn).toBe(21);
      expect(m.messageId).toBe('errorStringArray');
      // fix is a flat UTF-16 range + replacement text (Array<string> → string[]).
      expect(m.fix.range).toEqual([7, 20]);
      expect(m.fix.text).toBe('string[]');
    } finally {
      await rslint.close();
    }
  });

  test('lintText runs a local community plugin Program listener and returns its fix output', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-plugin-'));
    const pluginUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'local-plugin.mjs'),
    ).href;
    await writeFile(
      path.join(tmp, 'rslint.config.mjs'),
      [
        `import local from ${JSON.stringify(pluginUrl)};`,
        'export default [{',
        '  plugins: { local },',
        '  rules: {',
        "    'local/program-listener': 'error',",
        "    'local/prefer-array-some': 'error',",
        '  },',
        '}];',
        '',
      ].join('\n'),
    );

    const rslint = new Rslint({ cwd: tmp, fix: true });
    try {
      const [result] = await rslint.lintText(
        'const found = [1].filter(Boolean);\n',
        { filePath: 'probe.ts' },
      );
      expect(result.messages.map((message) => message.ruleId).sort()).toEqual([
        'local/prefer-array-some',
        'local/program-listener',
      ]);
      expect(result.fixableErrorCount).toBe(1);
      expect(result.output).toBe('const found = [1].some(Boolean);\n');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('rejects object-form community plugins in overrideConfig before linting', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          plugins: { local: { rules: { probe: {} } } },
          rules: { 'local/probe': 'error' },
        },
      ],
    });
    try {
      await expect(
        rslint.lintText('const value = 1;', { filePath: 'probe.ts' }),
      ).rejects.toThrow(/overrideConfig.*object-form.*cannot re-import/s);
    } finally {
      await rslint.close();
    }
  });

  test('lintText preserves the requested path when the Program uses a symlink alias', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-symlink-'));
    const realDir = path.join(tmp, 'real');
    const linkDir = path.join(tmp, 'link');
    const realTarget = path.join(realDir, 'src', 'a.ts');
    await mkdir(path.dirname(realTarget), { recursive: true });
    await writeFile(realTarget, 'let a: string[] = [];\n');
    await writeFile(
      path.join(realDir, 'tsconfig.json'),
      JSON.stringify({ include: ['src/a.ts'] }),
    );
    try {
      try {
        await symlink(realDir, linkDir, 'dir');
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: linkDir,
        overrideConfigFile: true,
        overrideConfig: arrayTypeConfig,
      });
      try {
        const results = await rslint.lintText('let a: Array<string> = [];\n', {
          filePath: realTarget,
        });
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(path.normalize(realTarget));
        expect(results[0].messages).toHaveLength(1);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText does not use a canonical-ancestor config fallback', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-virtual-symlink-'),
    );
    const realRoot = path.join(tmp, 'real-workspace');
    const realPackage = path.join(realRoot, 'packages', 'app');
    const aliasPackage = path.join(tmp, 'alias-app');
    await mkdir(realPackage, { recursive: true });
    await writeFile(
      path.join(realRoot, 'rslint.config.mjs'),
      [
        'export default [{',
        "  files: ['**/*.ts'],",
        "  plugins: ['@typescript-eslint'],",
        "  rules: { '@typescript-eslint/array-type': 'error' },",
        '}];',
        '',
      ].join('\n'),
    );

    try {
      try {
        await symlink(realPackage, aliasPackage, 'dir');
      } catch {
        return;
      }

      const virtualFile = path.join(aliasPackage, 'src', 'virtual.ts');
      const rslint = new Rslint({ cwd: tmp });
      try {
        await expect(
          rslint.lintText('let values: Array<string> = [];\n', {
            filePath: virtualFile,
          }),
        ).rejects.toThrow(/Could not find config file/);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('warn level maps to severity 1', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { '@typescript-eslint/array-type': 'warn' },
          plugins: ['@typescript-eslint'],
        },
      ],
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];', {
        filePath: 'gap-rslint-warn.ts',
      });
      const r = results[0];
      expect(r.errorCount).toBe(0);
      expect(r.warningCount).toBe(1);
      expect(r.messages[0].severity).toBe(1); // warn → 1
    } finally {
      await rslint.close();
    }
  });

  test('fix:true returns per-result output + fixable count', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
      fix: true,
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];', {
        filePath: 'gap-rslint-fix.ts',
      });
      const r = results[0];
      expect(r.output).toBe('let a: string[] = [];');
      expect(r.fixableErrorCount).toBe(1);
      // Single-pass (design §8): messages still report the pre-fix diagnostics
      // — the fixable error is in `output` AND still listed, unlike ESLint's
      // post-fix re-lint. Pinned so this reverse-of-ESLint behavior is explicit.
      expect(r.messages).toHaveLength(1);
      expect(r.errorCount).toBe(1);
    } finally {
      await rslint.close();
    }
  });

  test('a clean file yields a result with zero messages', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
    });
    try {
      const results = await rslint.lintText('let a: string[] = [];', {
        filePath: 'gap-rslint-clean.ts',
      });
      expect(results).toHaveLength(1);
      expect(results[0].messages).toHaveLength(0);
      expect(results[0].errorCount).toBe(0);
    } finally {
      await rslint.close();
    }
  });

  test('lintFiles globs files and returns one result per file', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
    });
    try {
      const results = await rslint.lintFiles('src/index.ts');
      expect(results).toHaveLength(1);
      expect(path.isAbsolute(results[0].filePath)).toBe(true);
      expect(results[0].filePath.endsWith(path.join('src', 'index.ts'))).toBe(
        true,
      );
    } finally {
      await rslint.close();
    }
  });

  test('outputFixes writes output to absolute paths, skips no-output/relative', async () => {
    const { readFile, rm, mkdtemp } = await import('node:fs/promises');
    const os = await import('node:os');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-outputfixes-'));
    const target = path.join(tmp, 'a.ts');
    const noOutput = path.join(tmp, 'b.ts');
    const base = {
      messages: [],
      errorCount: 0,
      warningCount: 0,
      fixableErrorCount: 0,
      fixableWarningCount: 0,
    };
    try {
      await Rslint.outputFixes([
        { ...base, filePath: target, output: 'let a: string[] = [];' },
        { ...base, filePath: noOutput }, // no output → skipped
        { ...base, filePath: '__rslint_relskip_probe__.ts', output: 'x' }, // relative → skipped
      ]);
      expect(await readFile(target, 'utf8')).toBe('let a: string[] = [];');
      let bWritten = true;
      try {
        await readFile(noOutput);
      } catch {
        bWritten = false;
      }
      expect(bWritten).toBe(false);
      // Relative path is skipped (isAbsolute guard): nothing written to cwd.
      let relWritten = true;
      try {
        await readFile(
          path.resolve(process.cwd(), '__rslint_relskip_probe__.ts'),
        );
      } catch {
        relWritten = false;
      }
      expect(relWritten).toBe(false);
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText converts a rule suggestion to ESLint shape', async () => {
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { eqeqeq: 'error' },
        },
      ],
    });
    try {
      const results = await rslint.lintText(
        'const a = 1;\nconst b = 2;\nconst c = a == b;',
        { filePath: 'gap-eqeqeq.ts' },
      );
      const m = results[0].messages.find((x) => x.ruleId === 'eqeqeq');
      expect(m).toBeDefined();
      expect(m.suggestions).toHaveLength(1);
      const s = m.suggestions[0];
      expect(s.messageId).toBe('replaceOperator');
      expect(s.desc).toBe("Use '===' instead of '=='.");
      expect(s.fix.range).toHaveLength(2);
      expect(s.fix.text).toBe('===');
    } finally {
      await rslint.close();
    }
  });

  test('lintText without filePath uses the <text> sentinel, outputFixes skips it', async () => {
    const { readdir } = await import('node:fs/promises');
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: arrayTypeConfig,
      fix: true,
    });
    try {
      const results = await rslint.lintText('let a: Array<string> = [];'); // no filePath
      expect(results[0].filePath).toBe('<text>');
      expect(path.isAbsolute(results[0].filePath)).toBe(false);
      const before = (await readdir(fixturesDir)).length;
      await Rslint.outputFixes(results); // must skip the non-absolute <text>
      const after = await readdir(fixturesDir);
      expect(after.length).toBe(before); // no phantom __text__.ts written
      expect(after.includes('__text__.ts')).toBe(false);
    } finally {
      await rslint.close();
    }
  });

  test('auto-discovers config and appends overrideConfig', async () => {
    const { mkdtemp, writeFile, rm } = await import('node:fs/promises');
    const os = await import('node:os');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-discover-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { eqeqeq: 'error' } }];\n",
      );
      // No overrideConfigFile → auto-discover; overrideConfig appends no-var.
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfig: [{ files: ['**/*.ts'], rules: { 'no-var': 'error' } }],
      });
      try {
        const results = await rslint.lintText('var c = 1 == 2;', {
          filePath: 'a.ts',
        });
        const ruleIds = new Set(results[0].messages.map((m) => m.ruleId));
        expect(ruleIds.has('eqeqeq')).toBe(true); // from discovered config
        expect(ruleIds.has('no-var')).toBe(true); // from appended overrideConfig
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('reuses one Rslint instance with request-scoped override predicates', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-predicate-inline-'),
    );
    const calls = [];
    const selectKeep = (filePath) => {
      calls.push(path.basename(filePath));
      return path.basename(filePath) === 'keep.ts';
    };
    try {
      await writeFile(path.join(tmp, 'keep.ts'), 'debugger;\n');
      await writeFile(path.join(tmp, 'drop.ts'), 'debugger;\n');
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          { files: [selectKeep], rules: { 'no-debugger': 'error' } },
        ],
      });
      try {
        for (let request = 0; request < 2; request++) {
          const results = await rslint.lintFiles('*.ts');
          const byName = new Map(
            results.map((result) => [path.basename(result.filePath), result]),
          );
          expect(
            byName.get('keep.ts').messages.map((message) => message.ruleId),
          ).toEqual(['no-debugger']);
          expect(byName.get('drop.ts').messages).toEqual([]);
        }
        expect(calls.sort()).toEqual([
          'drop.ts',
          'drop.ts',
          'keep.ts',
          'keep.ts',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('auto-discovered function selectors execute once per visited file', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-predicate-auto-'));
    const marker = path.join(tmp, 'predicate-calls.log');
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        [
          "import { appendFileSync } from 'node:fs';",
          "const marker = new URL('./predicate-calls.log', import.meta.url);",
          'export default [{',
          '  files: [(filePath) => {',
          '    appendFileSync(marker, `${filePath}\\n`);',
          "    return filePath.endsWith('keep.ts');",
          '  }],',
          "  rules: { 'no-debugger': 'error' },",
          '}];',
          '',
        ].join('\n'),
      );
      await writeFile(path.join(tmp, 'keep.ts'), 'debugger;\n');
      await writeFile(path.join(tmp, 'drop.ts'), 'debugger;\n');

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('*.ts');
        const byName = new Map(
          results.map((result) => [path.basename(result.filePath), result]),
        );
        expect(
          byName.get('keep.ts').messages.map((message) => message.ruleId),
        ).toEqual(['no-debugger']);
        expect(byName.get('drop.ts').messages).toEqual([]);
      } finally {
        await rslint.close();
      }

      const evaluated = (await readFile(marker, 'utf8'))
        .trim()
        .split('\n')
        .map((filePath) => path.basename(filePath))
        .sort();
      // ESLint evaluates ConfigArray before its input glob. The config module
      // is therefore visited too, but the selected lint target is never
      // rematched by native/plugin/fix consumers.
      expect(evaluated).toEqual(['drop.ts', 'keep.ts', 'rslint.config.mjs']);
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('explicit config predicates are selected once and reused by fix passes', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-predicate-explicit-'),
    );
    const configDir = path.join(tmp, 'config');
    const marker = path.join(configDir, 'predicate-calls.log');
    const configPath = path.join(configDir, 'custom.config.mjs');
    try {
      await mkdir(configDir, { recursive: true });
      await writeFile(
        configPath,
        [
          "import { appendFileSync } from 'node:fs';",
          "const marker = new URL('./predicate-calls.log', import.meta.url);",
          'const record = (kind, filePath) => {',
          '  appendFileSync(marker, `${kind}:${filePath}\\n`);',
          '  return filePath;',
          '};',
          'export default [{',
          "  files: [(filePath) => record('files', filePath).endsWith('.ts')],",
          "  ignores: [(filePath) => record('ignores', filePath).endsWith('drop.ts')],",
          "  rules: { 'no-regex-spaces': 'error' },",
          '}];',
          '',
        ].join('\n'),
      );
      await writeFile(path.join(tmp, 'keep.ts'), 'const keep = /a  b/;\n');
      await writeFile(path.join(tmp, 'drop.ts'), 'const drop = /a  b/;\n');

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: configPath,
        fix: true,
      });
      try {
        const results = await rslint.lintFiles('*.ts');
        const byName = new Map(
          results.map((result) => [path.basename(result.filePath), result]),
        );
        expect(byName.get('keep.ts').output).toBe('const keep = /a {2}b/;\n');
        expect(byName.get('drop.ts').messages).toEqual([]);
      } finally {
        await rslint.close();
      }

      const calls = (await readFile(marker, 'utf8')).trim().split('\n');
      expect(new Set(calls).size).toBe(calls.length);
      expect(calls.some((call) => call.endsWith('keep.ts'))).toBe(true);
      expect(calls.some((call) => call.endsWith('drop.ts'))).toBe(true);
      expect(calls.some((call) => call.startsWith('files:'))).toBe(true);
      expect(calls.some((call) => call.startsWith('ignores:'))).toBe(true);
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('fail-fast sibling discovery deletes a still-loading config transaction', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-config-fail-fast-'),
    );
    const hangingDir = path.join(tmp, 'hanging');
    const failingDir = path.join(tmp, 'failing');
    const loadSpy = rs.spyOn(ConfigModuleHost.prototype, 'loadConfigs');
    const deleteSpy = rs.spyOn(ConfigModuleHost.prototype, 'deleteSession');
    const rslint = new Rslint({ cwd: tmp });
    try {
      await mkdir(hangingDir, { recursive: true });
      await mkdir(failingDir, { recursive: true });
      await writeFile(
        path.join(hangingDir, 'rslint.config.mjs'),
        'await new Promise(() => {});\nexport default [];\n',
      );
      await writeFile(
        path.join(failingDir, 'rslint.config.mjs'),
        "await new Promise((resolve) => setTimeout(resolve, 100));\nthrow new Error('broken sibling');\n",
      );
      await writeFile(path.join(hangingDir, 'index.ts'), 'export {};\n');
      await writeFile(path.join(failingDir, 'index.ts'), 'export {};\n');

      let timeout;
      const error = await Promise.race([
        rslint.lintFiles(['hanging/index.ts', 'failing/*.ts']).then(
          () => new Error('lint unexpectedly succeeded'),
          (reason) => reason,
        ),
        new Promise((resolve) => {
          timeout = setTimeout(
            () => resolve(new Error('lintFiles timed out')),
            5_000,
          );
        }),
      ]);
      clearTimeout(timeout);
      expect(error.message).toContain('broken sibling');

      const hangingLoad = loadSpy.mock.calls.find(([request]) =>
        request.candidates.some((candidate) =>
          candidate.configPath.includes(`${path.sep}hanging${path.sep}`),
        ),
      );
      expect(hangingLoad).toBeDefined();
      const transactionId = hangingLoad[0].transactionId;
      expect(
        deleteSpy.mock.calls.some(([deletedId]) => deletedId === transactionId),
      ).toBe(true);
    } finally {
      await rslint.close();
      loadSpy.mockRestore();
      deleteSpy.mockRestore();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('no config candidate is fatal unless config discovery is disabled', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-no-candidate-'));
    const configured = new Rslint({
      cwd: tmp,
      overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
    });
    const syntaxOnly = new Rslint({ cwd: tmp });
    try {
      await expect(
        configured.lintText('debugger;\n', { filePath: 'configured.ts' }),
      ).rejects.toThrow(/Could not find config file/);
      await expect(
        syntaxOnly.lintText('const = ;\n', { filePath: 'syntax.ts' }),
      ).rejects.toThrow(/Could not find config file/);
    } finally {
      await Promise.all([configured.close(), syntaxOnly.close()]);
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('overrideConfigFile outside cwd still resolves files from cwd', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-override-cwd-'));
    const projectDir = path.join(tmp, 'project');
    const configDir = path.join(tmp, 'configs');
    try {
      await mkdir(path.join(projectDir, 'src'), { recursive: true });
      await mkdir(configDir, { recursive: true });
      const configFile = path.join(configDir, 'custom.config.mjs');
      await writeFile(
        configFile,
        "export default [{ files: ['src/**/*.ts'], rules: { 'no-console': 'error' } }];\n",
      );
      await writeFile(
        path.join(projectDir, 'src', 'index.ts'),
        'console.log("project");\n',
      );

      const rslint = new Rslint({
        cwd: projectDir,
        overrideConfigFile: configFile,
      });
      try {
        const results = await rslint.lintFiles('src/index.ts');
        expect(results).toHaveLength(1);
        expect(results[0].messages.map((m) => m.ruleId)).toContain(
          'no-console',
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test.each(['cjs', 'cts'])(
    'overrideConfigFile loads an explicitly selected .%s config',
    async (extension) => {
      const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-override-ext-'));
      const configPath = path.join(tmp, `custom.config.${extension}`);
      const config =
        extension === 'cts'
          ? `const config: Array<Record<string, unknown>> = [{
  files: ['**/*.ts'],
  rules: { 'no-debugger': 'error' },
}];
module.exports = config;`
          : `module.exports = [{
  files: ['**/*.ts'],
  rules: { 'no-debugger': 'error' },
}];`;
      try {
        await writeFile(configPath, config);
        await writeFile(path.join(tmp, 'test.ts'), 'debugger;\n');
        const rslint = new Rslint({
          cwd: tmp,
          overrideConfigFile: configPath,
        });
        try {
          const results = await rslint.lintFiles('test.ts');
          expect(results).toHaveLength(1);
          expect(
            results[0].messages.map((message) => message.ruleId),
          ).toContain('no-debugger');
        } finally {
          await rslint.close();
        }
      } finally {
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test('lintFiles returns one result per matched file, routed correctly', async () => {
    const { mkdtemp, writeFile, mkdir, rm } = await import('node:fs/promises');
    const os = await import('node:os');
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-multifile-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { 'no-var': 'error' } }];\n",
      );
      await mkdir(path.join(tmp, 'sub'), { recursive: true });
      await writeFile(path.join(tmp, 'dirty.ts'), 'var x = 1;\n'); // 1 no-var
      await writeFile(path.join(tmp, 'sub', 'clean.ts'), 'const y = 1;\n'); // 0
      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        const byBase = new Map(
          results.map((r) => [path.basename(r.filePath), r]),
        );
        expect(byBase.size).toBe(2);
        expect(byBase.get('dirty.ts').messages).toHaveLength(1);
        expect(byBase.get('dirty.ts').messages[0].ruleId).toBe('no-var');
        // clean file still produces a (zero-message) result.
        expect(byBase.get('clean.ts').messages).toHaveLength(0);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles preserves ESLint leading-negation unmatched behavior', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-extglob-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await mkdir(path.join(tmp, 'included'), { recursive: true });
      await mkdir(path.join(tmp, 'vendor'), { recursive: true });
      await writeFile(path.join(tmp, 'included', 'index.ts'), 'debugger;\n');
      await writeFile(path.join(tmp, 'vendor', 'index.ts'), 'debugger;\n');

      const rslint = new Rslint({ cwd: tmp });
      try {
        await expect(rslint.lintFiles('!(vendor)/**/*.ts')).rejects.toThrow(
          /All files matched/,
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles glob does not evaluate configs outside matched target branches', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-target-ancestor-config-'),
    );
    const marker = path.join(tmp, 'src', 'deep', 'loaded.marker');
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await mkdir(path.join(tmp, 'src', 'deep'), { recursive: true });
      await writeFile(path.join(tmp, 'src', 'index.ts'), 'debugger;\n');
      await writeFile(
        path.join(tmp, 'src', 'deep', 'rslint.config.mjs'),
        [
          "import { writeFileSync } from 'node:fs';",
          "writeFileSync(new URL('./loaded.marker', import.meta.url), 'loaded');",
          'export default [];',
          '',
        ].join('\n'),
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('src/*.ts');
        expect(results).toHaveLength(1);
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-debugger',
        ]);
        let evaluated = true;
        try {
          await readFile(marker);
        } catch {
          evaluated = false;
        }
        expect(evaluated).toBe(false);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles routes matched files through their nearest discovered config', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-nearest-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { 'no-console': 'error' } }];\n",
      );
      await mkdir(path.join(tmp, 'packages', 'app'), { recursive: true });
      await writeFile(
        path.join(tmp, 'packages', 'app', 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(path.join(tmp, 'root.ts'), 'console.log("root");\n');
      await writeFile(
        path.join(tmp, 'packages', 'app', 'index.ts'),
        'debugger;\nconsole.log("app");\n',
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        const byRel = new Map(
          results.map((r) => [path.relative(tmp, r.filePath), r]),
        );

        expect(byRel.get('root.ts').messages.map((m) => m.ruleId)).toContain(
          'no-console',
        );
        const appRuleIds =
          byRel
            .get(path.join('packages', 'app', 'index.ts'))
            ?.messages.map((m) => m.ruleId) ?? [];
        expect(appRuleIds).toContain('no-debugger');
        expect(appRuleIds).not.toContain('no-console');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles routes community plugins through one multi-config host', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-plugin-routing-'),
    );
    const xDir = path.join(tmp, 'packages', 'x');
    const yDir = path.join(tmp, 'packages', 'y');
    const pluginXUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgX', 'plugin-x.mjs'),
    ).href;
    const pluginYUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgY', 'plugin-y.mjs'),
    ).href;
    try {
      await mkdir(xDir, { recursive: true });
      await mkdir(yDir, { recursive: true });
      await writeFile(
        path.join(xDir, 'rslint.config.mjs'),
        `import plugin from ${JSON.stringify(pluginXUrl)};\n` +
          "export default [{ plugins: { px: plugin }, rules: { 'px/no-foo': 'error' } }];\n",
      );
      await writeFile(
        path.join(yDir, 'rslint.config.mjs'),
        `import plugin from ${JSON.stringify(pluginYUrl)};\n` +
          "export default [{ plugins: { py: plugin }, rules: { 'py/no-bar': 'error' } }];\n",
      );
      const source = 'const foo = 1; const bar = 2;\n';
      await writeFile(path.join(xDir, 'index.ts'), source);
      await writeFile(path.join(yDir, 'index.ts'), source);

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        const byRelativePath = new Map(
          results.map((result) => [
            path.relative(tmp, result.filePath),
            result,
          ]),
        );
        expect(
          byRelativePath
            .get(path.join('packages', 'x', 'index.ts'))
            .messages.map((message) => message.ruleId),
        ).toEqual(['px/no-foo']);
        expect(
          byRelativePath
            .get(path.join('packages', 'y', 'index.ts'))
            .messages.map((message) => message.ruleId),
        ).toEqual(['py/no-bar']);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('overrideConfig preserves authored routing for same-prefix community plugins', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-plugin-override-routing-'),
    );
    const xDir = path.join(tmp, 'packages', 'x');
    const yDir = path.join(tmp, 'packages', 'y');
    const pluginSource = (message) => `export default {
  rules: {
    check: {
      meta: { type: 'problem', schema: [] },
      create(context) {
        return {
          Identifier(node) {
            if (node.name === 'value') context.report({ node, message: ${JSON.stringify(message)} });
          },
        };
      },
    },
  },
};
`;
    try {
      await mkdir(xDir, { recursive: true });
      await mkdir(yDir, { recursive: true });
      await writeFile(path.join(xDir, 'plugin.mjs'), pluginSource('from-x'));
      await writeFile(path.join(yDir, 'plugin.mjs'), pluginSource('from-y'));
      const configSource =
        "import plugin from './plugin.mjs';\n" +
        "export default [{ plugins: { p: plugin }, rules: { 'p/check': 'error' } }];\n";
      await writeFile(path.join(xDir, 'rslint.config.mjs'), configSource);
      await writeFile(path.join(yDir, 'rslint.config.mjs'), configSource);
      await writeFile(path.join(xDir, 'index.ts'), 'const value = 1;\n');
      await writeFile(path.join(yDir, 'index.ts'), 'const value = 1;\n');

      const rslint = new Rslint({ cwd: tmp, overrideConfig: [{}] });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        const byRelativePath = new Map(
          results.map((result) => [
            path.relative(tmp, result.filePath),
            result,
          ]),
        );
        expect(
          byRelativePath
            .get(path.join('packages', 'x', 'index.ts'))
            .messages.find((message) => message.ruleId === 'p/check')?.message,
        ).toBe('from-x');
        expect(
          byRelativePath
            .get(path.join('packages', 'y', 'index.ts'))
            .messages.find((message) => message.ruleId === 'p/check')?.message,
        ).toBe('from-y');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles treats a broken nearest config as fatal', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-broken-fallback-'),
    );
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-console': 'error' } }];\n",
      );
      await mkdir(path.join(tmp, 'packages', 'app'), { recursive: true });
      await writeFile(
        path.join(tmp, 'packages', 'app', 'rslint.config.mjs'),
        'export default [;\n',
      );
      await writeFile(
        path.join(tmp, 'packages', 'app', 'index.ts'),
        'debugger;\nconsole.log("fallback");\n',
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        for (const target of ['**/*.ts', 'packages/app/index.ts']) {
          await expect(rslint.lintFiles(target)).rejects.toThrow(
            /failed to load nearest JavaScript config/,
          );
        }
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText treats a broken nearest JS config as fatal', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-text-broken-fallback-'),
    );
    const nested = path.join(tmp, 'packages', 'app');
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-console': 'error' } }];\n",
      );
      await mkdir(nested, { recursive: true });
      await writeFile(
        path.join(nested, 'rslint.config.mjs'),
        'export default [;\n',
      );
      // JSON configs are not candidates for the JS programmatic API fallback.
      await writeFile(
        path.join(nested, 'rslint.json'),
        JSON.stringify({ rules: { 'no-debugger': 'error' } }),
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        await expect(
          rslint.lintText('debugger;\nconsole.log("ancestor");\n', {
            filePath: path.join(nested, 'index.ts'),
          }),
        ).rejects.toThrow(/failed to load nearest JavaScript config/);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText throws the nearest JS config error when no ancestor loads', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-text-broken-boundary-'),
    );
    const nested = path.join(tmp, 'packages', 'app');
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        'export default [;\n',
      );
      await writeFile(
        path.join(tmp, 'rslint.json'),
        JSON.stringify({ rules: { 'no-console': 'error' } }),
      );
      await mkdir(nested, { recursive: true });
      const nestedConfig = path.join(nested, 'rslint.config.mjs');
      await writeFile(nestedConfig, 'export default [;\n');

      const rslint = new Rslint({ cwd: tmp });
      try {
        let thrown;
        try {
          await rslint.lintText('console.log("no fallback");\n', {
            filePath: path.join(nested, 'index.ts'),
          });
        } catch (error) {
          thrown = error;
        }
        expect(thrown).toBeInstanceOf(Error);
        expect(thrown.message).toContain(
          'failed to load nearest JavaScript config',
        );
        expect(thrown.message).toContain(nestedConfig);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles rejects the whole search when any nearest config is broken', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-broken-boundary-'),
    );
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        'export default [;\n',
      );
      await mkdir(path.join(tmp, 'broken'), { recursive: true });
      await writeFile(
        path.join(tmp, 'broken', 'index.ts'),
        'console.log("broken");\n',
      );
      await mkdir(path.join(tmp, 'healthy'), { recursive: true });
      await writeFile(
        path.join(tmp, 'healthy', 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(path.join(tmp, 'healthy', 'index.ts'), 'debugger;\n');

      const rslint = new Rslint({ cwd: tmp });
      try {
        await expect(rslint.lintFiles('**/*.ts')).rejects.toThrow(
          /failed to load nearest JavaScript config/,
        );
        await expect(
          rslint.lintFiles(['broken/index.ts', 'healthy/index.ts']),
        ).rejects.toThrow(/failed to load nearest JavaScript config/);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('nested configs resolve overrideConfig paths from each discovered config', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-nested-override-'),
    );
    const appSourceDir = path.join(tmp, 'packages', 'app', 'src');
    const libSourceDir = path.join(tmp, 'packages', 'lib', 'src');
    try {
      await mkdir(appSourceDir, { recursive: true });
      await mkdir(libSourceDir, { recursive: true });
      await writeFile(
        path.join(tmp, 'packages', 'app', 'rslint.config.mjs'),
        "export default [{ files: ['src/**/*.ts'], rules: { 'no-console': 'error' } }];\n",
      );
      await writeFile(
        path.join(tmp, 'packages', 'lib', 'rslint.config.mjs'),
        "export default [{ files: ['src/**/*.ts'], rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(
        path.join(tmp, 'packages', 'app', 'tsconfig.json'),
        JSON.stringify({ files: ['src/index.ts', 'src/ignored.ts'] }),
      );
      await writeFile(
        path.join(tmp, 'packages', 'lib', 'tsconfig.json'),
        JSON.stringify({ files: ['src/index.ts'] }),
      );
      const appSource = 'const values: string[] = [];\nconsole.log(values);\n';
      await writeFile(path.join(appSourceDir, 'index.ts'), appSource);
      await writeFile(path.join(appSourceDir, 'ignored.ts'), appSource);
      await writeFile(
        path.join(libSourceDir, 'index.ts'),
        'const values: string[] = [];\ndebugger;\n',
      );

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfig: [
          { ignores: ['src/ignored.ts'] },
          {
            files: ['src/**/*.ts'],
            plugins: ['@typescript-eslint'],
            languageOptions: {
              parserOptions: { project: ['./tsconfig.json'] },
            },
            rules: {
              '@typescript-eslint/array-type': [
                'error',
                { default: 'generic' },
              ],
            },
          },
        ],
      });
      try {
        const results = await rslint.lintFiles('packages/*/src/*.ts');
        expect(results).toHaveLength(2);
        const byPackage = new Map(
          results.map((result) => [
            path.relative(path.join(tmp, 'packages'), result.filePath),
            result.messages.map((message) => message.ruleId),
          ]),
        );
        const appRuleIds = byPackage.get(path.join('app', 'src', 'index.ts'));
        expect(appRuleIds).toContain('no-console');
        expect(appRuleIds).toContain('@typescript-eslint/array-type');
        expect(appRuleIds).not.toContain('no-debugger');

        const libRuleIds = byPackage.get(path.join('lib', 'src', 'index.ts'));
        expect(libRuleIds).toContain('no-debugger');
        expect(libRuleIds).toContain('@typescript-eslint/array-type');
        expect(libRuleIds).not.toContain('no-console');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles explicit file in parent-ignored subtree still uses nearest config', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-nearest-ignored-'),
    );
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        [
          'export default [',
          "  { ignores: ['packages/app/**/*'] },",
          "  { files: ['**/*.ts'], rules: { 'no-console': 'error' } },",
          '];',
          '',
        ].join('\n'),
      );
      await mkdir(path.join(tmp, 'packages', 'app'), { recursive: true });
      await writeFile(
        path.join(tmp, 'packages', 'app', 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(
        path.join(tmp, 'packages', 'app', 'index.ts'),
        'debugger;\nconsole.log("app");\n',
      );
      await writeFile(
        path.join(tmp, 'packages', 'app', 'automatic.ts'),
        'debugger;\n',
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles([
          'packages/app/*.ts',
          'packages/app/index.ts',
        ]);
        expect(
          results.map((result) => path.basename(result.filePath)).sort(),
        ).toEqual(['automatic.ts', 'index.ts']);
        for (const result of results) {
          const ruleIds = result.messages.map((message) => message.ruleId);
          expect(ruleIds).toContain('no-debugger');
          expect(ruleIds).not.toContain('no-console');
        }
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles glob does not enter a parent-ignored nested config', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-glob-parent-ignore-'),
    );
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        [
          'export default [',
          "  { ignores: ['packages/app/**'] },",
          "  { rules: { 'no-console': 'error' } },",
          '];',
          '',
        ].join('\n'),
      );
      await mkdir(path.join(tmp, 'packages', 'app'), { recursive: true });
      await writeFile(
        path.join(tmp, 'packages', 'app', 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(
        path.join(tmp, 'packages', 'app', 'index.ts'),
        'debugger;\n',
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        await expect(rslint.lintFiles('**/*.ts')).rejects.toThrow(
          /All files matched/,
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles scopes .gitignore to config ownership boundaries', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-gitignore-'));
    try {
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(path.join(tmp, '.gitignore'), 'ignored/\n');
      await mkdir(path.join(tmp, 'ignored'));
      await writeFile(
        path.join(tmp, 'ignored', 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(path.join(tmp, 'ignored', 'index.ts'), 'debugger;\n');
      await writeFile(path.join(tmp, 'visible.ts'), 'debugger;\n');

      const rslint = new Rslint({ cwd: tmp });
      try {
        const explicit = await rslint.lintFiles('ignored/index.ts');
        expect(explicit).toHaveLength(1);
        expect(path.relative(tmp, explicit[0].filePath)).toBe(
          path.join('ignored', 'index.ts'),
        );

        const results = await rslint.lintFiles('**/*.ts');
        expect(
          results.map((result) => path.relative(tmp, result.filePath)).sort(),
        ).toEqual([path.join('ignored', 'index.ts'), 'visible.ts'].sort());
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles includes dotfiles but skips default excluded directories', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-dotfiles-'));
    try {
      await mkdir(path.join(tmp, 'node_modules', 'pkg'), { recursive: true });
      await writeFile(path.join(tmp, '.hidden.ts'), 'debugger;\n');
      await writeFile(
        path.join(tmp, 'node_modules', 'pkg', 'index.ts'),
        'debugger;\n',
      );

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        expect(results).toHaveLength(1);
        expect(path.basename(results[0].filePath)).toBe('.hidden.ts');
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-debugger',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles includes a literal file symlink without following directory symlinks', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-file-symlink-'),
    );
    try {
      const target = path.join(tmp, 'target.ts');
      const link = path.join(tmp, 'link.ts');
      await writeFile(target, 'debugger;\n');
      try {
        await symlink(target, link, 'file');
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles('link.ts');
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(path.normalize(link));
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-debugger',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles preserves lexical aliases of one physical file', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-file-alias-dedupe-'),
    );
    try {
      const target = path.join(tmp, 'target.ts');
      const link = path.join(tmp, 'link.ts');
      await writeFile(target, 'debugger;\n');
      try {
        await symlink(target, link, 'file');
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles(['target.ts', 'link.ts']);
        expect(results.map((result) => result.filePath).sort()).toEqual(
          [path.normalize(target), path.normalize(link)].sort(),
        );
        for (const result of results) {
          expect(result.messages.map((message) => message.ruleId)).toEqual([
            'no-debugger',
          ]);
        }
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles keeps aliases governed by different lexical configs', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-file-alias-owner-'),
    );
    try {
      const shared = path.join(tmp, 'shared.ts');
      const ownerA = path.join(tmp, 'a');
      const ownerB = path.join(tmp, 'b');
      await mkdir(ownerA);
      await mkdir(ownerB);
      await writeFile(shared, "debugger; console.log('alias');\n");
      try {
        await symlink(shared, path.join(ownerA, 'target.ts'), 'file');
        await symlink(shared, path.join(ownerB, 'target.ts'), 'file');
      } catch {
        return;
      }
      await writeFile(
        path.join(ownerA, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(
        path.join(ownerB, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-console': 'error' } }];\n",
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles(['a/target.ts', 'b/target.ts']);
        const byPath = new Map(
          results.map((result) => [
            path.relative(tmp, result.filePath),
            result,
          ]),
        );
        expect(
          byPath
            .get(path.join('a', 'target.ts'))
            .messages.map((message) => message.ruleId),
        ).toEqual(['no-debugger']);
        expect(
          byPath
            .get(path.join('b', 'target.ts'))
            .messages.map((message) => message.ruleId),
        ).toEqual(['no-console']);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('fix:true keeps lexical alias outputs isolated by governing config', async () => {
    if (process.platform === 'win32') {
      return;
    }
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-file-alias-fix-'),
    );
    try {
      const shared = path.join(tmp, 'shared.ts');
      const ownerA = path.join(tmp, 'a');
      const ownerB = path.join(tmp, 'b');
      await mkdir(ownerA);
      await mkdir(ownerB);
      const source = "const pattern = /bar  baz/; const obj = { ['x']: 1 };\n";
      await writeFile(shared, source);
      try {
        await symlink(shared, path.join(ownerA, 'target.ts'), 'file');
        await symlink(shared, path.join(ownerB, 'target.ts'), 'file');
      } catch {
        return;
      }
      await writeFile(
        path.join(ownerA, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-regex-spaces': 'error' } }];\n",
      );
      await writeFile(
        path.join(ownerB, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-useless-computed-key': 'error' } }];\n",
      );

      const rslint = new Rslint({ cwd: tmp, fix: true });
      try {
        const results = await rslint.lintFiles(['a/target.ts', 'b/target.ts']);
        const byPath = new Map(
          results.map((result) => [
            path.relative(tmp, result.filePath),
            result,
          ]),
        );
        const resultA = byPath.get(path.join('a', 'target.ts'));
        const resultB = byPath.get(path.join('b', 'target.ts'));
        expect(resultA.messages.map((message) => message.ruleId)).toEqual([
          'no-regex-spaces',
        ]);
        expect(resultB.messages.map((message) => message.ruleId)).toEqual([
          'no-useless-computed-key',
        ]);
        expect(resultA.output).toBe(
          "const pattern = /bar {2}baz/; const obj = { ['x']: 1 };\n",
        );
        expect(resultB.output).toBe(
          "const pattern = /bar  baz/; const obj = { 'x': 1 };\n",
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles preserves a native POSIX backslash filename through compiler binding', async () => {
    if (process.platform === 'win32') {
      return;
    }
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-posix-backslash-file-'),
    );
    try {
      const target = path.join(tmp, 'a\\b.ts');
      await writeFile(target, 'debugger;\n');
      await mkdir(path.join(tmp, 'a'));
      await writeFile(path.join(tmp, 'a', 'b.ts'), 'export {};\n');
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('a\\b.ts');
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(path.normalize(target));
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-debugger',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles keeps lexical config-directory aliases independent', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-config-directory-alias-'),
    );
    try {
      const shared = path.join(tmp, 'shared');
      const ownerA = path.join(tmp, 'a');
      const ownerB = path.join(tmp, 'b');
      await mkdir(shared);
      await writeFile(path.join(shared, 'a.ts'), 'debugger;\n');
      await writeFile(path.join(shared, 'b.ts'), 'debugger;\n');
      await writeFile(
        path.join(shared, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      try {
        await symlink(
          shared,
          ownerA,
          process.platform === 'win32' ? 'junction' : 'dir',
        );
        await symlink(
          shared,
          ownerB,
          process.platform === 'win32' ? 'junction' : 'dir',
        );
      } catch {
        return;
      }

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles(['a/a.ts', 'b/b.ts']);
        expect(
          results.map((result) => path.relative(tmp, result.filePath)).sort(),
        ).toEqual([path.join('a', 'a.ts'), path.join('b', 'b.ts')]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles does not find config through a physical path fallback', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-file-alias-config-'),
    );
    try {
      const configured = path.join(tmp, 'configured');
      const entry = path.join(tmp, 'entry');
      await mkdir(configured);
      await mkdir(entry);
      const target = path.join(configured, 'target.ts');
      const link = path.join(entry, 'link.ts');
      await writeFile(target, 'debugger;\n');
      await writeFile(
        path.join(configured, 'rslint.config.mjs'),
        "export default [{ files: ['**/*.ts'], rules: { 'no-debugger': 'error' } }];\n",
      );
      try {
        await symlink(target, link, 'file');
      } catch {
        return;
      }

      const rslint = new Rslint({ cwd: tmp });
      try {
        await expect(rslint.lintFiles('entry/link.ts')).rejects.toThrow(
          /Could not find config file/,
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles keeps lexical files matching when a symlink target belongs to the Program', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-program-symlink-selector-'),
    );
    try {
      const physicalDir = path.join(tmp, 'physical');
      const physicalFile = path.join(physicalDir, 'index.ts');
      const linkFile = path.join(tmp, 'link.ts');
      await mkdir(physicalDir);
      await writeFile(physicalFile, 'console.log("value");\n');
      await writeFile(
        path.join(tmp, 'tsconfig.json'),
        JSON.stringify({ files: ['physical/index.ts'] }),
      );
      await writeFile(
        path.join(tmp, 'rslint.config.mjs'),
        `export default [{
          files: ['link.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { 'no-console': 'error' },
        }];\n`,
      );
      try {
        await symlink(physicalFile, linkFile, 'file');
      } catch {
        return;
      }

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles('link.ts');
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(path.normalize(linkFile));
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-console',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles keeps case-distinct files separate on a case-sensitive filesystem', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-case-sensitive-paths-'),
    );
    try {
      const upperDir = path.join(tmp, 'Foo');
      const lowerDir = path.join(tmp, 'foo');
      await mkdir(upperDir);
      try {
        await mkdir(lowerDir);
      } catch {
        return;
      }
      await writeFile(path.join(upperDir, 'index.ts'), 'debugger;\n');
      await writeFile(path.join(lowerDir, 'index.ts'), 'debugger;\n');

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles([
          'Foo/index.ts',
          'foo/index.ts',
        ]);
        expect(results).toHaveLength(2);
        expect(
          results.map((result) => path.relative(tmp, result.filePath)).sort(),
        ).toEqual([path.join('Foo', 'index.ts'), path.join('foo', 'index.ts')]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles treats a leading bang as a minimatch-negated search', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-literal-negative-case-'),
    );
    try {
      const upperDir = path.join(tmp, 'Foo');
      const lowerDir = path.join(tmp, 'foo');
      await mkdir(upperDir);
      try {
        await mkdir(lowerDir);
      } catch {
        return;
      }
      const target = path.join(upperDir, 'target.ts');
      const link = path.join(upperDir, 'link.ts');
      await writeFile(target, 'debugger;\n');
      try {
        await symlink(target, link, 'file');
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        await expect(
          rslint.lintFiles(['Foo/link.ts', '!foo/link.ts']),
        ).rejects.toThrow(/All files matched/);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles reports a leading-bang pattern as unmatched', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-literal-negative-'),
    );
    try {
      const target = path.join(tmp, 'target.ts');
      const link = path.join(tmp, 'link.ts');
      await writeFile(target, 'debugger;\n');
      try {
        await symlink(target, link, 'file');
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        await expect(rslint.lintFiles(['link.ts', '!link.ts'])).rejects.toThrow(
          /All files matched/,
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles preserves a valid alternate-case literal spelling', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-alternate-case-literal-'),
    );
    try {
      const actualDir = path.join(tmp, 'Project');
      const callerPath = path.join(tmp, 'project', 'a.ts');
      await mkdir(actualDir);
      await writeFile(path.join(actualDir, 'a.ts'), 'debugger;\n');
      try {
        await readFile(callerPath);
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles('project/a.ts');
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(path.normalize(callerPath));
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-debugger',
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles preserves each alternate-case literal spelling', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-alternate-case-literals-'),
    );
    try {
      const actualDir = path.join(tmp, 'Project');
      const alternateFile = path.join(tmp, 'project', 'b.ts');
      await mkdir(actualDir);
      await writeFile(path.join(actualDir, 'a.ts'), 'debugger;\n');
      await writeFile(path.join(actualDir, 'b.ts'), 'debugger;\n');
      try {
        await readFile(alternateFile);
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles([
          'Project/a.ts',
          'project/b.ts',
        ]);
        expect(results.map((result) => result.filePath).sort()).toEqual(
          [path.join(tmp, 'Project', 'a.ts'), alternateFile].sort(),
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles coalesces native case aliases of one config', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-config-case-alias-'),
    );
    try {
      const actualDir = path.join(tmp, 'Project');
      const alternateFile = path.join(tmp, 'project', 'b.ts');
      await mkdir(actualDir);
      await writeFile(path.join(actualDir, 'a.ts'), 'debugger;\n');
      await writeFile(path.join(actualDir, 'b.ts'), 'debugger;\n');
      await writeFile(
        path.join(actualDir, 'rslint.config.mjs'),
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      try {
        await readFile(alternateFile);
      } catch {
        return;
      }

      const rslint = new Rslint({ cwd: tmp });
      try {
        const results = await rslint.lintFiles([
          'Project/a.ts',
          'project/b.ts',
        ]);
        expect(results).toHaveLength(2);
        expect(
          results.map((result) => path.basename(result.filePath)).sort(),
        ).toEqual(['a.ts', 'b.ts']);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles does not recurse through a directory symlink cycle', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-directory-symlink-'),
    );
    try {
      const sourceDir = path.join(tmp, 'src');
      await mkdir(sourceDir, { recursive: true });
      await writeFile(path.join(sourceDir, 'index.ts'), 'debugger;\n');
      try {
        await symlink(
          tmp,
          path.join(sourceDir, 'loop'),
          process.platform === 'win32' ? 'junction' : 'dir',
        );
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
      });
      try {
        const results = await rslint.lintFiles('**/*.ts');
        expect(results).toHaveLength(1);
        expect(path.relative(tmp, results[0].filePath)).toBe(
          path.join('src', 'index.ts'),
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles does not use a physical ancestor config for a directory symlink', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-explicit-directory-symlink-'),
    );
    try {
      const realRoot = path.join(tmp, 'real');
      const realSubdir = path.join(realRoot, 'sub');
      const linkDir = path.join(tmp, 'link');
      await mkdir(realSubdir, { recursive: true });
      await writeFile(
        path.join(realRoot, 'rslint.config.mjs'),
        "export default [{ files: ['sub/**/*.ts'], rules: { 'no-debugger': 'error' } }];\n",
      );
      await writeFile(path.join(realSubdir, 'index.ts'), 'debugger;\n');
      try {
        await symlink(
          realSubdir,
          linkDir,
          process.platform === 'win32' ? 'junction' : 'dir',
        );
      } catch {
        return;
      }

      const rslint = new Rslint({ cwd: tmp });
      try {
        for (const pattern of ['link', 'link/**/*.ts']) {
          await expect(rslint.lintFiles(pattern)).rejects.toThrow(
            /Could not find config file/,
          );
        }
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('multi-edit fix: merged message.fix applied equals Go output', async () => {
    // no-extra-bind emits TWO fix edits (drop `.bind` + the `(this)` arg);
    // mergeFixes collapses them into one span. Applying that JS-merged fix to
    // the source must equal Go's in-band output — exercises both the multi-edit
    // merge branch and JS↔Go fix agreement.
    const code = 'const f = (function () { return 1; }).bind(this);';
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { 'no-extra-bind': 'error' },
        },
      ],
      fix: true,
    });
    try {
      const results = await rslint.lintText(code, { filePath: 'gap-bind.ts' });
      const r = results[0];
      const m = r.messages.find((x) => x.ruleId === 'no-extra-bind');
      expect(m).toBeDefined();
      expect(m.fix).toBeDefined();
      expect(m.fix.range).toHaveLength(2);
      const applied =
        code.slice(0, m.fix.range[0]) + m.fix.text + code.slice(m.fix.range[1]);
      expect(applied).toBe(r.output);
    } finally {
      await rslint.close();
    }
  });

  test('cross-layer oracle: low-level lint() and Rslint.lintText agree field-by-field', async () => {
    const code = 'let a: Array<string> = [];';
    const cfg = [
      {
        files: ['**/*.ts'],
        languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
        rules: { '@typescript-eslint/array-type': 'error' },
        plugins: ['@typescript-eslint'],
      },
    ];
    const filePath = path.join(fixturesDir, 'gap-oracle.ts');
    // Low-level lint() → wire (Go) shape.
    const wire = await lint({
      config: cfg,
      configDirectory: fixturesDir,
      workingDirectory: fixturesDir,
      files: [filePath],
      fileContents: { [filePath]: code },
    });
    const d = wire.diagnostics.find(
      (x) => x.ruleName === '@typescript-eslint/array-type',
    );
    // High-level Rslint.lintText → ESLint (JS) shape.
    const rslint = new Rslint({
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: cfg,
    });
    try {
      const results = await rslint.lintText(code, {
        filePath: 'gap-oracle.ts',
      });
      const m = results[0].messages.find(
        (x) => x.ruleId === '@typescript-eslint/array-type',
      );
      expect(d).toBeDefined();
      expect(m).toBeDefined();
      // Assert the reshape MAPPING (not two hardcoded equal constants):
      expect(m.severity).toBe(d.severity === 'error' ? 2 : 1);
      expect(m.ruleId).toBe(d.ruleName);
      expect(m.column).toBe(d.range.start.column);
      expect(m.endColumn).toBe(d.range.end.column);
      expect(m.fix.range[0]).toBe(d.fixes[0].startPos);
      expect(m.fix.range[1]).toBe(d.fixes[0].endPos);
      expect(m.fix.text).toBe(d.fixes[0].text);
    } finally {
      await rslint.close();
    }
  });

  test('lintFiles keeps fix.range BOM-stripped and re-prepends the BOM only to output', async () => {
    // A disk file whose bytes start with a UTF-8 BOM: Go reads it BOM-stripped,
    // so its fix offsets and Output carry no BOM. fix.range stays BOM-stripped
    // (matching ESLint v10, whose fix.range is relative to BOM-stripped source,
    // and the message column); only `output` gets the BOM re-prepended so
    // outputFixes writes back the real on-disk file. (lintText is unaffected —
    // its overlay keeps the BOM and Go's offsets already include it.)
    const BOM = String.fromCharCode(0xfeff);
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-bom-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      await writeFile(
        path.join(tmp, 'bom.ts'),
        BOM + 'let a: Array<string> = [];\n',
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: arrayTypeConfig,
        fix: true,
      });
      try {
        const results = await rslint.lintFiles('bom.ts');
        expect(results).toHaveLength(1);
        const m = results[0].messages[0];
        // Identical to the no-BOM lintText case (line 47): BOM-stripped [7, 20].
        expect(m.fix.range).toEqual([7, 20]);
        expect(m.fix.text).toBe('string[]');
        // Only output carries the BOM.
        expect(results[0].output).toBe(BOM + 'let a: string[] = [];\n');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles multi-edit fix on a BOM-prefixed file: merged range BOM-stripped, output keeps BOM', async () => {
    // no-extra-bind emits a multi-edit fix; on a BOM-prefixed disk file the
    // merged range must stay BOM-stripped (Go's coordinate space) so applying it
    // to the BOM-stripped body reproduces output minus its BOM. Exercises the
    // multi-edit merge branch together with the strip-then-re-prepend path.
    const BOM = String.fromCharCode(0xfeff);
    const body = 'const f = (function () { return 1; }).bind(this);\n';
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-bom-multi-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      await writeFile(path.join(tmp, 'bind.ts'), BOM + body);
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          {
            files: ['**/*.ts'],
            languageOptions: {
              parserOptions: { project: ['./tsconfig.json'] },
            },
            rules: { 'no-extra-bind': 'error' },
          },
        ],
        fix: true,
      });
      try {
        const results = await rslint.lintFiles('bind.ts');
        const m = results[0].messages.find((x) => x.ruleId === 'no-extra-bind');
        expect(m).toBeDefined();
        expect(m.fix).toBeDefined();
        // BOM-stripped range: applying it to the BOM-stripped body equals output
        // minus its re-prepended BOM.
        const applied =
          body.slice(0, m.fix.range[0]) +
          m.fix.text +
          body.slice(m.fix.range[1]);
        expect(results[0].output).toBe(BOM + applied);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles returns an ESLint warning result for an ignored literal', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-ignored-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      // Intentional syntax-error fixtures must be ignored by configuration;
      // parse failure itself is not an implicit ignore signal.
      await writeFile(path.join(tmp, 'ignored.ts'), 'const = ;\n');
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          { ignores: ['ignored.ts'] },
          {
            files: ['**/*.ts'],
            languageOptions: {
              parserOptions: { project: ['./tsconfig.json'] },
            },
            rules: { '@typescript-eslint/array-type': 'error' },
            plugins: ['@typescript-eslint'],
          },
        ],
      });
      try {
        const results = await rslint.lintFiles('ignored.ts');
        expect(results).toHaveLength(1);
        expect(results[0].warningCount).toBe(1);
        expect(results[0].messages).toEqual([
          {
            ruleId: null,
            severity: 1,
            message: 'File ignored because of a matching ignore pattern.',
          },
        ]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // Regression for the unref() fix: a script that lints and never calls close()
  // must still let the Node process exit. Without unref the resident `--api`
  // child + its stdio pipes keep the event loop alive and the process hangs.
  test('lintText without close() lets the process exit on its own (no hang)', async () => {
    const { spawn } = await import('node:child_process');
    const script = path.resolve(
      import.meta.dirname,
      'fixtures',
      'no-close-exit.mjs',
    );
    const child = spawn(process.execPath, [script], {
      cwd: path.resolve(import.meta.dirname, '..'),
      stdio: 'inherit',
    });
    const code = await new Promise((resolve) => {
      const timer = setTimeout(() => {
        child.kill('SIGKILL');
        resolve('TIMEOUT-HANG');
      }, 20000);
      child.on('exit', (c) => {
        clearTimeout(timer);
        resolve(c);
      });
    });
    // 0 = the fixture linted (it asserts exactly 1 message) AND the process
    // exited without close(); 'TIMEOUT-HANG' = unref regressed.
    expect(code).toBe(0);
  });

  test('community plugin host initializes and shuts down before lintText returns', async () => {
    const { spawn } = await import('node:child_process');
    const script = path.resolve(
      import.meta.dirname,
      'fixtures',
      'no-close-plugin-exit.mjs',
    );
    const child = spawn(process.execPath, [script], {
      cwd: path.resolve(import.meta.dirname, '..'),
      stdio: 'inherit',
    });
    const code = await new Promise((resolve) => {
      const timer = setTimeout(() => {
        child.kill('SIGKILL');
        resolve('TIMEOUT-HANG');
      }, 30000);
      child.on('exit', (exitCode) => {
        clearTimeout(timer);
        resolve(exitCode);
      });
    });
    expect(code).toBe(0);
  });

  test('community plugin host shuts down when the Go lint request fails', async () => {
    const { spawn } = await import('node:child_process');
    const script = path.resolve(
      import.meta.dirname,
      'fixtures',
      'no-close-plugin-error-exit.mjs',
    );
    const child = spawn(process.execPath, [script], {
      cwd: path.resolve(import.meta.dirname, '..'),
      stdio: 'inherit',
    });
    const code = await new Promise((resolve) => {
      const timer = setTimeout(() => {
        child.kill('SIGKILL');
        resolve('TIMEOUT-HANG');
      }, 30000);
      child.on('exit', (exitCode) => {
        clearTimeout(timer);
        resolve(exitCode);
      });
    });
    expect(code).toBe(0);
  });

  test('close shuts down an active community plugin host', async () => {
    const { spawn } = await import('node:child_process');
    const script = path.resolve(
      import.meta.dirname,
      'fixtures',
      'close-active-plugin-exit.mjs',
    );
    const child = spawn(process.execPath, [script], {
      cwd: path.resolve(import.meta.dirname, '..'),
      stdio: 'inherit',
    });
    const code = await new Promise((resolve) => {
      const timer = setTimeout(() => {
        child.kill('SIGKILL');
        resolve('TIMEOUT-HANG');
      }, 20000);
      child.on('exit', (exitCode) => {
        clearTimeout(timer);
        resolve(exitCode);
      });
    });
    expect(code).toBe(0);
  });

  // Fully in-memory (issue #1106): config object + in-memory tsconfig via
  // `virtualFiles`, type-aware rule, ZERO disk. Empty temp dir as cwd + path.join
  // keys so the tsconfig and the config's `project` resolve to one path on every OS.
  test('lintText runs type-aware rules with an in-memory tsconfig (zero disk)', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        // Explicit tsconfig `files` (not an `include` glob): a glob is expanded
        // against the overlay-over-real-FS, which would scan the cwd on disk.
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./src/probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: path.join(tmp, 'src', 'probe.ts') },
      );
      expect(results).toHaveLength(1);
      const messages = results[0].messages;
      // no-for-in-array is type-aware (it asks the TypeChecker whether `arr` is
      // an array), so it can only fire if the in-memory tsconfig built a real
      // program over the overlay — proving fully-in-memory type-aware linting.
      expect(messages).toHaveLength(1);
      expect(messages[0].ruleId).toBe('@typescript-eslint/no-for-in-array');
      expect(messages[0].severity).toBe(2);
      expect(messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // Cross-file in-memory type resolution: the linted buffer imports a symbol
  // from ANOTHER in-memory overlay file (dep.ts). no-for-in-array fires only if
  // the checker resolved `nums`'s array type across the in-memory import — i.e.
  // the overlay is one connected program, not isolated files.
  test('lintText resolves type info across in-memory dependency files', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true, moduleResolution: 'bundler' },
          files: ['./probe.ts', './dep.ts'],
        }),
        [path.join(tmp, 'dep.ts')]:
          'export const nums: number[] = [1, 2, 3];\n',
      },
    });
    try {
      const results = await rslint.lintText(
        "import { nums } from './dep';\nfor (const i in nums) {\n}\n",
        { filePath: path.join(tmp, 'probe.ts') },
      );
      expect(results).toHaveLength(1);
      const messages = results[0].messages;
      expect(messages).toHaveLength(1);
      expect(messages[0].ruleId).toBe('@typescript-eslint/no-for-in-array');
      expect(messages[0].severity).toBe(2);
      expect(messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // lintFiles + virtualFiles: glob a REAL on-disk source file but supply the
  // tsconfig in-memory via the instance overlay — the overlay is threaded into
  // lintFiles too (not just lintText), so type-aware rules run over disk files
  // with no tsconfig on disk.
  test('lintFiles runs type-aware rules with an in-memory tsconfig overlay', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    await writeFile(
      path.join(tmp, 'probe.ts'),
      'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
    );
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintFiles('probe.ts');
      expect(results).toHaveLength(1);
      const messages = results[0].messages;
      expect(messages).toHaveLength(1);
      expect(messages[0].ruleId).toBe('@typescript-eslint/no-for-in-array');
      expect(messages[0].severity).toBe(2);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // Path-style coverage: relative `virtualFiles` keys + relative filePath resolve
  // against cwd, same as absolute keys, on every OS.
  test('lintText accepts relative virtualFiles keys', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: 'probe.ts' },
      );
      expect(results[0].messages).toHaveLength(1);
      expect(results[0].messages[0].ruleId).toBe(
        '@typescript-eslint/no-for-in-array',
      );
      expect(results[0].messages[0].severity).toBe(2);
      expect(results[0].messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // A bare POSIX-absolute key with cwd '/' is re-anchored against cwd (→ `C:\…` on
  // Windows), matching where the config resolves — the cross-platform key fix.
  test('lintText re-anchors a POSIX-absolute virtualFiles key (cwd "/")', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        '/tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: '/probe.ts' },
      );
      expect(results[0].messages).toHaveLength(1);
      expect(results[0].messages[0].ruleId).toBe(
        '@typescript-eslint/no-for-in-array',
      );
      expect(results[0].messages[0].severity).toBe(2);
      expect(results[0].messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
    }
  });

  // The real fully-in-memory shape: cwd '/' (no process.cwd()) with all paths
  // relative — virtualFiles key, filePath, project, and the tsconfig `files`.
  // Relative paths anchor to cwd identically on every OS; Windows CI is the
  // real cross-platform check (a macOS host can't simulate the drive letter).
  test('lintText runs fully in-memory with cwd "/" and all relative paths', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const results = await rslint.lintText(
        'const arr = [1, 2, 3];\nfor (const i in arr) {\n}\n',
        { filePath: 'probe.ts' },
      );
      expect(results[0].messages).toHaveLength(1);
      expect(results[0].messages[0].ruleId).toBe(
        '@typescript-eslint/no-for-in-array',
      );
      expect(results[0].messages[0].severity).toBe(2);
      expect(results[0].messages[0].messageId).toBe('forInViolation');
    } finally {
      await rslint.close();
    }
  });

  // ESLint's lintText returns exactly one result — for the linted buffer. An
  // in-memory dependency file that matches the config and carries its own
  // violation must NOT leak a second result (which outputFixes would then write).
  test('lintText returns a single result even if an overlay dependency file has violations', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-virtual-'));
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        [path.join(tmp, 'tsconfig.json')]: JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts', './dep.ts'],
        }),
        // dep.ts itself violates no-for-in-array.
        [path.join(tmp, 'dep.ts')]:
          'export const xs = [1, 2, 3];\nfor (const k in xs) {\n}\n',
      },
    });
    try {
      const results = await rslint.lintText(
        "import { xs } from './dep';\nvoid xs;\n",
        { filePath: path.join(tmp, 'probe.ts') },
      );
      expect(results).toHaveLength(1);
      expect(results[0].filePath).toBe(path.join(tmp, 'probe.ts'));
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  // lintText reports BOM-INCLUSIVE offsets for BOM-prefixed code (known
  // limitation, one ahead of ESLint v10 which strips the BOM). Pinned without
  // hardcoding offsets: the same code with a leading BOM shifts every offset +1.
  test('lintText reports BOM-inclusive offsets for BOM-prefixed code (known limitation)', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/array-type': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const code = 'let a: Array<string> = [];\n';
      const [plain] = await rslint.lintText(code, { filePath: 'probe.ts' });
      const [bom] = await rslint.lintText('\uFEFF' + code, {
        filePath: 'probe.ts',
      });
      expect(plain.messages).toHaveLength(1);
      expect(bom.messages).toHaveLength(1);
      const pm = plain.messages[0];
      const bm = bom.messages[0];
      expect(pm.ruleId).toBe('@typescript-eslint/array-type');
      // The BOM shifts every offset by exactly one UTF-16 unit.
      expect(bm.column).toBe(pm.column + 1);
      expect(bm.fix.range[0]).toBe(pm.fix.range[0] + 1);
      expect(bm.fix.range[1]).toBe(pm.fix.range[1] + 1);
    } finally {
      await rslint.close();
    }
  });

  // errorCount/warningCount/fixable* are bucketed by severity; a file mixing an
  // error and a warning must split them, not collapse into one bucket.
  test('lintText splits error/warning counts for a file mixing severities', async () => {
    const rslint = new Rslint({
      cwd: '/',
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          plugins: ['@typescript-eslint'],
          rules: {
            '@typescript-eslint/array-type': 'error', // fixable error
            '@typescript-eslint/no-for-in-array': 'warn', // non-fixable warning
          },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['./probe.ts'],
        }),
      },
    });
    try {
      const [r] = await rslint.lintText(
        'const a: Array<number> = [1];\nfor (const k in a) {\n}\n',
        { filePath: 'probe.ts' },
      );
      expect(r.errorCount).toBe(1);
      expect(r.warningCount).toBe(1);
      expect(r.fixableErrorCount).toBe(1); // array-type is fixable
      expect(r.fixableWarningCount).toBe(0); // no-for-in-array has no fix
    } finally {
      await rslint.close();
    }
  });
});
