import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import { findJSConfig, loadConfigFile } from '@rslint/core/config-loader';
import { createTempDir, cleanupTempDir } from './helpers.js';

describe('loadConfigFile', () => {
  test('should load a .js config file with default export', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js':
        'export default [{ files: ["**/*.ts"], rules: { "no-console": "error" } }];',
    });
    try {
      const result = await loadConfigFile(
        path.join(tempDir, 'rslint.config.js'),
      );
      expect(Array.isArray(result)).toBe(true);
      expect((result as any[])[0].rules).toEqual({ 'no-console': 'error' });
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should load a .mjs config file', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs':
        'export default [{ files: ["**/*.js"], rules: {} }];',
    });
    try {
      const result = await loadConfigFile(
        path.join(tempDir, 'rslint.config.mjs'),
      );
      expect(Array.isArray(result)).toBe(true);
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should resolve thenable (Promise) default export', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js':
        'export default Promise.resolve([{ files: ["**/*.ts"], rules: { "no-console": "error" } }]);',
    });
    try {
      const result = await loadConfigFile(
        path.join(tempDir, 'rslint.config.js'),
      );
      expect(Array.isArray(result)).toBe(true);
      expect((result as any[])[0].rules).toEqual({ 'no-console': 'error' });
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should throw for unsupported extension', async () => {
    const tempDir = await createTempDir({
      'rslint.config.yaml': 'rules: {}',
    });
    try {
      await expect(
        loadConfigFile(path.join(tempDir, 'rslint.config.yaml')),
      ).rejects.toThrow('Unsupported config file extension');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});

describe('findJSConfig', () => {
  test('should return null when no config file exists', async () => {
    const tempDir = await createTempDir({ 'test.ts': '' });
    try {
      expect(findJSConfig(tempDir)).toBeNull();
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should find rslint.config.js', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer .js over .ts when both exist', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
      'rslint.config.ts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer .mjs over .ts when both exist', async () => {
    const tempDir = await createTempDir({
      'rslint.config.mjs': 'export default [];',
      'rslint.config.ts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.mjs');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should prefer .js over .mjs when both exist', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
      'rslint.config.mjs': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should find .ts config when no .js/.mjs exists', async () => {
    const tempDir = await createTempDir({
      'rslint.config.ts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.ts');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });

  test('should follow priority order: .js > .mjs > .ts > .mts', async () => {
    const tempDir = await createTempDir({
      'rslint.config.js': 'export default [];',
      'rslint.config.mjs': 'export default [];',
      'rslint.config.ts': 'export default [];',
      'rslint.config.mts': 'export default [];',
    });
    try {
      expect(findJSConfig(tempDir)).toContain('rslint.config.js');
    } finally {
      await cleanupTempDir(tempDir);
    }
  });
});
