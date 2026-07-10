import { describe, expect, test } from '@rstest/core';

import { RSLintService } from '../src/service/service.js';
import type {
  InboundRequestHandler,
  IpcMessage,
  RslintServiceInterface,
} from '../src/types.js';

class ReverseLintBackend implements RslintServiceInterface {
  inbound: InboundRequestHandler | null = null;
  lintPayloads: unknown[] = [];

  setInboundHandler(handler: InboundRequestHandler | null): void {
    this.inbound = handler;
  }

  async sendMessage(kind: string, data: any): Promise<any> {
    if (kind === 'handshake') return { version: '1.0.0', ok: true };
    if (kind === 'exit') return {};
    if (kind !== 'lint') throw new Error(`unexpected kind ${kind}`);

    this.lintPayloads.push(data);
    // Yield before dispatching the reverse request. Without service-level
    // serialization, a concurrent lint could replace the active handler here.
    await new Promise((resolve) => setTimeout(resolve, 5));
    if (!this.inbound) throw new Error('missing inbound handler');
    return this.inbound({
      id: 100,
      kind: 'pluginLint',
      data: { token: data.files[0] },
    } satisfies IpcMessage);
  }

  terminate(): void {}
}

describe('RSLintService reverse lint request scoping', () => {
  test('forwards eslintPlugins and dispatches pluginLint to the call handler', async () => {
    const backend = new ReverseLintBackend();
    const service = new RSLintService(backend);
    const eslintPlugins = [{ prefix: 'local', ruleNames: ['program'] }];

    await expect(
      service.lint(
        { files: ['a.ts'], eslintPlugins },
        { pluginLint: (request) => ({ request, ok: true }) },
      ),
    ).resolves.toEqual({ request: { token: 'a.ts' }, ok: true });
    expect(backend.lintPayloads[0]).toMatchObject({ eslintPlugins });
    await service.close();
  });

  test('serializes concurrent lint frames so handlers cannot overwrite each other', async () => {
    const backend = new ReverseLintBackend();
    const service = new RSLintService(backend);

    const [a, b] = await Promise.all([
      service.lint({ files: ['a.ts'] }, { pluginLint: () => ({ host: 'a' }) }),
      service.lint({ files: ['b.ts'] }, { pluginLint: () => ({ host: 'b' }) }),
    ]);

    expect(a).toEqual({ host: 'a' });
    expect(b).toEqual({ host: 'b' });
    await service.close();
  });
});
