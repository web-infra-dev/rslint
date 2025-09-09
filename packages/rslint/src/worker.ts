/// <reference lib="webworker" />

/**
 * Web Worker implementation for rslint in browser environment
 * This worker handles communication with the rslint binary or WASM implementation
 */

import type { IpcMessage } from './types.js';

function hasProp<K extends string>(
  obj: unknown,
  key: K,
): obj is Record<K, unknown> {
  return typeof obj === 'object' && obj !== null && key in obj;
}

function isIpcMessage(value: unknown): value is IpcMessage {
  if (!value || typeof value !== 'object') return false;
  const obj = value as { id?: unknown; kind?: unknown };
  return typeof obj.id === 'number' && typeof obj.kind === 'string';
}

// Global state for the worker
let rslintProcess: unknown = null;
let pendingMessages = new Map<
  number,
  { resolve: (data: unknown) => void; reject: (error: Error) => void }
>();

/**
 * Initialize the rslint process (could be WASM or other browser-compatible implementation)
 */
function initializeRslint(): void {
  try {
    // In a real implementation, this would load the rslint WASM module
    // or initialize a browser-compatible version of rslint
    // For now, we'll simulate the initialization

    // Example: Load WASM module
    // const rslintWasm = await import('./rslint.wasm');
    // rslintProcess = await rslintWasm.default();

    console.log('Rslint worker initialized');
  } catch (error) {
    console.error('Failed to initialize rslint:', error);
    throw error;
  }
}

/**
 * Send a message to the rslint process
 */
function sendToRslint(kind: string, data: unknown): unknown {
  if (!rslintProcess) {
    throw new Error('Rslint process not initialized');
  }

  // In a real implementation, this would call the appropriate method on the rslint process
  // For now, we'll simulate the response

  switch (kind) {
    case 'handshake':
      return { version: '1.0.0', status: 'ok' };

    case 'lint': {
      let files: unknown = undefined;
      if (hasProp(data, 'files')) {
        files = data.files;
      }
      const fileCount = Array.isArray(files)
        ? files.length
        : typeof files === 'string'
          ? 1
          : 0;
      // Simulate linting response
      return {
        diagnostics: [],
        errorCount: 0,
        fileCount,
        ruleCount: 0,
        duration: '0ms',
      };
    }

    case 'applyFixes': {
      let fileContent: unknown = undefined;
      let diagnostics: unknown = undefined;
      if (hasProp(data, 'fileContent')) fileContent = data.fileContent;
      if (hasProp(data, 'diagnostics')) diagnostics = data.diagnostics;
      const fixedContent =
        typeof fileContent === 'string' ? [fileContent] : [''];
      const unappliedCount = Array.isArray(diagnostics) ? diagnostics.length : 0;
      // Simulate apply fixes response
      return {
        fixedContent,
        wasFixed: false,
        appliedCount: 0,
        unappliedCount,
      };
    }

    case 'exit':
      rslintProcess = null;
      return { status: 'ok' };

    default:
      throw new Error(`Unknown message kind: ${kind}`);
  }
}

/**
 * Handle messages from the main thread
 */
async function handleMessage(evt: MessageEvent): Promise<void> {
  const raw = evt.data as unknown;
  if (!isIpcMessage(raw)) {
    return;
  }
  const { id, kind, data } = raw;

  try {
    // Ensure rslint is initialized
    if (!rslintProcess && kind !== 'exit') {
      initializeRslint();
    }

    // Send message to rslint and get response
    const response = await Promise.resolve(sendToRslint(kind, data));

    // Send response back to main thread
    self.postMessage({
      id,
      kind: 'response',
      data: response,
    });
  } catch (error) {
    // Send error back to main thread
    self.postMessage({
      id,
      kind: 'error',
      data: {
        message: error instanceof Error ? error.message : String(error),
      },
    });
  }
}

/**
 * Handle worker errors
 */
function handleError(error: ErrorEvent): void {
  console.error('Worker error:', error);

  // Send error to main thread for all pending messages
  for (const [id, pending] of pendingMessages) {
    self.postMessage({
      id,
      kind: 'error',
      data: {
        message: `Worker error: ${error.message}`,
      },
    });
  }
  pendingMessages.clear();
}

self.addEventListener('message', evt => {
  void handleMessage(evt);
});
self.addEventListener('error', handleError);

// Initialize the worker
console.log('Rslint worker started');
