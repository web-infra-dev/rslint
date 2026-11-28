import { projectFrame } from '../languages/js/project-frame.js';
import { describe, test, expect, rs } from 'rstack/test';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import path from 'node:path';
import os from 'node:os';
import { Worker } from 'node:worker_threads';
import { RSLintService } from '../../../rslint/src/service/service.js';
import { NodeRslintService } from '../../../rslint/src/internal/node.js';

import { createPluginLintHost } from '../../../rslint/src/eslint-plugin/host.js';
import { WorkerPool } from '../../../rslint/src/eslint-plugin/worker-pool.js';
import type { EslintPluginLintRequest } from '../../../rslint/src/eslint-plugin/plugin/plugin-lint-protocol.js';

import { LOCAL_CONFIG_DIR, localConfigs } from './worker-pool-e2e-helpers.js';
import { SKIP_WIN32_NAPI_TEARDOWN } from './win32-napi-teardown.js';

/**
 * Plugin-lint host — cancellation via AbortSignal. The LSP path bridges Go's
 * $/cancelRequest (a superseding keystroke / document close) to an AbortSignal;
 * host.lint must cancel the dispatched worker tasks so the worker stops instead
 * of running to completion.
 *
 * win32 teardown is gated by SKIP_WIN32_NAPI_TEARDOWN (same napi-teardown
 * reason as the other worker e2e suites); the flag is false so this runs too.
 */
