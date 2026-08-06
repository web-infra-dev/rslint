import { afterEach, describe, expect, test } from '@rstest/core';
import { spawn, type ChildProcessWithoutNullStreams } from 'node:child_process';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { PassThrough } from 'node:stream';
import { fileURLToPath, pathToFileURL } from 'node:url';

import { runEditorRuntime } from '../src/editor-runtime/editor-runtime.js';

interface JsonRpcMessage {
  jsonrpc: '2.0';
  id?: number | string;
  method?: string;
  params?: unknown;
  result?: unknown;
  error?: unknown;
}

const children = new Set<ChildProcessWithoutNullStreams>();
const tempDirectories = new Set<string>();

afterEach(async () => {
  for (const child of children) child.kill('SIGKILL');
  children.clear();
  await Promise.all(
    [...tempDirectories].map(async (directory) =>
      fs.rm(directory, { recursive: true, force: true }),
    ),
  );
  tempDirectories.clear();
});

function encodeLsp(message: JsonRpcMessage): Buffer {
  const body = Buffer.from(JSON.stringify(message), 'utf8');
  return Buffer.concat([
    Buffer.from(`Content-Length: ${body.length}\r\n\r\n`, 'ascii'),
    body,
  ]);
}

function waitForExit(
  child: ChildProcessWithoutNullStreams,
  stderr: () => string,
): Promise<number> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error(`editor runtime did not exit; stderr:\n${stderr()}`));
    }, 20_000);
    child.once('error', (error) => {
      clearTimeout(timer);
      reject(error);
    });
    child.once('exit', (code, signal) => {
      clearTimeout(timer);
      if (signal) {
        reject(
          new Error(
            `editor runtime exited via ${signal}; stderr:\n${stderr()}`,
          ),
        );
      } else {
        resolve(code ?? 1);
      }
    });
  });
}

