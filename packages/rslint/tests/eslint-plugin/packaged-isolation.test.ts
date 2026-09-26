import { describe, test, expect, beforeAll, afterAll } from 'rstack/test';
import { spawnSync } from 'node:child_process';
import { createRequire } from 'node:module';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';
import { platformTuple } from '../../src/native/platform-tuple.js';

/**
 * Packaged-layout isolation guard.
 *
 * Exercise the CLI's private host entry and the public entry used by API/LSP
 * consumers outside the repository's dependency resolution paths. The VS Code
 * extension resolves a project-local `@rslint/core`; it does not bundle a copy
 * of this runtime. Ordinary worker-pool tests can resolve workspace packages,
 * which would hide missing runtime dependencies in an installed package.
 *
 * This test reproduces the packaged layout under `os.tmpdir()` — OFF any
 * `@rslint/core` / workspace `node_modules` resolution path — and runs the host
 * in a SUBPROCESS (so it neither inherits this suite's `setWorkerEntryForTests`
 * override nor any in-process module cache). It asserts the worker loads the
 * native parser from a nested platform package and a plugin rule fires. The
 * lightweight host imports without that package, but creating workers must
 * still reject. The full public entry rejects during import instead. Neither
 * entry may fall back to a workspace platform package.
 *
 * The complete CLI case additionally proves that its arena and worker parser
 * share the staged native registry, preserving Go's snapshot after a disk edit.
 * It does not exercise termination during a native read.
 *
 * Requires `dist/`, the host Go binary (built by `pnpm build`) and the host
 * platform package's `.node` (built by `pnpm --filter @rslint/native build`).
 */

const require = createRequire(import.meta.url);
const TUPLE = platformTuple();
const PKG_BASE = `native-${TUPLE}`;
const NODE_FILE = `rslint.${TUPLE}.node`;
// Success is the validated process close, not a duration. This only releases
// an isolated packaged-layout process that has not exited for 30 minutes.
const PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS = 30 * 60_000;
const PACKAGED_OUTER_DEADLOCK_SENTINEL_MS = 35 * 60_000;

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
const HOST_ENTRIES = ['index.js', 'host.js'] as const;

// The runner imports the STAGED host by a path relative to its own location, so
// the worker's loader resolves the platform package by walking up from the
// staged worker into the staged nested node_modules — exactly the packaged
// resolution path.
const RUNNER = `import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
const here = path.dirname(fileURLToPath(import.meta.url));
const cfgDir = path.join(here, 'cfg');
const { createPluginLintHost } = await import(
  pathToFileURL(path.join(here, 'eslint-plugin', process.argv[2] ?? 'index.js')).href
);
const host = await createPluginLintHost([
  { configPath: path.join(cfgDir, 'rslint.config.mjs'), configDirectory: cfgDir },
]);
const res = await host.lint({
  files: [{ path: 'a.ts', text: 'const x = null;', configKey: cfgDir }],
  rules: { 'pkg/no-null': { options: [] } },
  collectFixes: false,
  suggestionsMode: 'off',
});
await host.shutdown();
const d = res.results?.[0]?.diagnostics ?? [];
if (d.length === 1 && d[0].ruleName === 'pkg/no-null') {
  console.log('PACKAGED_OK');
  process.exitCode = 0;
} else {
  console.error('UNEXPECTED ' + JSON.stringify(res));
  process.exitCode = 1;
}
`;

const SHARED_RUNNER = `import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
const here = path.dirname(fileURLToPath(import.meta.url));
const parentRequire = createRequire(path.join(here, 'dist', 'cli.js'));
const workerRequire = createRequire(path.join(here, 'dist', 'eslint-plugin', 'lint-worker.js'));
const nativePackage = ${JSON.stringify(`@rslint/${PKG_BASE}`)};
const resolved = parentRequire.resolve(nativePackage);
assert.equal(resolved, workerRequire.resolve(nativePackage));
assert.ok(resolved.startsWith(here + path.sep));
const native = parentRequire(nativePackage);
const file = path.join(here, 'cfg', 'input.ts');
const original = fs.readFileSync(file, 'utf8');
const changed = 'const changedOnDisk = false;';
let registered = 0;
const register = native.SourceArena.prototype.register;
native.SourceArena.prototype.register = function(slot, generation, length) {
  const lease = register.call(this, slot, generation, length);
  const source = native.parseSharedSource(file, { lease, offset: 0, length }, 'module', false);
  assert.equal(source.sourceText, original.slice(1));
  assert.equal(source.hadBom, true);
  registered++;
  // Observe the real CLI registration after Go published its snapshot, before
  // the adapter sends the capability to the real worker. No transport is mocked.
  fs.writeFileSync(file, changed);
  return lease;
};
const { run } = await import(pathToFileURL(path.join(here, 'dist', 'cli.js')).href);
const code = await run(parentRequire.resolve(nativePackage + '/bin'), ['--no-color', file], Date.now());
assert.equal(code, 1);
assert.equal(registered, 1, 'CLI must register shared source, not use inline fallback');
assert.equal(fs.readFileSync(file, 'utf8'), changed);
console.log('PACKAGED_SHARED_OK');
`;

