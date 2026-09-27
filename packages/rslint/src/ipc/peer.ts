/** Process bootstrap and native storage belong to the IPC session. */
import { spawn, type StdioOptions } from 'node:child_process';
import { IpcClient } from './client.js';
import { INHERITED_FD } from './protocol.generated.js';
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
    if (sources?.fd !== undefined) {
      while (stdio.length < INHERITED_FD) stdio.push('ignore');
      stdio[INHERITED_FD] = sources.fd;
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
