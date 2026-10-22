import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { afterEach, describe, expect, test } from 'rstack/test';
import { getNativeBinding } from '../src/native/binding.js';
import { runRustCLI } from '../src/rust/cli.js';
import { isRecord } from '../src/rust/engine.js';

const roots: string[] = [];
function fixture(
  files: Record<string, string | Buffer>,
  rules: Record<string, unknown> = { 'no-debugger': 'error' },
): string {
  const cwd = fs.realpathSync.native(
    fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rust-cli-')),
  );
  roots.push(cwd);
  for (const [name, content] of Object.entries(files)) {
    fs.mkdirSync(path.dirname(path.join(cwd, name)), { recursive: true });
    fs.writeFileSync(path.join(cwd, name), content);
  }
  fs.writeFileSync(
    path.join(cwd, 'rslint.config.mjs'),
    `export default [{ files: ['**/*.ts'], plugins: ${JSON.stringify([
      ...new Set(
        Object.keys(rules)
          .filter((name) => name.includes('/'))
          .map((name) => name.slice(0, name.lastIndexOf('/'))),
      ),
    ])}, rules: ${JSON.stringify(rules)} }];`,
  );
  return cwd;
}

afterEach(() => {
  for (const root of roots.splice(0))
    fs.rmSync(root, { recursive: true, force: true });
});

async function run(cwd: string, args: string[] = []) {
  let stdout = '';
  let stderr = '';
  const code = await runRustCLI(
    ['--config', 'rslint.config.mjs', '--no-color', ...args],
    {
      cwd,
      stdout: {
        write(value) {
          stdout += value;
        },
      },
      stderr: {
        write(value) {
          stderr += value;
        },
      },
    },
  );
  return { code, stdout, stderr };
}

async function withNativeHook(
  hook: (response: string, request: string) => string,
  callback: () => Promise<void>,
): Promise<void> {
  const binding: unknown = getNativeBinding();
  if (!isRecord(binding) || typeof binding.lintProjectsTsrsSync !== 'function')
    throw new Error('Configured Rust addon is required');
  const original = binding.lintProjectsTsrsSync;
  binding.lintProjectsTsrsSync = (request: string) =>
    hook(original(request), request);
  try {
    await callback();
  } finally {
    binding.lintProjectsTsrsSync = original;
  }
}

