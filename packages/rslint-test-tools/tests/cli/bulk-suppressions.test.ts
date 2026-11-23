import { describe, test, expect } from 'rstack/test';
import path from 'node:path';
import { existsSync } from 'node:fs';
import fs from 'node:fs/promises';
import { runCliProcess } from './spawn-cli.js';
import {
  cleanupTempDir,
  createTempDir,
  RSLINT_BIN,
  runRslint,
} from './js-config/helpers.js';

describe('CLI bulk suppressions', () => {
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
        const cwd = path.join(parent, 'LINKS', 'WO\u0301RK', 'STRASSE');
        const relative = 'src/f.js';
        const recorded = await check(cwd, '--suppress-all', relative);
        expect(recorded.exitCode, recorded.stderr).toBe(0);
        const json = await fs.readFile(
          path.join(cwd, 'rslint-suppressions.json'),
          'utf8',
        );
        expect(JSON.parse(json)).toEqual({
          [relative]: { 'no-debugger': { count: 2 } },
        });
        const checked = await check(path.join(project, 'Straße'), relative);
        expect(checked.exitCode, checked.stderr).toBe(0);
        const other = await check(cwd, 'src/F.js');
        expect(other.exitCode, other.stderr).toBe(1);
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
      'rslint.config.mjs': `export default [
        {rules:{'no-empty':'error','no-var':'error'}},
        {files:['other.js'],rules:{'no-debugger':'error'}},
        {files:['src/*.js'],rules:{'no-console':'error'}},
        {files:['entry.js'],rules:{'no-alert':'error'}}
      ];`,
      'src/app.js': '',
      'other.js': 'debugger;',
    });
    try {
      const alternate = existsSync(path.join(cwd, 'SRC')) ? 'SRC' : 'src';
      try {
        await fs.symlink(
          path.join(alternate, 'app.js'),
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
      const record = await check('--suppress-rule', 'no-debugger', '.');
      expect(record.exitCode, record.stderr).toBe(0);
      expect(JSON.parse(await fs.readFile(baseline, 'utf8'))).toEqual({
        'other.js': { 'no-debugger': { count: 1 } },
      });
      for (const rule of ['no-console', 'no-alert']) {
        const conflict = await check('--suppress-rule', rule, '.');
        expect(conflict.exitCode, conflict.stderr).toBe(2);
        expect(conflict.stderr).toContain('entry.js');
      }
      expect((await check('.')).exitCode).toBe(0);
      expect((await check('--suppress-rule', '[]', '.')).exitCode).toBe(0);
      expect((await check('--suppress-all', '.')).exitCode).toBe(2);

      await fs.writeFile(app, 'if (a) {}');
      expect(
        (await check('--suppress-all', 'src/app.js', './src/app.js')).exitCode,
      ).toBe(0);
      await fs.writeFile(app, 'if (a) {} if (b) {}');
      expect((await check('--suppress-all', 'entry.js')).exitCode).toBe(0);
      expect((await check('src/app.js')).exitCode).toBe(1);
      expect((await check('entry.js')).exitCode).toBe(0);
      if (alternate !== 'src') {
        for (const args of [
          ['SRC', 'src'],
          ['src/*.js', 'SRC/*.js'],
        ]) {
          const result = await check(...args);
          expect(result.exitCode, result.stderr).toBe(2);
          expect(result.stderr).toContain('SRC/app.js');
          expect(result.stderr).toContain('src/app.js');
        }
      }
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

  test('keeps entries across working directory aliases', async () => {
    const temp = await createTempDir({
      'rslint.config.mjs': "export default [{rules:{'no-debugger':'error'}}];",
      'shared.js': 'debugger;',
      'work/clean.js': '',
      'work/src/app.js': 'debugger;',
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
      const fromAlias = (...args: string[]) =>
        runCliProcess(process.execPath, [RSLINT_BIN, ...args], {
          cwd,
          env: { ...process.env, PWD: cwd },
        });
      const recorded = await fromAlias('--suppress-all', shared, 'src/app.js');
      expect(recorded.exitCode, recorded.stderr).toBe(0);
      const saved = await fs.readFile(baseline, 'utf8');
      expect(JSON.parse(saved)).toEqual({
        '../shared.js': { 'no-debugger': { count: 1 } },
        'src/app.js': { 'no-debugger': { count: 1 } },
      });

      const physical = await runRslint(['src/app.js', shared], work);
      expect(physical.exitCode, physical.stderr).toBe(0);
      for (const args of [[shared], ['--prune-suppressions', 'clean.js']]) {
        const result = await fromAlias(...args);
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

  test('preserves source and counts when a config redirects a selected directory', async () => {
    const source = 'export function f() { var x = 1; return x; }';
    const baseline = '{"current/file.js":{"no-var":{"count":1}}}';
    const cwd = await createTempDir({
      'current/file.js': source,
      'ignored/file.js': source,
      '.gitignore': '/ignored/\n',
      'rslint-suppressions.json': baseline,
      'rslint.config.mjs': `
        import fs from 'node:fs';
        import path from 'node:path';
        const at = name => path.join(import.meta.dirname, name);
        if (!fs.existsSync(at('saved'))) {
          fs.renameSync(at('current'), at('saved'));
          fs.symlinkSync(at('ignored'), at('current'), 'junction');
        }
        export default [{ rules: { 'no-var': 'error' } }];
      `,
    });
    try {
      const result = await runRslint(
        ['--fix', '--prune-suppressions', 'current/file.js'],
        cwd,
      );
      expect(result.exitCode).not.toBe(0);
      expect(result.stderr).toContain('changed after lint generation');
      expect(await fs.readFile(path.join(cwd, 'ignored/file.js'), 'utf8')).toBe(
        source,
      );
      expect(
        await fs.readFile(path.join(cwd, 'rslint-suppressions.json'), 'utf8'),
      ).toBe(baseline);
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
});
