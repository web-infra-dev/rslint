import fs from 'node:fs/promises';
import path from 'node:path';
import { describe, expect, test } from 'rstack/test';
import { createFixtureRoot, removeFixtureRoot } from '../src/fixture.ts';

describe('fixture root', () => {
  test('writes nested files under a real path and removes them', async () => {
    const root = await createFixtureRoot({ 'a/package.json': '{}' });
    try {
      expect(await fs.realpath(root)).toBe(root);
      expect(await fs.readFile(path.join(root, 'a/package.json'), 'utf8')).toBe(
        '{}',
      );
    } finally {
      await removeFixtureRoot(root);
    }
    await expect(fs.stat(root)).rejects.toThrow();
  });
});