describe.skipIf(SKIP_WIN32_NAPI_TEARDOWN && process.platform === 'win32')(
  'createPluginLintHost — cancellation via AbortSignal',
  () => {
    test('omitting typeServices transport degrades cleanly for scope-using rules', async () => {
      const directory = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-type-no-transport-'),
      );
      let host: Awaited<ReturnType<typeof createPluginLintHost>> | undefined;
      try {
        const configPath = path.join(directory, 'rslint.config.mjs');
        await writeFile(
          configPath,
          `export default [{plugins:{local:{rules:{scopeCheck:{meta:{schema:[]},create(context){
  const services = context.sourceCode.parserServices;
  return {
    Program(node) {
      const scope = context.sourceCode.getScope(node);
      if (!scope || services.program !== undefined) {
        throw new Error('unexpected parser services without transport');
      }
      context.report({ node, message: 'degraded-ok' });
    },
  };
}}}}}}];`,
        );
        host = await createPluginLintHost(
          [{ configPath, configDirectory: directory }],
          undefined,
          true,
        );
        const file = path.join(directory, 'index.ts');
        const result = await host.lint({
          files: [
            {
              path: file,
              frame: projectFrame(file, 'const value = 1;').toString('base64'),
              configKey: directory,
              typeServices: {
                session: 'typed-session',
                binding: {
                  generation: 1,
                  sourceFileName: file,
                },
              },
            },
          ],
          rules: { 'local/scopeCheck': { options: [] } },
          collectFixes: false,
        });
        expect(result.results[0].ruleErrors).toBeUndefined();
        expect(result.results[0].parseError).toBeUndefined();
        expect(result.results[0].diagnostics).toEqual([
          expect.objectContaining({ message: 'degraded-ok' }),
        ]);
      } finally {
        await host?.shutdown();
        await rm(directory, { recursive: true, force: true });
      }
    });

    test('keeps a Program probe cold and preserves its identity through activation', async () => {
      const directory = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-type-probe-'),
      );
      let host: Awaited<ReturnType<typeof createPluginLintHost>> | undefined;
      try {
        const configPath = path.join(directory, 'rslint.config.mjs');
        await writeFile(
          configPath,
          `let firstProgram, firstChecker, readOptions;
export default [{plugins:{typed:{rules:{probe:{meta:{schema:[]},create(context){
  const services = context.sourceCode.parserServices;
  const phase = context.options[0];
  if (!services.program || !services.experimentalDecorators ||
      typeof services.esTreeNodeToTSNodeMap.has !== 'function')
    throw new Error('missing lightweight parser services');
  if (phase === 'probe') {
    firstProgram = services.program;
    readOptions = firstProgram.getCompilerOptions;
  } else if (phase === 'retired') {
    let closed = false;
    try { readOptions(); } catch (error) { closed = /closed/.test(error.message); }
    if (!closed || services.program === firstProgram)
      throw new Error('retired Program remained usable');
  } else {
    if (services.program !== firstProgram || !readOptions().experimentalDecorators)
      throw new Error('Program identity or compiler options changed');
    const checker = services.program.getTypeChecker();
    firstChecker ??= checker;
    if (checker !== firstChecker) throw new Error('checker identity changed');
  }
  return {Program(node){context.report({node,message:phase});}};
}}}}}}];`,
        );
        host = await createPluginLintHost(
          [{ configPath, configDirectory: directory }],
          undefined,
          true,
        );
        const transport = rs.fn(async () => ({
          descriptor: {
            generation: 1,
            currentDirectory: directory,
            compilerOptions: { experimentalDecorators: true },
            sources: [],
          },
        }));
        for (const phase of ['probe', 'first use', 'second use', 'retired']) {
          if (phase === 'retired') {
            await host.lint({
              files: [],
              collectFixes: false,
              retiredTypeServices: ['first'],
            });
          }
          const file = path.join(directory, `${phase}.ts`);
          const result = await host.lint(
            {
              files: [
                {
                  path: file,
                  frame: projectFrame(file, 'export {};').toString('base64'),
                  configKey: directory,
                  typeServices: {
                    session: phase === 'retired' ? 'second' : 'first',
                    binding: {
                      generation: 1,
                      sourceFileName: file,
                    },
                    experimentalDecorators: true,
                  },
                },
              ],
              rules: { 'typed/probe': { options: [phase] } },
              collectFixes: false,
            },
            undefined,
            undefined,
            transport,
          );
          expect(result.results[0].ruleErrors).toBeUndefined();
          expect(result.results[0].diagnostics).toEqual([
            expect.objectContaining({ message: phase }),
          ]);
          expect(transport).toHaveBeenCalledTimes(phase === 'probe' ? 0 : 1);
        }
      } finally {
        await host?.shutdown();
        await rm(directory, { recursive: true, force: true });
      }
    });

    test.each(['shutdown', 'cancel'] as const)(
      '%s stops a synchronous type query without force termination',
      async (action) => {
        const directory = await mkdtemp(
          path.join(os.tmpdir(), 'rslint-type-shutdown-'),
        );
        let host: Awaited<ReturnType<typeof createPluginLintHost>> | undefined;
        let service: RSLintService | undefined;
        const terminate = rs.spyOn(Worker.prototype, 'terminate');
        const postMessage = rs.spyOn(Worker.prototype, 'postMessage');
        try {
          const configPath = path.join(directory, 'rslint.config.mjs');
          await writeFile(
            configPath,
            `export default [{plugins:{typed:{rules:{wait:{meta:{schema:[]},create(context){return {Program(){context.sourceCode.parserServices.program.getTypeChecker().getVoidType();}};}}}}}}];`,
          );
          await writeFile(
            path.join(directory, 'tsconfig.json'),
            '{"compilerOptions":{"noLib":true},"include":["*.ts"]}',
          );
          const file = path.join(directory, 'input.ts');
          await writeFile(file, 'export const value = 1;');
          host = await createPluginLintHost([
            { configPath, configDirectory: directory },
          ]);
          service = new RSLintService(
            new NodeRslintService({ workingDirectory: directory }),
          );
          const controller = new AbortController();
          let started!: () => void;
          const queryStarted = new Promise<void>((resolve) => {
            started = resolve;
          });
          const lint = service.lint(
            {
              files: [file],
              workingDirectory: directory,
              configDirectory: directory,
              pluginConfigDirectory: directory,
              config: [
                {
                  files: ['**/*.ts'],
                  plugins: ['typed'],
                  languageOptions: {
                    parserOptions: { project: ['./tsconfig.json'] },
                  },
                  rules: { 'typed/wait': 'error' },
                },
              ],
              eslintPlugins: [{ prefix: 'typed', ruleNames: ['wait'] }],
            },
            {
              pluginLint: (request, attachments) =>
                host!.lint(
                  request as EslintPluginLintRequest,
                  controller.signal,
                  attachments,
                  async (_query, signal) =>
                    new Promise((_resolve, reject) => {
                      started();
                      signal?.addEventListener(
                        'abort',
                        () =>
                          reject(new DOMException('cancelled', 'AbortError')),
                        { once: true },
                      );
                    }),
                ),
            },
          );
          await Promise.race([
            queryStarted,
            lint.then(() => {
              throw new Error('typed rule did not issue its query');
            }),
          ]);
          if (action === 'shutdown') await host.shutdown();
          else controller.abort();
          expect((await lint).errorCount).toBe(0);
          expect(terminate).not.toHaveBeenCalled();
          if (action === 'cancel') {
            expect(postMessage).toHaveBeenCalledWith(
              expect.objectContaining({ kind: 'retire-types' }),
            );
          }
        } finally {
          terminate.mockRestore();
          postMessage.mockRestore();
          await host?.shutdown();
          await service?.close();
          await rm(directory, { recursive: true, force: true });
        }
      },
    );

    test('reuses persistent type bridge after mid-query cancellation without leaking stale replies', async () => {
      const directory = await mkdtemp(
        path.join(os.tmpdir(), 'rslint-type-bridge-reuse-'),
      );
      let host: Awaited<ReturnType<typeof createPluginLintHost>> | undefined;
      const terminate = rs.spyOn(Worker.prototype, 'terminate');
      try {
        const configPath = path.join(directory, 'rslint.config.mjs');
        await writeFile(
          configPath,
          `export default [{plugins:{typed:{rules:{probe:{meta:{schema:[]},create(context){
  return {
    Program(node) {
      const opts = context.sourceCode.parserServices.program.getCompilerOptions();
      context.report({ node, message: String(opts.target ?? 'none') });
    },
  };
}}}}}}];`,
        );
        host = await createPluginLintHost(
          [{ configPath, configDirectory: directory }],
          undefined,
          true,
        );
        let firstStarted!: () => void;
        let resolveFirst!: (res: {
          descriptor: {
            generation: number;
            currentDirectory: string;
            compilerOptions: Record<string, unknown>;
            sources: [];
          };
        }) => void;
        const firstReady = new Promise<void>((resolve) => {
          firstStarted = resolve;
        });
        let callCount = 0;
        const transport = async () => {
          callCount++;
          if (callCount === 1) {
            firstStarted();
            return new Promise<{
              descriptor: {
                generation: number;
                currentDirectory: string;
                compilerOptions: Record<string, unknown>;
                sources: [];
              };
            }>((resolve) => {
              resolveFirst = resolve;
            });
          }
          return {
            descriptor: {
              generation: 1,
              currentDirectory: directory,
              compilerOptions: { target: 99 },
              sources: [],
            },
          };
        };

        const file1 = path.join(directory, 'a.ts');
        const ac1 = new AbortController();
        const task1 = host.lint(
          {
            files: [
              {
                path: file1,
                frame: projectFrame(file1, 'export const a = 1;').toString(
                  'base64',
                ),
                configKey: directory,
                typeServices: {
                  session: 'bridge-s1',
                  binding: { generation: 1, sourceFileName: file1 },
                },
              },
            ],
            rules: { 'typed/probe': { options: [] } },
            collectFixes: false,
          },
          ac1.signal,
          undefined,
          transport,
        );
        await firstReady;
        ac1.abort();
        // Deliver a late reply after cancellation to confirm no stale reply leaks into task 2.
        resolveFirst({
          descriptor: {
            generation: 1,
            currentDirectory: directory,
            compilerOptions: { target: 1 },
            sources: [],
          },
        });
        const res1 = await task1;
        expect(res1.results[0].cancelled).toBe(true);

        const file2 = path.join(directory, 'b.ts');
        const res2 = await host.lint(
          {
            files: [
              {
                path: file2,
                frame: projectFrame(file2, 'export const b = 2;').toString(
                  'base64',
                ),
                configKey: directory,
                typeServices: {
                  session: 'bridge-s2',
                  binding: { generation: 1, sourceFileName: file2 },
                },
              },
            ],
            rules: { 'typed/probe': { options: [] } },
            collectFixes: false,
          },
          undefined,
          undefined,
          transport,
        );
        expect(res2.results[0].cancelled).toBe(false);
        expect(res2.results[0].parseError).toBeUndefined();
        expect(res2.results[0].ruleErrors).toBeUndefined();
        expect(res2.results[0].diagnostics).toEqual([
          expect.objectContaining({ message: '99' }),
        ]);
        expect(terminate).not.toHaveBeenCalled();
      } finally {
        terminate.mockRestore();
        await host?.shutdown();
        await rm(directory, { recursive: true, force: true });
      }
    });

    const req: EslintPluginLintRequest = {
      files: [
        {
          path: 'a.ts',
          frame: projectFrame('a.ts', 'const x = null;').toString('base64'),
          configKey: LOCAL_CONFIG_DIR,
        },
      ],
      rules: { 'local/no-null': { options: [] } },
      collectFixes: false,
      suggestionsMode: 'off',
    };

    test('a pre-aborted signal cancels the dispatched task (no diagnostics)', async () => {
      const host = await createPluginLintHost(localConfigs);
      try {
        // Aborted before lint starts → onTaskDispatched sees signal.aborted and
        // cancels the task, so it produces no diagnostics.
        const ac = new AbortController();
        ac.abort();
        const cancelled = await host.lint(req, ac.signal);
        expect(cancelled.results).toHaveLength(1);
        expect(cancelled.results[0].diagnostics).toHaveLength(0);
        // Pin the positive cancellation signal, not just an empty diag array.
        expect(cancelled.results[0].cancelled).toBe(true);

        // Control: the SAME request without a signal runs to completion and the
        // local/no-null rule reports on the `null` literal — proving the empty
        // result above is the cancellation, not a plugin that failed to load.
        const normal = await host.lint(req);
        expect(normal.results).toHaveLength(1);
        expect(normal.results[0].diagnostics).toHaveLength(1);
        const diag = normal.results[0].diagnostics[0] as { ruleName?: string };
        expect(diag.ruleName).toBe('local/no-null');
      } finally {
        await host.shutdown();
      }
    });

    test('aborting after dispatch cancels via the onAbort listener (mid-flight)', async () => {
      const cancelSpy = rs.spyOn(WorkerPool.prototype, 'cancelTask');
      const host = await createPluginLintHost(localConfigs);
      try {
        const ac = new AbortController();
        // host.lint synchronously builds tasks + registers the abort listener,
        // then dispatches (onTaskDispatched pushes the id) before awaiting the
        // worker. Aborting now drives the onAbort-listener path (mid-flight),
        // distinct from the pre-abort branch above.
        const p = host.lint(req, ac.signal);
        ac.abort();
        const res = await p;
        expect(res.results).toHaveLength(1);
        // The cancel was routed to the worker pool (onAbort → cancelTask).
        expect(cancelSpy).toHaveBeenCalled();
      } finally {
        cancelSpy.mockRestore();
        await host.shutdown();
      }
    });
  },
);
