import { describe, test, expect } from '@rstest/core';
import { PassThrough } from 'node:stream';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { runEngine } from '../src/engine.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FAKE_BIN = path.resolve(__dirname, './fixtures/fake-ipc-binary.cjs');

/**
 * Runs the engine against the fake IPC binary, which echoes the `init`
 * payload it received back through an `output` frame — letting the tests
 * assert on what actually crossed the wire, not on engine internals.
 */
async function runWithSink(sink: PassThrough): Promise<{
  exitCode: number;
  payload: { runtime?: { stdoutIsTTY?: boolean } };
}> {
  let captured = '';
  sink.on('data', (d: Buffer) => {
    captured += d.toString();
  });
  const exitCode = await runEngine({
    binPath: process.execPath,
    goArgs: [FAKE_BIN],
    configs: [],
    stdout: sink,
    stderr: new PassThrough(),
  });
  return { exitCode, payload: JSON.parse(captured) };
}

describe('runEngine init payload TTY fact', () => {
  test('sends runtime.stdoutIsTTY=true when the output sink is a TTY', async () => {
    const sink = Object.assign(new PassThrough(), { isTTY: true });
    const { exitCode, payload } = await runWithSink(sink);
    expect(exitCode).toBe(0);
    expect(payload.runtime?.stdoutIsTTY).toBe(true);
  });

  test('sends runtime.stdoutIsTTY=false for a non-TTY sink', async () => {
    const { exitCode, payload } = await runWithSink(new PassThrough());
    expect(exitCode).toBe(0);
    expect(payload.runtime?.stdoutIsTTY).toBe(false);
  });
});
