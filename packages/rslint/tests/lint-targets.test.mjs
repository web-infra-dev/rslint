import { describe, expect, test } from 'rstack/test';
import { spawn } from 'node:child_process';
import { mkdir, mkdtemp, rm, writeFile } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';

const cliScript = path.resolve(import.meta.dirname, '../bin/rslint.js');
const cliTimeoutMs = 30_000;
const cliOutputLimitBytes = 1024 * 1024;
const cliTerminationGraceMs = 1_000;

async function writeFixture(root, files) {
  for (const [relative, content] of Object.entries(files)) {
    const fileName = path.join(root, ...relative.split('/'));
    await mkdir(path.dirname(fileName), { recursive: true });
    await writeFile(fileName, content);
  }
}

async function runCLI(cwd, args) {
  return new Promise((resolve, reject) => {
    const child = spawn(
      process.execPath,
      [
        cliScript,
        '--no-color',
        '--singleThreaded',
        '--format',
        'jsonline',
        ...args,
      ],
      {
        cwd,
        detached: process.platform !== 'win32',
        stdio: ['ignore', 'pipe', 'pipe'],
        windowsHide: true,
      },
    );
    const stdoutChunks = [];
    const stderrChunks = [];
    let outputBytes = 0;
    let settled = false;
    let failure;
    let timeoutTimer;
    let hardKillTimer;

    const terminateProcessTree = () => {
      if (child.pid === undefined) return;
      if (process.platform === 'win32') {
        const windowsTreeKill = spawn(
          // cspell:disable-next-line
          'taskkill',
          ['/pid', String(child.pid), '/t', '/f'],
          { stdio: 'ignore', windowsHide: true },
        );
        windowsTreeKill.once('error', () => child.kill('SIGKILL'));
        return;
      }
      try {
        process.kill(-child.pid, 'SIGTERM');
      } catch {
        child.kill('SIGTERM');
      }
      hardKillTimer = setTimeout(() => {
        try {
          process.kill(-child.pid, 'SIGKILL');
        } catch {
          child.kill('SIGKILL');
        }
      }, cliTerminationGraceMs);
      hardKillTimer.unref();
    };
    const requestFailure = (error) => {
      if (failure !== undefined || settled) return;
      failure = error;
      clearTimeout(timeoutTimer);
      if (child.pid === undefined) {
        settled = true;
        reject(error);
        return;
      }
      terminateProcessTree();
    };
    const collect = (chunks) => (chunk) => {
      outputBytes += chunk.length;
      if (outputBytes > cliOutputLimitBytes) {
        requestFailure(
          new Error(`rslint CLI output exceeded ${cliOutputLimitBytes} bytes`),
        );
        return;
      }
      chunks.push(chunk);
    };

    child.stdout.on('data', collect(stdoutChunks));
    child.stderr.on('data', collect(stderrChunks));
    child.once('error', requestFailure);
    child.once('close', (code, signal) => {
      if (settled) return;
      settled = true;
      clearTimeout(timeoutTimer);
      clearTimeout(hardKillTimer);
      const stdout = Buffer.concat(stdoutChunks).toString('utf8');
      const stderr = Buffer.concat(stderrChunks).toString('utf8');
      if (failure !== undefined) {
        reject(
          new Error(
            `${failure.message}\ncode=${String(code)} signal=${String(signal)}\nstdout:\n${stdout}\nstderr:\n${stderr}`,
          ),
        );
        return;
      }
      if (signal !== null || code === null) {
        reject(
          new Error(
            `rslint CLI exited abnormally: code=${String(code)} signal=${String(signal)}\nstdout:\n${stdout}\nstderr:\n${stderr}`,
          ),
        );
        return;
      }
      resolve({ code, stdout, stderr });
    });

    timeoutTimer = setTimeout(() => {
      requestFailure(new Error(`rslint CLI timed out after ${cliTimeoutMs}ms`));
    }, cliTimeoutMs);
    timeoutTimer.unref();
  });
}

function parseDiagnostics(stdout) {
  return stdout
    .split(/\r?\n/)
    .filter(Boolean)
    .map((line) => JSON.parse(line));
}

function absoluteDiagnosticPaths(cwd, diagnostics, ruleName) {
  return diagnostics
    .filter((diagnostic) => diagnostic.ruleName === ruleName)
    .map((diagnostic) => path.normalize(path.resolve(cwd, diagnostic.filePath)))
    .sort();
}

function normalizedDiagnostics(cwd, diagnostics) {
  return diagnostics
    .map(({ filePath, ruleName }) => ({
      filePath: path.normalize(path.resolve(cwd, filePath)),
      ruleName,
    }))
    .sort((left, right) => {
      const leftKey = `${left.ruleName}\0${left.filePath}`;
      const rightKey = `${right.ruleName}\0${right.filePath}`;
      return leftKey < rightKey ? -1 : leftKey > rightKey ? 1 : 0;
    });
}

