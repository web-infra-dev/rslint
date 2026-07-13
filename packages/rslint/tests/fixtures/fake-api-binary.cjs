#!/usr/bin/env node
// Minimal `--api` stand-in for NodeRslintService tests, speaking the IPC frame
// protocol ([4-byte u32 LE length][JSON {id,kind,data}]) over stdio:
//   - handshake → response {version, ok}
//   - crash     → process.exit(42)            (simulate an unexpected crash)
//   - exit      → response {} then exit 0      (normal close)
//   - reverse   → pluginLint request, then echo the Node response/error
//   - anything else (e.g. lint) → no reply     (stays in-flight, so the test
//     can kill/terminate while a request is pending)
// Lets the reject-all-pending logic be exercised without the real Go binary.

let buf = Buffer.alloc(0);
const reverseRequests = new Set();

function send(msg) {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const head = Buffer.alloc(4);
  head.writeUInt32LE(body.length, 0);
  process.stdout.write(Buffer.concat([head, body]));
}

function onMessage(msg) {
  if (msg.kind === 'handshake') {
    send({
      kind: 'response',
      id: msg.id,
      data: {
        version: '2.0.0',
        ok: true,
        capabilities: ['reversePluginLint'],
      },
    });
  } else if (msg.kind === 'crash') {
    process.exit(42);
  } else if (msg.kind === 'reverse') {
    // Deliberately reuse the outer request ID. Request IDs are independent in
    // each direction, so Node must route by frame kind rather than treating
    // this pluginLint frame as the response to `reverse`.
    reverseRequests.add(msg.id);
    send({
      kind: 'pluginLint',
      id: msg.id,
      data: { files: [{ path: 'probe.ts' }], rules: {} },
    });
  } else if (
    reverseRequests.has(msg.id) &&
    (msg.kind === 'response' || msg.kind === 'error')
  ) {
    reverseRequests.delete(msg.id);
    send({
      kind: 'response',
      id: msg.id,
      data: { reverseKind: msg.kind, reverseData: msg.data },
    });
  } else if (msg.kind === 'exit') {
    // Silent mode: exit WITHOUT sending the ack, simulating the peer exiting
    // before its 'exit' response is read — the close() race that must settle
    // the pending without an unhandledRejection.
    if (process.env.RSLINT_FAKE_EXIT_SILENT === '1') {
      process.exit(0);
    }
    send({ kind: 'response', id: msg.id, data: {} });
    process.exit(0);
  }
  // else: leave in-flight (no reply)
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
