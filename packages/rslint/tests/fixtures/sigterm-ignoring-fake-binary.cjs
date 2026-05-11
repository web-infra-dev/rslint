#!/usr/bin/env node
'use strict';

/**
 * Hybrid test fixture: behaves like long-running-fake-binary.cjs
 * (ack's `init` and parks) but DELIBERATELY IGNORES SIGTERM and
 * SIGINT. Used to verify engine.ts's `safeKillGo` SIGKILL backstop
 * actually fires end-to-end (the old `safe-kill.test.ts` only tested
 * a local copy of the pattern, not the engine's real implementation).
 *
 * Stderr markers (parent test waits on these):
 *   __FAKE_PID__:<pid>           — emitted once at startup
 *   __FAKE_INIT_OK__             — emitted after engine's init is ack'd
 *   __FAKE_RECEIVED_SIGTERM__    — confirms SIGTERM hit us (and was ignored)
 *
 * The whole point: engine.ts must send SIGTERM (we see the marker),
 * wait the documented grace window (~5s), then escalate to SIGKILL
 * (un-catchable — process just dies). The parent test asserts elapsed
 * time ≥ grace AND that PID was reaped.
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
    writeFrame({ kind: 'response', id: msg.id, data: { ok: true } });
    process.stderr.write(`__FAKE_INIT_OK__\n`);
    return;
  }
  // Even shutdown is ignored — we want to force the SIGKILL path.
  if (msg.kind === 'shutdown') {
    process.stderr.write(`__FAKE_IGNORING_SHUTDOWN__\n`);
    return;
  }
  writeFrame({
    kind: 'error',
    id: msg.id,
    data: { message: 'sigterm-ignoring-fake: ignoring kind=' + msg.kind },
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

// Critically, we DO NOT exit on stdin end. If the parent dies via
// SIGKILL, OS cleans us up via SIGPIPE during the next write — but
// that's also the path we want to verify, not bypass.

process.on('SIGTERM', () => {
  process.stderr.write('__FAKE_RECEIVED_SIGTERM__\n');
  // Deliberate no-op. SIGKILL (sent ~5s later by safeKillGo) cannot
  // be caught — that's the OS-level kill that ends this process.
});

process.on('SIGINT', () => {
  process.stderr.write('__FAKE_RECEIVED_SIGINT__\n');
  // Deliberate no-op.
});

process.stderr.write(`__FAKE_PID__:${process.pid}\n`);

// Keep alive forever; only SIGKILL can end us.
const keepAlive = setInterval(() => {
  /* tick */
}, 60_000);
keepAlive.unref();
