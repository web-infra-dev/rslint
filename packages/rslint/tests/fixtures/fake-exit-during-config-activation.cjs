#!/usr/bin/env node

// Starts config activation, then exits while the Node host is still preparing
// its plugin worker. runEngine must await and dispose that late worker.

const path = require('node:path');
const fs = require('node:fs');

const configPath = path.resolve(process.argv[2]);
const buildMarker = process.argv[3] ? path.resolve(process.argv[3]) : undefined;
const configDirectory = path.dirname(configPath);
let buffer = Buffer.alloc(0);
// The parent test succeeds on process exit. This only releases a fixture that
// has failed to observe its event-driven build marker for 30 minutes.
const CLEANUP_WATCHDOG_MS = 30 * 60_000;
const cleanupWatchdog = setTimeout(() => {
  process.stderr.write(
    'fake config-activation child timed out waiting for build marker\n',
  );
  process.exit(2);
}, CLEANUP_WATCHDOG_MS);
cleanupWatchdog.unref();

let parentInputClosed = false;
function exitOnParentDisconnect() {
  if (parentInputClosed) return;
  parentInputClosed = true;
  process.exit(2);
}

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
        protocolVersion: 1,
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
        protocolVersion: 1,
        transactionId: 'cli-exit-during-prepare',
        effectiveConfigIds: ['root'],
      },
    });
    const exitAfterBuildStarts = () => {
      if (buildMarker && fs.existsSync(buildMarker)) {
        process.exit(0);
      }
      setTimeout(exitAfterBuildStarts, 10);
    };
    exitAfterBuildStarts();
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
process.stdin.once('end', exitOnParentDisconnect);
process.stdin.once('close', exitOnParentDisconnect);
