/**
 * Unit tests for the pure helpers backing CompatPool. Run via rstest
 * (`pnpm --filter rslint test:unit`) — no VS Code runtime needed.
 *
 * Coverage:
 *
 *   - fingerprintConfigs: `(configPath, mtime, size, hasPlugin)`-based
 *     identity. Drives the tier-1 "unchanged → noop" check in
 *     CompatPool.reconfigure. Must change when (a) any config file's
 *     on-disk identity changes, (b) the set of configs changes,
 *     (c) the plugin-bearing state flips.
 *   - extractConfigDescriptors: filters out plugin-less configs, drops
 *     configs without a real `configPath`, converts URI directories to
 *     filesystem paths so the values byte-match what Go writes into
 *     `CompatLintFile.ConfigKey`. Producing the wrong path shape here
 *     silently breaks per-file routing across an entire workspace.
 *   - uriToPath: URL shape gotchas that bit us when first connecting
 *     Node's createRequire to LSP.
 */

import os from 'node:os';
import path from 'node:path';
import { existsSync, mkdirSync, utimesSync, writeFileSync } from 'node:fs';
import { pathToFileURL } from 'node:url';

import { describe, test, expect } from '@rstest/core';

import {
  extractConfigDescriptors,
  fingerprintConfigs,
  uriToPath,
  type NormalizedConfig,
} from '../src/compat-pool-helpers';

// ── Fixture infrastructure ─────────────────────────────────────────

const FIXTURE_ROOT = path.join(os.tmpdir(), `rslint-cph-test-${process.pid}`);

interface FixtureDescriptor {
  configDirectory: string; // URI
  configPath: string; // fs path
  dir: string; // fs path of the directory (test-only convenience)
}

function ensureFixture(
  suffix: string,
  content = 'export default [];\n',
): FixtureDescriptor {
  const dir = path.join(FIXTURE_ROOT, suffix);
  if (!existsSync(dir)) mkdirSync(dir, { recursive: true });
  const configPath = path.join(dir, 'rslint.config.mjs');
  writeFileSync(configPath, content);
  return {
    configDirectory: pathToFileURL(dir).toString(),
    configPath,
    dir,
  };
}

function makeConfig(
  fix: FixtureDescriptor,
  prefixes?: string[],
): NormalizedConfig {
  return {
    configDirectory: fix.configDirectory,
    configPath: fix.configPath,
    entries: prefixes
      ? [
          {
            eslintPlugins: prefixes.map((p) => ({
              prefix: p,
              ruleNames: [],
            })),
          },
        ]
      : [{}],
  };
}

// ── Tests ──────────────────────────────────────────────────────────

