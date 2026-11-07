// Portable test transport: blocking rule callers communicate with this broker
// through a bounded mailbox; the broker owns asynchronous child stdio.
import { spawn } from 'node:child_process';
import { parentPort, workerData } from 'node:worker_threads';
import { createInterface } from 'node:readline';
const control = new Int32Array(workerData.mailbox, 0, 2);
const body = new Uint8Array(workerData.mailbox, 8);
if (workerData.readCache) {
  const { SharedTypeCacheReader } = await import(workerData.readerModule);
  const reader = new SharedTypeCacheReader(workerData.readCache);
  const packet = reader.get(workerData.key);
  const expected = workerData.expected;
  const equal =
    packet?.length === expected.length &&
    packet.every((value, index) => value === expected[index]);
  Atomics.store(control, 1, equal ? 1 : 0);
  reader.close();
  Atomics.store(control, 0, 1);
  Atomics.notify(control, 0);
} else {
  const child = spawn(workerData.binary, [], {
    cwd: workerData.cwd,
    stdio: ['pipe', 'pipe', 'pipe'],
  });
  let pending = false,
    stderr = '';
  child.stderr.on('data', (chunk) => {
    stderr += chunk;
  });
  function reply(value) {
    const bytes = Buffer.from(JSON.stringify(value));
    if (bytes.length > body.length)
      return reply({ error: 'type service test mailbox exhausted' });
    body.set(bytes);
    Atomics.store(control, 1, bytes.length);
    Atomics.store(control, 0, 1);
    Atomics.notify(control, 0);
    pending = false;
  }
  createInterface({ input: child.stdout }).on('line', (line) =>
    reply(JSON.parse(line)),
  );
  child.on('error', (error) => {
    if (pending) reply({ error: error.message });
  });
  child.on('close', (code) => {
    if (pending)
      reply({ error: `type service host exited ${code}: ${stderr}` });
    parentPort.postMessage({ stopped: true });
  });
  parentPort.on('message', (request) => {
    if (request.method === 'stop') {
      child.stdin.end();
      return;
    }
    pending = true;
    child.stdin.write(`${JSON.stringify(request)}\n`);
  });
  process.on('exit', () => child.kill());
}
