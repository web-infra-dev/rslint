/// <reference lib="webworker" />

/**
 * Web Worker implementation for rslint in browser environment
 * This worker handles communication with the rslint binary or WASM implementation
 */

interface IpcMessage {
  id: number;
  kind: string;
  data: unknown;
}

// Global state for the worker
let rslintProcess: unknown = null;
let nextMessageId = 1;
let pendingMessages = new Map<
  number,
  { resolve: (data: unknown) => void; reject: (error: Error) => void }
>();

/**
 * Initialize the rslint process (could be WASM or other browser-compatible implementation)
 */
async function initializeRslint(): Promise<void> {
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
function sendToRslint(kind: string, data: unknown): Promise<unknown> {
  if (!rslintProcess) {
    return Promise.reject(new Error('Rslint process not initialized'));
  }

  // In a real implementation, this would call the appropriate method on the rslint process
  // For now, we'll simulate the response

  switch (kind) {
    case 'handshake':
      return Promise.resolve({ version: '1.0.0', status: 'ok' });

    case 'lint': {
      const files = (data as { files?: unknown[] }).files;
      // Simulate linting response
      return Promise.resolve({
        diagnostics: [],
        errorCount: 0,
        fileCount: files?.length || 0,
        ruleCount: 0,
        duration: '0ms',
      });
    }

    case 'applyFixes': {
      const fixData = data as {
        fileContent: string;
        diagnostics?: unknown[];
      };
      // Simulate apply fixes response
      return Promise.resolve({
        fixedContent: [fixData.fileContent],
        wasFixed: false,
        appliedCount: 0,
        unappliedCount: fixData.diagnostics?.length || 0,
      });
    }

    case 'exit':
      rslintProcess = null;
      return Promise.resolve({ status: 'ok' });

    default:
      return Promise.reject(new Error(`Unknown message kind: ${kind}`));
  }
}

/**
 * Handle messages from the main thread
 */
async function handleMessage(event: MessageEvent<IpcMessage>): Promise<void> {
  const { id, kind, data } = event.data;

  try {
    // Ensure rslint is initialized
    if (!rslintProcess && kind !== 'exit') {
      await initializeRslint();
    }

    // Send message to rslint and get response
    const response = await sendToRslint(kind, data);

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

// Set up event listeners
self.addEventListener('message', (event: MessageEvent<IpcMessage>) => {
  void handleMessage(event);
});
self.addEventListener('error', handleError);

// Initialize the worker
console.log('Rslint worker started');
