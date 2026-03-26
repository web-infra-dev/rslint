import { describe, test, expect } from '@rstest/core';
import { discoverConfigs } from '../src/utils/config-discovery.js';
import fs from 'node:fs';
import path from 'node:path';
import os from 'node:os';

function createTempDir(): string {
  return fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-discovery-test-'));
}

function cleanup(dir: string): void {
  fs.rmSync(dir, { recursive: true, force: true });
}

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
      const result = discoverConfigs(
        [path.join(tmp, 'packages', 'foo', 'a.ts')],
        [],
        tmp,
        null,
      );
      expect(result.size).toBe(1);
    } finally {
      cleanup(tmp);
    }
  });
});
