/** Process bootstrap and native storage belong to the IPC session. */
import { spawn, type StdioOptions } from 'node:child_process';
import { IpcClient } from './client.js';
import {
  createMemoryTransport,
  type MemoryTransport,
} from './memory-transport.js';
import { createRpcTransport, type RpcTransport } from './rpc-transport.js';

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

function optionalRpcTransport(): RpcTransport | undefined {
  try {
    return createRpcTransport();
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
  // Only prepare the optional RPC channel when shared memory is active and not
  // overridden by a custom memory-transport test seam, keeping stdio slots
  // fixed: fd 3 = MemoryArena, fd 4 = RpcChannel SHM, fd 5 = RpcChannel doorbell.
  let rpc =
    memory !== undefined && options.createMemoryTransport === undefined
      ? optionalRpcTransport()
      : undefined;
  try {
    const stdio: StdioOptions = ['pipe', 'pipe', 'inherit'];
    const fd = memory?.fd();
    let inheritedMemoryFd: number | undefined;
    if (fd !== undefined) {
      // Bootstrap reserves stdio[3] for MemoryArena; append RPC handles at 4 and 5.
      inheritedMemoryFd = stdio.length;
      stdio.push(fd);
    }
    let inheritedRpcFd: number | undefined;
    let inheritedRpcDoorbellFd: number | undefined;
    const rpcFd = rpc?.fd();
    const rpcDoorbellFd = rpc?.doorbellFd();
    if (
      rpcFd !== undefined &&
      rpcDoorbellFd !== undefined &&
      inheritedMemoryFd === 3
    ) {
      inheritedRpcFd = stdio.length;
      stdio.push(rpcFd);
      inheritedRpcDoorbellFd = stdio.length;
      stdio.push(rpcDoorbellFd);
    } else if (rpc && (rpcFd !== undefined || rpcDoorbellFd !== undefined)) {
      // Prepared Unix descriptors could not be placed at fixed slots 4/5:
      // close immediately so the process does not leak idle FDs.
      rpc.close();
      rpc = undefined;
    }
    const child = spawn(options.binPath, options.goArgs, {
      stdio,
      cwd: options.cwd ?? process.cwd(),
    });
    // Close parent's copy of the child's doorbell socket end immediately so
    // child death delivers EOF on the parent's reader thread.
    rpc?.afterSpawn();
    if (!child.stdin || !child.stdout) {
      child.kill();
      throw new Error('IPC process is missing stdin/stdout');
    }
    const client = new IpcClient(child.stdout, child.stdin, {
      memoryTransport: memory,
      inheritedMemoryFd,
      rpcTransport: rpc,
      inheritedRpcFd,
      inheritedRpcDoorbellFd,
      peerPid: child.pid,
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
    rpc?.close();
    memory?.close();
    throw error;
  }
}