describe('CLI basePath product contract', () => {
  test('automatic config scopes files, ignores, and project from its directory', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-base-path-'));
    try {
      await writeFixture(root, {
        'packages/app/rslint.config.mjs': `export default [
  { basePath: 'src', ignores: ['global-ignored.ts'] },
  {
    basePath: 'src',
    files: ['*.ts'],
    ignores: ['local-ignored.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    rules: { 'no-debugger': 'error' },
  },
];\n`,
        'packages/app/.gitignore': 'src/git-ignored.ts\n',
        'packages/app/src/tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['visible.ts'],
        }),
        'packages/app/src/visible.ts':
          "debugger;\nexport const broken: number = 'value';\n",
        'packages/app/src/global-ignored.ts': 'debugger;\n',
        'packages/app/src/local-ignored.ts': 'debugger;\n',
        'packages/app/src/git-ignored.ts': 'debugger;\n',
        'packages/app/src/deep/not-matched.ts': 'debugger;\n',
        'packages/app/outside-base.ts': 'debugger;\n',
      });

      const result = await runCLI(root, [
        '--type-check',
        path.join('packages', 'app'),
      ]);
      const visible = path.join(root, 'packages', 'app', 'src', 'visible.ts');

      expect(result.code).toBe(1);
      expect(
        normalizedDiagnostics(root, parseDiagnostics(result.stdout)),
      ).toEqual([
        { filePath: visible, ruleName: 'TypeScript(TS2322)' },
        { filePath: visible, ruleName: 'no-debugger' },
      ]);
      expect(result.stderr).not.toContain('warning:');
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });

  test('explicit external config resolves basePath from cwd without moving gitignore', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-external-'));
    const configDir = path.join(root, 'configs');
    const cwd = path.join(root, 'packages', 'app');
    const configPath = path.join(configDir, 'rslint.config.mjs');
    try {
      await writeFixture(root, {
        'configs/rslint.config.mjs': `export default [
  { basePath: 'src', ignores: ['config-ignored.ts'] },
  {
    basePath: 'src',
    files: ['**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    rules: { 'no-debugger': 'error' },
  },
];\n`,
        'packages/app/src/tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['visible.ts'],
        }),
        'configs/.gitignore': 'visible.ts\n',
        'packages/app/.gitignore': 'src/git-ignored.ts\n',
        'packages/app/src/visible.ts':
          "debugger;\nexport const broken: number = 'value';\n",
        'packages/app/src/git-ignored.ts': 'debugger;\n',
        'packages/app/src/config-ignored.ts': 'debugger;\n',
      });

      const configArgument = path.relative(cwd, configPath);
      const result = await runCLI(cwd, [
        '--type-check',
        '--config',
        configArgument,
      ]);
      const visible = path.join(cwd, 'src', 'visible.ts');

      expect(result.code).toBe(1);
      expect(
        normalizedDiagnostics(cwd, parseDiagnostics(result.stdout)),
      ).toEqual([
        { filePath: visible, ruleName: 'TypeScript(TS2322)' },
        { filePath: visible, ruleName: 'no-debugger' },
      ]);
      expect(result.stderr).not.toContain('warning:');
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });

  test('ancestor basePath does not let a global ignore prune the config root', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-base-path-'));
    try {
      await writeFixture(root, {
        'app/rslint.config.mjs': `export default [
  { basePath: '..', ignores: ['*'] },
  { files: ['**/*.js'], rules: { 'no-debugger': 'error' } },
];\n`,
        'app/root.js': 'debugger;\n',
        'app/nested/child.js': 'debugger;\n',
      });

      const result = await runCLI(root, ['app']);
      const expected = [
        path.join(root, 'app', 'root.js'),
        path.join(root, 'app', 'nested', 'child.js'),
      ].sort();

      expect(result.code).toBe(1);
      expect(
        absoluteDiagnosticPaths(
          root,
          parseDiagnostics(result.stdout),
          'no-debugger',
        ),
      ).toEqual(expected);
      expect(result.stderr).not.toContain('warning:');
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });
});

