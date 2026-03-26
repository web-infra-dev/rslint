import { describe, test, expect } from '@rstest/core';
import { isJSConfigFile, classifyArgs, parseArgs } from '../src/utils/args.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-args-test-'));
}

function cleanup(dir: string): void {
  fs.rmSync(dir, { recursive: true, force: true });
}

describe('isJSConfigFile', () => {
  test('returns true for .js', () => {
    expect(isJSConfigFile('rslint.config.js')).toBe(true);
  });

  test('returns true for .mjs', () => {
    expect(isJSConfigFile('rslint.config.mjs')).toBe(true);
  });

  test('returns true for .ts', () => {
    expect(isJSConfigFile('rslint.config.ts')).toBe(true);
  });

  test('returns true for .mts', () => {
    expect(isJSConfigFile('rslint.config.mts')).toBe(true);
  });

  test('returns false for .json', () => {
    expect(isJSConfigFile('rslint.json')).toBe(false);
  });

  test('returns false for .jsonc', () => {
    expect(isJSConfigFile('rslint.jsonc')).toBe(false);
  });

  test('returns false for no extension', () => {
    expect(isJSConfigFile('rslint')).toBe(false);
  });

  test('handles full paths', () => {
    expect(isJSConfigFile('/project/rslint.config.ts')).toBe(true);
  });

  test('returns false for .cjs', () => {
    expect(isJSConfigFile('rslint.config.cjs')).toBe(false);
  });
});

describe('classifyArgs', () => {
  test('empty positionals', () => {
    const result = classifyArgs([], '/tmp');
    expect(result.files).toEqual([]);
    expect(result.dirs).toEqual([]);
  });

  test('classifies existing directory', () => {
    const tmp = createTempDir();
    try {
      const realTmp = fs.realpathSync(tmp);
      const result = classifyArgs([tmp], '/');
      expect(result.dirs).toEqual([realTmp]);
      expect(result.files).toEqual([]);
    } finally {
      cleanup(tmp);
    }
  });

  test('classifies existing file', () => {
    const tmp = createTempDir();
    const filePath = path.join(tmp, 'test.ts');
    try {
      fs.writeFileSync(filePath, 'const x = 1;');
      const realFile = fs.realpathSync(filePath);
      const result = classifyArgs([filePath], '/');
      expect(result.files).toEqual([realFile]);
      expect(result.dirs).toEqual([]);
    } finally {
      cleanup(tmp);
    }
  });

  test('non-existent path treated as file', () => {
    const result = classifyArgs(['/nonexistent/path/file.ts'], '/');
    expect(result.files).toEqual(['/nonexistent/path/file.ts']);
    expect(result.dirs).toEqual([]);
  });

  test('mixed files and directories', () => {
    const tmp = createTempDir();
    const dir = path.join(tmp, 'src');
    const filePath = path.join(tmp, 'test.ts');
    try {
      fs.mkdirSync(dir);
      fs.writeFileSync(filePath, 'const x = 1;');
      const realDir = fs.realpathSync(dir);
      const realFile = fs.realpathSync(filePath);
      const result = classifyArgs([dir, filePath], '/');
      expect(result.dirs).toEqual([realDir]);
      expect(result.files).toEqual([realFile]);
    } finally {
      cleanup(tmp);
    }
  });

  test('resolves relative paths against cwd', () => {
    const tmp = createTempDir();
    const filePath = path.join(tmp, 'test.ts');
    try {
      fs.writeFileSync(filePath, 'const x = 1;');
      const realFile = fs.realpathSync(filePath);
      const result = classifyArgs(['test.ts'], tmp);
      expect(result.files).toEqual([realFile]);
    } finally {
      cleanup(tmp);
    }
  });

  test('resolves symlinks in path', () => {
    const tmp = createTempDir();
    const realDir = path.join(tmp, 'real');
    const linkDir = path.join(tmp, 'link');
    const filePath = path.join(realDir, 'test.ts');
    try {
      fs.mkdirSync(realDir);
      fs.writeFileSync(filePath, 'const x = 1;');
      fs.symlinkSync(realDir, linkDir);
      const result = classifyArgs([path.join(linkDir, 'test.ts')], '/');
      expect(result.files).toEqual([fs.realpathSync(filePath)]);
    } finally {
      cleanup(tmp);
    }
  });

  test('resolves symlinked directory arg', () => {
    const tmp = createTempDir();
    const realDir = path.join(tmp, 'real');
    const linkDir = path.join(tmp, 'link');
    try {
      fs.mkdirSync(realDir);
      fs.symlinkSync(realDir, linkDir);
      const result = classifyArgs([linkDir], '/');
      expect(result.dirs).toEqual([fs.realpathSync(realDir)]);
    } finally {
      cleanup(tmp);
    }
  });
});

describe('parseArgs positionals', () => {
  test('--format jsonline does not pollute positionals', () => {
    const result = parseArgs(['--format', 'jsonline', 'src/a.ts']);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('file before --format', () => {
    const result = parseArgs(['src/a.ts', '--format', 'jsonline']);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('--no-color does not consume next positional', () => {
    const result = parseArgs(['--no-color', 'src/a.ts']);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('--no-color --format jsonline before file', () => {
    const result = parseArgs([
      '--no-color',
      '--format',
      'jsonline',
      'src/a.ts',
    ]);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('--max-warnings value not in positionals', () => {
    const result = parseArgs(['--max-warnings', '10', 'src/a.ts']);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('multiple files with flags interspersed', () => {
    const result = parseArgs(['src/a.ts', '--format', 'jsonline', 'src/b.ts']);
    expect(result.positionals).toEqual(['src/a.ts', 'src/b.ts']);
  });

  test('--fix before file (boolean flag)', () => {
    const result = parseArgs(['--fix', 'src/a.ts']);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('--format=jsonline inline value', () => {
    const result = parseArgs(['--format=jsonline', 'src/a.ts']);
    expect(result.positionals).toEqual(['src/a.ts']);
  });

  test('no positionals with only flags', () => {
    const result = parseArgs(['--format', 'jsonline', '--no-color']);
    expect(result.positionals).toEqual([]);
  });

  test('-- separator treats everything after as positional', () => {
    const result = parseArgs(['--', '--not-a-flag']);
    expect(result.positionals).toEqual(['--not-a-flag']);
  });

  test('rest includes all non-config non-init args', () => {
    const result = parseArgs([
      '--format',
      'jsonline',
      '--no-color',
      'src/a.ts',
    ]);
    expect(result.rest).toContain('--format');
    expect(result.rest).toContain('jsonline');
    expect(result.rest).toContain('--no-color');
    expect(result.rest).toContain('src/a.ts');
  });

  test('--config is excluded from rest', () => {
    const result = parseArgs(['--config', 'custom.js', 'src/a.ts']);
    expect(result.config).toBe('custom.js');
    expect(result.rest).not.toContain('--config');
    expect(result.rest).not.toContain('custom.js');
    expect(result.rest).toContain('src/a.ts');
    expect(result.positionals).toEqual(['src/a.ts']);
  });
});
