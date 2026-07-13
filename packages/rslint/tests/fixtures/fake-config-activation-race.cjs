#!/usr/bin/env node

// Drives the reverse config protocol far enough to make the engine stage a
// mocked plugin host. The test-side factory rewrites the config while it is
// being created, so activation must fail and a subsequent pluginLint request
// must observe no published host.

const path = require('node:path');

const configPath = path.resolve(process.argv[2]);
const configDirectory = path.dirname(configPath);
let buf = Buffer.alloc(0);
let activationError = '';

function send(msg) {
  const body = Buffer.from(JSON.stringify(msg), 'utf8');
  const head = Buffer.alloc(4);
  head.writeUInt32LE(body.length, 0);
  process.stdout.write(Buffer.concat([head, body]));
}

function onMessage(msg) {
  if (msg.kind === 'init') {
    send({ kind: 'response', id: msg.id, data: { ok: true } });
    send({
      kind: 'loadConfigs',
      id: 200,
      data: {
        protocolVersion: 1,
        transactionId: 'cli-prepare-race',
        loadMode: 'cached',
        candidates: [
          {
            id: 'root',
            configPath,
            configDirectory,
          },
        ],
      },
    });
    return;
  }
  if (msg.kind === 'response' && msg.id === 200) {
    send({
      kind: 'activateConfigs',
      id: 201,
      data: {
        protocolVersion: 1,
        transactionId: 'cli-prepare-race',
        effectiveConfigIds: ['root'],
      },
    });
    return;
  }
  if (msg.kind === 'error' && msg.id === 200) {
    process.stderr.write(
      `loadConfigs unexpectedly failed: ${msg.data?.message}\n`,
    );
    process.exit(2);
  }
  if (msg.kind === 'response' && msg.id === 201) {
    process.stderr.write('activateConfigs unexpectedly succeeded\n');
    process.exit(2);
  }
  if (msg.kind === 'error' && msg.id === 201) {
    activationError = msg.data?.message ?? '';
    send({ kind: 'pluginLint', id: 202, data: {} });
    return;
  }
  if (msg.kind === 'response' && msg.id === 202) {
    send({
      kind: 'output',
      id: 0,
      data: { stream: 'stdout', text: activationError },
    });
    send({ kind: 'shutdown', id: 203, data: {} });
    return;
  }
  if (msg.kind === 'response' && msg.id === 203) {
    process.exit(0);
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
