import { describe, expect, test } from 'rstack/test';
import { spawn } from 'node:child_process';
import {
  mkdir,
  mkdtemp,
  readFile,
  realpath,
  writeFile,
} from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { pathToFileURL } from 'node:url';
import { cleanupTempDir } from './cleanup-temp-dir.js';

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

async function runCLI(cwd, args, options = {}) {
  const {
    scriptPath = cliScript,
    timeoutMs = cliTimeoutMs,
    graceMs = cliTerminationGraceMs,
    platform = process.platform,
    spawnImpl = spawn,
    setTimeoutImpl = setTimeout,
    onSpawn,
  } = options;
  return new Promise((resolve, reject) => {
    const child = spawnImpl(
      process.execPath,
      [
        scriptPath,
        '--no-color',
        '--singleThreaded',
        '--format',
        'jsonline',
        ...args,
      ],
      {
        cwd,
        detached: platform !== 'win32',
        stdio: ['ignore', 'pipe', 'pipe'],
        windowsHide: true,
      },
    );
    onSpawn?.(child);
    const stdoutChunks = [];
    const stderrChunks = [];
    let outputBytes = 0;
    let settled = false;
    let failure;
    let timeoutTimer;
    let pipeDrainTimer;

    let killDone = Promise.resolve();
    const destroyPipes = () => {
      child.stdout.destroy();
      child.stderr.destroy();
    };
    const terminateProcessTree = () => {
      if (child.pid === undefined) return;
      if (platform === 'win32') {
        if (child.exitCode !== null || child.signalCode !== null) {
          destroyPipes();
          return;
        }
        killDone = new Promise((done) => {
          let finished = false;
          let windowsTreeKill;
          const finish = () => {
            if (finished) return;
            finished = true;
            clearTimeout(killTimer);
            done();
          };
          const killTimer = setTimeout(() => {
            try {
              windowsTreeKill?.kill('SIGKILL');
            } catch {
              // Ignore teardown races if the tree-kill helper already exited.
            }
            if (child.exitCode === null && child.signalCode === null) {
              try {
                child.kill('SIGKILL');
              } catch {
                // Ignore if child already exited.
              }
            }
            destroyPipes();
            finish();
          }, graceMs);
          killTimer.unref();
          try {
            windowsTreeKill = spawnImpl(
              // cspell:disable-next-line
              'taskkill',
              ['/pid', String(child.pid), '/t', '/f'],
              { stdio: 'ignore', windowsHide: true },
            );
            windowsTreeKill.once('error', () => {
              if (child.exitCode === null && child.signalCode === null) {
                child.kill('SIGKILL');
              }
              destroyPipes();
              finish();
            });
            windowsTreeKill.once('close', (code) => {
              if (
                code !== 0 &&
                child.exitCode === null &&
                child.signalCode === null
              ) {
                child.kill('SIGKILL');
              }
              if (code !== 0) {
                destroyPipes();
              }
              finish();
            });
          } catch {
            if (child.exitCode === null && child.signalCode === null) {
              child.kill('SIGKILL');
            }
            destroyPipes();
            finish();
          }
        });
        pipeDrainTimer = setTimeout(() => {
          destroyPipes();
        }, graceMs);
        pipeDrainTimer.unref();
        return;
      }

      // POSIX: `detached: true` places the CLI and its descendants in process
      // group `-child.pid`. Signal the group even when the group leader has
      // already exited while a descendant still holds stdout/stderr open.
      const processGroupId = -child.pid;
      const isGroupAlive = () => {
        try {
          process.kill(processGroupId, 0);
          return true;
        } catch (err) {
          return err?.code === 'EPERM';
        }
      };
      let groupSignaled = false;
      try {
        process.kill(processGroupId, 'SIGTERM');
        groupSignaled = true;
      } catch (err) {
        if (err?.code === 'EPERM') {
          groupSignaled = true;
        } else if (child.exitCode === null && child.signalCode === null) {
          child.kill('SIGTERM');
        }
      }
      if (
        !groupSignaled &&
        (child.exitCode !== null || child.signalCode !== null)
      ) {
        destroyPipes();
        return;
      }
      killDone = new Promise((done) => {
        let finished = false;
        let pollTimer;
        let hardTimer;
        let reapDeadlineTimer;
        const finish = () => {
          if (finished) return;
          finished = true;
          clearInterval(pollTimer);
          clearTimeout(hardTimer);
          clearTimeout(reapDeadlineTimer);
          done();
        };
        const checkDone = () => {
          const leaderAlive =
            child.exitCode === null && child.signalCode === null;
          if (!leaderAlive && (!groupSignaled || !isGroupAlive())) {
            finish();
          }
        };
        pollTimer = setInterval(checkDone, 20);
        pollTimer.unref();
        hardTimer = setTimeout(() => {
          if (groupSignaled && isGroupAlive()) {
            try {
              process.kill(processGroupId, 'SIGKILL');
            } catch {
              // Group exited between probe and SIGKILL.
            }
          }
          if (child.exitCode === null && child.signalCode === null) {
            try {
              child.kill('SIGKILL');
            } catch {
              // Leader already exited.
            }
          }
          destroyPipes();
          if (!groupSignaled || !isGroupAlive()) {
            finish();
            return;
          }
          reapDeadlineTimer = setTimeout(finish, 2_000);
          reapDeadlineTimer.unref();
        }, graceMs);
        hardTimer.unref();
        checkDone();
      });
      pipeDrainTimer = setTimeout(() => {
        destroyPipes();
      }, graceMs);
      pipeDrainTimer.unref();
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
      clearTimeout(pipeDrainTimer);
      const settle = () => {
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
      };
      void killDone.then(settle, settle);
    });

    timeoutTimer = setTimeoutImpl(() => {
      requestFailure(new Error(`rslint CLI timed out after ${timeoutMs}ms`));
    }, timeoutMs);
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
      await cleanupTempDir(root);
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
      await cleanupTempDir(root);
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
      await cleanupTempDir(root);
    }
  });
});