describe('core editor runtime', () => {
  test('fails a missing native binary promptly without leaking the private listener', async () => {
    const cwd = await fs.mkdtemp(
      path.join(os.tmpdir(), 'rslint-editor-missing-'),
    );
    tempDirectories.add(cwd);
    const started = Date.now();
    await expect(
      runEditorRuntime({
        cwd,
        binaryPath: path.join(cwd, 'definitely-missing-rslint-binary'),
        stdin: new PassThrough(),
        stdout: new PassThrough(),
        stderr: new PassThrough(),
      }),
    ).rejects.toThrow();
    expect(Date.now() - started).toBeLessThan(3_000);
  });

  test('tears down the child when an authenticated peer violates private framing', async () => {
    if (process.platform === 'win32') return;
    const cwd = await fs.mkdtemp(
      path.join(os.tmpdir(), 'rslint-editor-invalid-'),
    );
    tempDirectories.add(cwd);
    const fakeBinary = path.join(cwd, 'fake-rslint');
    await fs.writeFile(
      fakeBinary,
      `#!/usr/bin/env node\n` +
        `const net = require('node:net');\n` +
        `process.on('SIGTERM', () => {});\n` +
        `const [host, rawPort] = process.env.RSLINT_RUNTIME_IPC_ADDRESS.split(':');\n` +
        `const socket = net.connect({ host, port: Number(rawPort) }, () => {\n` +
        `  const body = Buffer.from('null');\n` +
        `  const header = Buffer.alloc(4);\n` +
        `  header.writeUInt32LE(body.length, 0);\n` +
        `  socket.write(process.env.RSLINT_RUNTIME_IPC_TOKEN + '\\n');\n` +
        `  socket.write(Buffer.concat([header, body]));\n` +
        `});\n` +
        `socket.on('error', () => {});\n` +
        `setInterval(() => {}, 1000);\n`,
      { mode: 0o755 },
    );
    const stderr = new PassThrough();
    let stderrText = '';
    stderr.setEncoding('utf8');
    stderr.on('data', (chunk: string) => {
      stderrText += chunk;
    });

    const started = Date.now();
    await runEditorRuntime({
      cwd,
      binaryPath: fakeBinary,
      stdin: new PassThrough(),
      stdout: new PassThrough(),
      stderr,
    });
    expect(Date.now() - started).toBeLessThan(5_000);
    expect(stderrText).toMatch(/private editor-runtime transport failed/);
    expect(stderrText).toMatch(/IPC message must be an object/);
  });

  test('reaps a native child that ignores the sidecar termination signal', async () => {
    if (process.platform === 'win32') return;
    const cwd = await fs.mkdtemp(
      path.join(os.tmpdir(), 'rslint-editor-signal-'),
    );
    tempDirectories.add(cwd);
    const connectedMarker = path.join(cwd, 'connected');
    const fakeBinary = path.join(cwd, 'fake-rslint');
    await fs.writeFile(
      fakeBinary,
      `#!/usr/bin/env node\n` +
        `const fs = require('node:fs');\n` +
        `const net = require('node:net');\n` +
        `process.on('SIGTERM', () => {});\n` +
        `const [host, rawPort] = process.env.RSLINT_RUNTIME_IPC_ADDRESS.split(':');\n` +
        `const socket = net.connect({ host, port: Number(rawPort) }, () => {\n` +
        `  socket.write(process.env.RSLINT_RUNTIME_IPC_TOKEN + '\\n');\n` +
        `  fs.writeFileSync(${JSON.stringify(connectedMarker)}, 'connected');\n` +
        `});\n` +
        `socket.on('error', () => {});\n` +
        `setInterval(() => {}, 1000);\n`,
      { mode: 0o755 },
    );
    const signalListenersBefore = process.listenerCount('SIGTERM');
    const runtime = runEditorRuntime({
      cwd,
      binaryPath: fakeBinary,
      stdin: new PassThrough(),
      stdout: new PassThrough(),
      stderr: new PassThrough(),
    });
    const deadline = Date.now() + 5_000;
    for (;;) {
      try {
        await fs.access(connectedMarker);
        break;
      } catch (error) {
        if (
          (error as NodeJS.ErrnoException).code !== 'ENOENT' ||
          Date.now() >= deadline
        ) {
          throw error;
        }
        await new Promise((resolve) => setTimeout(resolve, 20));
      }
    }

    process.emit('SIGTERM', 'SIGTERM');
    expect(await runtime).toBe(143);
    expect(process.listenerCount('SIGTERM')).toBe(signalListenersBefore);
  });

  test('keeps private config IPC off the standard LSP stream', async () => {
    const cwd = await fs.mkdtemp(path.join(os.tmpdir(), 'rslint-editor-'));
    tempDirectories.add(cwd);
    await fs.writeFile(
      path.join(cwd, 'rslint.config.mjs'),
      `console.log('config console output');\n` +
        `process.stdout.write('config direct stdout\\n');\n` +
        `setInterval(() => {}, 60_000);\n` +
        `export default [];\n`,
    );
    const entry = fileURLToPath(
      new URL('../dist/editor-runtime.js', import.meta.url),
    );
    const child = spawn(process.execPath, [entry], {
      cwd,
      stdio: ['pipe', 'pipe', 'pipe'],
    });
    children.add(child);
    let stderr = '';
    child.stderr.setEncoding('utf8');
    child.stderr.on('data', (chunk: string) => {
      stderr += chunk;
    });

    let buffer = Buffer.alloc(0);
    const responses = new Map<
      number | string,
      { resolve(message: JsonRpcMessage): void; reject(error: Error): void }
    >();
    let protocolError: Error | undefined;
    child.stdout.on('data', (chunk: Buffer) => {
      buffer = Buffer.concat([buffer, chunk]);
      try {
        for (;;) {
          const headerEnd = buffer.indexOf('\r\n\r\n');
          if (headerEnd < 0) return;
          const header = buffer.subarray(0, headerEnd).toString('ascii');
          const match = /^Content-Length: (\d+)$/im.exec(header);
          if (!match) {
            throw new Error(`non-LSP bytes reached stdout: ${header}`);
          }
          const length = Number(match[1]);
          const frameEnd = headerEnd + 4 + length;
          if (buffer.length < frameEnd) return;
          const body = buffer.subarray(headerEnd + 4, frameEnd);
          buffer = buffer.subarray(frameEnd);
          const message = JSON.parse(body.toString('utf8')) as JsonRpcMessage;
          if (message.method && message.id !== undefined) {
            child.stdin.write(
              encodeLsp({ jsonrpc: '2.0', id: message.id, result: null }),
            );
            continue;
          }
          if (message.id !== undefined) {
            const pending = responses.get(message.id);
            if (pending) {
              responses.delete(message.id);
              if (message.error) {
                pending.reject(
                  new Error(
                    `LSP request failed: ${JSON.stringify(message.error)}`,
                  ),
                );
              } else {
                pending.resolve(message);
              }
            }
          }
        }
      } catch (error) {
        protocolError =
          error instanceof Error ? error : new Error(String(error));
      }
    });

    let nextId = 1;
    const request = (method: string, params?: unknown) => {
      const id = nextId++;
      const response = new Promise<JsonRpcMessage>((resolve, reject) => {
        responses.set(id, { resolve, reject });
      });
      child.stdin.write(
        encodeLsp(
          params === undefined
            ? { jsonrpc: '2.0', id, method }
            : { jsonrpc: '2.0', id, method, params },
        ),
      );
      return response;
    };
    const notify = (method: string, params: unknown) => {
      child.stdin.write(encodeLsp({ jsonrpc: '2.0', method, params }));
    };

    const rootUri = pathToFileURL(cwd).href;
    await request('initialize', {
      processId: process.pid,
      rootUri,
      workspaceFolders: [{ uri: rootUri, name: 'editor-runtime-test' }],
      capabilities: {
        workspace: {
          didChangeWatchedFiles: {
            dynamicRegistration: true,
            relativePatternSupport: true,
          },
        },
      },
    });
    notify('initialized', {});
    await request('shutdown');
    notify('exit', null);
    child.stdin.end();

    const code = await waitForExit(child, () => stderr);
    children.delete(child);
    expect(protocolError).toBeUndefined();
    expect(code).toBe(0);
  });
});
