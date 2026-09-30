#!/usr/bin/env node
// Minimal Go-binary stand-in for engine tests, speaking the IPC frame
// protocol ([4-byte u32 LE length][JSON {kind,id,data}]) over stdio:
//   1. answers optional `transportConfig`, then `init` → `response {ok:true}`,
//   2. sends the first output half as an acknowledged `output` request,
//   3. after its acknowledgement, sends the second half as a notification,
//   4. sends a `shutdown` request,
//   5. exits 0 on both acknowledgements or 2 if the peer rejects either.
// This mirrors the real binary's happy-path frame sequence so runEngine can
// be exercised end-to-end without Go.

// Own fd 1 so end() flushes the frames and closes the actual pipe. Node's
// special stdout can finish without producing EOF on Windows. Do not access
// process.stdout first: its lazy getter duplicates the Windows pipe handle.
const stdout = require('node:fs').createWriteStream(null, {
  fd: 1,
  autoClose: true,
});

let buf = Buffer.alloc(0);
let remainingText = '';
let configured = false;
const mode = process.argv[2];

function send(msg) {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const head = Buffer.alloc(4);
  head.writeUInt32LE(body.length, 0);
  stdout.write(Buffer.concat([head, body]));
}

function onMessage(msg) {
  if (msg.kind === 'transportConfig') {
    const assert = require('node:assert/strict');
    assert.equal(configured, false);
    assert.equal(msg.data, undefined);
    assert.equal(msg.attachments, undefined);
    assert.equal(msg.transport, undefined);
    configured = true;
    send({
      kind: 'response',
      id: msg.id,
      // A small non-default peer layout, independent of production defaults.
      data: {
        version: 1,
        slotCount: 3,
        slotSize: 4096,
        headerSize: 512,
        publicationStride: 32,
      },
    });
    return;
  }
  if (msg.kind === 'init') {
    if (mode === 'eof-before-init' || mode === 'eof-after-init') {
      if (mode === 'eof-after-init') {
        send({ kind: 'response', id: msg.id, data: { ok: true } });
      }
      // Remain alive on stdin after EOF. Tests either let the host terminate
      // this disconnected child or explicitly release a natural exit below.
      stdout.end();
      return;
    }
    if (mode === 'reject-init' || mode === 'reject-init-eof') {
      send({
        kind: 'error',
        id: msg.id,
        data: { message: 'injected init failure' },
      });
      if (mode === 'reject-init-eof') stdout.end();
      return;
    }
    if (mode === 'require-mapping') {
      const assert = require('node:assert/strict');
      assert.equal(configured, true);
      assert.equal(msg.transport?.mapping?.version, 1);
      if (process.platform !== 'win32') {
        assert.equal(msg.transport.mapping.fd, 3);
        // macOS may round the backing object up to a host page.
        assert.ok(require('node:fs').fstatSync(3).size >= 512 + 3 * 4096);
      } else {
        assert.equal(typeof msg.transport.mapping.handle, 'string');
        assert.equal(typeof msg.transport.mapping.processId, 'number');
      }
    }
    send({ kind: 'response', id: msg.id, data: { ok: true } });
    const text = JSON.stringify(msg.data);
    const split = Math.ceil(text.length / 2);
    remainingText = text.slice(split);
    send({
      kind: 'output',
      id: 999,
      data: { stream: 'stdout', text: text.slice(0, split) },
    });
  } else if (
    (mode === 'eof-before-init' ||
      mode === 'eof-after-init' ||
      mode === 'reject-init-eof') &&
    msg.kind === 'exit-after-eof'
  ) {
    // The parent sends this only after observing its readable EOF, making the
    // EOF-before-exit ordering deterministic without a timing-based sleep.
    process.exit(msg.data.code);
  } else if (msg.id === 999) {
    if (msg.kind !== 'response') process.exit(2);
    send({
      kind: 'output',
      id: 0,
      data: { stream: 'stdout', text: remainingText },
    });
    send({ kind: 'shutdown', id: 1000, data: {} });
  } else if (msg.id === 1000) {
    process.exit(msg.kind === 'response' ? 0 : 2);
  }
}

process.stdin.on('data', (chunk) => {
  buf = Buffer.concat([buf, chunk]);
  while (buf.length >= 4) {
    const len = buf.readUInt32LE(0);
    if (buf.length < 4 + len) break;
    const body = buf.subarray(4, 4 + len).toString('utf8');
    buf = buf.subarray(4 + len);
    onMessage(JSON.parse(body));
  }
});
