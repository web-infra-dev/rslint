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
  stopping = false,
  exitReason = null,
  stopError = null,
  stderr = '';
const MAX_STDERR_CHARS = 64 * 1024;
child.stdin.on('error', () => {});
child.stdout.on('error', () => {});
child.stderr.on('error', () => {});
child.stderr.on('data', (chunk) => {
  if (stderr.length < MAX_STDERR_CHARS) {
    stderr += String(chunk).slice(0, MAX_STDERR_CHARS - stderr.length);
  }
});
function reply(value) {
  const bytes = Buffer.from(JSON.stringify(value));
  if (bytes.length > body.length) {
    if (value?.mailboxOverflow) {
      throw new Error('type service test mailbox too small for error header');
    }
    return reply({
      error: 'type service test mailbox exhausted',
      childFatal: value?.childFatal,
      workerFatal: value?.workerFatal,
      mailboxOverflow: true,
    });
  }
  body.set(bytes);
  Atomics.store(control, 1, bytes.length);
  Atomics.store(control, 0, 2);
  Atomics.notify(control, 0);
  pending = false;
}
createInterface({ input: child.stdout }).on('line', (line) => {
  let message;
  try {
    message = JSON.parse(line);
  } catch (error) {
    exitReason ??= `invalid type service stdout ${JSON.stringify(line)}: ${error.message}`;
    stopError ??= exitReason;
    if (pending) reply({ error: exitReason, childFatal: true });
    child.kill();
    return;
  }
  reply(message);
});
child.on('error', (error) => {
  exitReason ??= `type service host failed: ${error.message}`;
  stopError ??= exitReason;
  if (pending) reply({ error: error.message, childFatal: true });
});
child.on('close', (code, signal) => {
  childClosed = true;
  const detail = signal !== null ? `signal ${signal}` : `${code}`;
  exitReason ??= `type service host exited ${detail}: ${stderr}`;
  if (code !== 0 || signal !== null) {
    stopError ??= exitReason;
  }
  if (pending) reply({ error: exitReason, childFatal: true });
  if (stopping) {
    parentPort.postMessage({ stopped: true, error: stopError ?? undefined });
  }
});
parentPort.on('message', (request) => {
  if (request.method === 'stop') {
    stopping = true;
    if (childClosed) {
      parentPort.postMessage({ stopped: true, error: stopError ?? undefined });
      return;
    }
    child.stdin.end();
    const forceKill = setTimeout(() => {
      if (!childClosed) {
        stopError ??= `type service host did not exit within 5000ms after stdin EOF: ${stderr}`;
        child.kill('SIGKILL');
      }
    }, 5_000);
    forceKill.unref();
    child.once('close', () => clearTimeout(forceKill));
    return;
  }
  if (childClosed || exitReason !== null) {
    reply({
      error: exitReason ?? `type service host already exited: ${stderr}`,
      childFatal: true,
    });
    return;
  }
  let serialized;
  try {
    serialized = JSON.stringify(request);
  } catch (error) {
    reply({ error: error.message });
    return;
  }
  pending = true;
  child.stdin.write(`${serialized}\n`);
});
process.on('exit', (code) => {
  if (!childClosed) child.kill();
  if (Atomics.load(control, 0) === 1) {
    reply({
      error: `type service broker worker exited unexpectedly with code ${code}`,
      workerFatal: true,
    });
  }
});
