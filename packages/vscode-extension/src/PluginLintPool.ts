/**
 * VS Code-side host for ESLint-plugin lint requests.
 *
 * The Go LSP server lints natively, but rules mounted via a config's
 * object-form `plugins` run in JS. So when Go encounters such a rule it sends a
 * server→client `rslint/pluginLint` request back to this extension;
 * we answer it from an in-process WorkerPool owned by `@rslint/core`'s
 * `createPluginLintHost`. This file is a thin lifecycle wrapper over that
 * host — the request→tasks→result boundary itself lives in `@rslint/core`
 * (`buildPluginLintTasks` / `buildPluginLintResult`), shared with the CLI
 * engine so the two paths never drift.
 *
 * The host is loaded via a dynamic `import()` of the RELATIVE specifier
 * `./eslint-plugin/index.js` (not the bare `@rslint/core/eslint-plugin`):
 * relative to the built `dist/main.js` it resolves to the extension's OWN
 * `dist/eslint-plugin/index.js`, which `scripts/build.js` stages into the
 * package (worker bundle + nested `@rslint/native-<tuple>` platform pkg). The bare
 * specifier only resolves in dev (workspace `node_modules`); the packaged
 * vsix ships no `@rslint/core`, so it must be the relative sibling. The host
 * is ESM (uses `import.meta.url` to spawn its sibling `lint-worker.js`) while
 * this extension is bundled to CJS — a static `require` of an ESM module
 * would fail, a dynamic `import()` loads it correctly. esbuild keeps the
 * specifier external (see scripts/build.js) so it is emitted verbatim.
 */

import { window } from 'vscode';
import type { CancellationToken } from 'vscode';
import type { Logger } from './logger';
import type {
  ConfigDescriptor,
  PluginLintHost,
  EslintPluginLintRequest,
  EslintPluginLintResult,
} from '@rslint/core/eslint-plugin';

/** Subset of the `@rslint/core/eslint-plugin` module surface we depend on. */
interface EslintPluginModule {
  createPluginLintHost(
    configs: ConfigDescriptor[],
    onLog?: (rec: { level: string; source: string; text: string }) => void,
  ): Promise<PluginLintHost>;
}

let modPromise: Promise<EslintPluginModule> | undefined;

/**
 * Load the ESM host entry once. The specifier is RELATIVE
 * (`./eslint-plugin/index.js`): esbuild keeps it external so it survives
 * verbatim into `dist/main.js`, where it resolves to the sibling
 * `dist/eslint-plugin/index.js` staged by `scripts/build.js` — the same path
 * in dev and in the packaged vsix, so the dev test exercises the packaged
 * mechanism rather than a dev-only one.
 */
async function loadModule(): Promise<EslintPluginModule> {
  // The host entry exists only in the built `dist/` (staged by build.js), never
  // under `src/`, so it is intentionally unresolvable at compile time; its
  // module shape is supplied by `modPromise`'s declared type (no cast needed —
  // the suppressed import is `any`, assignable straight into the typed slot).
  // @ts-expect-error -- runtime-only path, resolved relative to dist/main.js
  modPromise ??= import('./eslint-plugin/index.js').catch((err: unknown) => {
    // Don't cache a rejected load: a transient failure (e.g. a mid-rebuild
    // dist in the dev watch window) must stay retryable on the next ensure(),
    // which is the recovery ensure()'s catch documents.
    modPromise = undefined;
    throw err;
  });
  return modPromise;
}

/**
 * Latches the one-shot "host failed to load" warning at MODULE scope (not
 * per-instance) so a persistent failure (e.g. a broken vsix that didn't ship
 * the worker payload) surfaces once per session — not once per workspace folder
 * in a multi-root window, where each folder owns its own PluginLintPool.
 */
let warnedOnce = false;

export class PluginLintPool {
  private readonly logger: Logger;
  /** The live host, once initialized. */
  private host: PluginLintHost | undefined;
  /** Fingerprint of the inputs `host` was built from. */
  private fingerprint: string | undefined;
  /**
   * Serializes every lifecycle op (ensure/dispose). Each op chains onto the
   * previous one's settlement, so at most one host is ever being built or torn
   * down at a time — concurrent config reloads can't race two worker pools
   * into existence or leak one. `lint` awaits this too, so it always runs
   * against the settled current host, never a half-torn-down one.
   */
  private opChain: Promise<void> = Promise.resolve();
  private disposed = false;