describe('direct Rust CLI', () => {
  test('discovers a directory once and snapshots every admitted source before native lint', async () => {
    const cwd = fixture({
      'nested/input.ts': 'debugger;',
      'nested/clean.ts': 'export const answer = 42;',
    });
    const directory = path.join(cwd, 'nested');
    const original = fs.readdirSync;
    let directoryReads = 0;
    let inputs: unknown;
    fs.readdirSync = ((...args: unknown[]) => {
      if (args[0] === directory) directoryReads++;
      return Reflect.apply(original, fs, args);
    }) as typeof original;
    try {
      await withNativeHook(
        (response, request) => {
          inputs = JSON.parse(request).files;
          return response;
        },
        async () => {
          const result = await run(cwd, ['--format', 'jsonline']);
          expect(result.code).toBe(1);
          expect(result.stderr).toBe('');
          expect(JSON.parse(result.stdout)).toMatchObject({
            ruleName: 'no-debugger',
            filePath: 'nested/input.ts',
          });
        },
      );
    } finally {
      fs.readdirSync = original;
    }
    expect(directoryReads).toBe(1);
    expect(inputs).toEqual(
      expect.arrayContaining([
        expect.objectContaining({
          path: path.join(directory, 'input.ts'),
          text: 'debugger;',
        }),
        expect.objectContaining({
          path: path.join(directory, 'clean.ts'),
          text: 'export const answer = 42;',
        }),
      ]),
    );
  });

  test('uses the explicit config directory for source-only Programs without changing admission cwd', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' });
    const configDirectory = path.join(cwd, 'external-config');
    fs.mkdirSync(configDirectory);
    const config = path.join(configDirectory, 'rslint.config.mjs');
    fs.writeFileSync(
      config,
      `export default [{basePath:${JSON.stringify(cwd)},files:['input.ts'],rules:{'no-debugger':'error'}}];`,
    );
    let request: unknown;
    await withNativeHook(
      (response, raw) => {
        request = JSON.parse(raw);
        return response;
      },
      async () => {
        const result = await run(cwd, [
          '--config',
          config,
          '--format',
          'jsonline',
        ]);
        expect(result.code).toBe(1);
        expect(result.stderr).toBe('');
      },
    );
    expect(request).toMatchObject({
      cwd,
      sourceOnlyCurrentDirectory: configDirectory,
    });
    if (!isRecord(request) || !Array.isArray(request.files))
      throw new Error('Missing native request files');
    expect(
      request.files.find(
        (file: unknown) =>
          isRecord(file) && file.path === path.join(cwd, 'input.ts'),
      ),
    ).toMatchObject({ policy: { kind: 'sourceOnly' } });
  });
  test('reports complete JSON line locations and exit code, retaining clean inputs', async () => {
    const cwd = fixture({
      'input.ts': 'debugger;',
      'clean.ts': 'const value = 1;',
    });
    let admitted: unknown;
    await withNativeHook(
      (response, request) => {
        admitted = JSON.parse(request).files.map((file: { path: string }) =>
          path.basename(file.path),
        );
        return response;
      },
      async () => {
        const result = await run(cwd, ['--format=jsonline']);
        expect(result.code).toBe(1);
        expect(result.stderr).toBe('');
        expect(JSON.parse(result.stdout)).toEqual({
          ruleName: 'no-debugger',
          message: "Unexpected 'debugger' statement.",
          filePath: 'input.ts',
          range: {
            start: { line: 1, column: 1 },
            end: { line: 1, column: 10 },
          },
          severity: 'error',
        });
        expect(admitted).toEqual(['clean.ts', 'input.ts', 'rslint.config.mjs']);
      },
    );
  });

  test('emits no machine lifecycle text for a successful clean file', async () => {
    const cwd = fixture({ 'clean.ts': 'const value = 1;' });
    expect(await run(cwd, ['--format', 'jsonline'])).toEqual({
      code: 0,
      stdout: '',
      stderr: '',
    });
    expect(await run(cwd, ['--format', 'gitlab'])).toEqual({
      code: 0,
      stdout: '[]\n',
      stderr: '',
    });
  });

  test('quiet only filters presentation; warnings still fail the warning limit', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' }, { 'no-debugger': 'warn' });
    expect(
      await run(cwd, ['--format=jsonline', '--quiet', '--max-warnings=0']),
    ).toEqual({
      code: 1,
      stdout: '',
      stderr: 'Rslint found too many warnings (maximum: 0).\n',
    });
    expect(
      (await run(cwd, ['--format=jsonline', '--max-warnings=1'])).code,
    ).toBe(0);
    expect(
      (await run(cwd, ['--format=jsonline', '--max-warnings=-2'])).code,
    ).toBe(0);
  });

  test('diagnostic errors take precedence over exceeded warnings', async () => {
    const cwd = fixture(
      { 'input.ts': 'debugger; value!!;' },
      {
        'no-debugger': 'error',
        '@typescript-eslint/no-extra-non-null-assertion': 'warn',
      },
    );
    const result = await run(cwd, ['--format=jsonline', '--max-warnings=0']);
    expect(result.code).toBe(1);
    expect(result.stdout.trim().split('\n')).toHaveLength(2);
    expect(result.stderr).toBe('');
  });

  test('supports options after positional targets and repeated rule overrides', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' });
    expect(
      await run(cwd, [
        'input.ts',
        '--format=jsonline',
        '--rule',
        'no-debugger: warn',
        '--rule',
        'no-debugger: off',
      ]),
    ).toEqual({ code: 0, stdout: '', stderr: '' });
    const warning = await run(cwd, [
      'input.ts',
      '--rule',
      'no-debugger: ["warn"]',
      '--format=jsonline',
    ]);
    expect(warning.code).toBe(0);
    expect(JSON.parse(warning.stdout).severity).toBe('warn');
  });

  test('respects the option terminator for dash-prefixed filenames', async () => {
    const cwd = fixture({ '-input.ts': 'debugger;' });
    const result = await run(cwd, ['--format=jsonline', '--', '-input.ts']);
    expect(result.code).toBe(1);
    expect(JSON.parse(result.stdout).filePath).toBe('-input.ts');
  });

  test('escapes GitHub properties and data without ANSI', async () => {
    const cwd = fixture({ 'a,%x.ts': 'debugger;' });
    const result = await run(cwd, ['--format=github', '--force-color']);
    expect(result).toEqual({
      code: 1,
      stderr: '',
      stdout:
        "::error file=a%2C%25x.ts,line=1,endLine=1,col=1,endColumn=10,title=no-debugger::Unexpected 'debugger' statement.\n",
    });
  });

  test('GitLab retains severity, location and unique fingerprints for duplicate findings', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' });
    await withNativeHook(
      (response) => {
        const value = JSON.parse(response);
        value.files[0].messages.push(value.files[0].messages[0]);
        return JSON.stringify(value);
      },
      async () => {
        const result = await run(cwd, ['--format=gitlab']);
        expect(result.code).toBe(1);
        const issues = JSON.parse(result.stdout);
        expect(issues).toHaveLength(2);
        expect(issues[0]).toMatchObject({
          description: "Unexpected 'debugger' statement.",
          check_name: 'no-debugger',
          severity: 'major',
          location: {
            path: 'input.ts',
            lines: { begin: 1, end: 1 },
            positions: {
              begin: { line: 1, column: 1 },
              end: { line: 1, column: 10 },
            },
          },
        });
        // Frozen native CLI observation, including its historical MD5 input.
        expect(issues[0].fingerprint).toBe('06be158862d6a1fc8ba549386d725007');
        expect(issues[0].fingerprint).not.toBe(issues[1].fingerprint);
      },
    );
  });

  test('renders the original default code frame and truthful run summary', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;\n' });
    const result = await run(cwd);
    expect(result.code).toBe(1);
    expect(result.stderr).toBe('');
    expect(result.stdout).toContain(
      "start   Linting...\n\n  no-debugger  — [error] Unexpected 'debugger' statement.\n  ╭─┴──────────( input.ts:1:1 )─────\n  │ 1 │  debugger;\n  │ 2 │  \n  ╰────────────────────────────────\n\n",
    );
    expect(result.stdout).toMatch(
      /error   Lint failed with 1 error in .+ \(2 files, 1 rule, 1 thread\)\n$/,
    );
  });

  test('default quiet reports the warning budget without displaying warning frames', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' }, { 'no-debugger': 'warn' });
    const result = await run(cwd, ['--quiet', '--max-warnings=0']);
    expect(result.code).toBe(1);
    expect(result.stdout).not.toContain('no-debugger');
    expect(result.stdout).toMatch(
      /1 warning exceeded the configured limit of 0 \(2 files, 1 rule, 1 thread\)/,
    );
  });

  test('fix writes only final verified output and reports the final diagnostics', async () => {
    const source = "\uFEFF'😀'; value!!!!; debugger;";
    const cwd = fixture(
      { 'input.ts': source },
      {
        '@typescript-eslint/no-extra-non-null-assertion': 'error',
        'no-debugger': 'warn',
      },
    );
    const result = await run(cwd, ['--fix', '--format=jsonline']);
    expect(result.code).toBe(0);
    expect(result.stderr).toBe('');
    expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(
      "\uFEFF'😀'; value!; debugger;",
    );
    const message = JSON.parse(result.stdout);
    expect(message.ruleName).toBe('no-debugger');
    expect(message.range.start).toEqual({ line: 1, column: 15 });
  });

  test('without --fix source remains unchanged, including suggestions', async () => {
    const source = 'value!!; object?.prop!;';
    const cwd = fixture(
      { 'input.ts': source },
      {
        '@typescript-eslint/no-extra-non-null-assertion': 'error',
        '@typescript-eslint/no-non-null-asserted-optional-chain': 'warn',
      },
    );
    const result = await run(cwd, ['--format=jsonline']);
    expect(result.code).toBe(1);
    expect(result.stdout.trim().split('\n')).toHaveLength(2);
    expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(source);
  });

  test('preflights the complete fix delta before the first disk write', async () => {
    const cwd = fixture(
      { 'a.ts': 'a!!;', 'b.ts': 'b!!;' },
      { '@typescript-eslint/no-extra-non-null-assertion': 'error' },
    );
    await withNativeHook(
      (response) => {
        fs.writeFileSync(path.join(cwd, 'b.ts'), 'changed externally;');
        return response;
      },
      async () => {
        const result = await run(cwd, ['--fix', '--format=jsonline']);
        expect(result.code).toBe(1);
        expect(result.stdout).toBe('');
        expect(result.stderr).toContain('changed after lint generation');
        expect(fs.readFileSync(path.join(cwd, 'a.ts'), 'utf8')).toBe('a!!;');
        expect(fs.readFileSync(path.join(cwd, 'b.ts'), 'utf8')).toBe(
          'changed externally;',
        );
      },
    );
  });

  test('detects a symlink retarget even when replacement source bytes are identical', async () => {
    const cwd = fixture(
      { 'original.txt': 'value!!;', 'replacement.txt': 'value!!;' },
      { '@typescript-eslint/no-extra-non-null-assertion': 'error' },
    );
    const alias = path.join(cwd, 'input.ts');
    fs.symlinkSync('original.txt', alias);
    await withNativeHook(
      (response) => {
        fs.unlinkSync(alias);
        fs.symlinkSync('replacement.txt', alias);
        return response;
      },
      async () => {
        const result = await run(cwd, [
          'input.ts',
          '--fix',
          '--format=jsonline',
        ]);
        expect(result.code).toBe(1);
        expect(result.stderr).toContain('changed after lint generation');
        expect(fs.readFileSync(path.join(cwd, 'replacement.txt'), 'utf8')).toBe(
          'value!!;',
        );
      },
    );
  });

  test('does not commit edits when final native results fail validation', async () => {
    const cwd = fixture(
      { 'input.ts': 'value!!;' },
      { '@typescript-eslint/no-extra-non-null-assertion': 'error' },
    );
    await withNativeHook(
      (response) => {
        const value = JSON.parse(response);
        value.files[0].messages.push({
          ruleId: 'forged-rule',
          message: 'forged',
          severity: 2,
        });
        return JSON.stringify(value);
      },
      async () => {
        const result = await run(cwd, ['--fix', '--format=jsonline']);
        expect(result.code).toBe(1);
        expect(result.stdout).toBe('');
        expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(
          'value!!;',
        );
      },
    );
  });

  test('UTF-16 disk input is decoded once and fixed using the original UTF-8 output policy', async () => {
    const source = Buffer.concat([
      Buffer.from([0xff, 0xfe]),
      Buffer.from('value!!;', 'utf16le'),
    ]);
    const cwd = fixture(
      { 'input.ts': source },
      { '@typescript-eslint/no-extra-non-null-assertion': 'error' },
    );
    const result = await run(cwd, ['--fix', '--format=jsonline']);
    expect(result).toEqual({ code: 0, stdout: '', stderr: '' });
    expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(
      '\uFEFFvalue!;',
    );
  });

  test('unsupported rules fail rather than yielding a partial successful report', async () => {
    const cwd = fixture(
      { 'input.ts': 'debugger;' },
      { 'unimplemented/rule': 'error' },
    );
    const result = await run(cwd, ['--format=jsonline']);
    expect(result.code).toBe(1);
    expect(result.stdout).toBe('');
    expect(result.stderr).toContain('unimplemented/rule');
  });

  test('invalid CLI arguments fail before evaluating configuration', async () => {
    const cwd = fixture({ 'input.ts': '' });
    fs.writeFileSync(
      path.join(cwd, 'rslint.config.mjs'),
      'throw new Error("config must not run");',
    );
    for (const args of [
      ['--unknown'],
      ['--format=invalid'],
      ['--max-warnings=1.5'],
      ['--rule=no-colon'],
      ['--type-check'],
      ['--type-check-only'],
      ['--timing'],
      ['--init'],
      ['--trace=trace.out'],
      ['--cpuprof=cpu.out'],
    ]) {
      const result = await run(cwd, args);
      expect(result.code).toBe(2);
      expect(result.stdout).toBe('');
      expect(result.stderr).not.toContain('config must not run');
    }
  });

  test('help bypasses config loading and invalid output format', async () => {
    const cwd = fixture({});
    fs.writeFileSync(
      path.join(cwd, 'rslint.config.mjs'),
      'throw new Error("config must not run");',
    );
    const result = await run(cwd, ['--help', '--format=invalid']);
    expect(result.code).toBe(0);
    expect(result.stderr).toBe('');
    expect(result.stdout).toContain('Usage:');
    expect(result.stdout).toContain('not implemented');
  });

  test('missing explicit config reports the existing error and never starts linting', async () => {
    const cwd = fixture({});
    fs.unlinkSync(path.join(cwd, 'rslint.config.mjs'));
    const result = await run(cwd);
    expect(result).toEqual({
      code: 1,
      stdout: '',
      stderr: `Error: config file not found: ${path.join(cwd, 'rslint.config.mjs')}\n`,
    });
  });
});
