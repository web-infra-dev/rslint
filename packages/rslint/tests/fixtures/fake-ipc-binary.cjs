#!/usr/bin/env node
// Minimal Go-binary stand-in for engine tests, speaking the IPC frame
// protocol ([4-byte u32 LE length][JSON {kind,id,data}]) over stdio:
//   1. expects an `init` request → replies `response {ok:true}`,
//   2. echoes the received init payload back as two `output` notifications,
//   3. sends a `shutdown` request,
//   4. exits 0 on an acknowledgement or 2 if the peer rejects shutdown.
// This mirrors the real binary's happy-path frame sequence so runEngine can
// be exercised end-to-end without Go.

let buf = Buffer.alloc(0);

function send(msg) {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const head = Buffer.alloc(4);
  head.writeUInt32LE(body.length, 0);
  process.stdout.write(Buffer.concat([head, body]));
}

function onMessage(msg) {
  if (msg.kind === 'init') {
    send({ kind: 'response', id: msg.id, data: { ok: true } });
    const text = JSON.stringify(msg.data);
    const split = Math.ceil(text.length / 2);
    send({
      kind: 'output',
      id: 0,
      data: { stream: 'stdout', text: text.slice(0, split) },
    });
    send({
      kind: 'output',
      id: 0,
      data: { stream: 'stdout', text: text.slice(split) },
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
