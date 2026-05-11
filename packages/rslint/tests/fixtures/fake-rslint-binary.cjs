#!/usr/bin/env node
'use strict';

// Test fixture: stands in for the real `rslint` Go binary in unit tests
// that need to inspect what engine.ts puts on the wire. The fixture
// implements just enough of the runCLI handshake to:
//
//   1. read length-prefixed frames from stdin
//   2. when an `init` frame arrives, echo its parsed `data` field to
//      stderr as a JSON line `__FAKE_INIT__<json>`
//   3. reply `{ok:true}` so engine.ts proceeds
//   4. accept and ack `shutdown`, then exit 0
//
// Stderr capture (rather than stdout) keeps the IPC frame stream on
// stdout uncontaminated.

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
    // Echo the received init payload back to engine.ts as an `output`
    // notification — engine.ts forwards `output` frames to opts.stdout,
    // which the test then captures. We can't use stderr because
    // engine.ts spawns with stdio[2]='inherit', so the fake binary's
    // stderr goes to the test runner's stderr, not into the test's
    // captured stream.
    const payload = JSON.stringify(msg.data);
    writeFrame({
      kind: 'output',
      id: 0,
      data: { stream: 'stdout', text: '__FAKE_INIT__' + payload + '\n' },
    });
    writeFrame({ kind: 'response', id: msg.id, data: { ok: true } });
    // For test-isolation: exit immediately after replying. engine.ts
    // waits on child exit (it normally relies on the Go binary's
    // shutdownPeer reverse-request to clean up). Without an explicit
    // exit here the test would hang on engine.ts's `child.on('exit')`
    // await.
    //
    // Give the response frame one tick to drain through stdout before
    // exiting so engine.ts actually receives the {ok:true} ack.
    setImmediate(() => process.exit(0));
    return;
  }
  if (msg.kind === 'shutdown') {
    writeFrame({ kind: 'response', id: msg.id, data: { ok: true } });
    setImmediate(() => process.exit(0));
    return;
  }
  // Any other inbound: nack as an error response so engine.ts surfaces it.
  writeFrame({
    kind: 'error',
    id: msg.id,
    data: { message: 'fake-rslint-binary: unsupported kind=' + msg.kind },
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
      process.stderr.write('__FAKE_ERR__' + String(err) + '\n');
    }
  }
});

stdin.on('end', () => {
  process.exit(0);
});
