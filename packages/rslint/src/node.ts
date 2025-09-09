import { spawn, ChildProcess } from 'child_process';
import path from 'path';
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
 * Node.js implementation of RslintService using child processes
 */
export class NodeRslintService implements RslintServiceInterface {
  private nextMessageId: number;
  private pendingMessages: Map<number, PendingMessage>;
  private rslintPath: string;
  private process: ChildProcess;
  private chunks: Buffer[];
  private chunkSize: number;
  private expectedSize: number | null;

  constructor(options: RSlintOptions = {}) {
    this.nextMessageId = 1;
    this.pendingMessages = new Map();
    this.rslintPath =
      options.rslintPath || path.join(import.meta.dirname, '../bin/rslint');

    this.process = spawn(this.rslintPath, ['--api'], {
      stdio: ['pipe', 'pipe', 'inherit'],
      cwd: options.workingDirectory || process.cwd(),
      env: {
        ...process.env,
      },
    });

    // Set up binary message reading
    this.process.stdout!.on('data', (data: Buffer) => {
      this.handleChunk(data);
    });
    this.chunks = [];
    this.chunkSize = 0;
    this.expectedSize = null;
  }

  /**
   * Send a message to the rslint process
   */
  async sendMessage(kind: string, data: unknown): Promise<unknown> {
    return new Promise((resolve, reject) => {
      const id = this.nextMessageId++;
      const message: IpcMessage = { id, kind, data };

      // Register promise callbacks
      this.pendingMessages.set(id, { resolve, reject });

      // Write message length as 4 bytes in little endian
      const json = JSON.stringify(message);
      const length = Buffer.alloc(4);
      length.writeUInt32LE(json.length, 0);

      // Send message
      this.process.stdin!.write(
        Buffer.concat([length, Buffer.from(json, 'utf8')]),
      );
    });
  }

  /**
   * Handle incoming binary data chunks
   */
  private handleChunk(chunk: Buffer): void {
    this.chunks.push(chunk);
    this.chunkSize += chunk.length;

    // Process complete messages
    while (true) {
      // Read message length if we don't have it yet
      if (this.expectedSize === null) {
        if (this.chunkSize < 4) return;

        // Combine chunks to read the message length
        const combined = Buffer.concat(this.chunks);
        this.expectedSize = combined.readUInt32LE(0);

        // Remove length bytes from buffer
        this.chunks = [combined.slice(4)];
        this.chunkSize -= 4;
      }

      // Check if we have the full message
      if (this.chunkSize < this.expectedSize) return;

      // Read the message content
      const combined = Buffer.concat(this.chunks);
      const message = combined.slice(0, this.expectedSize).toString('utf8');

      // Handle the message
      try {
        const raw = JSON.parse(message) as unknown;
        if (NodeRslintService.isIpcMessage(raw)) {
          this.handleMessage(raw);
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

  private static isIpcMessage(value: unknown): value is IpcMessage {
    if (!value || typeof value !== 'object') return false;
    const obj = value as { id?: unknown; kind?: unknown };
    return typeof obj.id === 'number' && typeof obj.kind === 'string';
  }

  /**
   * Handle a complete message from rslint
   */
  private handleMessage(message: IpcMessage): void {
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
   * Terminate the rslint process
   */
  terminate(): void {
    if (this.process && !this.process.killed) {
      this.process.stdin!.end();
      this.process.kill();
    }
  }
}
