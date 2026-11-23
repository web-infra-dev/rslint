import { describe, test, expect, onTestFinished } from 'rstack/test';
import path from 'node:path';
import fs from 'node:fs/promises';
import { runCliProcess } from './spawn-cli.js';
import {
  cleanupTempDir,
  createTempDir,
  RSLINT_BIN,
  runRslint,
} from './js-config/helpers.js';

async function project(files: Record<string, string>) {
  const cwd = await createTempDir(files);
  onTestFinished(() => cleanupTempDir(cwd));
  return cwd;
}

async function readBaseline(dir: string) {
  return JSON.parse(
    await fs.readFile(path.join(dir, 'rslint-suppressions.json'), 'utf8'),
  );
}

describe('CLI bulk suppressions', () => {
  test('records and checks per-file error counts through the package CLI', async () => {
    const cwd = await project({
      'rslint.config.mjs':
        "export default [{rules: {'no-var':'error','no-debugger':'error','no-console':'warn'}}];",
      'a.js': 'var a = 1; debugger; console.log(a);',
      'eslint-suppressions.json': '{}',
    });
    const check = (...flags: string[]) =>
      runRslint(['a.js', '--format', 'jsonline', ...flags], cwd);
    expect(
      (await check('--suppress-rule', '[no-var,no-debugger]')).exitCode,
    ).toBe(0);
    expect(await readBaseline(cwd)).toEqual({
      'a.js': {
        'no-var': { count: 1 },
        'no-debugger': { count: 1 },
      },
    });
    expect(
      await fs.readFile(path.join(cwd, 'eslint-suppressions.json'), 'utf8'),
    ).toBe('{}');
    const warnings = await check('--max-warnings', '0');
    expect(warnings.exitCode).toBe(1);
    expect(JSON.parse(warnings.stdout)).toMatchObject({
      ruleName: 'no-console',
      severity: 'warn',
    });
    await fs.writeFile(path.join(cwd, 'a.js'), 'var a = 1;');
    const stale = await runRslint(
      ['a.js', '--no-color', '--suppressions-location', ''],
      cwd,
    );
    expect(stale.exitCode).toBe(2);
    expect(stale.stdout).toContain('error   Lint failed');
    expect(stale.stderr).toContain('--prune-suppressions');

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
    expect(await readBaseline(cwd)).toEqual({
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
  });

  test('keeps distinct counts for separately linted aliases', async () => {
    const cwd = await project({
      'rslint.config.mjs': `export default [
        {files:['link/*.js'],rules:{'no-console':['error',{allow:['warn']}]}},
        {files:['real/*.js'],rules:{'no-console':['error',{allow:['log']}]}}
      ];`,
      'real/file.js': 'console.log(1);\nconsole.warn(2);',
    });
    await fs.symlink(
      path.join(cwd, 'real'),
      path.join(cwd, 'link'),
      'junction',
    );
    const check = (...args: string[]) =>
      runRslint(['--format', 'jsonline', ...args], cwd);
    const baseline = path.join(cwd, 'rslint-suppressions.json');
    const names = ['real/file.js', 'link/file.js'];
    const linted = await check(...names);
    expect(linted.exitCode, linted.stderr).toBe(1);
    const selected = JSON.parse(linted.stdout).filePath;
    expect(names).toContain(selected);
    const together = await check('--suppress-all', ...names);
    expect(together.exitCode, together.stderr).toBe(0);
    expect(await readBaseline(cwd)).toEqual({
      [selected]: { 'no-console': { count: 1 } },
    });
    await fs.unlink(baseline);

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
  });

  test('keeps entries across working directory aliases', async () => {
    const temp = await project({
      'rslint.config.mjs': "export default [{rules:{'no-debugger':'error'}}];",
      'shared.js': 'debugger;',
      'work/src/app.js': 'debugger;',
    });
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
  });

  test('prunes after fixes and preserves files skipped by lint', async () => {
    const cwd = await project({
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
    const result = await runRslint(
      ['--fix', '--prune-suppressions', 'ok.js', 'broken.js', 'ignored.js'],
      cwd,
    );
    expect(result.exitCode).toBe(1);
    expect(await fs.readFile(path.join(cwd, 'ok.js'), 'utf8')).toContain(
      'let a',
    );
    expect(Object.keys(await readBaseline(cwd))).toEqual([
      'broken.js',
      'ignored.js',
    ]);

    await fs.writeFile(path.join(cwd, 'ok.js'), 'var a = 1;');
    await fs.writeFile(path.join(cwd, 'rslint-suppressions.json'), '{');
    const invalid = await runRslint(
      ['--fix', '--prune-suppressions', 'ok.js'],
      cwd,
    );
    expect(invalid.exitCode).toBe(2);
    expect(invalid.stderr).toContain('invalid suppressions');
    expect(await fs.readFile(path.join(cwd, 'ok.js'), 'utf8')).toContain(
      'let a',
    );
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
    const cwd = await project({
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
  });

  test('preserves recorded counts when a nested config fails to load', async () => {
    const saved = { 'nested/a.js': { 'no-var': { count: 1 } } };
    const cwd = await project({
      'rslint.config.mjs': "export default [{rules:{'no-debugger':'error'}}];",
      'nested/rslint.config.mjs': 'throw Error("broken config");',
      'nested/a.js': 'debugger;',
      'rslint-suppressions.json': JSON.stringify(saved),
    });
    const result = await runRslint(
      ['--prune-suppressions', 'nested/a.js'],
      cwd,
    );
    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain('configuration failed');
    expect(await readBaseline(cwd)).toEqual(saved);
  });

  test('preserves a type-aware count when the file is outside its TypeScript project', async () => {
    const rule = '@typescript-eslint/await-thenable';
    const cwd = await project({
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
    const update = await runRslint(
      ['--fix', '--prune-suppressions', 'a.ts'],
      cwd,
    );
    expect(update.exitCode).toBe(0);
    expect(await readBaseline(cwd)).toEqual({
      'a.ts': { [rule]: { count: 1 } },
    });
    expect((await runRslint(['a.ts'], cwd)).exitCode).toBe(0);
  });
});
