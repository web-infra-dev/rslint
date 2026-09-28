/** Process bootstrap and native storage belong to the IPC session. */
import { spawn, type StdioOptions } from 'node:child_process';
import { IpcClient } from './client.js';
import {
  createSourceTransport,
  type SourceTransport,
} from './source-transport.js';

export interface IpcProcessOptions {
  binPath: string;
  goArgs: string[];
  cwd?: string;
  /** Enable negotiated source storage only for attachment-aware hosts. */
  sharedSources?: boolean;
  /** @internal Test seam for optional native storage and inline fallback. */
  createSourceTransport?: () => SourceTransport | undefined;
}

function optionalSourceTransport(): SourceTransport | undefined {
  // Missing/older native addons and restricted hosts retain complete inline text.
  // This loader does not import the plugin runtime.
  try {
    return createSourceTransport();
  } catch {
    return undefined;
  }
}

export function spawnIpcProcess(options: IpcProcessOptions) {
  const sources =
    options.sharedSources === false
      ? undefined
      : options.createSourceTransport
        ? options.createSourceTransport()
        : optionalSourceTransport();
  try {
    const stdio: StdioOptions = ['pipe', 'pipe', 'inherit'];
    const fd = sources?.fd();
    let inheritedSourceFd: number | undefined;
    if (fd !== undefined) {
      inheritedSourceFd = stdio.length;
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
      sourceTransport: sources,
      inheritedSourceFd,
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
    sources?.close();
    throw error;
  }
}
