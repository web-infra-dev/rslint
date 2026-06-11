import { describe, test, expect, beforeAll, afterAll } from '@rstest/core';
import { execFileSync } from 'node:child_process';
import { createRequire } from 'node:module';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';

/**
 * Packaged-layout isolation guard.
 *
 * The VS Code extension does NOT ship `@rslint/core`; instead `build.js` stages
 * the built `dist/eslint-plugin/` worker bundle into the vsix together with a
 * nested `node_modules/@rslint/native/` (a tiny shim loader + the platform
 * `.node`), so the worker resolves `@rslint/native` from THERE rather than from
 * a workspace `node_modules`. The original bug was that this worked in dev (the
 * dev host has the workspace package) but silently produced zero diagnostics
 * once packaged. The normal worker-pool e2e suites can't catch a regression of
 * this because they resolve `@rslint/native` from the workspace.
 *
 * This test reproduces the packaged layout under `os.tmpdir()` — OFF any
 * `@rslint/core` / workspace `node_modules` resolution path — and runs the host
 * in a SUBPROCESS (so it neither inherits this suite's `setWorkerEntryForTests`
 * override nor any in-process module cache). It asserts the worker loads the
 * native parser from the nested shim and a plugin rule fires; a negative
 * control proves the host entry (`index.js`) itself — before any worker spawns
 * — can't fall back to a workspace `@rslint/native`.
 *
 * Requires `dist/eslint-plugin/` (built by `pnpm build`, the same prerequisite
 * the worker-pool e2e suites already document) and the host `@rslint/native`
 * `.node` (built by `pnpm --filter @rslint/native build`).
 */

const require = createRequire(import.meta.url);
const SHIM =
  "const binding = require('./rslint.node');\n" +
  'module.exports = binding;\n' +
  'module.exports.parse = binding.parse;\n';

// A self-contained `.mjs` plugin + config — `.mjs` needs no `jiti`, isolating
// this test to the native-parser resolution it is meant to guard.
const LOCAL_PLUGIN = `export default {
  meta: { name: 'lp', version: '1' },
  rules: {
    'no-null': {
      meta: { type: 'suggestion', schema: [], messages: { e: 'no null' } },
      create(c) {
        return { Literal(n) { if (n.raw === 'null') c.report({ node: n, messageId: 'e' }); } };
      },
    },
  },
};
`;
const CONFIG = `import lp from './local-plugin.mjs';
export default [{ plugins: { pkg: lp } }];
`;

// The runner imports the STAGED host by a path relative to its own location, so
// Node resolves `@rslint/native` by walking up from the staged worker into the
// staged nested node_modules — exactly the packaged resolution path.
const RUNNER = `import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
const here = path.dirname(fileURLToPath(import.meta.url));
const cfgDir = path.join(here, 'cfg');
const { createPluginLintHost } = await import(
  pathToFileURL(path.join(here, 'eslint-plugin', 'index.js')).href
);
const host = await createPluginLintHost([
  { configPath: path.join(cfgDir, 'rslint.config.mjs'), configDirectory: cfgDir },
]);
const res = await host.lint({
  files: [{ path: 'a.ts', text: 'const x = null;', configKey: cfgDir }],
  rules: { 'pkg/no-null': { options: [] } },
  fix: false,
  suggestionsMode: 'off',
});
await host.shutdown();
const d = res.results?.[0]?.diagnostics ?? [];
if (d.length === 1 && d[0].ruleName === 'pkg/no-null') {
  console.log('PACKAGED_OK');
  process.exit(0);
}
console.error('UNEXPECTED ' + JSON.stringify(res));
process.exit(1);
`;

/** Stage a packaged layout under `root`; omit the nested native for the negative control. */
function stage(root: string, opts: { withNative: boolean }): void {
  const coreEpDir = path.resolve(__dirname, '../../dist/eslint-plugin');
  if (!fs.existsSync(path.join(coreEpDir, 'index.js'))) {
    throw new Error(
      `built worker bundle missing at ${coreEpDir} — run \`pnpm build\` (or ` +
        '`pnpm --filter @rslint/core build:js`) before this test',
    );
  }
  const epDest = path.join(root, 'eslint-plugin');
  fs.cpSync(coreEpDir, epDest, { recursive: true });
  fs.writeFileSync(
    path.join(epDest, 'package.json'),
    JSON.stringify({ type: 'module' }),
  );

  if (opts.withNative) {
    const nativeDir = path.join(epDest, 'node_modules', '@rslint', 'native');
    fs.mkdirSync(nativeDir, { recursive: true });
    fs.writeFileSync(path.join(nativeDir, 'index.js'), SHIM);
    fs.writeFileSync(
      path.join(nativeDir, 'package.json'),
      JSON.stringify({
        name: '@rslint/native',
        type: 'commonjs',
        main: 'index.js',
      }),
    );
    const nativeSrcDir = path.dirname(require.resolve('@rslint/native'));
    const hostNode = fs
      .readdirSync(nativeSrcDir)
      .find((f) => /^rslint\..*\.node$/.test(f));
    if (!hostNode) {
      throw new Error(
        `no built @rslint/native .node in ${nativeSrcDir} — run ` +
          '`pnpm --filter @rslint/native build` before this test',
      );
    }
    fs.copyFileSync(
      path.join(nativeSrcDir, hostNode),
      path.join(nativeDir, 'rslint.node'),
    );
  }

  const cfgDir = path.join(root, 'cfg');
  fs.mkdirSync(cfgDir, { recursive: true });
  fs.writeFileSync(path.join(cfgDir, 'local-plugin.mjs'), LOCAL_PLUGIN);
  fs.writeFileSync(path.join(cfgDir, 'rslint.config.mjs'), CONFIG);
  fs.writeFileSync(path.join(root, 'runner.mjs'), RUNNER);
}

// Spawns a real worker that does native teardown, so it respects the same
// win32 kill-switch as the worker-pool e2e suites (flag is false → runs on
// win32 too, validating the napi-teardown mitigation).
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'packaged-layout isolation (vsix eslint-plugin worker)',
  () => {
    let tmp: string;
    beforeAll(() => {
      tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pkg-'));
    });
    afterAll(() => {
      if (tmp) fs.rmSync(tmp, { recursive: true, force: true });
    });

    test('worker loads @rslint/native from the nested shim and a plugin rule fires', () => {
      const root = path.join(tmp, 'ok');
      fs.mkdirSync(root, { recursive: true });
      stage(root, { withNative: true });
      // timeout + SIGKILL so a worker wedged in native teardown (the win32
      // abort this validates) fails loudly instead of hanging CI forever.
      const out = execFileSync('node', ['runner.mjs'], {
        cwd: root,
        encoding: 'utf8',
        timeout: 30_000,
        killSignal: 'SIGKILL',
      });
      expect(out).toContain('PACKAGED_OK');
    });

    test('without the nested @rslint/native the host fails (no workspace fallback)', () => {
      const root = path.join(tmp, 'no-native');
      fs.mkdirSync(root, { recursive: true });
      stage(root, { withNative: false });
      let stderr = '';
      let threw = false;
      try {
        execFileSync('node', ['runner.mjs'], {
          cwd: root,
          encoding: 'utf8',
          timeout: 30_000,
          killSignal: 'SIGKILL',
        });
      } catch (err) {
        threw = true;
        stderr = String((err as { stderr?: string }).stderr ?? '');
      }
      expect(threw).toBe(true);
      expect(stderr).toContain('@rslint/native');
      expect(stderr).toContain('ERR_MODULE_NOT_FOUND');
    });
  },
);
