import { describe, test, expect } from 'rstack/test';
import path from 'node:path';
import { existsSync } from 'node:fs';
import fs from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { runCliProcess, type CliProcessResult } from './spawn-cli.js';
const RSLINT_BIN = require.resolve('@rslint/core/bin');

/**
 * Helper function to run rslint CLI command
 */
async function runRslint(
  args: string[],
  cwd?: string,
): Promise<CliProcessResult> {
  return runCliProcess(process.execPath, [RSLINT_BIN, ...args], {
    cwd: cwd || process.cwd(),
  });
}

/**
 * Create a temporary directory with test files
 */
async function createTempDir(files: Record<string, string>): Promise<string> {
  const tempDir = await fs.mkdtemp(path.join(tmpdir(), 'rslint-test-'));

  for (const [filePath, content] of Object.entries(files)) {
    const fullPath = path.join(tempDir, filePath);
    await fs.mkdir(path.dirname(fullPath), { recursive: true });
    await fs.writeFile(fullPath, content, 'utf8');
  }

  return tempDir;
}

/**
 * Cleanup temporary directory
 */
async function cleanupTempDir(tempDir: string): Promise<void> {
  await fs.rm(tempDir, { recursive: true, force: true });
}

describe('CLI Configuration Tests', () => {
  test('records and checks per-file error counts through the package CLI', async () => {
    const cwd = await createTempDir({
      'rslint.config.mjs':
        "export default [{rules: {'no-var':'error','no-debugger':'error','no-console':'error'}}];",
      'a.js': 'var a = 1; debugger; console.log(a);',
      'eslint-suppressions.json': '{}',
    });
    const check = (...flags: string[]) =>
      runRslint(['a.js', '--format', 'jsonline', ...flags], cwd);
    try {
      expect((await check('--suppress-rule')).exitCode).toBe(2);
      expect((await check('--suppress-all=false')).exitCode).toBe(1);
      await expect(
        fs.access(path.join(cwd, 'rslint-suppressions.json')),
      ).rejects.toThrow();
      const baseline = path.join(cwd, 'rslint-suppressions.json');
      expect((await check('--suppress-rule', '[]')).exitCode).toBe(1);
      expect(JSON.parse(await fs.readFile(baseline, 'utf8'))).toEqual({});
      expect(
        (
          await check(
            '--suppress-rule',
            'no-var',
            '--suppress-rule',
            '[no-debugger, no-console]',
          )
        ).exitCode,
      ).toBe(0);
      expect(JSON.parse(await fs.readFile(baseline, 'utf8'))).toEqual({
        'a.js': {
          'no-var': { count: 1 },
          'no-debugger': { count: 1 },
          'no-console': { count: 1 },
        },
      });
      expect(
        await fs.readFile(path.join(cwd, 'eslint-suppressions.json'), 'utf8'),
      ).toBe('{}');
      await fs.writeFile(path.join(cwd, 'a.js'), 'var a = 1; var b = 2;');
      const exceeded = await check();
      expect(exceeded.exitCode).toBe(2);
      expect(
        exceeded.stdout
          .trim()
          .split('\n')
          .map((line) => JSON.parse(line).ruleName),
      ).toEqual(['no-var', 'no-var']);
      expect(exceeded.stderr).toContain('--prune-suppressions');
      expect((await check('--pass-on-unpruned-suppressions')).exitCode).toBe(1);
      expect((await check('--prune-suppressions')).exitCode).toBe(1);
      expect(JSON.parse(await fs.readFile(baseline, 'utf8'))).toEqual({
        'a.js': { 'no-var': { count: 1 } },
      });

      const directory = path.join(cwd, 'baselines');
      await fs.mkdir(directory);
      expect(
        (await check('--suppressions-location', directory, '--suppress-all'))
          .exitCode,
      ).toBe(0);
      expect(await fs.readdir(directory)).toEqual([
        expect.stringMatching(/^rslint-suppressions_[a-z0-9]+$/),
      ]);
      expect(
        (await check('--suppressions-location', directory + path.sep)).exitCode,
      ).toBe(0);
    } finally {
      await cleanupTempDir(cwd);
    }
  });

  test.skipIf(process.platform !== 'darwin')(
    'uses the on-disk working directory spelling for relative files',
    async ({ skip }) => {
      const temp = await createTempDir({
        'Project/Straße/rslint.config.mjs':
          "export default [{rules:{'no-debugger':'error'}}];",
        'Project/Straße/src/f.js': 'debugger; debugger;',
      });
      try {
        const parent = await fs.realpath(temp);
        const project = path.join(parent, 'Project');
        if (!existsSync(path.join(parent, 'PROJECT', 'STRASSE'))) skip();
        const links = path.join(parent, 'Links');
        const work = path.join(links, 'Wórk');
        await fs.mkdir(links);
        await fs.symlink(project, work, 'dir');
        const check = (cwd: string, ...args: string[]) =>
          runCliProcess(process.execPath, [RSLINT_BIN, ...args], {
            cwd,
            env: { ...process.env, PWD: cwd },
          });
        const counts = { 'no-debugger': { count: 2 } };
        const expected = { 'src/f.js': counts };
        for (const [actual, alternate, entries] of [
          [project, path.join(parent, 'PROJECT'), expected],
          [
            work,
            path.join(parent, 'LINKS', 'WO\u0301RK'),
            {
              ...expected,
              '../../Links/Wórk/Straße/src/f.js': counts,
            },
          ],
        ] as const) {
          const directory = path.join(actual, 'Straße');
          const cwd = path.join(alternate, 'STRASSE');
          const relative = 'src/f.js';
          const recorded = await check(
            cwd,
            '--suppress-all',
            path.join(directory, relative),
          );
          expect(recorded.exitCode, recorded.stderr).toBe(0);
          const json = await fs.readFile(
            path.join(cwd, 'rslint-suppressions.json'),
            'utf8',
          );
          expect(JSON.parse(json)).toEqual(entries);
          for (const invocation of [directory, cwd]) {
            const checked = await check(invocation, relative);
            expect(checked.exitCode, checked.stderr).toBe(0);
          }
          const otherFileName = await check(
            cwd,
            path.join(directory, 'src/F.js'),
          );
          expect(otherFileName.exitCode, otherFileName.stderr).toBe(1);
        }
      } finally {
        await cleanupTempDir(temp);
      }
    },
  );

  test('reports real findings when different file names overlap', async () => {
    const temp = await createTempDir({
      'rslint.config.mjs':
        "export default [{files:['**/*.js'],rules:{'no-debugger':'error'}}];",
      'repo/a/clean.js': '',
      'repo/b/file.js': 'debugger; debugger;',
    });
    try {
      const parent = await fs.realpath(temp);
      const repo = path.join(parent, 'repo');
      const link = path.join(parent, 'link');
      await fs.symlink(repo, link, 'junction');
      const cwd = path.join(link, 'a');
      const relative = '../b/file.js';
      const logical = path.join(link, 'b', 'file.js');
      const physical = path.join(repo, 'b', 'file.js');
      const paths = [relative, logical, physical];
      const reversed = [...paths].reverse();
      const baseline = path.join(cwd, 'rslint-suppressions.json');
      const check = (...args: string[]) =>
        runCliProcess(
          process.execPath,
          [RSLINT_BIN, '--format', 'jsonline', ...args],
          { cwd, env: { ...process.env, PWD: cwd } },
        );

      const selected = '../../link/b/file.js';
      const two = { 'no-debugger': { count: 2 } };
      const other = { [relative]: two };
      const seeded = JSON.stringify({
        [selected]: { 'no-debugger': { count: 1 } },
        ...other,
      });
      await fs.writeFile(baseline, seeded);
      for (const args of [[logical], [relative, logical], paths, reversed]) {
        const result = await check(...args);
        expect(result.exitCode, result.stderr).not.toBe(0);
        if (args.length === 3) {
          expect(result.exitCode).toBe(2);
          expect(result.stderr).toContain('same file');
        } else if (args.length === 1 || process.platform !== 'win32') {
          expect(result.exitCode).toBe(1);
        }
        expect(
          result.stdout
            .trim()
            .split('\n')
            .map((line) => JSON.parse(line).ruleName),
        ).toEqual(['no-debugger', 'no-debugger']);
      }
      expect(await fs.readFile(baseline, 'utf8')).toBe(seeded);

      await fs.rm(baseline);
      const recorded = await check('--suppress-all', ...paths);
      expect(recorded.exitCode, recorded.stderr).toBe(2);
      expect(existsSync(baseline)).toBe(false);
      expect((await check('--suppress-all', logical)).exitCode).toBe(0);
      expect(JSON.parse(await fs.readFile(baseline, 'utf8'))).toEqual({
        [selected]: two,
      });

      await fs.writeFile(baseline, seeded);
      await fs.writeFile(physical, '');
      const pruned = await check('--prune-suppressions', ...reversed);
      expect(pruned.exitCode, pruned.stderr).toBe(2);
      expect(await fs.readFile(baseline, 'utf8')).toBe(seeded);
      expect((await check('--prune-suppressions', logical)).exitCode).toBe(0);
      expect(JSON.parse(await fs.readFile(baseline, 'utf8'))).toEqual(other);
    } finally {
      await cleanupTempDir(temp);
    }
  });

  test('verifies suppression coverage during project scans', async ({
    skip,
  }) => {
    const cwd = await createTempDir({
      'rslint.config.mjs':
        "export default [{rules:{'no-debugger':'error','no-var':'error'}}];",
      'src/app.js': '',
      'other.js': 'debugger;',
    });
    try {
      try {
        await fs.symlink(
          path.join('src', 'app.js'),
          path.join(cwd, 'entry.js'),
          'file',
        );
      } catch (error) {
        if (
          process.platform === 'win32' &&
          (error as NodeJS.ErrnoException).code === 'EPERM'
        )
          skip();
        throw error;
      }
      const check = (...args: string[]) =>
        runRslint(['--format', 'jsonline', ...args], cwd);
      const app = path.join(cwd, 'src/app.js');
      const baseline = path.join(cwd, 'rslint-suppressions.json');
      expect((await check('--suppress-all', 'other.js')).exitCode).toBe(0);
      expect((await check('.')).exitCode).toBe(0);
      expect((await check('--suppress-rule', '[]', '.')).exitCode).toBe(0);
      expect((await check('--suppress-all', '.')).exitCode).toBe(2);

      await fs.writeFile(app, 'debugger;');
      expect(
        (await check('--suppress-all', 'src/app.js', './src/app.js')).exitCode,
      ).toBe(0);
      await fs.writeFile(app, 'debugger; debugger;');
      expect((await check('--suppress-all', 'entry.js')).exitCode).toBe(0);
      expect((await check('src/app.js')).exitCode).toBe(1);
      expect((await check('entry.js')).exitCode).toBe(0);
      const scan = await check('--pass-on-unpruned-suppressions');
      expect(scan.exitCode, scan.stderr).toBe(2);
      expect(scan.stderr).toContain('entry.js');
      expect(scan.stderr).toContain('src/app.js');
      const saved = await fs.readFile(baseline, 'utf8');
      const fixable = 'var a = 1;';
      await fs.writeFile(app, fixable);
      expect((await check('--fix', '--prune-suppressions', '.')).exitCode).toBe(
        2,
      );
      expect(await fs.readFile(app, 'utf8')).toBe(fixable);
      expect(await fs.readFile(baseline, 'utf8')).toBe(saved);
    } finally {
      await cleanupTempDir(cwd);
    }
  });

  test('keeps distinct counts for separately linted aliases', async () => {
    const cwd = await createTempDir({
      'rslint.config.mjs': `export default [
        {files:['link/*.js'],rules:{'no-console':['error',{allow:['warn']}]}},
        {files:['real/*.js'],rules:{'no-console':['error',{allow:['log']}]}}
      ];`,
      'real/file.js': 'console.log(1);\nconsole.warn(2);',
    });
    try {
      await fs.symlink(
        path.join(cwd, 'real'),
        path.join(cwd, 'link'),
        'junction',
      );
      const check = (...args: string[]) =>
        runRslint(['--format', 'jsonline', ...args], cwd);
      const baseline = path.join(cwd, 'rslint-suppressions.json');
      const recorded = await check('--suppress-all', 'link/file.js');
      expect(recorded.exitCode, recorded.stderr).toBe(0);
      const saved = await fs.readFile(baseline, 'utf8');
      expect(JSON.parse(saved)).toEqual({
        'link/file.js': { 'no-console': { count: 1 } },
      });

      const checked = await check('real/file.js');
      expect(checked.exitCode, checked.stderr).toBe(1);
      expect(JSON.parse(checked.stdout)).toMatchObject({
        ruleName: 'no-console',
        range: { start: { line: 2 } },
      });
      await fs.writeFile(path.join(cwd, 'real/file.js'), 'console.log(1);');
      const pruned = await check('--prune-suppressions', 'real/file.js');
      expect(pruned.exitCode, pruned.stderr).toBe(0);
      expect(await fs.readFile(baseline, 'utf8')).toBe(saved);
    } finally {
      await cleanupTempDir(cwd);
    }
  });

  test('keeps external entries across working directory aliases', async () => {
    const temp = await createTempDir({
      'rslint.config.mjs': "export default [{rules:{'no-debugger':'error'}}];",
      'shared.js': 'debugger;',
      'work/clean.js': '',
    });
    try {
      const root = await fs.realpath(temp);
      const work = path.join(root, 'work');
      const nested = path.join(root, 'nested');
      await fs.mkdir(nested);
      const cwd = path.join(nested, 'link');
      await fs.symlink(work, cwd, 'junction');
      const shared = path.join(root, 'shared.js');
      const baseline = path.join(work, 'rslint-suppressions.json');
      const recorded = await runRslint(['--suppress-all', shared], work);
      expect(recorded.exitCode, recorded.stderr).toBe(0);
      const saved = await fs.readFile(baseline, 'utf8');
      expect(JSON.parse(saved)).toEqual({
        '../shared.js': { 'no-debugger': { count: 1 } },
      });

      for (const args of [[shared], ['--prune-suppressions', 'clean.js']]) {
        const result = await runCliProcess(
          process.execPath,
          [RSLINT_BIN, ...args],
          { cwd, env: { ...process.env, PWD: cwd } },
        );
        expect(result.exitCode, result.stderr).toBe(0);
      }
      expect(await fs.readFile(baseline, 'utf8')).toBe(saved);
    } finally {
      await cleanupTempDir(temp);
    }
  });

  test('prunes after fixes and preserves files skipped by lint', async () => {
    const cwd = await createTempDir({
      'rslint.config.mjs':
        "export default [{ignores:['ignored.js']},{rules:{'no-var':'error'}}];",
      'ok.js': 'var a = 1;',
      'broken.js': 'const = ;',
      'ignored.js': 'var b = 2;',
      'rslint-suppressions.json': JSON.stringify(
        Object.fromEntries(
          ['ok.js', 'broken.js', 'ignored.js'].map((file) => [
            file,
            { 'no-var': { count: 1 } },
          ]),
        ),
      ),
    });
    try {
      const result = await runRslint(
        ['--fix', '--prune-suppressions', 'ok.js', 'broken.js', 'ignored.js'],
        cwd,
      );
      expect(result.exitCode).toBe(1);
      expect(await fs.readFile(path.join(cwd, 'ok.js'), 'utf8')).toContain(
        'let a',
      );
      expect(
        Object.keys(
          JSON.parse(
            await fs.readFile(
              path.join(cwd, 'rslint-suppressions.json'),
              'utf8',
            ),
          ),
        ),
      ).toEqual(['broken.js', 'ignored.js']);
    } finally {
      await cleanupTempDir(cwd);
    }
  });

  test('keeps plugin failures visible and preserves the baseline on an incomplete update', async () => {
    const saved = JSON.stringify({
      'a.js': {
        'no-debugger': { count: 1 },
        'local/healthy': { count: 1 },
        'local/throws': { count: 2 },
      },
      'fixed.js': { 'no-var': { count: 1 } },
    });
    const cwd = await createTempDir({
      'a.js': 'debugger;',
      'fixed.js': 'const value = 1;',
      'local.mjs': `export default { rules: {
        healthy: { create(ctx) { return { Program(node) { ctx.report({ node, message: 'healthy finding' }); } }; } },
        throws: { create(ctx) { return { Program(node) {
          ctx.report({ node, message: 'finding before failure' });
          throw Error('fixture failure');
        } }; } },
      } };`,
      'rslint.config.mjs':
        "import local from './local.mjs'; export default [{files:['a.js'],plugins:{local},rules:{'no-debugger':'error','local/healthy':'error','local/throws':'error'}},{files:['fixed.js'],rules:{'no-var':'error'}}];",
      'rslint-suppressions.json': saved,
    });
    try {
      const args = ['--format', 'jsonline', 'a.js'];
      const read = await runRslint(args, cwd);
      expect(read.exitCode).toBe(1);
      expect(read.stdout).toContain('fixture failure');
      expect(read.stdout).toContain('finding before failure');
      expect(read.stdout).not.toContain('no-debugger');
      expect(read.stdout).not.toContain('healthy finding');
      const together = await runRslint([...args, 'fixed.js'], cwd);
      expect(together.exitCode).toBe(2);
      expect(together.stdout).toContain('fixture failure');
      expect(together.stderr).toContain('Unused suppressions');
      const write = await runRslint(
        ['--prune-suppressions', ...args, 'fixed.js'],
        cwd,
      );
      expect(write.exitCode).toBe(2);
      expect(write.stderr).toContain('incomplete');
      expect(
        await fs.readFile(path.join(cwd, 'rslint-suppressions.json'), 'utf8'),
      ).toBe(saved);
    } finally {
      await cleanupTempDir(cwd);
    }
  });

  test('preserves a type-aware count when the file is outside its TypeScript project', async () => {
    const rule = '@typescript-eslint/await-thenable';
    const cwd = await createTempDir({
      'rslint.config.mjs': `export default [{
        files: ['**/*.ts'], plugins: ['@typescript-eslint'],
        languageOptions: { parserOptions: { projectService: false, project: ['./tsconfig.json'] } },
        rules: { '${rule}': 'error', 'no-var': 'error' },
      }];`,
      'tsconfig.json': JSON.stringify({ files: ['other.ts'] }),
      'other.ts': 'export {};',
      'a.ts': 'async function f() { var value = 1; await value; }',
      'rslint-suppressions.json': JSON.stringify({
        'a.ts': { [rule]: { count: 1 }, 'no-var': { count: 1 } },
      }),
    });
    try {
      const update = await runRslint(
        ['--fix', '--prune-suppressions', 'a.ts'],
        cwd,
      );
      expect(update.exitCode).toBe(0);
      expect(
        JSON.parse(
          await fs.readFile(path.join(cwd, 'rslint-suppressions.json'), 'utf8'),
        ),
      ).toEqual({
        'a.ts': { [rule]: { count: 1 } },
      });
      expect((await runRslint(['a.ts'], cwd)).exitCode).toBe(0);
    } finally {
      await cleanupTempDir(cwd);
    }
  });

  test('should show help when --help flag is used', async () => {
    const result = await runRslint(['--help']);
    expect(result.exitCode).toBe(0);
  });

  test('should show help when -h flag is used', async () => {
    const result = await runRslint(['-h']);
    expect(result.exitCode).toBe(0);
  });

  test('should prioritize help over an invalid output format', async () => {
    const result = await runRslint(['--help', '--format', 'stylish']);
    expect(result.exitCode).toBe(0);
    expect(result.stderr).toContain('Usage:');
    expect(result.stderr).not.toContain('invalid output format');
  });

  test('should prioritize init over an invalid lint output format', async () => {
    const tempDir = await createTempDir({});
    try {
      const result = await runRslint(
        ['--init', '--format', 'stylish'],
        tempDir,
      );
      expect(result.exitCode).toBe(0);
      expect(result.stderr).not.toContain('invalid output format');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should reject an invalid output format before linting', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `throw new Error('config must not be evaluated');`,
    });
    try {
      const result = await runRslint(['--format', 'stylish'], tempDir);
      expect(result.exitCode).toBe(2);
      expect(result.stdout).toBe('');
      expect(result.stderr).toContain('invalid output format "stylish"');
      expect(result.stderr).not.toContain('config must not be evaluated');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should use default config when no config specified', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        language: 'javascript',
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20; // This should trigger no-unsafe-member-access
      `,
    });

    try {
      const result = await runRslint([], tempDir);

      // Should find and use the default module config.
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should use custom config when --config flag is specified', async () => {
    const tempDir = await createTempDir({
      'custom-config.mjs': `export default [${JSON.stringify({
        language: 'javascript',
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-unsafe-assignment': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = getValue();
        let b: string = a; // This should trigger no-unsafe-assignment
      `,
    });

    try {
      const result = await runRslint(
        ['--config', 'custom-config.mjs'],
        tempDir,
      );

      expect(result.stdout).toContain('no-unsafe-assignment');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should reject a legacy JSON config with migration guidance', async () => {
    const tempDir = await createTempDir({
      'legacy.jsonc': '[{ this is intentionally not valid JSONC',
      'test.ts': 'debugger;\n',
    });

    try {
      const result = await runRslint(['--config', 'legacy.jsonc'], tempDir);
      expect(result.exitCode).toBe(1);
      expect(result.stdout).toBe('');
      expect(result.stderr).toContain('unsupported config file');
      expect(result.stderr).toContain('JavaScript or TypeScript modules');
      expect(result.stderr).toContain('rslint --init');
      expect(result.stderr).not.toContain('error parsing');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should output in jsonline format when --format jsonline is used', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20;
      `,
    });

    try {
      const result = await runRslint(['--format', 'jsonline'], tempDir);

      // Should output valid JSON lines
      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((line) => line.trim());
      for (const line of lines) {
        // eslint-disable-next-line
        expect(() => JSON.parse(line)).not.toThrow();
      }
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should output in github workflow format when --format github is used', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-explicit-any': 'warn',
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test%,file.ts': `
        let a: any = 10;
        a.b = 20;
      `,
    });

    try {
      const result = await runRslint(['--format', 'github'], tempDir);

      const lines = result.stdout
        .trim()
        .split('\n')
        .filter((line) => line.trim());

      expect(lines.length).toBe(2);
      expect(lines[0]).toBe(
        '::warning file=test%25%2Cfile.ts,line=2,endLine=2,col=16,endColumn=19,title=@typescript-eslint/no-explicit-any::Unexpected any. Specify a different type.',
      );
      expect(lines[1]).toBe(
        '::error file=test%25%2Cfile.ts,line=3,endLine=3,col=11,endColumn=12,title=@typescript-eslint/no-unsafe-member-access::Unsafe member access .b on an `any` value.',
      );
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should output in gitlab code quality format when --format gitlab is used', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-explicit-any': 'warn',
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20;
      `,
    });

    try {
      const result = await runRslint(['--format', 'gitlab'], tempDir);

      const issues = JSON.parse(result.stdout);
      expect(Array.isArray(issues)).toBe(true);
      expect(issues).toHaveLength(2);

      expect(issues[0]).toMatchObject({
        check_name: '@typescript-eslint/no-explicit-any',
        severity: 'minor',
        location: {
          path: 'test.ts',
          lines: { begin: 2, end: 2 },
        },
      });
      expect(issues[1]).toMatchObject({
        check_name: '@typescript-eslint/no-unsafe-member-access',
        severity: 'major',
        location: {
          path: 'test.ts',
          lines: { begin: 3, end: 3 },
        },
      });

      // Fingerprints must be present and distinct.
      expect(issues[0].fingerprint).toBeTruthy();
      expect(issues[0].fingerprint).not.toBe(issues[1].fingerprint);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should output an empty array for --format gitlab with no diagnostics', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {},
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `export const a = 1;\n`,
    });

    try {
      const result = await runRslint(['--format', 'gitlab'], tempDir);
      expect(JSON.parse(result.stdout)).toEqual([]);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should only report errors when --quiet flag is used', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-unsafe-member-access': 'error',
          '@typescript-eslint/return-await': 'warn',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20;
        async function invalidInTryCatch1() {
        try {
          return Promise.reject('try');
        } catch (e) {
          // Doesn't execute due to missing await.
          }
        }
      `,
    });

    try {
      const result = await runRslint(['--quiet'], tempDir);
      // Should contain error reports but not verbose output
      expect(result.stdout).toContain('no-unsafe-member-access');
      // Should not contain summary information typically shown in verbose mode
      expect(result.stdout).not.toContain('return-await');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle an invalid module config gracefully', async () => {
    const tempDir = await createTempDir({
      'invalid.config.mjs': 'export default [INVALID SYNTAX',
      'test.ts': 'export const a = 1;',
    });

    try {
      const result = await runRslint(
        ['--config', 'invalid.config.mjs'],
        tempDir,
      );

      expect(result.exitCode).not.toBe(0);
      expect(result.stderr || result.stdout).toContain('failed to load config');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle non-existent config file gracefully', async () => {
    const tempDir = await createTempDir({
      'test.ts': 'export const a = 1;',
    });

    try {
      const result = await runRslint(
        ['--config', 'non-existent.json'],
        tempDir,
      );

      expect(result.exitCode).not.toBe(0);
      expect(result.stderr || result.stdout).toMatch(/config|file|not found/i);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should handle directory with no matching files', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
        },
        include: ['**/*.ts'],
      }),
      'readme.md': '# Test project',
      'package.json': '{"name": "test"}',
    });

    try {
      const result = await runRslint([], tempDir);

      expect(result.exitCode).toBe(0);
      // Should not report any errors since no files match the pattern
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit with non-zero code when linting errors are found', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        let a: any = 10;
        a.b = 20; // This should trigger an error
      `,
    });

    try {
      const result = await runRslint([], tempDir);

      expect(result.exitCode).not.toBe(0);
      expect(result.stdout).toContain('no-unsafe-member-access');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should exit with zero code when no linting errors are found', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': `export default [${JSON.stringify({
        files: ['**/*.ts'],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          'prefer-const': 'off',
          '@typescript-eslint/no-unsafe-member-access': 'error',
          'no-console': 'off',
        },
        plugins: ['@typescript-eslint'],
      })}];`,
      'tsconfig.json': JSON.stringify({
        compilerOptions: {
          target: 'ES2020',
          module: 'ESNext',
          strict: true,
        },
        include: ['**/*.ts'],
      }),
      'test.ts': `
        const a = "hello";
        console.log(a.length); // This is safe
      `,
    });

    try {
      const result = await runRslint([], tempDir);

      expect(result.exitCode).toBe(0);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});
