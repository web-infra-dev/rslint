/** Process bootstrap and native storage belong to the IPC session. */
import { spawn } from 'node:child_process';
import { IpcClient } from './client.js';
import {
  createMemoryTransport,
  type MemoryTransport,
} from './memory-transport.js';

export interface IpcProcessOptions {
  binPath: string;
  goArgs: string[];
  cwd?: string;
  /** Disable the optional shared backend while retaining complete inline bytes. */
  sharedMemory?: boolean;
  /** @internal Test seam for optional native storage and inline fallback. */
  createMemoryTransport?: () => MemoryTransport | undefined;
}

export function spawnIpcProcess(options: IpcProcessOptions) {
  const child = spawn(options.binPath, options.goArgs, {
    stdio: ['pipe', 'pipe', 'inherit'],
    cwd: options.cwd ?? process.cwd(),
  });
  if (!child.stdin || !child.stdout) {
    child.kill();
    throw new Error('IPC process is missing stdin/stdout');
  }
  const client = new IpcClient(child.stdout, child.stdin, {
    createMemoryTransport:
      options.sharedMemory === false
        ? undefined
        : (options.createMemoryTransport ?? createMemoryTransport),
  });
  // These pipes belong to the child until its close event, beyond the IPC
  // client's lifetime. A pending write can emit EPIPE after client.close()
  // detaches its listeners during termination. The active client still owns
  // error reporting; these guards only prevent a late unhandled event.
  const ignorePipeError = () => {
    // The child retains ownership of queued pipe errors until close.
  };
  child.stdin.on('error', ignorePipeError);
  child.stdout.on('error', ignorePipeError);
  child.once('close', () => {
    child.stdin?.off('error', ignorePipeError);
    child.stdout?.off('error', ignorePipeError);
  });
  return { child, client };
}