describe('fingerprintConfigs', () => {
  test('stable and order-independent for the same set of configs', () => {
    const a = ensureFixture('fp-stable-a');
    const b = ensureFixture('fp-stable-b');
    const fpA = fingerprintConfigs([
      makeConfig(a, ['uc']),
      makeConfig(b, ['imp']),
    ]);
    const fpB = fingerprintConfigs([
      makeConfig(b, ['imp']),
      makeConfig(a, ['uc']),
    ]);
    expect(fpA).toBe(fpB);
  });

  test('changes when a config file is edited (mtime/size shift)', () => {
    const f = ensureFixture('fp-edit', 'export default [];\n');
    const before = fingerprintConfigs([makeConfig(f, ['uc'])]);
    writeFileSync(f.configPath, 'export default [\n  { rules: {} },\n];\n');
    const after = fingerprintConfigs([makeConfig(f, ['uc'])]);
    expect(after).not.toBe(before);
  });

  test('changes when the set of configs changes', () => {
    const a = ensureFixture('fp-set-a');
    const b = ensureFixture('fp-set-b');
    expect(fingerprintConfigs([makeConfig(a, ['uc'])])).not.toBe(
      fingerprintConfigs([makeConfig(a, ['uc']), makeConfig(b, ['imp'])]),
    );
  });

  test('changes when a config flips from plugin-bearing to empty', () => {
    const f = ensureFixture('fp-toggle');
    const withPlugin = fingerprintConfigs([makeConfig(f, ['uc'])]);
    const empty = fingerprintConfigs([makeConfig(f)]);
    expect(withPlugin).not.toBe(empty);
  });

  test('stable when nothing on disk or in the input changed', () => {
    const f = ensureFixture('fp-noop');
    // Pin mtime so two stat() calls within the same test see the same
    // value even on filesystems with sub-second mtime resolution.
    const now = new Date('2026-01-01T00:00:00Z');
    utimesSync(f.configPath, now, now);
    const a = fingerprintConfigs([makeConfig(f, ['uc'])]);
    const b = fingerprintConfigs([makeConfig(f, ['uc'])]);
    expect(a).toBe(b);
  });

  test('missing config file produces a sentinel fingerprint that differs from a present one', () => {
    const ghost: NormalizedConfig = {
      configDirectory: pathToFileURL(
        path.join(FIXTURE_ROOT, 'fp-missing-ghost-dir'),
      ).toString(),
      configPath: path.join(
        FIXTURE_ROOT,
        'fp-missing-ghost-dir',
        'rslint.config.mjs',
      ),
      entries: [
        {
          eslintPlugins: [{ prefix: 'uc', ruleNames: [] }],
        },
      ],
    };
    const present = ensureFixture('fp-missing-present');
    const fpGhost = fingerprintConfigs([ghost]);
    const fpPresent = fingerprintConfigs([makeConfig(present, ['uc'])]);
    expect(fpGhost).not.toBe(fpPresent);
    // Sentinel stays stable across calls (statSync miss is deterministic).
    expect(fingerprintConfigs([ghost])).toBe(fpGhost);
  });

  test('empty input has a stable distinct fingerprint', () => {
    const empty = fingerprintConfigs([]);
    expect(fingerprintConfigs([])).toBe(empty);
    const present = ensureFixture('fp-empty-vs-present');
    expect(empty).not.toBe(fingerprintConfigs([makeConfig(present, ['uc'])]));
  });
});

describe('extractConfigDescriptors', () => {
  test('emits descriptors only for configs that have at least one plugin', () => {
    const empty = ensureFixture('extract-empty');
    const withPlugin = ensureFixture('extract-withPlugin');
    const configs: NormalizedConfig[] = [
      makeConfig(empty), // no eslintPlugins
      makeConfig(withPlugin, ['uc']),
    ];
    const descriptors = extractConfigDescriptors(configs);
    expect(descriptors).toHaveLength(1);
    expect(descriptors[0].configPath).toBe(withPlugin.configPath);
  });

  test('converts URI configDirectory to filesystem path (byte-matches Go)', () => {
    // CRITICAL: Go writes filesystem PATHS into each lintBatch file's
    // configKey. If extractConfigDescriptors shipped URI strings,
    // every worker `Map<configDirectory, LoadedPlugins>` lookup would
    // miss and per-file routing would silently fall back to "no
    // rules resolve". The previous architecture had exactly this bug.
    const f = ensureFixture('extract-uri-to-path');
    const out = extractConfigDescriptors([makeConfig(f, ['uc'])]);
    expect(out).toHaveLength(1);
    // Forward-slashed fs path on every platform.
    expect(out[0].configDirectory.includes('\\')).toBe(false);
    expect(out[0].configDirectory.startsWith('file://')).toBe(false);
  });

  test('skips configs whose URI does not parse', () => {
    const f = ensureFixture('extract-bad-uri');
    const corrupt: NormalizedConfig = {
      configDirectory: 'not a url',
      configPath: f.configPath,
      entries: [{ eslintPlugins: [{ prefix: 'uc', ruleNames: [] }] }],
    };
    expect(extractConfigDescriptors([corrupt])).toHaveLength(0);
  });

  test('skips configs without a configPath (legacy callers)', () => {
    const corrupt: NormalizedConfig = {
      configDirectory: 'file:///some/dir',
      configPath: '',
      entries: [{ eslintPlugins: [{ prefix: 'uc', ruleNames: [] }] }],
    };
    expect(extractConfigDescriptors([corrupt])).toHaveLength(0);
  });

  test('keeps every plugin-bearing config in a monorepo set', () => {
    // Two distinct configs in a monorepo each contribute their own
    // ConfigDescriptor → the worker imports both, builds two
    // LoadedPlugins, and per-file dispatch picks the right one via
    // `configKey`. Naive de-duping (one-per-prefix) would collapse
    // these and break monorepo routing.
    const pkgA = ensureFixture('extract-pkg-a');
    const pkgB = ensureFixture('extract-pkg-b');
    const out = extractConfigDescriptors([
      makeConfig(pkgA, ['uc']),
      makeConfig(pkgB, ['uc']), // same prefix, different dir
    ]);
    expect(out).toHaveLength(2);
    const paths = out.map((d) => d.configPath).sort();
    expect(paths).toEqual([pkgA.configPath, pkgB.configPath].sort());
  });
});

