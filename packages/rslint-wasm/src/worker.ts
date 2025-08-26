import '../wasm_exec';
import * as memfs from 'memfs';
import { Buffer } from 'buffer';
/// <reference lib="webworker" />

/**
 * Web Worker implementation for rslint in browser environment
 * This worker handles communication with the rslint binary or WASM implementation
 */

declare interface IpcMessage {
  id: number;
  kind: string;
  data: any;
}

// Global state for the worker
let nextMessageId = 1;
let pendingMessages = new Map<
  number,
  { resolve: (data: any) => void; reject: (error: Error) => void }
>();

/**
 * Initialize the rslint process (could be WASM or other browser-compatible implementation)
 */
async function initializeRslint(wasmURL: string): Promise<void> {
  try {
    console.log('Initializing rslint with WASM URL:', wasmURL);
    const go = new Go();
    go.argv = ['rslint', '--api'];
    console.log('go', go.importObject);
    const result = await WebAssembly.instantiateStreaming(
      fetch(wasmURL),
      go.importObject,
    );
    go.run(result.instance);

    console.log('Rslint worker initialized');
  } catch (error) {
    console.error('Failed to initialize rslint:', error);
    throw error;
  }
}

let stdin: any[] = [];
let resumeStdin;
let stdinPos = 0;
let stderr = '';
let decoder = new TextDecoder();
let servicePromise: Promise<void> | undefined;
async function ensureServiceIsRunning() {
  if (!servicePromise) {
    throw new Error('serivce is not running');
  } else {
    return await servicePromise;
  }
}

function encodeMessage(message: any): Uint8Array {
  // Serialize to JSON
  const json = JSON.stringify(message);

  // Encode JSON string as UTF-8
  const encoder = new TextEncoder();
  const jsonBytes = encoder.encode(json);

  // Allocate 4 bytes for length + payload
  const buffer = new Uint8Array(4 + jsonBytes.length);
  const view = new DataView(buffer.buffer);

  // Write length as little endian
  view.setUint32(0, jsonBytes.length, true);

  // Copy payload after the 4-byte header
  buffer.set(jsonBytes, 4);

  return buffer;
}
/**
 * Handle messages from the main thread
 */
async function handleRequest(event: MessageEvent): Promise<void> {
  console.log('handleRequest', event);
  const { id, kind, data } = event.data as IpcMessage;

  /**
   * Send a message to the rslint process
   */
  async function sendMessage(kind: string, data: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = nextMessageId++;
      const message: IpcMessage = { id, kind, data };

      // Register promise callbacks
      pendingMessages.set(id, { resolve, reject });

      // Write message length as 4 bytes in little endian
      const buffer = encodeMessage(message);
      stdin.push(buffer);
      if (resumeStdin) resumeStdin();
    });
  }

  async function handleInit() {
    const rslint_config = [
      {
        language: 'javascript',
        files: [],
        languageOptions: {
          parserOptions: {
            projectService: false,
            project: ['./tsconfig.json'],
          },
        },
        rules: {
          '@typescript-eslint/no-unsafe-member-access': 'error',
        },
        plugins: ['@typescript-eslint'],
      },
    ];
    const tsconfig = {
      compilerOptions: {
        strictNullChecks: true,
      },
      include: ['**/*.ts'],
    };
    const inner_fs = memfs.Volume.fromJSON({
      '/rslint.json': JSON.stringify(rslint_config),
      '/tsconfig.json': JSON.stringify(tsconfig),
      '/index.ts': 'let a:any; a.b = 10',
    });
    //globalThis.fs = inner_fs.fs;
    let fs = globalThis.fs;
    let process = globalThis.process;
    process.cwd = () => {
      console.log('process.cwd');
      return '/';
    };

    fs.writeSync = (fd, buffer) => {
      if (fd === 1) {
        self.postMessage(buffer);
      } else if (fd === 2) {
        stderr += decoder.decode(buffer);
        let parts = stderr.split('\n');
        if (parts.length > 1) console.log(parts.slice(0, -1).join('\n'));
        stderr = parts[parts.length - 1];
      } else {
        throw new Error('Bad write');
      }
      return buffer.length;
    };
    fs.stat = (path, callback) => {
      inner_fs.stat(path, (err, data) => {
        callback(err, data);
      });
    };
    fs.readFile = (path, callback) => {
      inner_fs.readFile(path, (err, data) => {
        callback(err, data);
      });
    };
    fs.readFileSync = (path, options) => {
      inner_fs.readFileSync(path, options);
    };
    fs.open = (path, flags, mode, callback) => {
      console.log('open', path, flags, mode);
      inner_fs.open(path, flags, mode, (err, data) => {
        callback(err, data);
      });
    };
    fs.openSync = (path, flags, mode) => {
      inner_fs.openSync(path, flags, mode);
    };
    fs.fstat = (path, callback) => {
      inner_fs.fstat(path, callback);
    };
    fs.close = (fd, callback) => {
      inner_fs.close(fd, (err, data) => {
        callback(err, data);
      });
    };

    fs.read = (
      fd: number,
      buffer: Uint8Array,
      offset: number,
      length: number,
      position: null,
      callback: (err: Error | null, count?: number) => void,
    ) => {
      if (
        fd !== 0 ||
        offset !== 0 ||
        length !== buffer.length ||
        position !== null
      ) {
        // convert uint8array to buffer
        inner_fs.read(
          fd,
          buffer,
          offset,
          length,
          position as any,
          (err, data, ...rest) => {
            console.log('read', err, data, rest);
            callback(err, data);
          },
        );
      }

      if (stdin.length === 0) {
        resumeStdin = () =>
          fs.read(fd, buffer, offset, length, position, callback);
        return;
      }

      let first = stdin[0];
      let count = Math.max(0, Math.min(length, first.length - stdinPos));
      buffer.set(first.subarray(stdinPos, stdinPos + count), offset);
      stdinPos += count;
      if (stdinPos === first.length) {
        stdin.shift();
        stdinPos = 0;
      }
      callback(null, count);
    };

    await initializeRslint(data.wasmURL);
    self.postMessage(
      encodeMessage({
        id,
        kind: 'response',
        data: {
          status: 'ok',
        },
      }),
    );
    return;
  }
  try {
    // handle worker setup
    if (kind == 'init') {
      servicePromise = handleInit();
      return;
    } else {
      await ensureServiceIsRunning();
      sendMessage(kind, data);
    }
  } catch (error) {
    // Send error back to main thread
    self.postMessage(
      encodeMessage({
        id,
        kind: 'error',
        data: {
          message: error instanceof Error ? error.message : String(error),
        },
      }),
    );
  }
}

/**
 * Handle worker errors
 */
function handleError(error: ErrorEvent): void {
  console.error('Worker error:', error);

  // Send error to main thread for all pending messages
  for (const [id, pending] of pendingMessages) {
    self.postMessage(
      encodeMessage({
        id,
        kind: 'error',
        data: {
          message: `Worker error: ${error.message}`,
        },
      }),
    );
  }
  pendingMessages.clear();
}

// Set up event listeners
self.addEventListener('message', handleRequest);
self.addEventListener('error', handleError);

// Initialize the worker
console.log('Rslint worker started');
