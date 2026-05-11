/**
 * #1 regression: the ESM-only `@rslint/eslint-plugin-runner` package must
 * never be STATICALLY imported by CompatPool. The extension bundles to
 * CJS with the package external; a static import compiles to a top-level
 * `require("@rslint/eslint-plugin-runner")`, which throws ERR_REQUIRE_ESM
 * on VS Code hosts whose Node can't require ESM (Node < 22 / VS Code
 * <= 1.100), crashing extension activation. It must be reached only via a
 * lazy dynamic `import()`. CI missed the original bug because its suites
 * run on Node 22+ (where require-of-ESM is permitted).
 */
import { describe, test, expect } from '@rstest/core';
import path from 'node:path';
import { existsSync } from 'node:fs';
import { build } from 'esbuild';

describe('CompatPool ESM-only runner import', () => {
  test('CJS bundle reaches the runner only via dynamic import, never top-level require', async () => {
    // rstest transforms test files into a temp dir, so import.meta.url is
    // unreliable for locating src/. The unit suite runs from the package
    // root (pnpm), so resolve the source against cwd.
    const entry = path.resolve(process.cwd(), 'src/CompatPool.ts');
    expect(existsSync(entry)).toBe(true);

    const result = await build({
      entryPoints: [entry],
      bundle: true,
      format: 'cjs',
      platform: 'node',
      write: false,
      // Same externals the real extension build uses (scripts/build.js).
      external: ['@rslint/eslint-plugin-runner', '@rslint/core', 'vscode'],
      logLevel: 'silent',
    });
    const out = result.outputFiles[0]!.text;

    // The crash signature: a top-level (eager) require of the ESM package.
    expect(out).not.toMatch(
      /require\(\s*["']@rslint\/eslint-plugin-runner["']\s*\)/,
    );
    // The package must still be referenced — via a lazy dynamic import.
    expect(out).toMatch(
      /import\(\s*["']@rslint\/eslint-plugin-runner["']\s*\)/,
    );
  });
});
