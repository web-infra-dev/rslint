// Portable test transport: blocking rule callers communicate with this broker
// through a bounded mailbox; the broker owns asynchronous child stdio.
import { spawn } from 'node:child_process';
import { parentPort, workerData } from 'node:worker_threads';
import { createInterface } from 'node:readline';
const control = new Int32Array(workerData.mailbox, 0, 2);
const body = new Uint8Array(workerData.mailbox, 8);

const child = spawn(workerData.binary, [], {
  cwd: workerData.cwd,
  stdio: ['pipe', 'pipe', 'pipe'],
});
let pending = false,
  childClosed = false,
  exitReason = null,
  stderr = '';
child.stdin.on('error', () => {});
child.stdout.on('error', () => {});
child.stderr.on('error', () => {});
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
  exitReason ??= `type service host failed: ${error.message}`;
  if (pending) reply({ error: error.message });
});
child.on('close', (code) => {
  childClosed = true;
  exitReason ??= `type service host exited ${code}: ${stderr}`;
  if (pending) reply({ error: exitReason });
  parentPort.postMessage({ stopped: true });
});
parentPort.on('message', (request) => {
  if (request.method === 'stop') {
    if (childClosed) {
      parentPort.postMessage({ stopped: true });
      return;
    }
    child.stdin.end();
    const forceKill = setTimeout(() => {
      if (!childClosed) child.kill('SIGKILL');
    }, 5_000);
    forceKill.unref();
    child.once('close', () => clearTimeout(forceKill));
    return;
  }
  if (childClosed || exitReason !== null) {
    reply({
      error: exitReason ?? `type service host already exited: ${stderr}`,
    });
    return;
  }
  pending = true;
  child.stdin.write(`${JSON.stringify(request)}\n`);
});
process.on('exit', () => {
  if (!childClosed) child.kill();
});