describe('uriToPath', () => {
  test('extracts pathname from file:// URLs', () => {
    expect(uriToPath('file:///projects/a')).toBe('/projects/a');
  });

  test('returns empty for invalid input rather than throwing', () => {
    expect(uriToPath('')).toBe('');
    expect(uriToPath('not even a url')).toBe('');
  });

  // ── A9 regression coverage ────────────────────────────────────
  //
  // Go's internal/lsp.uriToPath percent-decodes the path and strips
  // the leading `/` from `/C:/...` on Windows. The extension's
  // configsByDir keys MUST byte-match what Go writes into each
  // lintBatch file's configKey, otherwise per-file routing silently
  // miss-resolves on workspaces whose path contains a space, CJK
  // character, or Windows drive letter.

  test('percent-decodes spaces (%20) — Go side does this with url.ParseRequestURI', () => {
    expect(uriToPath('file:///Users/John%20Doe/project')).toBe(
      '/Users/John Doe/project',
    );
  });

  test('percent-decodes CJK paths (UTF-8)', () => {
    expect(uriToPath('file:///%E6%B5%8B%E8%AF%95/%E9%A1%B9%E7%9B%AE')).toBe(
      '/测试/项目',
    );
  });

  test('strips Windows drive-letter leading slash (Windows only)', () => {
    const p = uriToPath('file:///C:/Users/proj');
    expect(p.includes('\\')).toBe(false);
    expect(p.includes('%')).toBe(false);
    if (process.platform === 'win32') {
      // The actual invariant: on Windows, `/C:/Users/proj` must
      // become `C:/Users/proj` (drive letter without the leading
      // slash). On POSIX both `fileURLToPath` and our shim leave the
      // string alone, so the assertion below is Windows-specific.
      // Without this branch, the file:line:262 negative assertions
      // (`!\` + `!%`) trivially pass on POSIX CI regardless of the
      // drive-letter behavior.
      expect(p).toBe('C:/Users/proj');
    } else {
      // On POSIX: `fileURLToPath('file:///C:/Users/proj')` returns
      // `/C:/Users/proj`. Pin that so a future change that strips
      // the slash on POSIX too is noticed.
      expect(p).toBe('/C:/Users/proj');
    }
  });

  test('handles file URL with no path part', () => {
    const p = uriToPath('file:///');
    // Pin the result rather than just `typeof === string && length>0`.
    // Both fileURLToPath('file:///') (POSIX) and the Windows variant
    // produce '/' — anchor that explicitly so a regression returning
    // an arbitrary non-empty string doesn't slip through.
    expect(p).toBe('/');
  });
});
