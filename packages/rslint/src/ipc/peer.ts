/** Process bootstrap and native storage belong to the IPC session. */
import { spawn, type StdioOptions } from 'node:child_process';
import { IpcClient } from './client.js';
import {
  createSourceTransport,
  type SourceTransport,
} from './source-transport.js';

export interface IpcPeerOptions {
  binPath: string;
  goArgs: string[];
  cwd?: string;
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

export function spawnIpcPeer(options: IpcPeerOptions) {
  const sources = options.createSourceTransport
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
      throw new Error('IPC peer is missing stdin/stdout');
    }
    const client = new IpcClient(child.stdout, child.stdin, {
      sourceTransport: sources,
      inheritedSourceFd,
    });
    return {
      child,
      client,
      close: () => {
        client.close();
      },
    };
  } catch (error) {
    sources?.close();
    throw error;
  }
}
