/**
 * Unit tests for `applyParentIgnoresFilter`. These exercise the
 * LSP-side parent-ignores wiring without spinning up VS Code:
 *
 *   - the helper takes the LSP-shaped record (carrying both
 *     `configDirectoryFsPath` and `configDirectoryUri`)
 *   - delegates to the shared `filterConfigsByParentIgnores`
 *     (`packages/rslint/src/utils/config-discovery.ts`, already unit-
 *     tested in `packages/rslint/tests/config-discovery.test.ts` —
 *     we don't re-test that algorithm here, only the LSP-side
 *     projection)
 *   - returns the surviving entries with their URI form intact
 *
 * The bug this regression-guards (review Finding 2):
 *
 *   CLI runs `filterConfigsByParentIgnores(loaded)` at cli.ts:253;
 *   LSP previously did NOT, so a root config with
 *   `{ ignores: ['fixtures/**'] }` and a nested
 *   `fixtures/sub/rslint.config.mjs` would: CLI silently skip the
 *   nested config; LSP load it and route `fixtures/sub/x.ts` through
 *   it. Editor / CLI diagnostics diverged on the same file.
 */
import { describe, test, expect } from '@rstest/core';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';

import { applyParentIgnoresFilter } from '../src/workspace-config';

function mkRecord(fsPath: string, configFile: string, entries: unknown[]) {
  return {
    configDirectoryFsPath: fsPath,
    configDirectoryUri: `file://${fsPath}`,
    configPath: configFile,
    entries: entries as never,
  };
}

describe('applyParentIgnoresFilter', () => {
  test('single config is returned as-is', () => {
    const root = '/abs/proj';
    const result = applyParentIgnoresFilter([
      mkRecord(root, `${root}/rslint.config.mjs`, [
        { ignores: ['ignored/**'] },
      ]),
    ]);
    expect(result).toHaveLength(1);
    // URI form is preserved on the way out — Go's `jsConfigs` map
    // keys off this, NOT the fs path.
    expect(result[0].configDirectoryUri).toBe(`file://${root}`);
  });

  test('empty input returns empty', () => {
    expect(applyParentIgnoresFilter([])).toHaveLength(0);
  });

  test('nested config under root ignore is dropped', () => {
    // The shared helper uses fs.realpathSync on configDirectory to
    // resolve symlinks. mkdir a real tmpdir tree so realpath works
    // and ancestor-startsWith checks succeed.
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pi-'));
    try {
      const rootDir = tmp;
      const ignoredDir = path.join(tmp, 'ignored', 'sub');
      fs.mkdirSync(ignoredDir, { recursive: true });
      const result = applyParentIgnoresFilter([
        mkRecord(rootDir, path.join(rootDir, 'rslint.config.mjs'), [
          { ignores: ['ignored/**'] },
        ]),
        mkRecord(ignoredDir, path.join(ignoredDir, 'rslint.config.mjs'), [
          { files: ['*.ts'] },
        ]),
      ]);
      expect(result).toHaveLength(1);
      expect(result[0].configPath).toMatch(
        /rslint-pi-.+\/rslint\.config\.mjs$/,
      );
      // The nested one is gone.
      expect(
        result.some((r) => r.configPath.includes(`${path.sep}sub${path.sep}`)),
      ).toBe(false);
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  test('nested config NOT under any ignore is kept', () => {
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pi-'));
    try {
      const rootDir = tmp;
      const sibDir = path.join(tmp, 'packages', 'sib');
      fs.mkdirSync(sibDir, { recursive: true });
      const result = applyParentIgnoresFilter([
        mkRecord(rootDir, path.join(rootDir, 'rslint.config.mjs'), [
          // Different glob — does NOT cover packages/sib
          { ignores: ['ignored/**'] },
        ]),
        mkRecord(sibDir, path.join(sibDir, 'rslint.config.mjs'), [
          { files: ['*.ts'] },
        ]),
      ]);
      expect(result).toHaveLength(2);
      // URIs preserved on both survivors.
      const uris = result.map((r) => r.configDirectoryUri).sort();
      expect(uris).toEqual([`file://${rootDir}`, `file://${sibDir}`].sort());
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  test('input order of survivors is preserved', () => {
    // Helper is documented to filter, not reorder. Important because
    // downstream `getConfigForURI`-equivalent lookups assume a stable
    // ordering across reconfigure events; reorder would race the
    // VS Code extension's fingerprint comparison and produce
    // spurious pool reconfigures.
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pi-'));
    try {
      const a = path.join(tmp, 'a');
      const b = path.join(tmp, 'b');
      const c = path.join(tmp, 'c');
      fs.mkdirSync(a, { recursive: true });
      fs.mkdirSync(b, { recursive: true });
      fs.mkdirSync(c, { recursive: true });
      const result = applyParentIgnoresFilter([
        mkRecord(a, `${a}/rslint.config.mjs`, []),
        mkRecord(b, `${b}/rslint.config.mjs`, []),
        mkRecord(c, `${c}/rslint.config.mjs`, []),
      ]);
      expect(result.map((r) => r.configDirectoryFsPath)).toEqual([a, b, c]);
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });

  test('preserves both `configDirectoryFsPath` and `configDirectoryUri` on survivors', () => {
    // Regression guard: an earlier inline implementation projected to
    // the helper's shape then forgot to project back, dropping the
    // URI form. Without the URI, Go's `jsConfigs` map keys off `""`
    // and every file routes through the same (empty) entry.
    const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pi-'));
    try {
      const r = applyParentIgnoresFilter([
        mkRecord(tmp, `${tmp}/rslint.config.mjs`, []),
      ]);
      expect(r[0].configDirectoryFsPath).toBe(tmp);
      expect(r[0].configDirectoryUri).toBe(`file://${tmp}`);
      expect(r[0].configPath).toBe(`${tmp}/rslint.config.mjs`);
    } finally {
      fs.rmSync(tmp, { recursive: true, force: true });
    }
  });
});
