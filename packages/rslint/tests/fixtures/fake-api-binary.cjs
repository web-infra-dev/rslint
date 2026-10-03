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

// Own fd 1 so end() flushes the frames and closes the actual pipe. Node's
// special stdout can finish without producing EOF on Windows. Do not access
// process.stdout first: its lazy getter duplicates the Windows pipe handle.
const stdout = require('node:fs').createWriteStream(null, {
  fd: 1,
  autoClose: true,
});

let buf = Buffer.alloc(0);
const reverseRequests = new Set();

function send(msg) {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const head = Buffer.alloc(4);
  head.writeUInt32LE(body.length, 0);
  stdout.write(Buffer.concat([head, body]));
}

function onMessage(msg) {
  if (msg.kind === 'transportConfig') {
    // An older API peer does not implement optional memory negotiation.
    send({ kind: 'error', id: msg.id, data: { message: 'unknown request' } });
  } else if (msg.kind === 'handshake') {
    send({
      kind: 'response',
      id: msg.id,
      data: {
        version: '3.1.0',
        ok: true,
        capabilities: ['reversePluginLint'],
      },
    });
  } else if (msg.kind === 'crash') {
    process.exit(42);
  } else if (msg.kind === 'close-output') {
    // Keep stdin open and the process alive. Transport EOF alone must reject
    // requests instead of waiting indefinitely for the child's exit event.
    stdout.end();
  } else if (msg.kind === 'reverse-close-output') {
    // This command is a notification: there is no outstanding host request.
    // The host handler deliberately never settles; EOF must still retire the
    // connection and stop the resident child from keeping its host alive.
    send({
      kind: 'pluginLint',
      id: 1,
      data: { files: [{ path: 'probe.ts' }], rules: {} },
    });
    stdout.end();
  } else if (msg.kind === 'reverse' || msg.kind === 'reverse-bytes') {
    // Deliberately reuse the outer request ID. Request IDs are independent in
    // each direction, so Node must route by frame kind rather than treating
    // this pluginLint frame as the response to `reverse`.
    reverseRequests.add(msg.id);
    send({
      kind: msg.kind === 'reverse-bytes' ? 'arbitraryBinary' : 'pluginLint',
      id: msg.id,
      data: { files: [{ path: 'probe.ts' }], rules: {} },
      ...(msg.kind === 'reverse-bytes'
        ? { attachments: [{ bytes: 'AP+A' }, { text: '' }] }
        : {}),
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
    if (msg.data?.reject) {
      send({ kind: 'error', id: msg.id, data: { message: 'exit rejected' } });
      stdout.end();
      return;
    }
    // Silent mode: exit WITHOUT sending the ack, simulating the peer exiting
    // before its 'exit' response is read — the close() race that must settle
    // the pending without an unhandledRejection.
    if (process.env.RSLINT_FAKE_EXIT_SILENT === '1') {
      process.exit(0);
    }
    send({ kind: 'response', id: msg.id, data: {} });
    // fs.WriteStream writes asynchronously, so flush the final ack and close
    // its descriptor before exiting.
    stdout.once('close', () => process.exit(0));
    stdout.end();
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