  constructor(logger: Logger) {
    this.logger = logger;
  }

  /** Append `op` to the serialized lifecycle chain and await its turn. */
  private async enqueue(op: () => Promise<void>): Promise<void> {
    const run = this.opChain.then(op, op);
    // Keep the chain alive even if `op` throws — swallow on the chain copy so a
    // single failed op doesn't poison every subsequent one. Callers awaiting
    // the returned promise still observe the rejection.
    this.opChain = run.catch(() => undefined);
    return run;
  }

  /**
   * Ensure a host exists that matches `descriptors`/`fingerprint`, rebuilding
   * if the fingerprint changed (config file or lockfile mtime/size moved).
   *
   * Empty `descriptors` still builds a host — it spawns no workers (zero
   * overhead) but can still answer a `lint` request with empty results, which
   * keeps the request handler total even before any plugin config is seen.
   */
  async ensure(
    descriptors: ConfigDescriptor[],
    fingerprint: string,
  ): Promise<void> {
    return this.enqueue(async () => {
      if (this.disposed) return;
      // Already serving exactly this fingerprint — reuse the live host.
      if (this.host && this.fingerprint === fingerprint) return;

      // Tear down the previous host before standing up the new one so we never
      // hold two worker pools at once.
      const old = this.host;
      this.host = undefined;
      this.fingerprint = undefined;
      if (old) {
        await old.shutdown().catch((err: unknown) => {
          this.logger.error('Error shutting down previous plugin host', err);
        });
      }
      if (this.disposed) return; // disposed mid-teardown — don't spin a new pool

      try {
        const mod = await loadModule();
        const host = await mod.createPluginLintHost(descriptors, (rec) => {
          const text = `[rslint:plugin] ${rec.text}`;
          if (rec.level === 'error') this.logger.error(text);
          else this.logger.debug(text);
        });
        if (this.disposed) {
          // Disposed while initializing — shut the fresh pool back down.
          await host.shutdown().catch(() => undefined);
          return;
        }
        this.host = host;
        this.fingerprint = fingerprint;
      } catch (err: unknown) {
        // Init failed: either the host module couldn't be loaded (a packaging
        // regression — the vsix didn't ship `dist/eslint-plugin/` or its
        // native `.node`), or a referenced plugin failed to import. Leave host
        // unset so we serve empty until the next config change retries, rather
        // than wedging diagnostics.
        this.host = undefined;
        this.fingerprint = undefined;
        this.logger.error('Failed to initialize ESLint-plugin host', err);
        // Make the failure visible — but ONLY when a config actually mounted
        // plugins (an empty-descriptor host builds no worker and failing is
        // not a user-facing problem), and only once per session so a
        // persistent failure doesn't re-warn on every reload.
        if (descriptors.length > 0 && !warnedOnce) {
          warnedOnce = true;
          void window.showWarningMessage(
            'Rslint: failed to load the ESLint-plugin host; rules mounted via a config’s `plugins` will report no diagnostics. See the Rslint output channel for details.',
          );
        }
      }
    });
  }

  /**
   * Answer one reverse `rslint/pluginLint` request. If no host is up
   * (init pending / failed, or never configured) return empty results so Go's
   * plugin-rule diagnostics simply come back empty rather than erroring.
   */
  async lint(
    req: EslintPluginLintRequest,
    token?: CancellationToken,
  ): Promise<EslintPluginLintResult> {
    // Wait for any in-flight lifecycle op so we lint against the settled host,
    // not one mid-rebuild.
    await this.opChain;
    const host = this.host;
    if (!host) return { results: [] };
    // Bridge the LSP CancellationToken → AbortSignal for the core host, so a
    // superseding keystroke / close (Go sends $/cancelRequest) stops the worker
    // instead of letting it run to completion.
    let signal: AbortSignal | undefined;
    if (token) {
      const ac = new AbortController();
      if (token.isCancellationRequested) ac.abort();
      else
        token.onCancellationRequested(() => {
          ac.abort();
        });
      signal = ac.signal;
    }
    return host.lint(req, signal);
  }

  /** Shut down the worker pool. Idempotent. */
  async dispose(): Promise<void> {
    return this.enqueue(async () => {
      this.disposed = true;
      const host = this.host;
      this.host = undefined;
      this.fingerprint = undefined;
      if (host) {
        await host.shutdown().catch((err: unknown) => {
          this.logger.error('Error disposing ESLint-plugin host', err);
        });
      }
    });
  }
}