describe('CLI lint target contracts', () => {
  test('projectService migration preserves explicit projects across separately inserted TypeScript presets', async () => {
    const root = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-cli-service-migrate-'),
    );
    try {
      await writeFixture(root, {
        'rslint.json': JSON.stringify([
          {
            plugins: ['@typescript-eslint'],
            languageOptions: { parserOptions: { project: './custom.json' } },
            rules: { '@typescript-eslint/no-for-in-array': 'error' },
          },
          {
            plugins: ['@typescript-eslint'],
            rules: { 'no-console': 'error' },
          },
        ]),
        'tsconfig.json': JSON.stringify({ files: [] }),
        'custom.json': JSON.stringify({ files: ['probe.ts'] }),
        'probe.ts':
          'const values = [1]; for (const key in values) { console.log(key); }\n',
        'node_modules/@rslint/core/package.json': JSON.stringify({
          name: '@rslint/core',
          type: 'module',
          exports: './index.js',
        }),
        'node_modules/@rslint/core/index.js': `export * from ${JSON.stringify(
          pathToFileURL(path.resolve(import.meta.dirname, '../dist/index.js'))
            .href,
        )};\n`,
      });
      const migration = await runCLI(root, ['--init']);
      expect(migration.code, migration.stderr).toBe(0);

      const result = await runCLI(root, ['probe.ts']);
      expect(result.code, result.stderr).toBe(1);
      const diagnostics = parseDiagnostics(result.stdout);
      expect(diagnostics.map(({ ruleName }) => ruleName).sort()).toEqual([
        '@typescript-eslint/no-for-in-array',
        'no-console',
      ]);
    } finally {
      await cleanupTempDir(root);
    }
  });

  test('projectService migration preserves disabled, explicit, and unconfigured file scopes', async () => {
    const root = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-cli-service-migrate-scopes-'),
    );
    try {
      await writeFixture(root, {
        'rslint.json': JSON.stringify([
          {
            files: ['tools/**/*.ts'],
            plugins: ['@typescript-eslint'],
            languageOptions: { parserOptions: { projectService: false } },
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-console': 'error',
            },
          },
          {
            files: ['src/**/*.ts'],
            plugins: ['@typescript-eslint'],
            languageOptions: {
              parserOptions: { project: ['./tsconfig.json'] },
            },
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-console': 'error',
            },
          },
          {
            files: ['gap/**/*.ts'],
            plugins: ['@typescript-eslint'],
            rules: {
              '@typescript-eslint/no-for-in-array': 'error',
              'no-console': 'error',
            },
          },
        ]),
        'tsconfig.json': JSON.stringify({
          files: ['tools/probe.ts', 'src/probe.ts', 'gap/probe.ts'],
        }),
        'tools/probe.ts':
          'export const values = [1]; for (const key in values) { console.log(key); }\n',
        'src/probe.ts':
          'export const values = [1]; for (const key in values) { console.log(key); }\n',
        'gap/probe.ts':
          'export const values = [1]; for (const key in values) { console.log(key); }\n',
        'node_modules/@rslint/core/package.json': JSON.stringify({
          name: '@rslint/core',
          type: 'module',
          exports: './index.js',
        }),
        'node_modules/@rslint/core/index.js': `export * from ${JSON.stringify(
          pathToFileURL(path.resolve(import.meta.dirname, '../dist/index.js'))
            .href,
        )};\n`,
      });
      const migration = await runCLI(root, ['--init']);
      expect(migration.code, migration.stderr).toBe(0);
      const migrated = await readFile(
        path.join(root, 'rslint.config.ts'),
        'utf8',
      );
      expect(migrated).toContain('projectService: false');
      expect(migrated).toContain("project: ['./tsconfig.json']");
      expect(migrated).not.toContain('projectService: true');

      const result = await runCLI(root, ['tools', 'src', 'gap']);
      expect(result.code, result.stderr).toBe(1);
      expect(
        normalizedDiagnostics(root, parseDiagnostics(result.stdout)),
      ).toEqual([
        {
          filePath: path.join(root, 'src', 'probe.ts'),
          ruleName: '@typescript-eslint/no-for-in-array',
        },
        {
          filePath: path.join(root, 'gap', 'probe.ts'),
          ruleName: 'no-console',
        },
        {
          filePath: path.join(root, 'src', 'probe.ts'),
          ruleName: 'no-console',
        },
        {
          filePath: path.join(root, 'tools', 'probe.ts'),
          ruleName: 'no-console',
        },
      ]);
    } finally {
      await cleanupTempDir(root);
    }
  });

  test.each([
    { packageType: 'module', configName: 'rslint.config.js' },
    { packageType: 'commonjs', configName: 'rslint.config.mjs' },
  ])(
    'projectService migration preserves a runtime null boundary in $configName',
    async ({ packageType, configName }) => {
      const root = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-cli-service-migrate-root-'),
      );
      try {
        await writeFixture(root, {
          'package.json': JSON.stringify({ type: packageType }),
          'rslint.json': JSON.stringify([
            {
              plugins: ['@typescript-eslint'],
              languageOptions: {
                parserOptions: {
                  projectService: true,
                  tsconfigRootDir: path.join(root, 'pkg'),
                },
              },
              rules: {
                '@typescript-eslint/no-for-in-array': 'error',
                'no-console': 'error',
              },
            },
            {
              files: ['pkg/**/*.ts'],
              plugins: ['@typescript-eslint'],
              languageOptions: { parserOptions: { tsconfigRootDir: null } },
            },
          ]),
          'tsconfig.json': JSON.stringify({ files: ['pkg/probe.ts'] }),
          'pkg/probe.ts':
            'export const values = [1]; for (const key in values) { console.log(key); }\n',
          'node_modules/@rslint/core/package.json': JSON.stringify({
            name: '@rslint/core',
            type: 'module',
            exports: './index.js',
          }),
          'node_modules/@rslint/core/index.js': `export * from ${JSON.stringify(
            pathToFileURL(path.resolve(import.meta.dirname, '../dist/index.js'))
              .href,
          )};\n`,
        });
        const migration = await runCLI(root, ['--init']);
        expect(migration.code, migration.stderr).toBe(0);
        expect(await readFile(path.join(root, configName), 'utf8')).toContain(
          'tsconfigRootDir: null',
        );

        const result = await runCLI(root, ['pkg/probe.ts']);
        expect(result.code, result.stderr).toBe(1);
        expect(
          parseDiagnostics(result.stdout)
            .map(({ ruleName }) => ruleName)
            .sort(),
        ).toEqual(['@typescript-eslint/no-for-in-array', 'no-console']);
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test.each([
    { mode: [], expectedRules: ['no-debugger'] },
    {
      mode: ['--type-check'],
      expectedRules: ['TypeScript(TS2322)', 'no-debugger'],
    },
    { mode: ['--type-check-only'], expectedRules: ['TypeScript(TS2322)'] },
  ])(
    'projectService discovers the selected file project and preserves full type context (%j)',
    async ({ mode, expectedRules }) => {
      const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-service-'));
      try {
        await writeFixture(root, {
          'rslint.config.mjs': `export default [{
  files: ['**/*.ts'],
  languageOptions: { parserOptions: { projectService: true } },
  rules: { 'no-debugger': 'error' },
}];\n`,
          'tsconfig.json': JSON.stringify({
            files: ['unrelated.ts', 'packages/app/src/target.ts'],
          }),
          'unrelated.ts': 'export const wrong: number = "root";\n',
          'packages/app/tsconfig.json': JSON.stringify({
            files: ['src/target.ts', 'src/peer.ts'],
          }),
          'packages/app/src/target.ts': 'debugger;\nexport {};\n',
          'packages/app/src/peer.ts': 'export const wrong: number = "peer";\n',
        });
        const result = await runCLI(root, [
          ...mode,
          'packages/app/src/target.ts',
        ]);
        expect(result.code).toBe(1);
        const diagnostics = parseDiagnostics(result.stdout);
        expect(diagnostics.map(({ ruleName }) => ruleName).sort()).toEqual(
          [...expectedRules].sort(),
        );
        expect(
          absoluteDiagnosticPaths(root, diagnostics, 'no-debugger'),
        ).toEqual(
          mode.includes('--type-check-only')
            ? []
            : [path.join(root, 'packages', 'app', 'src', 'target.ts')],
        );
        expect(
          absoluteDiagnosticPaths(root, diagnostics, 'TypeScript(TS2322)'),
        ).toEqual(
          mode.length === 0
            ? []
            : [path.join(root, 'packages', 'app', 'src', 'peer.ts')],
        );
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test.each([
    ['invocation directory without presets', false, false, false],
    ['invocation directory with presets', true, false, false],
    ['explicit config directory', false, true, true],
  ])(
    'projectService CLI root uses %s',
    async (_name, preset, explicitRoot, typed) => {
      const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-root-'));
      const coreURL = pathToFileURL(
        path.resolve(import.meta.dirname, '../dist/index.js'),
      ).href;
      try {
        await writeFixture(root, {
          'rslint.config.mjs': `import {ts} from ${JSON.stringify(coreURL)}; export default [${preset ? 'ts.configs.base,' : ''}{plugins:['@typescript-eslint'],languageOptions:{parserOptions:{projectService:true${explicitRoot ? `,tsconfigRootDir:${JSON.stringify(root)}` : ''}}},rules:{'no-debugger':'error','@typescript-eslint/no-for-in-array':'error'}}];`,
          'tsconfig.json': JSON.stringify({ files: ['pkg/tool.ts'] }),
          'pkg/tool.ts':
            'const values = [1]; for (const key in values) {} debugger;',
        });
        for (const args of [
          ['tool.ts'],
          ['.'],
          ['--config', '../rslint.config.mjs', 'tool.ts'],
        ]) {
          const result = await runCLI(path.join(root, 'pkg'), args);
          expect(result.code).toBe(1);
          const rules = parseDiagnostics(result.stdout).map(
            ({ ruleName }) => ruleName,
          );
          expect(rules).toContain('no-debugger');
          expect(rules.includes('@typescript-eslint/no-for-in-array')).toBe(
            typed,
          );
        }
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test.each(['directory', 'trailing separator', 'dot segments'])(
    'projectService normalizes the CLI root boundary with %s',
    async (spelling) => {
      const root = await realpath(
        await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-root-spelling-')),
      );
      const boundary = path.join(root, 'pkg');
      const tsconfigRootDir =
        spelling === 'trailing separator'
          ? boundary + path.sep
          : spelling === 'dot segments'
            ? boundary + path.sep + '..' + path.sep + 'pkg' + path.sep + '.'
            : boundary;
      try {
        await writeFixture(root, {
          'rslint.config.mjs': `export default ${JSON.stringify([
            {
              files: ['**/*.ts'],
              plugins: ['@typescript-eslint'],
              languageOptions: {
                parserOptions: { projectService: true, tsconfigRootDir },
              },
              rules: {
                '@typescript-eslint/no-for-in-array': 'error',
                'no-debugger': 'error',
              },
            },
          ])};`,
          'tsconfig.json': JSON.stringify({ files: ['pkg/tool.ts'] }),
          'pkg/tool.ts':
            'const values = [1];\nfor (const key in values) {}\ndebugger;\n',
        });
        const result = await runCLI(root, ['pkg/tool.ts']);
        expect(result.code).toBe(1);
        expect(result.stderr).toBe('');
        expect(
          parseDiagnostics(result.stdout).map(({ ruleName }) => ruleName),
        ).toEqual(['no-debugger']);
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test.each(
    [
      '.',
      '',
      ...(process.platform === 'win32' ? [] : ['C:/foreign-root']),
    ].flatMap((value) =>
      ['matched', 'unmatched', 'overridden', 'null reset'].map((mode) => ({
        value,
        mode,
      })),
    ),
  )(
    'projectService validates final root $value after $mode configuration',
    async ({ value, mode }) => {
      const root = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-root-validation-'),
      );
      try {
        const invalid = {
          ...(mode === 'unmatched' ? { files: ['other/**'] } : {}),
          languageOptions: { parserOptions: { tsconfigRootDir: value } },
        };
        const config = [
          {
            languageOptions: { parserOptions: { projectService: true } },
            rules: { 'no-debugger': 'error' },
          },
          invalid,
          ...(mode === 'overridden' || mode === 'null reset'
            ? [
                {
                  languageOptions: {
                    parserOptions: {
                      tsconfigRootDir: mode === 'null reset' ? null : root,
                    },
                  },
                },
              ]
            : []),
        ];
        await writeFixture(root, {
          'rslint.config.mjs': `export default ${JSON.stringify(config)};`,
          'tsconfig.json': JSON.stringify({ files: ['src.ts'] }),
          'src.ts': 'debugger;',
        });
        const result = await runCLI(root, ['src.ts']);
        expect(result.code).toBe(1);
        if (mode === 'matched') {
          expect(result.stderr).toContain(
            'tsconfigRootDir must be an absolute path',
          );
          expect(result.stdout).not.toContain('no-debugger');
        } else {
          expect(result.stderr).not.toContain('tsconfigRootDir');
          expect(
            parseDiagnostics(result.stdout).map(({ ruleName }) => ruleName),
          ).toEqual(['no-debugger']);
        }
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test('projectService keeps automatic, explicit, and disabled targets separate in a broad CLI run', async () => {
    const root = await mkdtemp(
      path.join(os.tmpdir(), 'rslint-cli-service-mixed-'),
    );
    try {
      await writeFixture(root, {
        'rslint.config.mjs': `export default [
  {
    files: ['**/*.ts'],
    plugins: ['@typescript-eslint'],
    languageOptions: { parserOptions: { projectService: true } },
    rules: { '@typescript-eslint/no-for-in-array': 'error', 'no-debugger': 'error' },
  },
  {
    files: ['legacy/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: './custom.json' } },
  },
  {
    files: ['scripts/*.ts'],
    languageOptions: { parserOptions: { projectService: false, project: false } },
  },
];\n`,
        'app/tsconfig.json': JSON.stringify({ files: ['probe.ts'] }),
        'app/probe.ts': 'const values = [1]; for (const key in values) {}\n',
        'custom.json': JSON.stringify({ files: ['legacy/probe.ts'] }),
        'legacy/tsconfig.json': JSON.stringify({ files: [] }),
        'legacy/probe.ts': 'const values = [1]; for (const key in values) {}\n',
        'scripts/probe.ts':
          'const values = [1]; for (const key in values) {}\ndebugger;\n',
      });
      const result = await runCLI(root, ['.']);
      expect(result.code).toBe(1);
      const diagnostics = parseDiagnostics(result.stdout);
      expect(normalizedDiagnostics(root, diagnostics)).toEqual([
        {
          filePath: path.join(root, 'app', 'probe.ts'),
          ruleName: '@typescript-eslint/no-for-in-array',
        },
        {
          filePath: path.join(root, 'legacy', 'probe.ts'),
          ruleName: '@typescript-eslint/no-for-in-array',
        },
        {
          filePath: path.join(root, 'scripts', 'probe.ts'),
          ruleName: 'no-debugger',
        },
      ]);
    } finally {
      await cleanupTempDir(root);
    }
  });

  test.each(
    ['ordinary', 'false', 'null', 'service'].flatMap((selection) =>
      ['--type-check', '--type-check-only'].map((mode) => ({
        selection,
        mode,
      })),
    ),
  )(
    'program-wide explicit projects use actual root contexts from $selection targets in $mode',
    async ({ selection, mode }) => {
      const root = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-cli-declared-roots-'),
      );
      try {
        const parserOptions = {
          tsconfigRootDir: path.join(root, 'actual'),
          ...(selection === 'ordinary'
            ? {}
            : { project: selection === 'null' ? null : false }),
          ...(selection === 'service' ? { projectService: true } : {}),
        };
        await writeFixture(root, {
          'rslint.config.mjs': `export default ${JSON.stringify([
            {
              files: ['actual/*.ts'],
              plugins: ['@typescript-eslint'],
              languageOptions: {
                parserOptions: { project: './declared.json' },
              },
              rules: {
                '@typescript-eslint/no-for-in-array': 'error',
                'no-debugger': 'error',
              },
            },
            {
              basePath: 'unused-origin',
              files: ['never.ts'],
              languageOptions: {
                parserOptions: { project: './additional.json' },
              },
            },
            { files: ['actual/probe.ts'], languageOptions: { parserOptions } },
          ])};`,
          // No declared.json exists at the owner: every mode must retain the
          // selected target's root even when service or clear owns its binding.
          'actual/declared.json': JSON.stringify({
            files: ['probe.ts', 'peer.ts'],
          }),
          'actual/additional.json': JSON.stringify({
            files: ['unselected.ts'],
          }),
          'actual/tsconfig.json': JSON.stringify({
            files: ['probe.ts', 'service-peer.ts'],
          }),
          'actual/probe.ts':
            'const values = [1];\nfor (const key in values) {}\ndebugger;\nexport {};\n',
          'actual/peer.ts': 'export const broken: number = "explicit";\n',
          'actual/unselected.ts':
            'export const broken: number = "additional";\n',
          'actual/service-peer.ts':
            'export const broken: number = "service";\n',
        });
        const result = await runCLI(root, [mode, 'actual/probe.ts']);
        expect(result.code).toBe(1);
        expect(result.stderr).toBe('');
        const expected = [
          { filePath: 'actual/peer.ts', ruleName: 'TypeScript(TS2322)' },
          { filePath: 'actual/unselected.ts', ruleName: 'TypeScript(TS2322)' },
        ];
        if (selection === 'service') {
          expected.push({
            filePath: 'actual/service-peer.ts',
            ruleName: 'TypeScript(TS2322)',
          });
        }
        if (mode === '--type-check') {
          expected.push({
            filePath: 'actual/probe.ts',
            ruleName: 'no-debugger',
          });
          if (selection === 'ordinary' || selection === 'service') {
            expected.push({
              filePath: 'actual/probe.ts',
              ruleName: '@typescript-eslint/no-for-in-array',
            });
          }
        }
        expect(
          normalizedDiagnostics(root, parseDiagnostics(result.stdout)),
        ).toEqual(normalizedDiagnostics(root, expected));
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test.each([
    {
      name: 'broad lint',
      args: ['.'],
      lintFiles: ['x.ts', 'y.ts'],
      typeFiles: [],
    },
    {
      name: 'focused lint',
      args: ['src/y.ts'],
      lintFiles: ['y.ts'],
      typeFiles: [],
    },
    {
      name: 'program-wide check',
      args: ['--type-check-only', '.'],
      lintFiles: [],
      typeFiles: ['a/peer.ts', 'b/peer.ts'],
    },
  ])(
    'scoped root contexts keep overlapping projects separate in $name',
    async ({ args, lintFiles, typeFiles }) => {
      const root = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-cli-root-contexts-'),
      );
      try {
        await writeFixture(root, {
          'rslint.config.mjs': `export default ${JSON.stringify([
            { ignores: ['a/**', 'b/**', 'rslint.config.mjs'] },
            {
              files: ['src/*.ts'],
              plugins: ['@typescript-eslint'],
              languageOptions: {
                parserOptions: { project: './declared.json' },
              },
              rules: {
                '@typescript-eslint/no-for-in-array': 'error',
                'no-debugger': 'error',
              },
            },
            ...['x', 'y'].map((name, index) => ({
              files: [`src/${name}.ts`],
              languageOptions: {
                parserOptions: {
                  tsconfigRootDir: path.join(root, index === 0 ? 'a' : 'b'),
                },
              },
            })),
          ])};`,
          ...Object.fromEntries(
            ['a', 'b'].map((name) => [
              `${name}/declared.json`,
              JSON.stringify({
                compilerOptions: { paths: { values: ['./values.ts'] } },
                files: ['../src/x.ts', '../src/y.ts', 'peer.ts'],
              }),
            ]),
          ),
          'a/values.ts': 'export const values = [1];\n',
          'b/values.ts': 'export const values = { item: 1 };\n',
          'a/peer.ts': 'export const broken: number = "a";\n',
          'b/peer.ts': 'export const broken: number = "b";\n',
          'src/x.ts':
            "import { values } from 'values';\nfor (const key in values) {}\ndebugger;\n",
          'src/y.ts':
            "import { values } from 'values';\nfor (const key in values) {}\ndebugger;\n",
        });
        const result = await runCLI(root, args);
        expect(result.code).toBe(1);
        expect(result.stderr).toBe('');
        const expected = [
          ...lintFiles.map((name) => ({
            filePath: `src/${name}`,
            ruleName: 'no-debugger',
          })),
          ...(lintFiles.includes('x.ts')
            ? [
                {
                  filePath: 'src/x.ts',
                  ruleName: '@typescript-eslint/no-for-in-array',
                },
              ]
            : []),
          ...typeFiles.map((filePath) => ({
            filePath,
            ruleName: 'TypeScript(TS2322)',
          })),
        ];
        expect(
          normalizedDiagnostics(root, parseDiagnostics(result.stdout)),
        ).toEqual(normalizedDiagnostics(root, expected));
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

  test.each([
    {
      name: 'false target',
      parserOptions: { project: false },
      selected: true,
      declared: false,
      mode: '--type-check-only',
      expected: [],
    },
    {
      name: 'null target',
      parserOptions: { project: null },
      selected: true,
      declared: false,
      mode: '--type-check-only',
      expected: [],
    },
    {
      name: 'disabled service target',
      parserOptions: { projectService: false },
      selected: true,
      declared: false,
      mode: '--type-check-only',
      expected: [],
    },
    {
      name: 'clear target with explicit declarations',
      parserOptions: { project: false },
      selected: true,
      declared: true,
      mode: '--type-check-only',
      expected: ['explicit-error.ts'],
    },
    {
      name: 'no service targets in lint',
      parserOptions: { projectService: true },
      selected: false,
      declared: false,
      mode: null,
      expected: [],
    },
    {
      name: 'no service targets in type-check',
      parserOptions: { projectService: true },
      selected: false,
      declared: false,
      mode: '--type-check-only',
      expected: ['default-error.ts'],
    },
    {
      name: 'no targets with explicit declarations in lint',
      parserOptions: {},
      selected: false,
      declared: true,
      mode: null,
      expected: [],
    },
    {
      name: 'no targets with explicit declarations in type-check',
      parserOptions: {},
      selected: false,
      declared: true,
      mode: '--type-check-only',
      expected: ['explicit-error.ts'],
    },
  ])(
    'project scope preserves $name',
    async ({ parserOptions, selected, declared, mode, expected }) => {
      const root = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-cli-project-scope-'),
      );
      try {
        await writeFixture(root, {
          'rslint.config.mjs': `export default ${JSON.stringify([
            {
              ignores: selected
                ? ['default-error.ts', 'explicit-error.ts', 'rslint.config.mjs']
                : ['**'],
            },
            ...(declared
              ? [
                  {
                    files: ['unused.ts'],
                    languageOptions: {
                      parserOptions: { project: './declared.json' },
                    },
                  },
                ]
              : []),
            {
              files: selected ? ['target.ts'] : ['absent/**'],
              languageOptions: { parserOptions },
              rules: { 'no-debugger': 'error' },
            },
          ])};`,
          'tsconfig.json': JSON.stringify({
            files: ['target.ts', 'default-error.ts'],
          }),
          'declared.json': JSON.stringify({
            files: ['target.ts', 'explicit-error.ts'],
          }),
          'target.ts': 'debugger;\nexport {};\n',
          'default-error.ts': 'export const broken: number = "default";\n',
          'explicit-error.ts': 'export const broken: number = "explicit";\n',
        });
        const result = await runCLI(root, [...(mode ? [mode] : []), '.']);
        expect(result.code).toBe(expected.length === 0 ? 0 : 1);
        expect(result.stderr).toBe('');
        expect(
          normalizedDiagnostics(root, parseDiagnostics(result.stdout)),
        ).toEqual(
          normalizedDiagnostics(
            root,
            expected.map((filePath) => ({
              filePath,
              ruleName: 'TypeScript(TS2322)',
            })),
          ),
        );
      } finally {
        await cleanupTempDir(root);
      }
    },
  );

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
      await cleanupTempDir(root);
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
      await cleanupTempDir(root);
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
      await cleanupTempDir(root);
    }
  });
});

describe('runCLI process-tree teardown', () => {
  test.skipIf(process.platform === 'win32')(
    'terminates a detached POSIX descendant holding stdout open after its leader exits',
    async () => {
      const root = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-cli-group-kill-'),
      );
      let descendantPid;
      let postExitTimer;
      try {
        const fakeCli = path.join(root, 'fake-cli.mjs');
        await writeFile(
          fakeCli,
          `import { spawn } from 'node:child_process';
const pidFile = process.argv.at(-1);
const ignoreSigterm = process.argv.includes('--ignore-sigterm');
const child = spawn(
  process.execPath,
  [
    '-e',
    \`import('node:fs').then(({ writeFileSync, writeSync }) => {
      if (\${String(ignoreSigterm)}) process.on('SIGTERM', () => {});
      writeFileSync(\${JSON.stringify(pidFile)}, String(process.pid));
      writeSync(3, 'ready');
      setInterval(() => {}, 60_000);
    });\`,
  ],
  { stdio: ['ignore', 'inherit', 'inherit', 'pipe'] },
);
child.stdio[3].once('data', () => {
  process.exit(0);
});
`,
        );

        for (const [label, extraArgs] of [
          ['sigterm', []],
          ['sigkill', ['--ignore-sigterm']],
        ]) {
          const pidFile = path.join(root, `descendant-${label}.pid`);
          let leader;
          let leaderExitCode = null;
          await expect(
            runCLI(root, [...extraArgs, pidFile], {
              scriptPath: fakeCli,
              timeoutMs: 80,
              graceMs: 120,
              onSpawn: (child) => {
                leader = child;
              },
              setTimeoutImpl: (fn) => {
                const fallback = setTimeout(fn, 30_000);
                leader.once('exit', (code) => {
                  leaderExitCode = code;
                  clearTimeout(fallback);
                  postExitTimer = setTimeout(fn, 10);
                  postExitTimer.unref();
                });
                return fallback;
              },
            }),
          ).rejects.toThrow('rslint CLI timed out after 80ms');

          expect(leaderExitCode).toBe(0);
          descendantPid = Number(await readFile(pidFile, 'utf8'));
          expect(Number.isInteger(descendantPid)).toBe(true);
          let alive = true;
          for (let i = 0; i < 100; i++) {
            try {
              process.kill(descendantPid, 0);
            } catch (err) {
              if (err?.code === 'ESRCH') {
                alive = false;
                break;
              }
            }
            await new Promise((resolve) => setTimeout(resolve, 20));
          }
          expect(alive).toBe(false);
          descendantPid = undefined;
        }
      } finally {
        clearTimeout(postExitTimer);
        if (descendantPid !== undefined) {
          try {
            process.kill(descendantPid, 'SIGKILL');
          } catch {
            // Already reaped.
          }
        }
        await cleanupTempDir(root);
      }
    },
  );

  test('bounds a stalled Windows tree-kill helper so runCLI reports its timeout promptly', async () => {
    const root = await mkdtemp(path.join(os.tmpdir(), 'rslint-cli-win-stall-'));
    let cliChild;
    let stalledTreeKiller;
    try {
      const fakeCli = path.join(root, 'fake-win-cli.mjs');
      await writeFile(fakeCli, 'setInterval(() => {}, 120_000);\n');
      const customSpawn = (command, args, opts) => {
        // cspell:disable-next-line
        if (command === 'taskkill') {
          stalledTreeKiller = spawn(
            process.execPath,
            ['-e', 'setInterval(() => {}, 120_000);'],
            { stdio: 'ignore', windowsHide: true },
          );
          // Simulate the target CLI exiting and emitting `close` while tree-kill hangs.
          cliChild?.kill('SIGKILL');
          return stalledTreeKiller;
        }
        cliChild = spawn(command, args, opts);
        return cliChild;
      };

      const startedAt = Date.now();
      await expect(
        runCLI(root, [], {
          scriptPath: fakeCli,
          timeoutMs: 60,
          graceMs: 100,
          platform: 'win32',
          spawnImpl: customSpawn,
        }),
      ).rejects.toThrow('rslint CLI timed out after 60ms');

      expect(Date.now() - startedAt).toBeLessThan(30_000);
      expect(stalledTreeKiller).toBeDefined();
      expect(
        stalledTreeKiller.killed ||
          stalledTreeKiller.exitCode !== null ||
          stalledTreeKiller.signalCode !== null,
      ).toBe(true);
    } finally {
      for (const proc of [cliChild, stalledTreeKiller]) {
        if (proc && proc.exitCode === null && proc.signalCode === null) {
          const closed = new Promise((resolve) => proc.once('close', resolve));
          proc.kill('SIGKILL');
          await closed;
        }
      }
      await cleanupTempDir(root);
    }
  });
});
