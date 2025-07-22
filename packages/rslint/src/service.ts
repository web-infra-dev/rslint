import { spawn, ChildProcess } from 'child_process';
import path from 'path';

/**
 * Types for rslint IPC protocol
 */
interface Position {
  line: number;
  column: number;
}

interface Range {
  start: Position;
  end: Position;
}

export interface Diagnostic {
  ruleName: string;
  message: string;
  filePath: string;
  range: Range;
  severity?: string;
}

export interface LintResponse {
  diagnostics: Diagnostic[];
  errorCount: number;
  fileCount: number;
  ruleCount: number;
  duration: string;
}

export interface LintOptions {
  files?: string[];
  tsconfig?: string;
  workingDirectory?: string;
  ruleOptions?: Record<string, string>;
  fileContents?: Record<string, string>; // Map of file paths to their contents for VFS
}

interface RSlintOptions {
  rslintPath?: string;
  workingDirectory?: string;
}

interface PendingMessage {
  resolve: (data: any) => void;
  reject: (error: Error) => void;
}

/**
 * Wrapper for the rslint binary communication via IPC
 */
export class RSLintService {
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

    this.process = spawn(this.rslintPath, ['--ipc'], {
      stdio: ['pipe', 'pipe', 'inherit'],
      cwd: options.workingDirectory || process.cwd(),
      env: {
        ...process.env,
        RSLINT_IPC: '1', // Alternative way to enable IPC mode
      },
    });

    // Set up binary message reading
    this.process.stdout!.on('data', data => this.handleChunk(data));
    this.chunks = [];
    this.chunkSize = 0;
    this.expectedSize = null;
  }

  /**
   * Send a message to the rslint process
   */
  private sendMessage(kind: string, data: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const id = this.nextMessageId++;
      const message = { id, kind, data };

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
        const parsed = JSON.parse(message);
        this.handleMessage(parsed);
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
   * Handle a complete message from rslint
   */
  private handleMessage(message: {
    id: number;
    kind: string;
    data: any;
  }): void {
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
   * Run the linter on specified files
   */
  async lint(options: LintOptions = {}): Promise<LintResponse> {
    const { files, tsconfig, workingDirectory, ruleOptions, fileContents } =
      options;
    // Send handshake
    await this.sendMessage('handshake', { version: '1.0.0' });

    // Send lint request
    return await this.sendMessage('lint', {
      files,
      tsconfig,
      workingDirectory,
      ruleOptions,
      fileContents,
      format: 'jsonline',
    });
  }

  /**
   * Close the rslint process
   */
  close(): Promise<void> {
    return new Promise(resolve => {
      this.sendMessage('exit', {}).finally(() => {
        this.process.stdin!.end();
        resolve();
      });
    });
  }
}
