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
    expect(stale.stdout).toContain('Lint failed');
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

  test('keeps the physical target when pruning through a working directory alias', async () => {
    const temp = await project({
      'rslint.config.mjs': `export default [
        {rules:{'no-debugger':'error'}},
        {files:['nested/link/**/*.js'],rules:{'no-debugger':'off'}}
      ];`,
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
    const physicalFile = path.join(work, 'src/app.js');
    const baseline = path.join(work, 'rslint-suppressions.json');
    const fromAlias = (...args: string[]) =>
      runCliProcess(process.execPath, [RSLINT_BIN, ...args], {
        cwd,
        env: { ...process.env, PWD: cwd },
      });
    expect((await fromAlias('src/app.js')).exitCode).toBe(1);
    const recorded = await fromAlias('--suppress-all', shared, '.');
    expect(recorded.exitCode, recorded.stderr).toBe(0);
    const saved = await fs.readFile(baseline, 'utf8');
    expect(JSON.parse(saved)).toEqual({
      '../shared.js': { 'no-debugger': { count: 1 } },
      'src/app.js': { 'no-debugger': { count: 1 } },
    });

    const pruned = await fromAlias(
      '--prune-suppressions',
      path.join(cwd, 'src/app.js'),
    );
    expect(pruned.exitCode, pruned.stderr).toBe(0);
    expect(await fs.readFile(baseline, 'utf8')).toBe(saved);
    const physical = await runRslint([physicalFile, shared], work);
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

  test('keeps a failed plugin rule visible and updates independent counts', async () => {
    const config =
      "import local from './local.mjs'; export default [{files:['a.js'],plugins:{local},rules:{'no-debugger':'error','local/healthy':'error','local/throws':'error'}},{files:['fixed.js'],rules:{'no-var':'error'}}];";
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
      'rslint.config.mjs': config,
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
    expect(write.exitCode).toBe(1);
    expect(write.stdout).toContain('finding before failure');
    expect(await readBaseline(cwd)).toEqual({
      'a.js': JSON.parse(saved)['a.js'],
    });
    // A warning-level plugin failure in a.js must not block recording fixed.js.
    await fs.writeFile(
      path.join(cwd, 'rslint.config.mjs'),
      config.replace("'local/throws':'error'", "'local/throws':'warn'"),
    );
    await fs.writeFile(path.join(cwd, 'fixed.js'), 'var value = 1;');
    const record = await runRslint(
      [
        '--suppress-rule',
        'no-var',
        '--pass-on-unpruned-suppressions',
        ...args,
        'fixed.js',
      ],
      cwd,
    );
    expect(record.exitCode).toBe(1);
    expect((await readBaseline(cwd))['fixed.js']).toEqual({
      'no-var': { count: 1 },
    });

    // A file-level plugin failure still leaves the selected native rule count available.
    await fs.writeFile(path.join(cwd, 'a.js'), 'debugger; debugger;');
    await fs.writeFile(
      path.join(cwd, 'local.mjs'),
      `export default { rules: { healthy: { create() { return {}; } },
        throws: { create() { return { get Program() { throw Error('listener getter failed'); } }; } } } };`,
    );
    const partial = await runRslint(
      ['--suppress-rule', 'no-debugger', ...args],
      cwd,
    );
    expect(partial.exitCode).toBe(2);
    expect(partial.stdout).toContain('listener getter failed');
    expect((await readBaseline(cwd))['a.js']).toEqual({
      ...JSON.parse(saved)['a.js'],
      'no-debugger': { count: 2 },
    });
  });

  test('preserves recorded counts when a nested config fails to load', async () => {
    const count = { 'no-console': { count: 1 } };
    const cwd = await project({
      'rslint.config.mjs':
        "export default [{rules:{'no-console':['error',{allow:['log']}]}}];",
      'nested/rslint.config.mjs':
        "export default [{rules:{'no-console':['error',{allow:['warn']}]}}];",
      'nested/a.js': 'console.log(1);\nconsole.warn(2);',
      'nested/ok/rslint.config.mjs':
        "export default [{rules:{'no-console':'error'}}];",
      'nested/ok/b.js': 'console.log(1);',
    });
    const check = (...flags: string[]) =>
      runRslint(['--format', 'jsonline', ...flags, '.'], cwd);
    expect((await check('--suppress-all')).exitCode).toBe(0);
    expect(await readBaseline(cwd)).toEqual({
      'nested/a.js': count,
      'nested/ok/b.js': count,
    });
    await fs.writeFile(
      path.join(cwd, 'nested/rslint.config.mjs'),
      'throw Error("broken config");',
    );
    const result = await check('--pass-on-unpruned-suppressions');
    expect(result.exitCode).toBe(2);
    expect(result.stderr).toContain('configuration failed');
    expect(JSON.parse(result.stdout)).toMatchObject({
      filePath: 'nested/a.js',
      ruleName: 'no-console',
      range: { start: { line: 2 } },
    });
    await fs.writeFile(path.join(cwd, 'nested/ok/b.js'), '');
    expect((await check('--prune-suppressions')).exitCode).toBe(2);
    expect(await readBaseline(cwd)).toEqual({ 'nested/a.js': count });

    // Without the root config, nested/a.js has no owner to fall back to.
    await fs.unlink(path.join(cwd, 'rslint.config.mjs'));
    await fs.unlink(path.join(cwd, 'rslint-suppressions.json'));
    expect((await check()).exitCode).toBe(0);
    expect((await check('--prune-suppressions')).exitCode).toBe(2);
    expect(await readBaseline(cwd)).toEqual({});
    expect((await check('--pass-on-unpruned-suppressions')).exitCode).toBe(2);
    await fs.writeFile(path.join(cwd, 'nested/ok/b.js'), 'console.log(1);');
    expect((await check('--suppress-all')).exitCode).toBe(2);
    expect(await readBaseline(cwd)).toEqual({ 'nested/ok/b.js': count });
  });

  test('keeps directory-scan counts when a literal file tries a different config', async () => {
    const cwd = await project({
      'rslint.config.mjs':
        "export default [{ignores:['sub/rslint.config.js']},{rules:{'no-console':'error','no-debugger':'error'}}];",
      'sub/rslint.config.js': 'throw Error("broken config");',
      'sub/rslint.config.mjs':
        "export default [{rules:{'no-debugger':'error'}}];",
      'sub/a.js': 'console.log(1);',
      'sub/b.js': 'debugger;',
    });
    const check = () =>
      runRslint(
        ['--suppress-all', '--format', 'jsonline', '.', 'sub/a.js'],
        cwd,
      );
    const result = await check();
    expect(result.exitCode).toBe(2);
    expect(JSON.parse(result.stdout)).toMatchObject({
      filePath: 'sub/a.js',
      ruleName: 'no-console',
    });
    expect(await readBaseline(cwd)).toEqual({
      'sub/b.js': { 'no-debugger': { count: 1 } },
    });
    await fs.unlink(path.join(cwd, 'sub/rslint.config.mjs'));
    expect(JSON.parse((await check()).stdout)).toMatchObject({
      filePath: 'sub/a.js',
      ruleName: 'no-console',
    });
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
