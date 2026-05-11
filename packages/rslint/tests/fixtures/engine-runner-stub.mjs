/**
 * Test fixture: minimal host that invokes `runEngine` against a fake
 * Go binary so an integration test can spawn this as a real child
 * process and exercise the signal-forwarding path in engine.ts.
 *
 * ESM (.mjs) so we can import `runEngine` directly from the built
 * `dist/engine.js` (also ESM) without dynamic-import or package-
 * exports gymnastics. This keeps the stub a pure test fixture — no
 * production config (package.json exports, src/index.ts re-exports)
 * is altered to make the test work.
 *
 * Inputs (via env):
 *   FAKE_BIN_PATH — absolute path to a binary that conforms to
 *                   long-running-fake-binary.cjs's contract.
 *
 * Stderr markers (used by the integration test):
 *   __STUB_EXITED__:<code>  — written from the engine return path;
 *                             may be skipped if the process is
 *                             killed by a non-cooperative signal.
 */
import { runEngine } from '../../dist/engine.js';

const fakeBinPath = process.env.FAKE_BIN_PATH;
if (!fakeBinPath) {
  process.stderr.write('engine-runner-stub: FAKE_BIN_PATH unset\n');
  process.exit(2);
}

try {
  // Engine wires SIGINT/SIGTERM/SIGHUP listeners inside runEngine
  // BEFORE awaiting the init handshake. The test waits for the fake
  // binary's __FAKE_INIT_OK__ marker before delivering a signal —
  // that marker is strictly after engine registered its handler.
  const result = await runEngine({
    binPath: fakeBinPath,
    goArgs: [],
    eslintPluginEntries: [],
    workerConfigs: [],
    configs: [],
    // Inherit so the fake binary's stderr (__FAKE_PID__,
    // __FAKE_INIT_OK__, __FAKE_SIGTERM__) reaches the test parent.
    stderr: process.stderr,
    stdout: process.stdout,
  });
  process.stderr.write(`__STUB_EXITED__:${result.exitCode}\n`);
  process.exit(result.exitCode);
} catch (err) {
  process.stderr.write(`engine-runner-stub crashed: ${err.stack ?? err}\n`);
  process.exit(99);
}
