import { Rslint, ts } from '@rslint/core';
import { lint } from '@rslint/core/internal';
import { describe, test, expect } from 'rstack/test';
import { spawn } from 'node:child_process';
import path from 'node:path';
import os from 'node:os';
import { pathToFileURL } from 'node:url';
import { convertPathToPattern } from 'tinyglobby';
import {
  writeFile,
  rm,
  mkdtemp,
  mkdir,
  readFile,
  cp,
  symlink,
} from 'node:fs/promises';

const serviceConfig = {
  ...ts.configs.base,
  languageOptions: { parserOptions: { projectService: true } },
};

const fixturesDir = path.resolve(import.meta.dirname, '../fixtures');
const eslintPluginFixturesDir = path.resolve(
  import.meta.dirname,
  'eslint-plugin/fixtures',
);
// These fixtures prove that the public API lets Node exit. Success is the
// validated close event; this only releases a process still alive after 30m.
const EXIT_FIXTURE_DEAD_PROCESS_WATCHDOG_MS = 30 * 60_000;
const EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS = 35 * 60_000;
const EXIT_FIXTURE_OUTPUT_LIMIT_BYTES = 1024 * 1024;

async function runExitFixture(scriptName, successMarker) {
  const script = path.resolve(import.meta.dirname, 'fixtures', scriptName);
  const child = spawn(process.execPath, [script], {
    cwd: path.resolve(import.meta.dirname, '..'),
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  const stdoutChunks = [];
  const stderrChunks = [];
  let spawnError;
  let timedOut = false;
  let outputBytes = 0;
  let outputOverflow = false;

  child.stdout.on('data', (chunk) => {
    outputBytes += chunk.length;
    if (outputBytes > EXIT_FIXTURE_OUTPUT_LIMIT_BYTES) {
      outputOverflow = true;
      child.kill('SIGKILL');
      return;
    }
    stdoutChunks.push(chunk);
  });
  child.stderr.on('data', (chunk) => {
    outputBytes += chunk.length;
    if (outputBytes > EXIT_FIXTURE_OUTPUT_LIMIT_BYTES) {
      outputOverflow = true;
      child.kill('SIGKILL');
      return;
    }
    stderrChunks.push(chunk);
  });
  child.on('error', (error) => {
    spawnError = error;
  });

  const timer = setTimeout(() => {
    timedOut = true;
    child.kill('SIGKILL');
  }, EXIT_FIXTURE_DEAD_PROCESS_WATCHDOG_MS);
  timer.unref();

  const { code, signal } = await new Promise((resolve) => {
    child.on('close', (code, signal) => {
      resolve({ code, signal });
    });
  });
  clearTimeout(timer);
  const stdout = Buffer.concat(stdoutChunks).toString('utf8');
  const stderr = Buffer.concat(stderrChunks).toString('utf8');

  const details =
    `fixture=${scriptName}, code=${String(code)}, signal=${String(signal)}, ` +
    `timedOut=${String(timedOut)}, outputOverflow=${String(outputOverflow)}, ` +
    `spawnError=${String(spawnError)}\n` +
    `stdout:\n${stdout}\nstderr:\n${stderr}`;
  expect(timedOut, details).toBe(false);
  expect(outputOverflow, details).toBe(false);
  expect(spawnError, details).toBeUndefined();
  expect(signal, details).toBeNull();
  expect(code, details).toBe(0);
  expect(stdout.trim(), details).toBe(successMarker);
  expect(stderr.trim(), details).toBe('');
  expect(
    `${stdout}\n${stderr}`,
    `fixture emitted an abnormal worker marker\n${details}`,
  ).not.toMatch(
    /CLOSE_GATE_LEAKED|task_timeout|worker_crashed|timed out after|EXPECTED-NATIVE-ABORT/,
  );
}

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
      expect(result.messages.map((message) => message.ruleId)).toEqual([
        'local/program-listener',
      ]);
      expect(result.fixableErrorCount).toBe(0);
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

  test('lintText resolves config from the canonical ancestor of a virtual file', async () => {
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
        const [result] = await rslint.lintText(
          'let values: Array<string> = [];\n',
          { filePath: virtualFile },
        );
        expect(result.filePath).toBe(path.normalize(virtualFile));
        expect(result.messages.map((message) => message.ruleId)).toEqual([
          '@typescript-eslint/array-type',
        ]);
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

  test('fix:true returns output with final post-fix diagnostics', async () => {
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
      expect(r.fixableErrorCount).toBe(0);
      expect(r.messages).toHaveLength(0);
      expect(r.errorCount).toBe(0);
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

  test('no config candidate uses overrideConfig or syntax-only linting', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-no-candidate-'));
    const configured = new Rslint({
      cwd: tmp,
      overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
    });
    const syntaxOnly = new Rslint({ cwd: tmp });
    try {
      const configuredResults = await configured.lintText('debugger;\n', {
        filePath: 'configured.ts',
      });
      expect(
        configuredResults[0].messages.map((message) => message.ruleId),
      ).toContain('no-debugger');

      const syntaxResults = await syntaxOnly.lintText('const = ;\n', {
        filePath: 'syntax.ts',
      });
      expect(syntaxResults).toHaveLength(1);
      expect(syntaxResults[0].messages).toHaveLength(1);
      expect(syntaxResults[0].messages[0].ruleId).toBe('TypeScript(TS1134)');
    } finally {
      await Promise.all([configured.close(), syntaxOnly.close()]);
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('overrideConfigFile outside cwd resolves files from its config directory', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-override-cwd-'));
    const projectDir = path.join(tmp, 'project');
    const configDir = path.join(tmp, 'configs');
    try {
      await mkdir(path.join(projectDir, 'src'), { recursive: true });
      await mkdir(configDir, { recursive: true });
      const configFile = path.join(configDir, 'custom.config.mjs');
      await writeFile(
        configFile,
        "export default [{ files: ['../project/src/**/*.ts'], rules: { 'no-console': 'error' } }];\n",
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

  test('overrideConfigFile keeps multiple directory Git scopes independent', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-multi-root-'));
    const invocationDir = path.join(tmp, 'invocation');
    const configDir = path.join(tmp, 'config');
    const firstDir = path.join(tmp, 'first');
    const secondDir = path.join(tmp, 'second');
    try {
      await Promise.all(
        [invocationDir, configDir, firstDir, secondDir].map((directory) =>
          mkdir(directory, { recursive: true }),
        ),
      );
      const configFile = path.join(configDir, 'custom.config.mjs');
      await writeFile(
        configFile,
        "export default [{ rules: { 'no-debugger': 'error' } }];\n",
      );
      await Promise.all([
        writeFile(path.join(firstDir, '.gitignore'), 'ignored.ts\n'),
        writeFile(path.join(firstDir, 'ignored.ts'), 'debugger;\n'),
        writeFile(path.join(firstDir, 'visible.ts'), 'debugger;\n'),
        writeFile(path.join(secondDir, '.gitignore'), 'hidden.ts\n'),
        writeFile(path.join(secondDir, 'ignored.ts'), 'debugger;\n'),
        writeFile(path.join(secondDir, 'hidden.ts'), 'debugger;\n'),
      ]);

      const rslint = new Rslint({
        cwd: invocationDir,
        overrideConfigFile: configFile,
      });
      try {
        const results = await rslint.lintFiles([
          `${convertPathToPattern(firstDir)}/**/*.ts`,
          `${convertPathToPattern(secondDir)}/**/*.ts`,
        ]);
        expect(results.map((result) => result.filePath)).toEqual([
          path.join(firstDir, 'visible.ts'),
          path.join(secondDir, 'ignored.ts'),
        ]);
        for (const result of results) {
          expect(result.messages.map((message) => message.ruleId)).toContain(
            'no-debugger',
          );
        }
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

  test('overrideConfigFile rejects a legacy JSON config before loading it', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-json-config-'));
    const configPath = path.join(tmp, 'rslint.jsonc');
    await writeFile(configPath, '{ intentionally malformed legacy config');
    const rslint = new Rslint({ cwd: tmp, overrideConfigFile: configPath });
    try {
      await expect(
        rslint.lintText('debugger;\n', { filePath: 'test.ts' }),
      ).rejects.toThrow(/JS\/TS config module/);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

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

  test('lintFiles discovers config for a positive extglob', async () => {
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
        const results = await rslint.lintFiles('!(vendor)/**/*.ts');
        expect(results).toHaveLength(1);
        expect(path.relative(tmp, results[0].filePath)).toBe(
          path.join('included', 'index.ts'),
        );
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

  test('repeated API calls fresh-load changed config and plugin topology', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-fresh-config-'),
    );
    const configPath = path.join(tmp, 'rslint.config.mjs');
    const pluginXUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgX', 'plugin-x.mjs'),
    ).href;
    const pluginYUrl = pathToFileURL(
      path.join(eslintPluginFixturesDir, 'cfgY', 'plugin-y.mjs'),
    ).href;
    const writeConfig = async (pluginUrl, ruleName) => {
      await writeFile(
        configPath,
        `import plugin from ${JSON.stringify(pluginUrl)};\n` +
          `export default [{ plugins: { p: plugin }, rules: { ${JSON.stringify(`p/${ruleName}`)}: 'error' } }];\n`,
      );
    };
    try {
      await writeConfig(pluginXUrl, 'no-foo');
      const rslint = new Rslint({ cwd: tmp });
      try {
        const source = 'const foo = 1; const bar = 2;\n';
        const first = await rslint.lintText(source, { filePath: 'index.ts' });
        expect(first[0].messages.map((message) => message.ruleId)).toEqual([
          'p/no-foo',
        ]);

        await writeConfig(pluginYUrl, 'no-bar');
        const second = await rslint.lintText(source, {
          filePath: 'index.ts',
        });
        expect(second[0].messages.map((message) => message.ruleId)).toEqual([
          'p/no-bar',
        ]);
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

  test('lintFiles falls back from a broken nested config to the nearest loaded ancestor', async () => {
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
          const results = await rslint.lintFiles(target);
          expect(results).toHaveLength(1);
          const ruleIds = results[0].messages.map((message) => message.ruleId);
          expect(ruleIds).toContain('no-console');
          expect(ruleIds).not.toContain('no-debugger');
        }
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintText falls back from a broken nearest JS config to a loaded ancestor', async () => {
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
        const [result] = await rslint.lintText(
          'debugger;\nconsole.log("ancestor");\n',
          { filePath: path.join(nested, 'index.ts') },
        );
        const ruleIds = result.messages.map((message) => message.ruleId);
        expect(ruleIds).toContain('no-console');
        expect(ruleIds).not.toContain('no-debugger');
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
          'all discovered JavaScript configs failed to load',
        );
        expect(thrown.message).toContain(nestedConfig);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles skips a broken subtree when another selected config loads', async () => {
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
        const results = await rslint.lintFiles('**/*.ts');
        expect(results).toHaveLength(1);
        expect(path.relative(tmp, results[0].filePath)).toBe(
          path.join('healthy', 'index.ts'),
        );
        expect(results[0].messages.map((message) => message.ruleId)).toEqual([
          'no-debugger',
        ]);

        const explicitResults = await rslint.lintFiles([
          'broken/index.ts',
          'healthy/index.ts',
        ]);
        expect(explicitResults).toHaveLength(1);
        expect(path.relative(tmp, explicitResults[0].filePath)).toBe(
          path.join('healthy', 'index.ts'),
        );
        expect(
          explicitResults[0].messages.map((message) => message.ruleId),
        ).toEqual(['no-debugger']);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('nested configs preserve the cwd base of inline overrideConfig paths', async () => {
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
          { ignores: ['packages/app/src/ignored.ts'] },
          {
            files: ['packages/*/src/**/*.ts'],
            plugins: ['@typescript-eslint'],
            languageOptions: {
              parserOptions: { project: ['./packages/*/tsconfig.json'] },
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
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(
          path.join(tmp, 'packages', 'app', 'index.ts'),
        );
        const ruleIds = results[0].messages.map((m) => m.ruleId);
        expect(ruleIds).toContain('no-debugger');
        expect(ruleIds).not.toContain('no-console');
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
        expect(await rslint.lintFiles('**/*.ts')).toEqual([]);
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles lets a literal target bypass Git-hidden config discovery but prunes a glob target', async () => {
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

        const rootedGlob = await rslint.lintFiles('ignored/**/*.ts');
        expect(rootedGlob).toHaveLength(1);
        expect(path.relative(tmp, rootedGlob[0].filePath)).toBe(
          path.join('ignored', 'index.ts'),
        );

        const results = await rslint.lintFiles('**/*.ts');
        expect(
          results.map((result) => path.relative(tmp, result.filePath)).sort(),
        ).toEqual(['visible.ts']);
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

  test('lintFiles coalesces aliases of one physical file under the same config', async () => {
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

  test('lintFiles rejects one physical file governed by different configs', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-api-file-alias-owner-'),
    );
    try {
      const shared = path.join(tmp, 'shared.ts');
      const ownerA = path.join(tmp, 'a');
      const ownerB = path.join(tmp, 'b');
      await mkdir(ownerA);
      await mkdir(ownerB);
      await writeFile(shared, 'debugger;\n');
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
        await expect(
          rslint.lintFiles(['a/target.ts', 'b/target.ts']),
        ).rejects.toThrow(
          /resolve to the same file.*governed by different configs/,
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles rejects aliases of one physical config directory', async () => {
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
        await expect(rslint.lintFiles(['a/a.ts', 'b/b.ts'])).rejects.toThrow(
          /Config directories .* resolve to the same filesystem location/,
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles finds a config through the unique physical path fallback', async () => {
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
        const results = await rslint.lintFiles('entry/link.ts');
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

  test('lintFiles applies negative patterns case-sensitively to a literal file symlink', async () => {
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
        const results = await rslint.lintFiles(['Foo/link.ts', '!foo/link.ts']);
        expect(results).toHaveLength(1);
        expect(results[0].filePath).toBe(path.normalize(link));
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles excludes a literal file symlink matched by a negative pattern', async () => {
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
        expect(await rslint.lintFiles(['link.ts', '!link.ts'])).toHaveLength(0);
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

  test('lintFiles preserves inline cwd matching through a physical ancestor config', async () => {
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
      await writeFile(
        path.join(realSubdir, 'index.ts'),
        "debugger;\nconsole.log('value');\n",
      );
      try {
        await symlink(
          realSubdir,
          linkDir,
          process.platform === 'win32' ? 'junction' : 'dir',
        );
      } catch {
        return;
      }

      const rslint = new Rslint({
        cwd: tmp,
        overrideConfig: [
          {
            files: ['link/**/*.ts'],
            rules: { 'no-console': 'error' },
          },
        ],
      });
      try {
        for (const pattern of ['link', 'link/**/*.ts']) {
          const results = await rslint.lintFiles(pattern);
          expect(results).toHaveLength(1);
          expect(results[0].filePath).toBe(path.join(linkDir, 'index.ts'));
          expect(results[0].messages.map((message) => message.ruleId)).toEqual([
            'no-debugger',
            'no-console',
          ]);
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
    const options = {
      cwd: fixturesDir,
      overrideConfigFile: true,
      overrideConfig: [
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: { 'no-extra-bind': 'error' },
        },
      ],
    };
    const inspector = new Rslint(options);
    const fixer = new Rslint({ ...options, fix: true });
    try {
      const [inspected] = await inspector.lintText(code, {
        filePath: 'gap-bind.ts',
      });
      const [fixed] = await fixer.lintText(code, { filePath: 'gap-bind.ts' });
      const m = inspected.messages.find((x) => x.ruleId === 'no-extra-bind');
      expect(m).toBeDefined();
      expect(m.fix).toBeDefined();
      expect(m.fix.range).toHaveLength(2);
      const applied =
        code.slice(0, m.fix.range[0]) + m.fix.text + code.slice(m.fix.range[1]);
      expect(applied).toBe(fixed.output);
      expect(fixed.messages).toEqual([]);
    } finally {
      await Promise.all([inspector.close(), fixer.close()]);
    }
  });

  test('low-level --api accepts a JSON config payload', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-api-config-'));
    const filePath = path.join(tmp, 'api-json-config.ts');
    // A malformed legacy file beside the request must be irrelevant: `config`
    // is a JSON wire payload, not a request to discover or read a JSON file.
    await writeFile(path.join(tmp, 'rslint.json'), '{ not valid JSON');
    try {
      const response = await lint({
        config: [{ rules: { 'no-debugger': 'error' } }],
        configDirectory: tmp,
        workingDirectory: tmp,
        files: [filePath],
        fileContents: { [filePath]: 'debugger;\n' },
      });

      expect(response.errorCount).toBe(1);
      expect(response.ruleCount).toBe(1);
      expect(response.diagnostics).toHaveLength(1);
      expect(response.diagnostics[0].ruleName).toBe('no-debugger');
    } finally {
      await rm(tmp, { recursive: true, force: true });
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

  test('lintFiles keeps the BOM in fixed output', async () => {
    // `output` is the whole file, so it keeps a mark the fix did not remove.
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
        expect(results[0].messages).toEqual([]);
        expect(results[0].output).toBe(BOM + 'let a: string[] = [];\n');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles multi-edit fix on a BOM-prefixed file keeps BOM', async () => {
    // no-extra-bind emits a multi-edit fix; final output keeps the mark.
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
        expect(results[0].messages).toEqual([]);
        expect(results[0].output).toBe(
          BOM + 'const f = (function () { return 1; });\n',
        );
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('unicode-bom fix removes the mark from a disk file through lintFiles', async () => {
    // The mark reaches the rule as ctx.HasBOM rather than as text, and the fix
    // is ESLint's [-1, 0] — a range one position ahead of the source. Output is
    // the whole file, so this is where that range takes effect.
    const BOM = String.fromCharCode(0xfeff);
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-bom-rule-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      await writeFile(path.join(tmp, 'bom.ts'), BOM + 'let a = 1;\n');
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          { files: ['**/*.ts'], rules: { 'unicode-bom': 'error' } },
        ],
        fix: true,
      });
      try {
        const results = await rslint.lintFiles('bom.ts');
        expect(results[0].messages).toEqual([]);
        expect(results[0].output).toBe('let a = 1;\n');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('unicode-bom fix removes a mark the caller passed to lintText', async () => {
    // Caller-supplied source is stripped of its mark before parsing exactly as
    // a file read off disk is, so both routes reach the rule the same way and
    // produce the same output.
    const BOM = String.fromCharCode(0xfeff);
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-bom-text-'));
    try {
      await writeFile(path.join(tmp, 'tsconfig.json'), '{}');
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          { files: ['**/*.ts'], rules: { 'unicode-bom': 'error' } },
        ],
        fix: true,
      });
      try {
        const results = await rslint.lintText(BOM + 'let a = 1;\n', {
          filePath: path.join(tmp, 'buffer.ts'),
        });
        expect(results[0].messages).toEqual([]);
        expect(results[0].output).toBe('let a = 1;\n');
      } finally {
        await rslint.close();
      }
    } finally {
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test('lintFiles returns no result when every match is config-ignored', async () => {
    // Regression guard for the empty-lintedFiles wire case: when all glob
    // matches are excluded by config `ignores`, Go returns an empty (non-nil)
    // lintedFiles array — NOT an omitted field — so the class yields zero
    // results instead of falling back to the glob matches and seeding phantom
    // empty results.
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
        expect(results).toEqual([]);
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
  test(
    'lintText without close() lets the process exit on its own (no hang)',
    async () => {
      await runExitFixture('no-close-exit.mjs', 'FIXTURE_OK:no-close');
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'community plugin host initializes and shuts down before lintText returns',
    async () => {
      await runExitFixture(
        'no-close-plugin-exit.mjs',
        'FIXTURE_OK:no-close-plugin',
      );
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'community plugin host shuts down when the Go lint request fails',
    async () => {
      await runExitFixture(
        'no-close-plugin-error-exit.mjs',
        'FIXTURE_OK:no-close-plugin-error',
      );
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test(
    'close shuts down an active community plugin host',
    async () => {
      await runExitFixture(
        'close-active-plugin-exit.mjs',
        'FIXTURE_OK:close-active-plugin',
      );
    },
    EXIT_FIXTURE_OUTER_DEADLOCK_SENTINEL_MS,
  );

  test.each(['explicit', 'mixed', 'service'])(
    'projectService lintFiles keeps overlapping %s programs from duplicating targets',
    async (mode) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-modes-'),
      );
      const code =
        'export const values = [1, 2];\nfor (const key in values) {}\n';
      await writeFile(
        path.join(tmp, 'tsconfig.json'),
        JSON.stringify({ files: ['a.ts', 'b.ts', 'c.ts'] }),
      );
      for (const file of ['a.ts', 'b.ts', 'c.ts']) {
        await writeFile(path.join(tmp, file), code);
      }
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          {
            files: ['**/*.ts'],
            rules: { '@typescript-eslint/no-for-in-array': 'error' },
          },
          ...['a.ts', 'b.ts'].map((file) => {
            const service =
              mode === 'service' || (mode === 'mixed' && file === 'b.ts');
            return {
              files: [file],
              languageOptions: {
                parserOptions: service
                  ? { projectService: true }
                  : { projectService: false, project: './tsconfig.json' },
              },
            };
          }),
        ],
      });
      try {
        const results = await rslint.lintFiles(['a.ts', 'b.ts']);
        expect(
          results.map((result) => path.basename(result.filePath)).sort(),
        ).toEqual(['a.ts', 'b.ts']);
        for (const result of results) {
          expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
            '@typescript-eslint/no-for-in-array',
          ]);
        }
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each([
    {
      name: 'explicit files',
      project: { files: ['gap.js', 'source.ts'] },
      owned: true,
    },
    {
      name: 'triple-slash references',
      project: { files: ['source.ts'] },
      source: '/// <reference path="./gap.js" />\n',
      // TypeScript may load this source, but rslint requires a configured root.
      owned: false,
    },
    {
      name: 'ordinary imports',
      project: { files: ['source.ts'] },
      source: "import './gap.js';\n",
      owned: false,
    },
    {
      name: 'include globs',
      project: { include: ['**/*'] },
      owned: false,
    },
  ])(
    'projectService lintFiles handles JavaScript ownership through $name',
    async ({ project, source = 'export {};\n', owned }) => {
      const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-service-js-'));
      await writeFile(
        path.join(tmp, 'tsconfig.json'),
        JSON.stringify({ compilerOptions: { allowJs: false }, ...project }),
      );
      await writeFile(path.join(tmp, 'source.ts'), source);
      await writeFile(
        path.join(tmp, 'gap.js'),
        'const values = [1, 2];\nfor (const key in values) {}\ndebugger;\n',
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          {
            files: ['**/*.js'],
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-debugger': 'error',
            },
          },
        ],
      });
      try {
        const [result] = await rslint.lintFiles(['gap.js']);
        expect(result.messages.map(({ ruleId }) => ruleId).sort()).toEqual(
          owned
            ? ['@typescript-eslint/no-for-in-array', 'no-debugger']
            : ['no-debugger'],
        );
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test('projectService lintText selects the nearest overlay project over a different config-directory project', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-service-overlay-'),
    );
    const nested = path.join(tmp, 'packages', 'app');
    await mkdir(nested, { recursive: true });
    await writeFile(
      path.join(nested, 'tsconfig.json'),
      JSON.stringify({
        compilerOptions: { paths: { values: ['../../scalar.ts'] } },
        files: ['probe.ts'],
      }),
    );
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        serviceConfig,
        {
          files: ['**/*.ts'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { paths: { values: ['./scalar.ts'] } },
          files: ['packages/app/probe.ts'],
        }),
        'scalar.ts': 'export const values = 1;\n',
        'packages/app/tsconfig.json': JSON.stringify({
          compilerOptions: { paths: { values: ['./array.ts'] } },
          files: ['probe.ts'],
        }),
        'packages/app/array.ts': 'export const values = [1, 2, 3];\n',
        'packages/app/probe.ts': 'export const diskLikeBuffer = 1;\n',
      },
    });
    try {
      const results = await rslint.lintText(
        "import { values } from 'values';\nfor (const key in values) {}\n",
        { filePath: 'packages/app/probe.ts' },
      );
      expect(results).toHaveLength(1);
      expect(results[0].filePath).toBe(path.join(nested, 'probe.ts'));
      expect(results[0].messages.map(({ ruleId }) => ruleId)).toEqual([
        '@typescript-eslint/no-for-in-array',
      ]);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test.each([false, null])(
    'projectService lintText accepts project:%s clearing inherited explicit paths',
    async (project) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-clear-'),
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          {
            languageOptions: { parserOptions: { project: './missing.json' } },
          },
          {
            files: ['**/*.ts'],
            languageOptions: { parserOptions: { project } },
            rules: { '@typescript-eslint/no-for-in-array': 'error' },
          },
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({ files: ['probe.ts'] }),
        },
      });
      try {
        const [result] = await rslint.lintText(
          'const values = [1];\nfor (const key in values) {}\n',
          { filePath: 'probe.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
          '@typescript-eslint/no-for-in-array',
        ]);
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each([{ project: './tsconfig.json' }, { project: [] }])(
    'projectService lintText rejects a project conflict inherited from separate config entries (%j)',
    async ({ project }) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-conflict-'),
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          {
            files: ['**/*.ts'],
            languageOptions: { parserOptions: { project } },
            rules: { 'no-debugger': 'error' },
          },
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({ files: ['probe.ts'] }),
        },
      });
      try {
        await expect(
          rslint.lintText('debugger;\n', { filePath: 'probe.ts' }),
        ).rejects.toThrow(/project.*projectService/);
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test('projectService lintText resolves an ancestor when the nearest config excludes the buffer', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-service-ancestor-'),
    );
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        serviceConfig,
        {
          files: ['**/*.ts'],
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({ files: ['nested/probe.ts'] }),
        'nested/tsconfig.json': JSON.stringify({ files: ['other.ts'] }),
        'nested/other.ts': 'export {};\n',
      },
    });
    try {
      const [result] = await rslint.lintText(
        'const values = [1];\nfor (const key in values) {}\n',
        { filePath: 'nested/probe.ts' },
      );
      expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
        '@typescript-eslint/no-for-in-array',
      ]);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test.each([false, null])(
    'projectService:%s lets lintText use an explicit project instead of the nearest tsconfig',
    async (projectService) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-explicit-'),
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          {
            files: ['**/*.ts'],
            languageOptions: {
              parserOptions: { projectService, project: './custom.json' },
            },
            rules: {
              'no-debugger': 'error',
              '@typescript-eslint/no-for-in-array': 'error',
            },
          },
        ],
        virtualFiles: {
          'custom.json': JSON.stringify({
            compilerOptions: { paths: { values: ['./scalar.ts'] } },
            files: ['nested/probe.ts'],
          }),
          'scalar.ts': 'export const values = 1;\n',
          'nested/tsconfig.json': JSON.stringify({
            compilerOptions: { paths: { values: ['./array.ts'] } },
            files: ['probe.ts'],
          }),
          'nested/array.ts': 'export const values = [1];\n',
        },
      });
      try {
        const [result] = await rslint.lintText(
          "import { values } from 'values';\nfor (const key in values) {}\ndebugger;\n",
          { filePath: 'nested/probe.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
          'no-debugger',
        ]);
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each([
    ...['none', 'direct', 'helper', 'function', 'rules'].flatMap((preset) =>
      ['omitted', 'root', 'cwd', 'null', 'undefined'].map((boundary) => [
        `${preset} / ${boundary}`,
        preset,
        boundary,
        !['cwd', 'undefined'].includes(boundary),
      ]),
    ),
    ...['js', 'ts', 'mts', 'cjs', 'cts'].map((extension) => [
      `direct ${extension}`,
      'direct',
      'omitted',
      true,
      extension,
    ]),
    ['object spread', 'object', 'omitted', true],
    ['JSON copy', 'json', 'omitted', true],
    ['imported config', 'shared', 'omitted', true],
    ['imported config plus direct preset', 'shared-direct', 'omitted', true],
    ['custom config filename', 'none', 'omitted', true, 'mjs', 'custom.mjs'],
  ])(
    'projectService config directory default: %s',
    async (_name, preset, boundary, typed, extension = 'mjs', customName) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-root-inference-'),
      );
      const cwd = path.join(tmp, 'pkg');
      const coreURL = pathToFileURL(
        path.resolve(import.meta.dirname, '../dist/index.js'),
      ).href;
      await mkdir(cwd);
      await writeFile(
        path.join(tmp, 'tsconfig.json'),
        JSON.stringify({ files: ['pkg/tool.ts'] }),
      );
      const source =
        'const values = [1]; for (const key in values) {} debugger;';
      await writeFile(path.join(cwd, 'tool.ts'), source);
      await writeFile(
        path.join(tmp, 'helper.mjs'),
        `import {ts} from ${JSON.stringify(coreURL)}; export const base = ts.configs.base; export const get = () => ts.configs.base;`,
      );
      await mkdir(path.join(tmp, 'shared'));
      await writeFile(
        path.join(tmp, 'shared/rslint.config.mjs'),
        `import {ts} from ${JSON.stringify(coreURL)}; export default [ts.configs.base];`,
      );
      const prefix = `import {ts} from ${JSON.stringify(coreURL)}; import {base, get} from './helper.mjs'; ${preset.startsWith('shared') ? "import shared from './shared/rslint.config.mjs';" : ''}`;
      const configName = customName ?? `rslint.config.${extension}`;
      const entry = {
        none: '{}',
        direct: 'ts.configs.base',
        helper: 'base',
        function: 'get()',
        object: '{...ts.configs.base}',
        json: 'JSON.parse(JSON.stringify(ts.configs.base))',
        shared: '...shared',
        'shared-direct': '...shared,ts.configs.base',
        rules:
          '{rules: {...ts.configs.recommended.find(entry => entry.rules).rules}}',
      }[preset];
      const options =
        boundary === 'root'
          ? `, tsconfigRootDir: ${JSON.stringify(tmp)}`
          : ['cwd', 'null', 'undefined'].includes(boundary)
            ? `, tsconfigRootDir: ${JSON.stringify(cwd)}`
            : '';
      const reset = ['null', 'undefined'].includes(boundary)
        ? `,{languageOptions:{parserOptions:{tsconfigRootDir:${boundary}}}}`
        : '';
      const configExpression = `[${entry},{plugins:['@typescript-eslint'], languageOptions:{parserOptions:{projectService:true${options}}},rules:{'no-debugger':'error','@typescript-eslint/no-for-in-array':'error'}}${reset}]`;
      await writeFile(
        path.join(tmp, configName),
        ['cjs', 'cts'].includes(extension)
          ? `module.exports = import(${JSON.stringify(coreURL)}).then(({ts}) => ${configExpression});`
          : `${prefix} export default ${configExpression};`,
      );
      const instance = new Rslint({
        cwd,
        // CJS/CTS are supported through explicit config selection.
        ...(customName || ['cjs', 'cts'].includes(extension)
          ? { overrideConfigFile: path.join(tmp, configName) }
          : {}),
      });
      try {
        for (const results of [
          await instance.lintFiles(['tool.ts']),
          await instance.lintText(source, { filePath: 'tool.ts' }),
        ]) {
          const ids = results[0].messages.map((message) => message.ruleId);
          expect(ids).toContain('no-debugger');
          expect(ids.includes('@typescript-eslint/no-for-in-array')).toBe(
            typed,
          );
        }
      } finally {
        await instance.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each([
    ['automatic parent config', false, 'auto', true],
    ['automatic nested config', true, 'auto', false],
    ['explicit parent config', true, 'parent', true],
    ['explicit nested config', true, 'nested', false],
    ['explicit external config', true, 'external', true],
    ['inline only', true, 'inline', undefined],
    ['parent config with inline basePath suffix', false, 'suffix', true],
  ])(
    'projectService config owner and cwd: %s',
    async (_name, nestedConfig, mode, ownerTyped) => {
      const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-owner-root-'));
      const pkg = path.join(tmp, 'pkg');
      const source =
        'const values = [1]; for (const key in values) {} debugger;';
      const entries = [
        {
          plugins: ['@typescript-eslint'],
          languageOptions: { parserOptions: { projectService: true } },
          rules: {
            'no-debugger': 'error',
            '@typescript-eslint/no-for-in-array': 'error',
          },
        },
      ];
      try {
        await mkdir(pkg);
        await writeFile(
          path.join(tmp, 'tsconfig.json'),
          JSON.stringify({ files: ['pkg/tool.ts'] }),
        );
        await writeFile(
          path.join(pkg, 'tsconfig.json'),
          JSON.stringify({ files: ['other.ts'] }),
        );
        await writeFile(path.join(pkg, 'other.ts'), 'export {};');
        await writeFile(path.join(pkg, 'tool.ts'), source);
        await writeFile(
          path.join(tmp, 'rslint.config.mjs'),
          `export default ${JSON.stringify(entries)};`,
        );
        if (nestedConfig) {
          await writeFile(
            path.join(pkg, 'rslint.config.mjs'),
            `export default ${JSON.stringify(entries)};`,
          );
        }
        const external = path.join(tmp, 'tooling/custom.mjs');
        if (mode === 'external') {
          await mkdir(path.dirname(external));
          await writeFile(
            external,
            `export default ${JSON.stringify(entries)};`,
          );
        }
        for (const cwd of [tmp, pkg]) {
          const instance = new Rslint({
            cwd,
            ...(mode === 'parent' || mode === 'nested'
              ? {
                  overrideConfigFile: path.join(
                    mode === 'parent' ? tmp : pkg,
                    'rslint.config.mjs',
                  ),
                }
              : {}),
            ...(mode === 'external' ? { overrideConfigFile: external } : {}),
            ...(mode === 'inline'
              ? { overrideConfigFile: true, overrideConfig: entries }
              : {}),
            ...(mode === 'suffix'
              ? {
                  overrideConfig: [
                    {
                      basePath: pkg,
                      files: [['**/*', '**/*.ts']],
                      languageOptions: {
                        parserOptions: { projectService: true },
                      },
                    },
                  ],
                }
              : {}),
          });
          try {
            const filePath = path.relative(cwd, path.join(pkg, 'tool.ts'));
            for (const results of [
              await instance.lintFiles([filePath]),
              await instance.lintText(source, { filePath }),
            ]) {
              const ids = results[0].messages.map(({ ruleId }) => ruleId);
              expect(ids).toContain('no-debugger');
              expect(ids.includes('@typescript-eslint/no-for-in-array')).toBe(
                ownerTyped ?? cwd === tmp,
              );
            }
          } finally {
            await instance.close();
          }
        }
      } finally {
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each(['default', 'narrow', 'null reset', 'disabled'])(
    'projectService AND files, basePath and ignores with %s root policy',
    async (mode) => {
      const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-policy-match-'));
      const source =
        'const values = [1]; for (const key in values) {} debugger;';
      const files = [
        'pkg/src/match.ts',
        'pkg/src/ignored.ts',
        'pkg/src/global.ts',
        'pkg/other.ts',
        'pkg/src/match.js',
      ];
      const rootOptions =
        mode === 'default' ? {} : { tsconfigRootDir: path.join(tmp, 'pkg') };
      const instance = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          { ignores: ['**/global.ts'] },
          {
            plugins: ['@typescript-eslint'],
            languageOptions: { parserOptions: { project: false } },
            rules: {
              'no-debugger': 'error',
              '@typescript-eslint/no-for-in-array': 'error',
            },
          },
          {
            basePath: 'pkg',
            files: [['src/**', '**/*.ts']],
            ignores: ['**/ignored.ts'],
            languageOptions: {
              parserOptions: { projectService: true, ...rootOptions },
            },
          },
          ...(mode === 'null reset'
            ? [
                {
                  languageOptions: { parserOptions: { tsconfigRootDir: null } },
                },
              ]
            : []),
          ...(mode === 'disabled'
            ? [
                {
                  languageOptions: { parserOptions: { projectService: false } },
                },
              ]
            : []),
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({
            compilerOptions: { allowJs: true },
            files,
          }),
        },
      });
      try {
        await mkdir(path.join(tmp, 'pkg/src'), { recursive: true });
        for (const file of files) await writeFile(path.join(tmp, file), source);
        const results = await instance.lintFiles(files);
        expect(results).toHaveLength(files.length - 1);
        for (const result of results) {
          const name = path
            .relative(tmp, result.filePath)
            .replaceAll(path.sep, '/');
          expect(name).not.toBe('pkg/src/global.ts');
          const ids = result.messages.map(({ ruleId }) => ruleId);
          expect(ids).toContain('no-debugger');
          expect(ids.includes('@typescript-eslint/no-for-in-array')).toBe(
            name === 'pkg/src/match.ts' &&
              ['default', 'null reset'].includes(mode),
          );
        }
      } finally {
        await instance.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test('projectService root policies isolate concurrent API instances and fresh reloads', async () => {
    const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-root-isolation-'));
    const coreURL = pathToFileURL(
      path.resolve(import.meta.dirname, '../dist/index.js'),
    ).href;
    const instances = [];
    try {
      for (const name of ['a', 'b']) {
        const root = path.join(tmp, name);
        await mkdir(path.join(root, 'pkg'), { recursive: true });
        await writeFile(
          path.join(root, 'tsconfig.json'),
          JSON.stringify({ files: ['pkg/tool.ts'] }),
        );
        await writeFile(
          path.join(root, 'pkg/tool.ts'),
          'const values = [1]; for (const key in values) {}',
        );
        await writeFile(
          path.join(root, 'rslint.config.mjs'),
          `import {ts} from ${JSON.stringify(coreURL)}; export default [ts.configs.base,{languageOptions:{parserOptions:{projectService:true}},rules:{'@typescript-eslint/no-for-in-array':'error'}}];`,
        );
        instances.push(new Rslint({ cwd: path.join(root, 'pkg') }));
      }
      const run = () =>
        Promise.all(
          instances.map((instance) => instance.lintFiles(['tool.ts'])),
        );
      for (const result of await run())
        expect(result[0].messages.map((message) => message.ruleId)).toEqual([
          '@typescript-eslint/no-for-in-array',
        ]);
      // A now narrows its root explicitly; B retains its own config default.
      await writeFile(
        path.join(tmp, 'a/rslint.config.mjs'),
        `export default [{plugins:['@typescript-eslint'],languageOptions:{parserOptions:{projectService:true,tsconfigRootDir:${JSON.stringify(path.join(tmp, 'a/pkg'))}}},rules:{'@typescript-eslint/no-for-in-array':'error'}}];`,
      );
      const [a, b] = await run();
      expect(a[0].messages).toEqual([]);
      expect(b[0].messages.map((message) => message.ruleId)).toEqual([
        '@typescript-eslint/no-for-in-array',
      ]);
    } finally {
      await Promise.all(instances.map((instance) => instance.close()));
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test.each(
    [true, false].flatMap((declared) =>
      [
        'none',
        'service',
        'root',
        'project paths',
        'project false',
        'project null',
      ].map((option) => ({
        declared,
        option,
      })),
    ),
  )(
    'unmatched $option preserves owner project declarations (declared=$declared)',
    async ({ declared, option }) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-matched-policy-'),
      );
      const cwd = path.join(tmp, 'literal[owner]');
      await mkdir(cwd);
      const source =
        'export const result = value.member;\nfor (const key in [1]) {}';
      const unusedOptions = {
        none: {},
        service: { projectService: true },
        root: { tsconfigRootDir: cwd },
        'project paths': { project: './missing.json' },
        'project false': { project: false },
        'project null': { project: null },
      }[option];
      const instance = new Rslint({
        cwd,
        overrideConfigFile: true,
        overrideConfig: [
          {
            plugins: ['@typescript-eslint'],
            rules: {
              '@typescript-eslint/no-unsafe-member-access': 'error',
              '@typescript-eslint/no-for-in-array': 'error',
            },
          },
          ...(declared
            ? [
                {
                  languageOptions: {
                    parserOptions: { project: './unsafe.json' },
                  },
                },
                {
                  languageOptions: {
                    parserOptions: { project: './safe.json' },
                  },
                },
              ]
            : []),
          {
            files: ['unused.ts'],
            languageOptions: { parserOptions: unusedOptions },
          },
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({
            files: ['target.ts', 'unsafe.d.ts'],
          }),
          'unsafe.json': JSON.stringify({
            files: ['target.ts', 'unsafe.d.ts'],
          }),
          'safe.json': JSON.stringify({ files: ['target.ts', 'safe.d.ts'] }),
          'unsafe.d.ts': 'declare const value: any;',
          'safe.d.ts': 'declare const value: { member: number };',
        },
      });
      try {
        await writeFile(path.join(cwd, 'target.ts'), source);
        if (option === 'project paths') {
          await expect(instance.lintFiles(['target.ts'])).rejects.toThrow(
            /missing\.json/,
          );
          await expect(
            instance.lintText(source, { filePath: 'target.ts' }),
          ).rejects.toThrow(/missing\.json/);
          return;
        }
        for (const results of [
          await instance.lintFiles(['target.ts']),
          await instance.lintText(source, { filePath: 'target.ts' }),
        ]) {
          expect(results[0].messages.map(({ ruleId }) => ruleId)).toEqual([
            '@typescript-eslint/no-unsafe-member-access',
            '@typescript-eslint/no-for-in-array',
          ]);
        }
      } finally {
        await instance.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each([true, false])(
    'projectService false preserves scoped owner declarations but disables the implicit default (declared=%s)',
    async (declared) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-disabled-'),
      );
      const instance = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          {
            plugins: ['@typescript-eslint'],
            languageOptions: { parserOptions: { projectService: false } },
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-debugger': 'error',
            },
          },
          ...(declared
            ? [
                {
                  files: ['unused.ts'],
                  languageOptions: {
                    parserOptions: { project: './custom.json' },
                  },
                },
              ]
            : []),
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({ files: ['target.ts'] }),
          'custom.json': JSON.stringify({ files: ['target.ts'] }),
        },
      });
      try {
        const [result] = await instance.lintText(
          'const values = [1];\nfor (const key in values) {}\ndebugger;\n',
          { filePath: 'target.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
          ...(declared ? ['@typescript-eslint/no-for-in-array'] : []),
          'no-debugger',
        ]);
      } finally {
        await instance.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each(['./custom.json', './custom*.json'])(
    'explicit project policy keeps the config directory literal for %s',
    async (project) => {
      const tmp = await mkdtemp(path.join(os.tmpdir(), 'rslint-literal-base-'));
      const cwd = path.join(tmp, 'literal[owner]');
      await mkdir(cwd);
      const instance = new Rslint({
        cwd,
        overrideConfigFile: true,
        overrideConfig: [
          {
            plugins: ['@typescript-eslint'],
            languageOptions: {
              parserOptions: { projectService: false, project },
            },
            rules: { '@typescript-eslint/no-unsafe-member-access': 'error' },
          },
        ],
        virtualFiles: {
          'custom.json': JSON.stringify({ files: ['target.ts', 'types.d.ts'] }),
          'types.d.ts': 'declare const value: any;',
        },
      });
      try {
        const [result] = await instance.lintText(
          'export const result = value.member;',
          { filePath: 'target.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
          '@typescript-eslint/no-unsafe-member-access',
        ]);
      } finally {
        await instance.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test('projectService lintText lets a null root boundary restore ancestor discovery', async () => {
    const tmp = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-service-root-reset-'),
    );
    const rslint = new Rslint({
      cwd: tmp,
      overrideConfigFile: true,
      overrideConfig: [
        serviceConfig,
        {
          languageOptions: {
            parserOptions: { tsconfigRootDir: path.join(tmp, 'nested') },
          },
        },
        {
          files: ['**/*.ts'],
          languageOptions: { parserOptions: { tsconfigRootDir: null } },
          rules: { '@typescript-eslint/no-for-in-array': 'error' },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({ files: ['nested/probe.ts'] }),
      },
    });
    try {
      const [result] = await rslint.lintText(
        'const values = [1];\nfor (const key in values) {}\n',
        { filePath: 'nested/probe.ts' },
      );
      expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
        '@typescript-eslint/no-for-in-array',
      ]);
    } finally {
      await rslint.close();
      await rm(tmp, { recursive: true, force: true });
    }
  });

  test.each(['projectService', 'project', 'tsconfigRootDir'])(
    'projectService lintText preserves inherited %s when a later value is undefined',
    async (option) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-undefined-'),
      );
      const parserOptions =
        option === 'project'
          ? { projectService: false, project: './custom.json' }
          : {
              projectService: true,
              ...(option === 'tsconfigRootDir'
                ? { tsconfigRootDir: path.join(tmp, 'nested') }
                : {}),
            };
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          { languageOptions: { parserOptions } },
          {
            files: ['**/*.ts'],
            languageOptions: { parserOptions: { [option]: undefined } },
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-debugger': 'error',
            },
          },
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({
            compilerOptions: {
              paths: {
                values: [
                  option === 'tsconfigRootDir' ? './array.ts' : './scalar.ts',
                ],
              },
            },
            files: ['nested/probe.ts'],
          }),
          'custom.json': JSON.stringify({
            compilerOptions: { paths: { values: ['./array.ts'] } },
            files: ['nested/probe.ts'],
          }),
          'array.ts': 'export const values = [1];\n',
          'scalar.ts': 'export const values = 1;\n',
          ...(option === 'projectService'
            ? {
                'nested/tsconfig.json': JSON.stringify({
                  compilerOptions: { paths: { values: ['../array.ts'] } },
                  files: ['probe.ts'],
                }),
              }
            : {}),
        },
      });
      try {
        const [result] = await rslint.lintText(
          "import { values } from 'values';\nfor (const key in values) {}\ndebugger;\n",
          { filePath: 'nested/probe.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId).sort()).toEqual(
          option === 'tsconfigRootDir'
            ? ['no-debugger']
            : ['@typescript-eslint/no-for-in-array', 'no-debugger'],
        );
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each([
    {
      name: 'later paths',
      overrides: [{ project: ['./second.json'] }],
      typed: true,
    },
    { name: 'later empty array', overrides: [{ project: [] }], typed: true },
    {
      name: 'empty array without paths',
      overrides: [{ project: [] }],
      declared: false,
      typed: false,
    },
    { name: 'false', overrides: [{ project: false }], typed: false },
    { name: 'null', overrides: [{ project: null }], typed: false },
    {
      name: 'paths after false',
      overrides: [{ project: false }, { project: ['./second.json'] }],
      typed: true,
    },
    {
      name: 'service disabled with later paths',
      overrides: [{ projectService: false, project: ['./second.json'] }],
      typed: true,
    },
  ])(
    'explicit lintText preserves declaration order with $name',
    async ({ overrides, typed, declared = true }) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-project-array-override-'),
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          {
            plugins: ['@typescript-eslint'],
            languageOptions: {
              parserOptions: declared ? { project: ['./first.json'] } : {},
            },
          },
          {
            files: ['**/*.ts'],
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-debugger': 'error',
            },
          },
          ...overrides.map((parserOptions) => ({
            languageOptions: { parserOptions },
          })),
        ],
        virtualFiles: {
          'tsconfig.json': JSON.stringify({
            compilerOptions: { paths: { values: ['./array.ts'] } },
            files: ['probe.ts'],
          }),
          'first.json': JSON.stringify({
            compilerOptions: { paths: { values: ['./array.ts'] } },
            files: ['probe.ts'],
          }),
          'second.json': JSON.stringify({
            compilerOptions: { paths: { values: ['./scalar.ts'] } },
            files: ['probe.ts'],
          }),
          'scalar.ts': 'export const values = 1;\n',
          'array.ts': 'export const values = [1];\n',
        },
      });
      try {
        const [result] = await rslint.lintText(
          "import { values } from 'values';\nfor (const key in values) {}\ndebugger;\n",
          { filePath: 'probe.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
          ...(typed ? ['@typescript-eslint/no-for-in-array'] : []),
          'no-debugger',
        ]);
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each(['missing', 'excluded'])(
    'projectService lintText keeps gap syntax rules with a %s project',
    async (mode) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-service-unowned-'),
      );
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          serviceConfig,
          {
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-debugger': 'error',
            },
          },
        ],
        ...(mode === 'excluded'
          ? {
              virtualFiles: {
                'tsconfig.json': JSON.stringify({ files: ['covered.ts'] }),
                'covered.ts': 'export {};',
              },
            }
          : {}),
      });
      try {
        const [result] = await rslint.lintText(
          'const values = [1];\nfor (const key in values) {}\ndebugger;\n',
          { filePath: 'unowned.ts' },
        );
        expect(result.messages.map(({ ruleId }) => ruleId)).toEqual([
          'no-debugger',
        ]);
        const [invalid] = await rslint.lintText('const = ;', {
          filePath: 'unowned.js',
        });
        expect(
          invalid.messages.some((message) =>
            message.ruleId?.startsWith('TypeScript(TS'),
          ),
        ).toBe(true);
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

  test.each(
    Object.keys(ts.configs).flatMap((preset) =>
      ['project', 'service'].flatMap((mode) =>
        ['before', 'after'].map((order) => ({ preset, mode, order })),
      ),
    ),
  )(
    'TypeScript $preset preset $order caller $mode selection',
    async ({ preset, mode, order }) => {
      const tmp = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-preset-policy-'),
      );
      const caller = {
        languageOptions: {
          parserOptions:
            mode === 'service'
              ? { projectService: true }
              : { project: './custom.json' },
        },
      };
      const presetEntries = [ts.configs[preset]].flat();
      const rslint = new Rslint({
        cwd: tmp,
        overrideConfigFile: true,
        overrideConfig: [
          ...(order === 'before'
            ? [...presetEntries, caller]
            : [caller, ...presetEntries]),
          {
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              '@typescript-eslint/no-unnecessary-condition': 'error',
            },
          },
        ],
        virtualFiles: {
          'custom.json': JSON.stringify({
            compilerOptions: { strict: false },
            files: ['pkg/file.ts'],
          }),
          'pkg/tsconfig.json': JSON.stringify({
            compilerOptions: { strict: true },
            files: ['file.ts'],
          }),
        },
      });
      try {
        const [result] = await rslint.lintText(
          'export function keep(x: string | undefined) { return x != null; }\nconst values = [1];\nfor (const key in values) {}\n',
          { filePath: 'pkg/file.ts' },
        );
        const ruleIds = result.messages.map(({ ruleId }) => ruleId);
        expect(ruleIds).toContain('@typescript-eslint/no-for-in-array');
        expect(
          ruleIds.includes('@typescript-eslint/no-unnecessary-condition'),
        ).toBe(mode === 'project');
      } finally {
        await rslint.close();
        await rm(tmp, { recursive: true, force: true });
      }
    },
  );

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

  // A leading BOM is not part of the text an offset indexes, matching ESLint
  // v10. Pinned without hardcoding offsets: the same code with and without a
  // BOM reports identical columns and fix ranges.
  test('lintText reports BOM-stripped offsets for BOM-prefixed code', async () => {
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
      // The BOM leaves every offset untouched.
      expect(bm.column).toBe(pm.column);
      expect(bm.line).toBe(pm.line);
      expect(bm.fix.range[0]).toBe(pm.fix.range[0]);
      expect(bm.fix.range[1]).toBe(pm.fix.range[1]);
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
