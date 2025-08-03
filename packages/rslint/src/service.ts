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
  messageId: string;
  message: string;
  filePath: string;
  range: Range;
  severity?: string;
  suggestions: any[];
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
  config?: string; // Path to rslint.json config file
  workingDirectory?: string;
  ruleOptions?: Record<string, any>;
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

    // Debug: Log the binary path and check if it exists
    console.error('RSLint binary path:', this.rslintPath);
    console.error('import.meta.dirname:', import.meta.dirname);
    console.error('Binary exists:', require('fs').existsSync(this.rslintPath));
    if (!require('fs').existsSync(this.rslintPath)) {
      // Try to list what's in the directory
      const binDir = path.join(import.meta.dirname, '../bin');
      console.error('Bin directory:', binDir);
      try {
        console.error(
          'Bin directory contents:',
          require('fs').readdirSync(binDir),
        );
      } catch (e: any) {
        console.error('Failed to read bin directory:', e.message);
      }
    }

    this.process = spawn(this.rslintPath, ['--api'], {
      stdio: ['pipe', 'pipe', 'inherit'],
      cwd: options.workingDirectory || process.cwd(),
      env: {
        ...process.env,
      },
    });

    // Set up binary message reading
    this.process.stdout!.on('data', data => {
      this.handleChunk(data);
    });
    this.chunks = [];
    this.chunkSize = 0;
    this.expectedSize = null;

    // Handle process errors
    this.process.on('error', err => {
      console.error('RSLint process error:', err);
      this.rejectAllPending(err);
    });

    this.process.on('exit', (code, signal) => {
      const err = new Error(
        `RSLint process exited with code ${code}, signal ${signal}`,
      );
      if (code !== 0) {
        console.error(err.message);
      }
      this.rejectAllPending(err);
    });

    // Handle stdout/stderr close events
    this.process.stdout!.on('close', () => {
      this.rejectAllPending(new Error('RSLint process stdout closed'));
    });
  }

  /**
   * Send a message to the rslint process
   */
  private async sendMessage(kind: string, data: any): Promise<any> {
    return new Promise((resolve, reject) => {
      // Check if process is still alive
      if (this.process.killed || this.process.exitCode !== null) {
        reject(new Error('RSLint process is not running'));
        return;
      }

      const id = this.nextMessageId++;
      const message = { id, kind, data };

      // Register promise callbacks with timeout
      const timeoutId = setTimeout(() => {
        this.pendingMessages.delete(id);
        reject(new Error(`Message ${id} (${kind}) timed out after 30 seconds`));
      }, 30000); // 30 second timeout

      this.pendingMessages.set(id, {
        resolve: result => {
          clearTimeout(timeoutId);
          resolve(result);
        },
        reject: error => {
          clearTimeout(timeoutId);
          reject(error);
        },
      });

      try {
        // Write message length as 4 bytes in little endian
        const json = JSON.stringify(message);
        const jsonBuffer = Buffer.from(json, 'utf8');
        const length = Buffer.alloc(4);
        length.writeUInt32LE(jsonBuffer.length, 0); // Use byte length, not string length

        // Send message
        const success = this.process.stdin!.write(
          Buffer.concat([length, jsonBuffer]),
        );

        if (!success) {
          console.warn('Write buffer is full, may cause backpressure');
        }
      } catch (error) {
        clearTimeout(timeoutId);
        this.pendingMessages.delete(id);
        reject(error);
      }
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
   * Reject all pending messages
   */
  private rejectAllPending(error: Error): void {
    for (const pending of this.pendingMessages.values()) {
      pending.reject(error);
    }
    this.pendingMessages.clear();
  }

  /**
   * Run the linter on specified files
   */
  async lint(options: LintOptions = {}): Promise<LintResponse> {
    const { files, config, workingDirectory, ruleOptions, fileContents } =
      options;
    // Send handshake
    await this.sendMessage('handshake', { version: '1.0.0' });

    // Send lint request
    return this.sendMessage('lint', {
      files,
      config,
      workingDirectory,
      ruleOptions,
      fileContents,
      format: 'jsonline',
    });
  }

  /**
   * Close the rslint process
   */
  async close(): Promise<void> {
    return new Promise(resolve => {
      // Set a timeout to force cleanup if the process doesn't respond
      const timeout = setTimeout(() => {
        console.warn(
          'RSLint process did not respond to exit message, forcing cleanup',
        );
        this.forceCleanup();
        resolve();
      }, 5000); // 5 second timeout

      this.sendMessage('exit', {})
        .then(() => {
          clearTimeout(timeout);
          this.forceCleanup();
          resolve();
        })
        .catch(() => {
          clearTimeout(timeout);
          this.forceCleanup();
          resolve();
        });
    });
  }

  /**
   * Force cleanup of the process and resources
   */
  private forceCleanup(): void {
    try {
      // Reject all pending messages
      this.rejectAllPending(new Error('Service shutting down'));

      // Close stdin if it's still open
      if (this.process.stdin && !this.process.stdin.destroyed) {
        this.process.stdin.end();
      }

      // Kill the process if it's still running
      if (!this.process.killed) {
        this.process.kill('SIGTERM');

        // Force kill after a short delay if SIGTERM doesn't work
        setTimeout(() => {
          if (!this.process.killed) {
            this.process.kill('SIGKILL');
          }
        }, 1000);
      }

      // Clean up buffers
      this.chunks = [];
      this.chunkSize = 0;
      this.expectedSize = null;
    } catch (error) {
      console.error('Error during force cleanup:', error);
    }
  }
}
