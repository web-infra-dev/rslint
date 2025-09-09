/// <reference lib="webworker" />

import type {
  RslintServiceInterface,
  LintOptions,
  LintResponse,
  ApplyFixesRequest,
  ApplyFixesResponse,
  RSlintOptions,
  PendingMessage,
  IpcMessage,
} from './types.js';

/**
 * Browser implementation of RslintService using web workers
 */
export class BrowserRslintService implements RslintServiceInterface {
  private nextMessageId: number;
  private pendingMessages: Map<number, PendingMessage>;
  private worker!: Worker | null;
  private workerUrl: string;
  private chunks: Uint8Array[];
  private chunkSize: number;
  private expectedSize: number | null;

  constructor(options: RSlintOptions & { workerUrl: string; wasmUrl: string }) {
    this.nextMessageId = 1;
    this.pendingMessages = new Map();
    this.chunks = [];
    this.chunkSize = 0;
    this.expectedSize = null;

    // In browser, we need to use a web worker that can run the rslint binary
    // This would typically be a WASM version or a worker that can spawn processes
    this.workerUrl = options.workerUrl;
    void this.ensureWorker(options.wasmUrl);
  }

  /**
   * Initialize the web worker
   */
  private ensureWorker(wasmUrl: string): Worker {
    if (!this.worker) {
      this.worker = new Worker(this.workerUrl, { name: 'rslint-worker.js' });

      this.worker.onmessage = (event: MessageEvent) => {
        const data: unknown = event.data;
        if (data instanceof Uint8Array) {
          this.handlePacket(data);
        }
      };

      this.worker.onerror = error => {
        console.error('Worker error:', error);
        // Reject all pending messages
        for (const [id, pending] of this.pendingMessages) {
          pending.reject(new Error(`Worker error: ${error.message}`));
        }
        this.pendingMessages.clear();
      };
      this.worker.postMessage({
        kind: 'init',
        data: { version: '1.0.0', wasmURL: wasmUrl },
      });
    }
    return this.worker;
  }

  /**
   * Handle incoming binary data chunks
   */
  private handlePacket(chunk: Uint8Array): void {
    this.chunks.push(chunk);
    this.chunkSize += chunk.length;

    // Process complete messages
    while (true) {
      // Read message length if we don't have it yet
      if (this.expectedSize === null) {
        if (this.chunkSize < 4) return;

        // Combine chunks to read the message length
        const combined = this.combineChunks();
        const dataView = new DataView(
          combined.buffer,
          combined.byteOffset,
          combined.byteLength,
        );
        this.expectedSize = dataView.getUint32(0, true); // true for little-endian

        // Remove length bytes from buffer
        this.chunks = [combined.slice(4)];
        this.chunkSize -= 4;
      }

      // Check if we have the full message
      if (this.chunkSize < this.expectedSize) return;

      // Read the message content
      const combined = this.combineChunks();
      const messageBytes = combined.slice(0, this.expectedSize);
      const message = new TextDecoder().decode(messageBytes);

      // Handle the message
      try {
        const raw = JSON.parse(message) as unknown;
        if (BrowserRslintService.isIpcMessage(raw)) {
          this.handleResponse(raw);
        }
      } catch (err) {
        console.error('Error parsing message:', err);
      }

      // Reset for next message
      this.chunks = [combined.slice(this.expectedSize)];
      this.chunkSize = this.chunks[0].length;
      this.expectedSize = null;
    }
  }

  /**
   * Combine multiple Uint8Array chunks into a single Uint8Array
   */
  private combineChunks(): Uint8Array {
    if (this.chunks.length === 1) {
      return this.chunks[0];
    }

    const totalLength = this.chunks.reduce(
      (sum, chunk) => sum + chunk.length,
      0,
    );
    const combined = new Uint8Array(totalLength);
    let offset = 0;

    for (const chunk of this.chunks) {
      combined.set(chunk, offset);
      offset += chunk.length;
    }

    return combined;
  }

  private static isIpcMessage(value: unknown): value is IpcMessage {
    if (!value || typeof value !== 'object') return false;
    const obj = value as { id?: unknown; kind?: unknown };
    return typeof obj.id === 'number' && typeof obj.kind === 'string';
  }

  /**
   * Send a message to the worker
   */
  async sendMessage(kind: string, data: unknown): Promise<unknown> {
    return new Promise((resolve, reject) => {
      const id = this.nextMessageId++;
      const message: IpcMessage = { id, kind, data };

      // Register promise callbacks
      this.pendingMessages.set(id, { resolve, reject });

      // Send message to worker
      this.worker!.postMessage(message);
    });
  }

  /**
   * Handle messages from the worker
   */
  private handleResponse(message: IpcMessage): void {
    const { id, kind, data } = message;
    const pending = this.pendingMessages.get(id);
    if (!pending) return;

    this.pendingMessages.delete(id);

    if (kind === 'error') {
      let msg = 'Unknown error';
      if (data && typeof data === 'object') {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
        const m = (data as Record<string, unknown>).message;
        if (typeof m === 'string') msg = m;
      }
      pending.reject(new Error(msg));
    } else {
      pending.resolve(data);
    }
  }

  /**
   * Terminate the worker
   */
  terminate(): void {
    if (this.worker) {
      // Reject all pending messages
      for (const [id, pending] of this.pendingMessages) {
        pending.reject(new Error('Service terminated'));
      }
      this.pendingMessages.clear();

      this.worker.terminate();
      this.worker = null;
    }
  }
}
