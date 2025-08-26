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
  private worker: Worker | null;
  private workerUrl: string;

  constructor(options: RSlintOptions = {}) {
    this.nextMessageId = 1;
    this.pendingMessages = new Map();
    this.worker = null;

    // In browser, we need to use a web worker that can run the rslint binary
    // This would typically be a WASM version or a worker that can spawn processes
    this.workerUrl =
      options.rslintPath || new URL('./worker.js', import.meta.url).href;
  }

  /**
   * Initialize the web worker
   */
  private async ensureWorker(): Promise<Worker> {
    if (!this.worker) {
      this.worker = new Worker(this.workerUrl);

      this.worker.onmessage = event => {
        this.handleWorkerMessage(event.data);
      };

      this.worker.onerror = error => {
        console.error('Worker error:', error);
        // Reject all pending messages
        for (const [id, pending] of this.pendingMessages) {
          pending.reject(new Error(`Worker error: ${error.message}`));
        }
        this.pendingMessages.clear();
      };
    }
    return this.worker;
  }

  /**
   * Send a message to the worker
   */
  async sendMessage(kind: string, data: any): Promise<any> {
    const worker = await this.ensureWorker();

    return new Promise((resolve, reject) => {
      const id = this.nextMessageId++;
      const message: IpcMessage = { id, kind, data };

      // Register promise callbacks
      this.pendingMessages.set(id, { resolve, reject });

      // Send message to worker
      worker.postMessage(message);
    });
  }

  /**
   * Handle messages from the worker
   */
  private handleWorkerMessage(message: IpcMessage): void {
    const { id, kind, data } = message;
    const pending = this.pendingMessages.get(id);
    if (!pending) return;

    this.pendingMessages.delete(id);

    if (kind === 'error') {
      pending.reject(new Error(data.message));
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
