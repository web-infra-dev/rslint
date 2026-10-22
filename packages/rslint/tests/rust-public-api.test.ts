import childProcess from 'node:child_process';
import fs from 'node:fs';
import { syncBuiltinESMExports } from 'node:module';
import os from 'node:os';
import path from 'node:path';
import { afterEach, describe, expect, test } from 'rstack/test';
import { Rslint, type RslintOptions } from '../src/api/rslint.js';
import { getNativeBinding } from '../src/native/binding.js';
import { isRecord } from '../src/rust/engine.js';

const roots: string[] = [];
function fixture(files: Record<string, string> = {}): string {
  const root = fs.realpathSync.native(
    fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-rust-public-api-')),
  );
  roots.push(root);
  for (const [name, text] of Object.entries(files)) {
    fs.mkdirSync(path.dirname(path.join(root, name)), { recursive: true });
    fs.writeFileSync(path.join(root, name), text);
  }
  return root;
}
afterEach(() => {
  for (const root of roots.splice(0))
    fs.rmSync(root, { recursive: true, force: true });
});

function options(cwd: string): RslintOptions {
  return {
    cwd,
    overrideConfigFile: true,
    overrideConfig: [{ rules: { 'no-debugger': 'error' } }],
  };
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

describe('public API uses the direct Rust engine', () => {
  test('construction and linting spawn no native service process', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' });
    const originalSpawn = childProcess.spawn;
    const originalSpawnSync = childProcess.spawnSync;
    const originalExec = childProcess.exec;
    const originalExecFile = childProcess.execFile;
    const originalExecSync = childProcess.execSync;
    const originalExecFileSync = childProcess.execFileSync;
    const originalFork = childProcess.fork;
    const forbidden = (): never => {
      throw new Error('Public Rust API must not spawn a child process');
    };
    childProcess.spawn = forbidden;
    childProcess.spawnSync = forbidden;
    childProcess.exec = forbidden;
    childProcess.execFile = forbidden;
    childProcess.execSync = forbidden;
    childProcess.execFileSync = forbidden;
    childProcess.fork = forbidden;
    syncBuiltinESMExports();
    let calls = 0;
    try {
      await withNativeHook(
        (response) => {
          calls++;
          return response;
        },
        async () => {
          const instance = new Rslint(options(cwd));
          expect(calls).toBe(0);
          try {
            const [result] = await instance.lintFiles('input.ts');
            expect(result.errorCount).toBe(1);
            expect(calls).toBe(1);
          } finally {
            await instance.close();
          }
        },
      );
    } finally {
      childProcess.spawn = originalSpawn;
      childProcess.spawnSync = originalSpawnSync;
      childProcess.exec = originalExec;
      childProcess.execFile = originalExecFile;
      childProcess.execSync = originalExecSync;
      childProcess.execFileSync = originalExecFileSync;
      childProcess.fork = originalFork;
      syncBuiltinESMExports();
    }
  });

  test('retains clean disk inputs, public diagnostic fields and per-file severity', async () => {
    const cwd = fixture({ 'z.ts': 'debugger;', 'a.ts': 'export {};' });
    const instance = new Rslint({
      ...options(cwd),
      overrideConfig: [{ rules: { 'no-debugger': 'warn' } }],
    });
    try {
      const result = await instance.lintFiles(['z.ts', 'a.ts']);
      expect(result.map((file) => path.basename(file.filePath))).toEqual([
        'a.ts',
        'z.ts',
      ]);
      expect(result[0]).toEqual({
        filePath: path.join(cwd, 'a.ts'),
        messages: [],
        errorCount: 0,
        warningCount: 0,
        fixableErrorCount: 0,
        fixableWarningCount: 0,
      });
      expect(result[1]).toEqual({
        filePath: path.join(cwd, 'z.ts'),
        messages: [
          {
            ruleId: 'no-debugger',
            severity: 1,
            message: "Unexpected 'debugger' statement.",
            messageId: 'no-debugger',
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 10,
          },
        ],
        errorCount: 0,
        warningCount: 1,
        fixableErrorCount: 0,
        fixableWarningCount: 0,
      });
      expect(fs.readFileSync(path.join(cwd, 'z.ts'), 'utf8')).toBe('debugger;');
    } finally {
      await instance.close();
    }
  });

  test('evaluates the selected config afresh on each lint operation', async () => {
    const cwd = fixture({
      'input.ts': 'debugger;',
      'config.mjs': 'export default [{rules:{"no-debugger":"error"}}];',
    });
    const instance = new Rslint({ cwd, overrideConfigFile: 'config.mjs' });
    try {
      expect((await instance.lintFiles('input.ts'))[0].errorCount).toBe(1);
      fs.writeFileSync(
        path.join(cwd, 'config.mjs'),
        'export default [{rules:{"no-debugger":"off"}}];',
      );
      expect((await instance.lintFiles('input.ts'))[0].messages).toEqual([]);
    } finally {
      await instance.close();
    }
  });

  test('lintText fixes BOM and Unicode buffers without creating the synthetic path', async () => {
    const cwd = fixture();
    const source = "\uFEFF'😀'; value!!!!; debugger;";
    const instance = new Rslint({
      ...options(cwd),
      fix: true,
      overrideConfig: [
        {
          plugins: ['@typescript-eslint'],
          rules: {
            '@typescript-eslint/no-extra-non-null-assertion': 'error',
            'no-debugger': 'error',
          },
        },
      ],
    });
    try {
      const [result] = await instance.lintText(source);
      expect(result.filePath).toBe('<text>');
      expect(result.output).toBe("\uFEFF'😀'; value!; debugger;");
      expect(result.messages).toHaveLength(1);
      expect(result.messages[0]).toMatchObject({
        ruleId: 'no-debugger',
        column: 15,
      });
      await Rslint.outputFixes([result]);
      expect(fs.readdirSync(cwd)).toEqual([]);
    } finally {
      await instance.close();
    }
  });

  test('the code buffer wins over the same overlay path and dependencies are not lint targets', async () => {
    const cwd = fixture();
    const instance = new Rslint({
      ...options(cwd),
      virtualFiles: { 'input.ts': 'debugger;', 'dependency.ts': 'debugger;' },
    });
    try {
      const result = await instance.lintText('export {};', {
        filePath: 'input.ts',
      });
      expect(result).toHaveLength(1);
      expect(result[0].filePath).toBe(path.join(cwd, 'input.ts'));
      expect(result[0].messages).toEqual([]);
      expect(fs.readdirSync(cwd)).toEqual([]);
    } finally {
      await instance.close();
    }
  });

  test('resolves cross-file type information from an entirely in-memory project', async () => {
    const cwd = fixture();
    const instance = new Rslint({
      ...options(cwd),
      overrideConfig: [
        {
          files: ['**/*.ts'],
          plugins: ['@typescript-eslint'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: {
            '@typescript-eslint/no-for-in-array': 'error',
            'no-debugger': 'error',
          },
        },
      ],
      virtualFiles: {
        'tsconfig.json': JSON.stringify({
          compilerOptions: { strict: true, moduleResolution: 'bundler' },
          files: ['src/input.ts', 'src/dep.ts'],
        }),
        'src/dep.ts': 'export const nums: number[] = [1, 2]; debugger;',
      },
    });
    try {
      const result = await instance.lintText(
        "import { nums } from './dep'; for (const key in nums) {}",
        { filePath: 'src/input.ts' },
      );
      expect(result).toHaveLength(1);
      expect(result[0].messages).toHaveLength(1);
      expect(result[0].messages[0]).toMatchObject({
        ruleId: '@typescript-eslint/no-for-in-array',
        messageId: 'forInViolation',
        severity: 2,
      });
      expect(fs.readdirSync(cwd)).toEqual([]);
    } finally {
      await instance.close();
    }
  });

  test('lintFiles uses overlays without adding their dependency paths to the result', async () => {
    const cwd = fixture({ 'input.ts': 'export {};' });
    const instance = new Rslint({
      ...options(cwd),
      virtualFiles: {
        'input.ts': 'debugger;',
        'dependency.ts': 'debugger;',
      },
    });
    try {
      const result = await instance.lintFiles('*.ts');
      expect(result).toHaveLength(1);
      expect(result[0].errorCount).toBe(1);
      expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(
        'export {};',
      );
      expect(fs.existsSync(path.join(cwd, 'dependency.ts'))).toBe(false);
    } finally {
      await instance.close();
    }
  });

  test('keeps a nonexistent target under a directory alias in lexical coordinates', async () => {
    const cwd = fixture({ 'real/.keep': '' });
    fs.symlinkSync(
      path.join(cwd, 'real'),
      path.join(cwd, 'alias'),
      process.platform === 'win32' ? 'junction' : 'dir',
    );
    const instance = new Rslint(options(cwd));
    try {
      const [result] = await instance.lintText('debugger;', {
        filePath: 'alias/nested/input.ts',
      });
      expect(result.filePath).toBe(path.join(cwd, 'alias/nested/input.ts'));
      expect(result.errorCount).toBe(1);
      expect(fs.existsSync(path.join(cwd, 'real/nested/input.ts'))).toBe(false);
    } finally {
      await instance.close();
    }
  });

  test('the buffer overrides a canonical alias overlay and preserves typed project ownership', async () => {
    const cwd = fixture({
      'real/input.ts': 'export {};',
      'tsconfig.json': JSON.stringify({ files: ['real/input.ts'] }),
    });
    fs.symlinkSync(
      path.join(cwd, 'real'),
      path.join(cwd, 'alias'),
      process.platform === 'win32' ? 'junction' : 'dir',
    );
    const instance = new Rslint({
      ...options(cwd),
      overrideConfig: [
        {
          plugins: ['@typescript-eslint'],
          languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
          rules: {
            '@typescript-eslint/no-for-in-array': 'error',
            'no-debugger': 'error',
          },
        },
      ],
      virtualFiles: { 'real/input.ts': 'debugger;' },
    });
    try {
      const result = await instance.lintText(
        'const values = [1]; for (const key in values) {}',
        { filePath: 'alias/input.ts' },
      );
      expect(result).toHaveLength(1);
      expect(result[0].filePath).toBe(path.join(cwd, 'alias/input.ts'));
      expect(result[0].messages.map((message) => message.ruleId)).toEqual([
        '@typescript-eslint/no-for-in-array',
      ]);
      expect(fs.readFileSync(path.join(cwd, 'real/input.ts'), 'utf8')).toBe(
        'export {};',
      );
    } finally {
      await instance.close();
    }
  });

  test('rejects unsupported rules and discovery without invoking an old backend', async () => {
    const cwd = fixture({ 'input.ts': 'debugger;' });
    const unsupported = new Rslint({
      ...options(cwd),
      overrideConfig: [{ rules: { 'not-a-real-rule': 'error' } }],
    });
    const discovery = new Rslint({ cwd });
    try {
      await expect(unsupported.lintFiles('input.ts')).rejects.toThrow(
        'not-a-real-rule',
      );
      await expect(discovery.lintFiles('input.ts')).rejects.toThrow(
        'automatic configuration discovery',
      );
    } finally {
      await Promise.all([unsupported.close(), discovery.close()]);
    }
  });

  test('serializes concurrent operations and close waits for all accepted requests', async () => {
    const cwd = fixture();
    const instance = new Rslint(options(cwd));
    const first = instance.lintText('debugger;', { filePath: 'first.ts' });
    const second = instance.lintText('export {};', { filePath: 'second.ts' });
    const closing = instance.close();
    const results = await Promise.all([first, second]);
    await closing;
    expect(results.map(([result]) => result.errorCount)).toEqual([1, 0]);
    await expect(instance.lintText('debugger;')).rejects.toThrow(
      'rslint service is closing',
    );
    await expect(instance.lintFiles([])).rejects.toThrow(
      'rslint service is closing',
    );
    await instance[Symbol.asyncDispose]();
    await instance.close();
  });

  test('a rejected request does not poison the queue or close', async () => {
    const cwd = fixture();
    const instance = new Rslint(options(cwd));
    const invalid = instance.lintText(null as unknown as string);
    const valid = instance.lintText('debugger;');
    await expect(invalid).rejects.toThrow('lintText code must be a string');
    expect((await valid)[0].errorCount).toBe(1);
    await instance.close();
  });

  test('handles empty selections and outputFixes without implicitly writing lint results', async () => {
    const cwd = fixture({ 'input.ts': 'value!!;' });
    const instance = new Rslint({
      ...options(cwd),
      fix: true,
      overrideConfig: [
        {
          plugins: ['@typescript-eslint'],
          rules: { '@typescript-eslint/no-extra-non-null-assertion': 'error' },
        },
      ],
    });
    try {
      expect(await instance.lintFiles([])).toEqual([]);
      const result = await instance.lintFiles('input.ts');
      expect(result[0].output).toBe('value!;');
      expect(result[0].messages).toEqual([]);
      expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(
        'value!!;',
      );
      await Rslint.outputFixes(result);
      expect(fs.readFileSync(path.join(cwd, 'input.ts'), 'utf8')).toBe(
        'value!;',
      );
    } finally {
      await instance.close();
    }
  });
});
