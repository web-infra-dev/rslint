#!/usr/bin/env node

// Exercise the CLI's two config responses over the real framed IPC transport.
// Output requests are acknowledged gates, so tests can hold worker preparation
// while this peer proves which independent requests have already completed.
const path = require('node:path');

const configPath = path.resolve(process.argv[2]);
const mode = process.argv[3];
const transactionId = 'cli-worker-warmup';
const selection = {
  protocolVersion: 3,
  transactionId,
  effectiveConfigIds: ['root'],
};
const load = {
  protocolVersion: 3,
  transactionId,
  loadMode: 'cached',
  candidates: [
    { id: 'root', configPath, configDirectory: path.dirname(configPath) },
  ],
};
const pending = new Map();
let nextId = 200;
let buffer = Buffer.alloc(0);

function send(message) {
  const body = Buffer.from(JSON.stringify(message), 'utf8');
  const header = Buffer.alloc(4);
  header.writeUInt32LE(body.length, 0);
  process.stdout.write(Buffer.concat([header, body]));
}

function request(kind, data) {
  const id = nextId++;
  return new Promise((resolve) => {
    pending.set(id, resolve);
    send({ kind, id, data });
  });
}

function report(event, data = {}) {
  return request('output', {
    stream: 'stdout',
    text: `${JSON.stringify({ event, ...data })}\n`,
  });
}

async function run() {
  const loaded = await request('loadConfigs', load);
  if (loaded.kind !== 'response') throw new Error('config load failed');

  if (mode === 'blocked-prepare') {
    let settled = false;
    const preparation = request('prepareConfigs', selection).then((result) => {
      settled = true;
      return result;
    });
    await report('waiters-started');
    await new Promise((resolve) => setImmediate(resolve));
    await report('waiters-pending', { settled });
    await report('completed', { preparation: await preparation });
  } else {
    const preparation = await request('prepareConfigs', selection);
    await report('prepared', { preparation });

    if (mode === 'exit') {
      process.exit(0);
    }
    if (mode === 'concurrent') {
      const duplicates = await Promise.all([
        request('prepareConfigs', selection),
        request('prepareConfigs', selection),
      ]);
      const conflicts = await Promise.all([
        request('prepareConfigs', {
          ...selection,
          effectiveConfigIds: ['other'],
        }),
        request('activateConfigs', {
          ...selection,
          transactionId: 'different-transaction',
        }),
      ]);
      const loadAfterActivation = await request('loadConfigs', load);
      await report('repeated', { duplicates, conflicts, loadAfterActivation });

      const settled = [false, false, false];
      const waiters = [
        request('activateConfigs', selection),
        request('pluginLint', { request: 'first' }),
        request('pluginLint', { request: 'second' }),
      ].map((promise, index) =>
        promise.then((result) => {
          settled[index] = true;
          return result;
        }),
      );
      await report('waiters-started');
      await new Promise((resolve) => setImmediate(resolve));
      await report('waiters-pending', { settled });
      await report('completed', { results: await Promise.all(waiters) });
    } else {
      const activation = await request('activateConfigs', selection);
      const plugin =
        mode === 'failure-without-tasks'
          ? undefined
          : await request('pluginLint', {});
      await report('completed', { activation, plugin });
    }
  }
  await request('shutdown', {});
  process.exit(0);
}

function onMessage(message) {
  if (message.kind === 'init') {
    send({ kind: 'response', id: message.id, data: { ok: true } });
    void run().catch((error) => {
      process.stderr.write(`${error.stack}\n`);
      process.exit(2);
    });
    return;
  }
  const resolve = pending.get(message.id);
  if (resolve) {
    pending.delete(message.id);
    resolve(message);
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
process.stdin.once('end', () => process.exit(2));
