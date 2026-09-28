/** Process bootstrap and native storage belong to the IPC session. */
import { spawn, type StdioOptions } from 'node:child_process';
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

function optionalMemoryTransport(): MemoryTransport | undefined {
  // Missing/older native addons and restricted hosts retain complete inline attachments.
  // This loader does not import the plugin runtime.
  try {
    return createMemoryTransport();
  } catch {
    return undefined;
  }
}

export function spawnIpcProcess(options: IpcProcessOptions) {
  const memory =
    options.sharedMemory === false
      ? undefined
      : options.createMemoryTransport
        ? options.createMemoryTransport()
        : optionalMemoryTransport();
  try {
    const stdio: StdioOptions = ['pipe', 'pipe', 'inherit'];
    const fd = memory?.fd();
    let inheritedMemoryFd: number | undefined;
    if (fd !== undefined) {
      // Bootstrap reserves the first extra stdio slot; append future handles after it.
      inheritedMemoryFd = stdio.length;
      stdio.push(fd);
    }
    const child = spawn(options.binPath, options.goArgs, {
      stdio,
      cwd: options.cwd ?? process.cwd(),
    });
    if (!child.stdin || !child.stdout) {
      child.kill();
      throw new Error('IPC process is missing stdin/stdout');
    }
    const client = new IpcClient(child.stdout, child.stdin, {
      memoryTransport: memory,
      inheritedMemoryFd,
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
  } catch (error) {
    memory?.close();
    throw error;
  }
}
