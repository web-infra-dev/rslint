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
 * Exercise the CLI/API private host entry and the public entry used by LSP
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
 * The termination case builds a separate feature-only addon to stop a real
 * worker inside parseSharedBytes while its native reader holds the mapping.
 *
 * Requires `dist/`, the host Go binary (built by `pnpm build`) and the host
 * platform package's `.node` (built by `pnpm --filter @rslint/native build`).
 * The termination cases also require Cargo; their debug/release feature addons
 * use a separate target directory and never replace the platform package.
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
let configured = 0;
const configure = native.MemoryArena.prototype.configure;
native.MemoryArena.prototype.configure = function(config) {
  assert.equal(configured, 0);
  assert.equal(registered, 0);
  const fd = this.fd();
  if (fd !== undefined && fd !== null) assert.equal(fs.fstatSync(fd).size, 0);
  for (const field of ['version', 'slotCount', 'slotSize', 'headerSize', 'publicationStride']) {
    assert.equal(Number.isInteger(config[field]) && config[field] > 0, true);
  }
  configure.call(this, config);
  configured++;
  assert.equal(this.descriptor().version, config.version);
};
const register = native.MemoryArena.prototype.register;
native.MemoryArena.prototype.register = function(batches) {
  assert.equal(configured, 1);
  const lease = register.call(this, batches);
  const length = batches.reduce((total, batch) => total + batch.length, 0);
  const source = native.parseSharedBytes(file, { lease, offset: 0, length }, 'module', false);
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
assert.equal(configured, 1, 'CLI must configure storage through its Go peer');
assert.equal(registered, 1, 'CLI must register shared source, not use inline fallback');
assert.equal(fs.readFileSync(file, 'utf8'), changed);
console.log('PACKAGED_SHARED_OK');
`;

const API_RUNNER = `import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
const here = path.dirname(fileURLToPath(import.meta.url));
const cfgDir = path.join(here, 'cfg');
const { Rslint } = await import('./dist/index.js');
const lint = new Rslint({ cwd: cfgDir });
try {
  for (let call = 0; call < 2; call++) {
    const results = await lint.lintText(
      fs.readFileSync(path.join(cfgDir, 'input.ts'), 'utf8'),
      { filePath: 'input.ts' },
    );
    assert.equal(results.length, 1);
    assert.equal(results[0].messages.length, 1);
    assert.equal(results[0].messages[0].ruleId, 'pkg/no-null');
    assert.equal(results[0].messages[0].message, 'no null');
  }
} finally {
  await lint.close();
}
console.log('PACKAGED_API_OK');
`;

const TERMINATION_RUNNER = `import assert from 'node:assert/strict';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath, pathToFileURL } from 'node:url';
const here = path.dirname(fileURLToPath(import.meta.url));
const entry = path.join(here, 'eslint-plugin', 'index.js');
const require = createRequire(entry);
const nativePath = require.resolve(${JSON.stringify(`@rslint/${PKG_BASE}`)});
const native = require(nativePath);
const { WorkerPool } = await import(pathToFileURL(entry).href);
const closeBeforeResume = process.argv[2] === 'close';
const cfgDir = path.join(here, 'cfg');
const text = '\\ufeffconst pinned = null; // café 😀 retained through arena.close()\\r\\n';
const expected = native.parse('pinned.ts', text.slice(1), 'module', false).program;
const { arena, source } = native.createWorkerTerminationFixture(text);
const logs = [];
const pool = new WorkerPool({
  configs: [{ configPath: path.join(cfgDir, 'rslint.config.mjs'), configDirectory: cfgDir }],
  workerCount: 1,
  warmupWorkerCount: 1,
  taskTimeoutMs: 1,
  retryCap: 0,
  onLog: (record) => logs.push(record),
});
try {
  await pool.init();
  const response = pool.lintBatch([{
    filePath: 'pinned.ts',
    sharedSource: source,
    configKey: cfgDir,
    rules: { 'pkg/no-null': { options: [] } },
    collectFixes: false,
    suggestionsMode: 'off',
  }]);
  // This event is emitted only after with_bytes has pinned the lease and the
  // real native parser entry has decoded UTF-8/BOM. Blocking this JS thread
  // keeps the task timer from firing until the worker has entered the barrier.
  native.waitForWorkerParse();
  // The pool timeout callback requests Worker.terminate() before this promise
  // continuation runs. The worker is still blocked inside the native call.
  const result = await response;
  assert.equal(result.length, 1);
  assert.equal(result[0].parseError, 'task_timeout');
  assert.deepEqual(result[0].diagnostics, []);
  assert.equal(arena.release(source.lease), false, 'an active native reader must prevent reuse');
  assert.throws(() => native.parseSharedBytes('pinned.ts', source, 'module', false), /invalid or expired/);
  native.republishWorkerFixture(arena);
  assert.throws(() => arena.register([{ slot: 0, generation: 2, length: source.length }]), /invalid or expired/);
  if (closeBeforeResume) {
    arena.close();
    assert.throws(() => arena.descriptor(), /invalid or expired/);
  }
  // Node termination is pending while synchronous native code runs. Releasing
  // this barrier lets Oxc finish, including after the optional arena.close().
  native.resumeWorkerParse();
  await pool.shutdown();
  assert.equal(native.workerParsedProgram(), expected, 'Oxc must finish from the pinned bytes');
  if (!closeBeforeResume) {
    assert.equal(arena.release(source.lease), true, 'reader exit must restore capacity in the same arena');
    assert.equal(arena.release(source.lease), false, 'reclamation must succeed exactly once');
    const lease = arena.register([{ slot: 0, generation: 2, length: source.length }]);
    assert.throws(() => native.parseSharedBytes('pinned.ts', source, 'module', false), /invalid or expired/);
    assert.equal(native.parseSharedBytes('pinned.ts', { ...source, lease }, 'module', false).parsed.program, expected);
    assert.equal(arena.release(lease), true);
  }
  assert.equal(logs.some((record) => record.level === 'error'), false);
  console.log('PACKAGED_TERMINATION_OK');
} finally {
  arena.close();
  native.resumeWorkerParse();
  await pool.shutdown();
}
`;

function stageNative(
  root: string,
  withBinary = false,
  testAddon?: string,
): void {
  const nativeDir = path.join(root, 'node_modules', '@rslint', PKG_BASE);
  fs.mkdirSync(nativeDir, { recursive: true });
  const srcPkgDir = path.dirname(
    require.resolve(`@rslint/${PKG_BASE}/package.json`),
  );
  const srcNode = testAddon ?? path.join(srcPkgDir, NODE_FILE);
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

/** A feature addon never enters the platform package used by production builds. */
function buildTerminationAddon(profile: 'debug' | 'release'): string {
  const repoRoot = path.resolve(__dirname, '../../../..');
  const target = path.join(repoRoot, 'target', 'worker-termination-test');
  const build = spawnSync(
    'cargo',
    [
      'build',
      '--locked',
      ...(profile === 'release' ? ['--release'] : []),
      '-p',
      'rslint-native',
      '--features',
      'test-worker-termination',
      '--target-dir',
      target,
    ],
    {
      cwd: repoRoot,
      // A release job may export a cross-compilation target. This addon must
      // use Cargo's host build and its ordinary debug/release output directory.
      env: { ...process.env, CARGO_BUILD_TARGET: undefined },
      encoding: 'utf8',
      timeout: PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS,
      killSignal: 'SIGKILL',
      maxBuffer: 16 * 1024 * 1024,
    },
  );
  if (build.error || build.signal || build.status !== 0) {
    throw new Error(
      `worker termination addon build failed: ${build.error ?? build.signal ?? build.status}\n${build.stderr}`,
    );
  }
  const library =
    process.platform === 'win32'
      ? 'rslint_native.dll'
      : process.platform === 'darwin'
        ? 'librslint_native.dylib'
        : 'librslint_native.so';
  return path.join(target, profile, library);
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

    test.each(
      (['debug', 'release'] as const).flatMap((profile) =>
        (['close', 'reuse'] as const).map((ending) => ({ profile, ending })),
      ),
    )(
      'terminating a real worker preserves native reads and supports $ending ($profile)',
      ({ profile, ending }) => {
        const root = path.join(tmp, `worker-termination-${profile}-${ending}`);
        fs.mkdirSync(root, { recursive: true });
        stage(root, { withNative: false });
        stageNative(
          path.join(root, 'eslint-plugin'),
          false,
          buildTerminationAddon(profile),
        );
        fs.writeFileSync(path.join(root, 'runner.mjs'), TERMINATION_RUNNER);
        const result = spawnSync(
          process.execPath,
          [path.join(root, 'runner.mjs'), ending],
          {
            cwd: root,
            encoding: 'utf8',
            timeout: PACKAGED_CHILD_DEAD_PROCESS_WATCHDOG_MS,
            killSignal: 'SIGKILL',
            maxBuffer: 16 * 1024 * 1024,
            env: { ...process.env, NODE_PATH: '' },
          },
        );
        expect(result.error).toBeUndefined();
        expect(result.signal, result.stderr).toBeNull();
        expect(result.status, result.stderr).toBe(0);
        expect(result.stdout.trim()).toBe('PACKAGED_TERMINATION_OK');
        expect(result.stderr.trim()).toBe('');
      },
      PACKAGED_OUTER_DEADLOCK_SENTINEL_MS,
    );

    test.each(['CLI', 'API'] as const)(
      'complete %s uses its staged host and worker',
      (mode) => {
        const root = path.join(tmp, `complete-${mode}`);
        fs.mkdirSync(root, { recursive: true });
        fs.cpSync(
          path.resolve(__dirname, '../../dist'),
          path.join(root, 'dist'),
          { recursive: true },
        );
        fs.copyFileSync(
          path.resolve(__dirname, '../../package.json'),
          path.join(root, 'package.json'),
        );
        if (mode === 'API') {
          // Preserve the real exports map so a package self-reference still
          // resolves, but fail if the API evaluates the full public runtime.
          fs.writeFileSync(
            path.join(root, 'dist', 'eslint-plugin', 'index.js'),
            "throw new Error('API loaded the full public plugin runtime');\n",
          );
        }
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
        fs.writeFileSync(
          path.join(root, 'runner.mjs'),
          mode === 'API' ? API_RUNNER : SHARED_RUNNER,
        );
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
        if (mode === 'CLI') {
          expect(result.stdout).toContain('no null');
        }
        expect(result.stdout).toContain(
          mode === 'API' ? 'PACKAGED_API_OK' : 'PACKAGED_SHARED_OK',
        );
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
