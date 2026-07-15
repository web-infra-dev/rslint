#!/usr/bin/env node

// Starts config activation, then exits while the Node host is still preparing
// its plugin worker. runEngine must await and dispose that late worker.

const path = require('node:path');

const configPath = path.resolve(process.argv[2]);
const configDirectory = path.dirname(configPath);
let buffer = Buffer.alloc(0);

function send(message) {
  const body = Buffer.from(JSON.stringify(message), 'utf8');
  const header = Buffer.alloc(4);
  header.writeUInt32LE(body.length, 0);
  process.stdout.write(Buffer.concat([header, body]));
}

function onMessage(message) {
  if (message.kind === 'init') {
    send({ kind: 'response', id: message.id, data: { ok: true } });
    send({
      kind: 'loadConfigs',
      id: 300,
      data: {
        transactionId: 'cli-exit-during-prepare',
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
  if (message.kind === 'response' && message.id === 300) {
    send({
      kind: 'activateConfigs',
      id: 301,
      data: {
        transactionId: 'cli-exit-during-prepare',
        effectiveConfigIds: ['root'],
      },
    });
    setTimeout(() => process.exit(0), 100);
  }
}

process.stdin.on('data', (chunk) => {
  buffer = Buffer.concat([buffer, chunk]);
  while (buffer.length >= 4) {
    const length = buffer.readUInt32LE(0);
    if (buffer.length < 4 + length) break;
    const body = buffer.subarray(4, 4 + length).toString('utf8');
    buffer = buffer.subarray(4 + length);
    onMessage(JSON.parse(body));
  }
});