function stageNative(root: string, withBinary = false): void {
  const nativeDir = path.join(root, 'node_modules', '@rslint', PKG_BASE);
  fs.mkdirSync(nativeDir, { recursive: true });
  const srcPkgDir = path.dirname(
    require.resolve(`@rslint/${PKG_BASE}/package.json`),
  );
  const srcNode = path.join(srcPkgDir, NODE_FILE);
  if (!fs.existsSync(srcNode)) {
    throw new Error(
      `no built ${NODE_FILE} in ${srcPkgDir} — run ` +
        '`pnpm --filter @rslint/native build` before this test',
    );
  }
  fs.copyFileSync(srcNode, path.join(nativeDir, NODE_FILE));
  const exports: Record<string, string> = { '.': `./${NODE_FILE}` };
  if (withBinary) {
    const binary = require.resolve(`@rslint/${PKG_BASE}/bin`);
    const name = path.basename(binary);
    fs.copyFileSync(binary, path.join(nativeDir, name));
    exports['./bin'] = `./${name}`;
  }
  fs.writeFileSync(
    path.join(nativeDir, 'package.json'),
    JSON.stringify({ name: `@rslint/${PKG_BASE}`, exports }),
  );
}

/** Stage a packaged layout under `root`; omit the nested native for the negative control. */
function stage(root: string, opts: { withNative: boolean }): void {
  const coreEpDir = path.resolve(__dirname, '../../dist/eslint-plugin');
  if (
    [...HOST_ENTRIES, 'lint-worker.js'].some(
      (entry) => !fs.existsSync(path.join(coreEpDir, entry)),
    )
  ) {
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
    // Stage the host platform package `@rslint/native-<tuple>` (minimal
    // package.json + the `.node` under its real name) — what the worker's
    // loader resolves at runtime.
    stageNative(epDest);
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
  'packaged-layout isolation (plugin host and worker)',
  () => {
    let tmp: string;
    beforeAll(() => {
      tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'rslint-pkg-'));
    });
    afterAll(() => {
      if (tmp) fs.rmSync(tmp, { recursive: true, force: true });
    });

    test(
      'complete CLI and worker share the staged native source registry',
      () => {
        const root = path.join(tmp, 'shared-cli');
        fs.mkdirSync(root, { recursive: true });
        fs.cpSync(
          path.resolve(__dirname, '../../dist'),
          path.join(root, 'dist'),
          { recursive: true },
        );
        fs.writeFileSync(
          path.join(root, 'package.json'),
          JSON.stringify({ type: 'module' }),
        );
        stageNative(root, true);
        fs.cpSync(
          path.dirname(require.resolve('picomatch/package.json')),
          path.join(root, 'node_modules', 'picomatch'),
          { recursive: true },
        );
        const cfgDir = path.join(root, 'cfg');
        fs.mkdirSync(cfgDir);
        fs.writeFileSync(path.join(cfgDir, 'local-plugin.mjs'), LOCAL_PLUGIN);
        fs.writeFileSync(
          path.join(cfgDir, 'rslint.config.mjs'),
          `import lp from './local-plugin.mjs';
export default [{ files: ['**/*.ts'], plugins: { pkg: lp }, rules: { 'pkg/no-null': 'error' } }];`,
        );
        fs.writeFileSync(
          path.join(cfgDir, 'input.ts'),
          '\ufeffconst sample = null; // café 😀\r\n',
        );
        fs.writeFileSync(path.join(root, 'runner.mjs'), SHARED_RUNNER);
        const result = spawnSync(
          process.execPath,
          [path.join(root, 'runner.mjs')],
          {
            cwd: cfgDir,
            encoding: 'utf8',
            timeout: PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS,
            killSignal: 'SIGKILL',
            maxBuffer: 16 * 1024 * 1024,
            env: { ...process.env, NODE_PATH: '' },
          },
        );
        expect(result.error).toBeUndefined();
        expect(result.signal).toBeNull();
        expect(result.status, result.stderr).toBe(0);
        expect(result.stdout).toContain('no null');
        expect(result.stdout).toContain('PACKAGED_SHARED_OK');
        expect(result.stderr.trim()).toBe('');
      },
      PACKAGED_OUTER_DEADLOCK_SENTINEL_MS,
    );

    test.each(HOST_ENTRIES)(
      '%s worker loads the nested platform package and a plugin rule fires',
      (entry) => {
        const root = path.join(tmp, `ok-${entry}`);
        fs.mkdirSync(root, { recursive: true });
        stage(root, { withNative: true });
        if (entry === 'host.js') {
          // Nonempty hosts must also work without the full public runtime.
          fs.rmSync(path.join(root, 'eslint-plugin', 'index.js'));
        }
        // timeout + SIGKILL so a worker wedged in native teardown (the win32
        // abort this validates) fails loudly instead of hanging CI forever.
        // Clear NODE_PATH: rstest injects it pointing at the pnpm virtual store
        // (which holds the workspace platform packages). Clearing it ensures
        // only the staged dependency tree is available to either entry.
        const result = spawnSync(
          process.execPath,
          [path.join(root, 'runner.mjs'), entry],
          {
            // Worker paths follow the entry module, not the caller's directory.
            cwd: path.join(root, 'cfg'),
            encoding: 'utf8',
            timeout: PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS,
            killSignal: 'SIGKILL',
            maxBuffer: 16 * 1024 * 1024,
            env: { ...process.env, NODE_PATH: '' },
          },
        );
        expect(result.error).toBeUndefined();
        expect(result.signal).toBeNull();
        expect(result.status).toBe(0);
        expect(result.stdout.trim()).toBe('PACKAGED_OK');
        expect(result.stderr.trim()).toBe('');
      },
      PACKAGED_OUTER_DEADLOCK_SENTINEL_MS,
    );

    test('lightweight host imports without loading the worker runtime', () => {
      const root = path.join(tmp, 'host-without-native');
      fs.mkdirSync(root, { recursive: true });
      stage(root, { withNative: false });
      // An empty host needs neither entry. Nonempty configurations load the
      // worker, while the full public index remains unnecessary.
      fs.rmSync(path.join(root, 'eslint-plugin', 'index.js'));
      fs.rmSync(path.join(root, 'eslint-plugin', 'lint-worker.js'));
      const result = spawnSync(
        process.execPath,
        [
          '--input-type=module',
          '-e',
          `import { createPluginLintHost } from './eslint-plugin/host.js';
const host = await createPluginLintHost([]);
await host.shutdown();
console.log('HOST_OK');`,
        ],
        {
          cwd: root,
          encoding: 'utf8',
          timeout: PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS,
          killSignal: 'SIGKILL',
          env: { ...process.env, NODE_PATH: '' },
        },
      );
      expect(result.error).toBeUndefined();
      expect(result.signal).toBeNull();
      expect(result.status).toBe(0);
      expect(result.stdout.trim()).toBe('HOST_OK');
      expect(result.stderr.trim()).toBe('');
    });

    test.each(HOST_ENTRIES)(
      '%s rejects without the nested platform package (no workspace fallback)',
      (entry) => {
        const root = path.join(tmp, `no-native-${entry}`);
        fs.mkdirSync(root, { recursive: true });
        stage(root, { withNative: false });
        const result = spawnSync(process.execPath, ['runner.mjs', entry], {
          cwd: root,
          encoding: 'utf8',
          timeout: PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS,
          killSignal: 'SIGKILL',
          maxBuffer: 16 * 1024 * 1024,
          env: { ...process.env, NODE_PATH: '' },
        });
        expect(result.error).toBeUndefined();
        expect(result.signal).toBeNull();
        expect(result.status).toBe(1);
        // The public entry fails at import; the lightweight host propagates
        // its worker's initialization error and still drains every worker.
        expect(result.stderr).toContain('failed to load the native parser');
        expect(result.stderr).toContain(`@rslint/${PKG_BASE}`);
        expect(`${result.stdout}\n${result.stderr}`).not.toMatch(
          /task_timeout|worker_crashed|EXPECTED-NATIVE-ABORT/,
        );
      },
      PACKAGED_OUTER_DEADLOCK_SENTINEL_MS,
    );
  },
);
