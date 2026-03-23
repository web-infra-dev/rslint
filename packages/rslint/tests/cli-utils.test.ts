import { describe, test, expect } from '@rstest/core';
import { isJSConfigFile, classifyArgs, discoverConfigs } from '../src/cli.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-cli-test-'));
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
      // Should resolve to realpath
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

describe('discoverConfigs', () => {
  test('no files/dirs uses cwd', () => {
    const tmp = createTempDir();
    try {
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs([], [], tmp, null);
      expect(result.size).toBe(1);
      const configPath = [...result.keys()][0];
      expect(configPath).toBe(path.join(tmp, 'rslint.config.js'));
    } finally {
      cleanup(tmp);
    }
  });

  test('explicit config overrides discovery', () => {
    const tmp = createTempDir();
    const configFile = path.join(tmp, 'custom.config.js');
    try {
      fs.writeFileSync(configFile, 'export default []');
      const result = discoverConfigs(
        ['/some/file.ts'],
        ['/some/dir'],
        tmp,
        configFile,
      );
      // Should only have the explicit config, ignore files/dirs
      expect(result.size).toBe(1);
      expect([...result.keys()][0]).toBe(configFile);
    } finally {
      cleanup(tmp);
    }
  });

  test('deduplicates files in same directory', () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    try {
      fs.mkdirSync(src);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      // Two files in same directory should find same config
      const result = discoverConfigs(
        [path.join(src, 'a.ts'), path.join(src, 'b.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('different directories find different configs', () => {
    const tmp = createTempDir();
    const foo = path.join(tmp, 'packages', 'foo');
    const bar = path.join(tmp, 'packages', 'bar');
    try {
      fs.mkdirSync(foo, { recursive: true });
      fs.mkdirSync(bar, { recursive: true });
      fs.writeFileSync(path.join(foo, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(path.join(bar, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs(
        [path.join(foo, 'a.ts'), path.join(bar, 'b.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(2);
    } finally {
      cleanup(tmp);
    }
  });

  test('no config found returns empty map', () => {
    const tmp = createTempDir();
    const deep = path.join(tmp, 'a', 'b');
    try {
      fs.mkdirSync(deep, { recursive: true });
      const result = discoverConfigs(
        [path.join(deep, 'file.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(0);
    } finally {
      cleanup(tmp);
    }
  });

  test('directory arg uses dir as start point', () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    try {
      fs.mkdirSync(src);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      const result = discoverConfigs([], [src], tmp, null);
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('files and dirs finding same config deduplicated', () => {
    const tmp = createTempDir();
    const src = path.join(tmp, 'src');
    const lib = path.join(tmp, 'lib');
    try {
      fs.mkdirSync(src);
      fs.mkdirSync(lib);
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      // File in src/ and dir lib/ both find the same root config
      const result = discoverConfigs(
        [path.join(src, 'a.ts')],
        [lib],
        tmp,
        null,
      );
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });

  test('no args discovers nested configs in monorepo', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.ts'),
        'export default []',
      );
      // No args: should find root + foo + bar configs
      const result = discoverConfigs([], [], tmp, null);
      expect(result.size).toBe(3);
    } finally {
      cleanup(tmp);
    }
  });

  test('dir arg discovers nested configs within scope', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'foo', 'rslint.config.js'),
        'export default []',
      );
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.js'),
        'export default []',
      );
      // Dir arg packages/: finds root (upward) + foo + bar (scan)
      const result = discoverConfigs(
        [],
        [path.join(tmp, 'packages')],
        tmp,
        null,
      );
      expect(result.size).toBe(3);
    } finally {
      cleanup(tmp);
    }
  });

  test('file args do not trigger nested config scan', () => {
    const tmp = createTempDir();
    try {
      fs.mkdirSync(path.join(tmp, 'packages', 'foo'), { recursive: true });
      fs.mkdirSync(path.join(tmp, 'packages', 'bar'), { recursive: true });
      fs.writeFileSync(path.join(tmp, 'rslint.config.js'), 'export default []');
      fs.writeFileSync(
        path.join(tmp, 'packages', 'bar', 'rslint.config.js'),
        'export default []',
      );
      // File arg from foo only — should NOT discover bar's config
      const result = discoverConfigs(
        [path.join(tmp, 'packages', 'foo', 'a.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(1); // only root config (upward from foo)
    } finally {
      cleanup(tmp);
    }
  });
});
