#!/usr/bin/env node
'use strict';

/**
 * Test fixture: the second-half of `fake-rslint-binary.cjs`. Same IPC
 * framing, but does NOT exit after replying to `init`. Instead it parks
 * on its event loop and waits for one of:
 *
 *   - a `shutdown` request from the engine    → ack and exit 0
 *   - SIGTERM (engine.ts's safeKillGo path)   → exit 143-ish
 *   - SIGINT (rare; usually engine handles it first)
 *   - stdin EOF (Node parent died, pipe closed) → exit 0
 *
 * Lets the integration tests in `cli-signal-integration.test.ts`
 * exercise the path where engine.ts is mid-`await childExit` when the
 * user hits Ctrl-C — the engine handler must forward the signal to the
 * child, this fixture cooperates, and the whole tree exits.
 *
 * Logs its PID to stderr as `__FAKE_PID__:<pid>` so the test can
 * verify (via `process.kill(pid, 0)`) that the child is actually gone
 * after the parent returns.
 */

const HEADER_BYTES = 4;
const stdin = process.stdin;
const stdout = process.stdout;

let buffered = Buffer.alloc(0);

function writeFrame(msg) {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const header = Buffer.alloc(HEADER_BYTES);
  header.writeUInt32LE(body.length, 0);
  stdout.write(Buffer.concat([header, body]));
}

function handleFrame(msg) {
  if (msg.kind === 'init') {
    // Ack init and announce readiness via stderr. Unlike
    // fake-rslint-binary.cjs we do NOT setImmediate-exit — the whole
    // point is to stay alive so the signal-forwarding path can be
    // exercised.
    writeFrame({ kind: 'response', id: msg.id, data: { ok: true } });
    process.stderr.write(`__FAKE_INIT_OK__\n`);
    return;
  }
  if (msg.kind === 'shutdown') {
    writeFrame({ kind: 'response', id: msg.id, data: { ok: true } });
    setImmediate(() => process.exit(0));
    return;
  }
  // Unknown kinds: nack so engine surfaces it.
  writeFrame({
    kind: 'error',
    id: msg.id,
    data: { message: 'long-running-fake-binary: unsupported kind=' + msg.kind },
  });
}

stdin.on('data', (chunk) => {
  buffered = Buffer.concat([buffered, chunk]);
  while (buffered.length >= HEADER_BYTES) {
    const len = buffered.readUInt32LE(0);
    if (buffered.length < HEADER_BYTES + len) break;
    const bodyBuf = buffered.subarray(HEADER_BYTES, HEADER_BYTES + len);
    buffered = buffered.subarray(HEADER_BYTES + len);
    try {
      const msg = JSON.parse(bodyBuf.toString('utf8'));
      handleFrame(msg);
    } catch (err) {
      process.stderr.write('__FAKE_ERR__:' + String(err) + '\n');
    }
  }
});

// Parent died → pipe closed → graceful exit. This is the path a
// SIGKILL'd Node parent would leave us with.
stdin.on('end', () => {
  process.stderr.write('__FAKE_EOF__\n');
  process.exit(0);
});

process.on('SIGTERM', () => {
  process.stderr.write('__FAKE_SIGTERM__\n');
  process.exit(143);
});

process.on('SIGINT', () => {
  process.stderr.write('__FAKE_SIGINT__\n');
  process.exit(130);
});

// Announce PID so the parent test can poll `process.kill(pid, 0)`
// after teardown to confirm the OS reaped this child.
process.stderr.write(`__FAKE_PID__:${process.pid}\n`);

// Keep the event loop alive in case stdin sees no data and signals
// never fire. Without this, an empty stdin (no init at all) would let
// Node exit on its own.
const keepAlive = setInterval(() => {
  /* tick */
}, 60_000);
keepAlive.unref();
