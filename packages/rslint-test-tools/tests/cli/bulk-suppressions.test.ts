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

    const directory = path.join(cwd, 'rslint-suppressions.json');
    await fs.unlink(directory);
    await fs.mkdir(directory);
    expect((await check('--suppress-all')).exitCode).toBe(0);
    expect(await fs.readdir(directory)).toEqual([
      expect.stringMatching(/^rslint-suppressions_[a-z0-9]+$/),
    ]);
    expect(
      (await check('--suppressions-location', directory + path.sep)).exitCode,
    ).toBe(0);
  });

  test('counts by the reported target when aliases share a TypeScript source', async () => {
    const cwd = await project({
      'tsconfig.json': JSON.stringify({ files: ['real/file.ts'] }),
      'rslint.config.mjs': `export default [
        {languageOptions:{parserOptions:{projectService:false,project:['./tsconfig.json']}}},
        {files:['link/*.ts'],rules:{'no-console':['error',{allow:['warn']}]}},
        {files:['real/*.ts'],rules:{'no-console':['error',{allow:['log']}]}}
      ];`,
      'real/file.ts': 'console.log(1);\nconsole.warn(2);',
    });
    await fs.symlink(
      path.join(cwd, 'real'),
      path.join(cwd, 'link'),
      'junction',
    );
    const check = (...args: string[]) =>
      runRslint(['--format', 'jsonline', ...args], cwd);
    const baseline = path.join(cwd, 'rslint-suppressions.json');
    const names = ['real/file.ts', 'link/file.ts'];
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

    const recorded = await check('--suppress-all', 'link/file.ts');
    expect(recorded.exitCode, recorded.stderr).toBe(0);
    const saved = await fs.readFile(baseline, 'utf8');
    expect(JSON.parse(saved)).toEqual({
      'link/file.ts': { 'no-console': { count: 1 } },
    });

    const checked = await check('real/file.ts');
    expect(checked.exitCode, checked.stderr).toBe(1);
    expect(JSON.parse(checked.stdout)).toMatchObject({
      ruleName: 'no-console',
      range: { start: { line: 2 } },
    });
    await fs.writeFile(
      path.join(cwd, 'tsconfig.json'),
      JSON.stringify({ files: ['link/file.ts'] }),
    );
    const sameTarget = await check('link/file.ts');
    expect(sameTarget.exitCode, sameTarget.stderr).toBe(0);
    await fs.writeFile(path.join(cwd, 'real/file.ts'), 'console.log(1);');
    const pruned = await check('--prune-suppressions', 'real/file.ts');
    expect(pruned.exitCode, pruned.stderr).toBe(0);
    expect(await fs.readFile(baseline, 'utf8')).toBe(saved);
  });

  test('resolves the source and baseline from the same symlinked working directory', async () => {
    const temp = await project({
      'rslint.config.mjs':
        "export default [{rules:{'no-debugger':'error','no-var':'error'}}];",
      'work/empty.js': '',
      'sibling.js': 'var physical = 1;',
      'nested/sibling.js': 'var logical = 1; debugger;',
      'base.json': '{}',
    });
    const root = await fs.realpath(temp);
    const nested = path.join(root, 'nested');
    const cwd = path.join(nested, 'link');
    await fs.symlink(path.join(root, 'work'), cwd, 'junction');
    const result = await runCliProcess(
      process.execPath,
      [
        RSLINT_BIN,
        '--fix',
        '--suppress-all',
        '--suppressions-location',
        '../base.json',
        '../sibling.js',
      ],
      { cwd, env: { ...process.env, PWD: cwd } },
    );
    expect(result.exitCode, result.stderr).toBe(0);
    expect(await fs.readFile(path.join(nested, 'sibling.js'), 'utf8')).toBe(
      'let logical = 1; debugger;',
    );
    expect(await fs.readFile(path.join(root, 'sibling.js'), 'utf8')).toBe(
      'var physical = 1;',
    );
    expect(
      JSON.parse(await fs.readFile(path.join(nested, 'base.json'), 'utf8')),
    ).toEqual({
      '../nested/sibling.js': { 'no-debugger': { count: 1 } },
    });
    expect(await fs.readFile(path.join(root, 'base.json'), 'utf8')).toBe('{}');
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
    expect(await fs.readFile(path.join(cwd, 'ok.js'), 'utf8')).toBe(
      'var a = 1;',
    );
  });

  test('updates verified rules when a plugin rule or file fails', async () => {
    const ruleCounts = {
      'no-debugger': { count: 1 },
      'local/healthy': { count: 1 },
      'local/fails': { count: 2 },
    };
    const fileCounts = { 'no-debugger': { count: 2 } };
    const cwd = await project({
      'rule.js': 'debugger;',
      'file.js': 'debugger;',
      'healthy.js': 'const value = 1;',
      'local.mjs': `export default {rules:{
        healthy: {create(ctx) { return {Program(node) { ctx.report({node, message:'healthy finding'}); }}; }},
        fails: {create(ctx) {
          if (ctx.filename.endsWith('file.js')) {
            return {get Program() { throw Error('file failure'); }};
          }
          return {Program(node) {
            ctx.report({node, message:'finding before failure'});
            throw Error('rule failure');
          }};
        }}
      }};`,
      'rslint.config.mjs': `import local from './local.mjs'; export default [
        {files:['rule.js','file.js'],plugins:{local},rules:{
          'no-debugger':'error','local/healthy':'error','local/fails':'error'
        }},
        {files:['healthy.js'],rules:{'no-var':'error'}}
      ];`,
      'rslint-suppressions.json': JSON.stringify({
        'rule.js': ruleCounts,
        'file.js': fileCounts,
        'healthy.js': { 'no-var': { count: 1 } },
      }),
    });
    const check = (...args: string[]) =>
      runRslint(['--format', 'jsonline', ...args], cwd);
    const rule = await check('rule.js');
    expect(rule.exitCode, rule.stderr).toBe(1);
    expect(rule.stdout).toContain('rule failure');
    expect(rule.stdout).toContain('finding before failure');
    expect(rule.stdout).not.toContain('healthy finding');
    expect(rule.stdout).not.toContain('no-debugger');

    const prune = await check(
      '--prune-suppressions',
      'rule.js',
      'file.js',
      'healthy.js',
    );
    expect(prune.exitCode).toBe(2);
    expect(prune.stdout).toContain('file failure');
    expect(await readBaseline(cwd)).toEqual({
      'rule.js': ruleCounts,
      'file.js': fileCounts,
    });

    await fs.writeFile(
      path.join(cwd, 'file.js'),
      'debugger; debugger; debugger;',
    );
    await fs.writeFile(path.join(cwd, 'healthy.js'), 'var value = 1;');
    const all = await check('--suppress-all', 'file.js', 'healthy.js');
    expect(all.exitCode).toBe(2);
    expect(all.stdout).toContain('no-debugger');
    expect(await readBaseline(cwd)).toEqual({
      'rule.js': ruleCounts,
      'file.js': fileCounts,
      'healthy.js': { 'no-var': { count: 1 } },
    });

    const named = await check('--suppress-rule', 'no-debugger', 'file.js');
    expect(named.exitCode).toBe(2);
    expect(named.stdout).toContain('file failure');
    expect(named.stdout).not.toContain('no-debugger');
    expect((await readBaseline(cwd))['file.js']).toEqual({
      'no-debugger': { count: 3 },
    });
  });

  test('refuses to update suppressions when a discovered config fails to load', async () => {
    const count = { 'no-console': { count: 1 } };
    const cwd = await project({
      'rslint.config.mjs':
        "export default [{ignores:['nested/skip.js']},{rules:{'no-console':['error',{allow:['log']}]}}];",
      'nested/rslint.config.mjs':
        "export default [{rules:{'no-console':['error',{allow:['warn']}]}}];",
      'nested/a.js': 'console.log(1);\nconsole.warn(2);',
      'nested/skip.js': 'console.log(1);',
      'nested/ok/rslint.config.mjs':
        "export default [{rules:{'no-console':'error'}}];",
      'nested/ok/b.js': 'console.log(1);',
    });
    const record = await runRslint(['--suppress-all', 'nested/'], cwd);
    expect(record.exitCode, record.stderr).toBe(0);
    expect(await readBaseline(cwd)).toEqual({
      'nested/a.js': count,
      'nested/skip.js': count,
      'nested/ok/b.js': count,
    });
    await fs.writeFile(
      path.join(cwd, 'nested/rslint.config.mjs'),
      'throw Error("broken config");',
    );
    await fs.writeFile(path.join(cwd, 'nested/ok/b.js'), '');
    const prune = await runRslint(
      ['--prune-suppressions', '--format', 'jsonline', 'nested/'],
      cwd,
    );
    expect(prune.exitCode).toBe(2);
    expect(prune.stderr).toContain('configuration failed');
    expect(JSON.parse(prune.stdout)).toMatchObject({
      filePath: 'nested/a.js',
      ruleName: 'no-console',
      range: { start: { line: 2 } },
    });
    expect(await readBaseline(cwd)).toEqual({
      'nested/a.js': count,
      'nested/skip.js': count,
      'nested/ok/b.js': count,
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