describe('CLI lint target contracts', () => {
  test('external config keeps authored paths and invocation target scope', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-external-'));
    const configDir = path.join(root, 'configs');
    const cwd = path.join(root, 'packages', 'app');
    const configPath = path.join(configDir, 'rslint.config.mjs');
    try {
      await writeFixture(root, {
        'configs/rslint.config.mjs': `export default [
  { ignores: ['../packages/app/src/config-ignored.ts'] },
  {
    files: ['../packages/app/src/**/*.ts'],
    languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
    rules: { 'no-debugger': 'error' },
  },
  { files: ['*.ts'], rules: { 'no-console': 'error' } },
];\n`,
        'configs/tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true },
          files: ['../packages/app/src/visible.ts'],
        }),
        'configs/.gitignore': 'visible.ts\nrslint.config.mjs\n',
        'configs/config-only.ts': 'console.log("config");\n',
        'packages/app/.gitignore': 'src/git-ignored.ts\n',
        'packages/app/src/visible.ts':
          "debugger;\nexport const broken: number = 'value';\n",
        'packages/app/src/git-ignored.ts': 'debugger;\n',
        'packages/app/src/config-ignored.ts': 'debugger;\n',
      });

      const configArgument = path.relative(cwd, configPath);
      const implicit = await runCLI(cwd, [
        '--type-check',
        '--config',
        configArgument,
      ]);
      const explicit = await runCLI(cwd, [
        '--type-check',
        '--config',
        configArgument,
        '.',
      ]);
      const visible = path.join(cwd, 'src', 'visible.ts');
      const expected = [
        { filePath: visible, ruleName: 'TypeScript(TS2322)' },
        { filePath: visible, ruleName: 'no-debugger' },
      ];

      for (const result of [implicit, explicit]) {
        expect(result.code).toBe(1);
        expect(
          normalizedDiagnostics(cwd, parseDiagnostics(result.stdout)),
        ).toEqual(expected);
        expect(result.stderr).not.toContain('warning:');
      }
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });

  test('multiple files and directories form one deduplicated union', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-union-'));
    try {
      await writeFixture(root, {
        'rslint.config.mjs':
          "export default [{ rules: { 'no-debugger': 'error' } }];\n",
        'standalone-a.js': 'debugger;\n',
        'standalone-b.js': 'debugger;\n',
        'first/one.js': 'debugger;\n',
        'first/nested/two.js': 'debugger;\n',
        'second/three.js': 'debugger;\n',
        'unselected.js': 'debugger;\n',
      });

      const result = await runCLI(root, [
        '--config',
        'rslint.config.mjs',
        'standalone-a.js',
        'standalone-b.js',
        'first',
        'first/nested',
        'first/one.js',
        'second',
      ]);
      expect(result.code).toBe(1);
      expect(
        absoluteDiagnosticPaths(
          root,
          parseDiagnostics(result.stdout),
          'no-debugger',
        ),
      ).toEqual(
        [
          'standalone-a.js',
          'standalone-b.js',
          'first/one.js',
          'first/nested/two.js',
          'second/three.js',
        ]
          .map((relative) => path.join(root, ...relative.split('/')))
          .sort(),
      );
      expect(result.stderr).not.toContain('warning:');
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });

  test('explicit file outcomes distinguish syntax, ignored, and missing files', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-files-'));
    try {
      await writeFixture(root, {
        'rslint.config.mjs': `export default [{
  files: ['**/*.ts'],
  rules: { 'no-debugger': 'error' },
}];\n`,
        '.gitignore': 'rslint.config.mjs\nignored.ts\n',
        'outside-files.js': 'debugger;\nconst = ;\n',
        'linted.ts': 'debugger;\n',
        'ignored.ts': 'debugger;\n',
      });

      const result = await runCLI(root, [
        '--config',
        'rslint.config.mjs',
        'outside-files.js',
        'linted.ts',
        'ignored.ts',
        'missing.ts',
      ]);
      const diagnostics = parseDiagnostics(result.stdout);
      expect(result.code).toBe(1);
      const syntaxDiagnostics = diagnostics.filter(({ ruleName }) =>
        /^TypeScript\(TS\d+\)$/.test(ruleName),
      );
      expect(syntaxDiagnostics.length).toBeGreaterThan(0);
      expect(
        syntaxDiagnostics.every(
          ({ filePath }) =>
            path.resolve(root, filePath) ===
            path.join(root, 'outside-files.js'),
        ),
      ).toBe(true);
      expect(absoluteDiagnosticPaths(root, diagnostics, 'no-debugger')).toEqual(
        [path.join(root, 'linted.ts')],
      );
      expect(diagnostics).toHaveLength(syntaxDiagnostics.length + 1);
      expect(result.stderr).toContain(
        'ignored.ts is ignored because of a matching ignore pattern',
      );
      expect(result.stderr).toContain('missing.ts was not found, skipping');
      expect(result.stderr).not.toContain('outside-files.js');
      expect(result.stderr).not.toContain('linted.ts');
    } finally {
      await rm(root, { recursive: true, force: true });
    }
  });
});
