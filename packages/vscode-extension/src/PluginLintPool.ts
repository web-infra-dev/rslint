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
 * `@rslint/core/eslint-plugin` is loaded via a dynamic `import()` (not a
 * static import): it is ESM (uses `import.meta.url` to resolve its sibling
 * `lint-worker.js`) while this extension is bundled to CJS, and esbuild
 * keeps it external (see scripts/build.js) so the worker file stays next to
 * `index.js` inside the installed package. A static `require` of an ESM
 * module would fail; a dynamic `import()` loads it correctly.
 */

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
 * Load the ESM `@rslint/core/eslint-plugin` entry once. esbuild preserves
 * this dynamic import verbatim in its CJS output (verified) because the
 * specifier is external — so it resolves to the installed package's ESM
 * `dist/eslint-plugin/index.js` at runtime.
 */
async function loadModule(): Promise<EslintPluginModule> {
  modPromise ??=
    import('@rslint/core/eslint-plugin') as Promise<EslintPluginModule>;
  return modPromise;
}

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
        // Init failed (e.g. a referenced plugin failed to import). Leave host
        // unset so we serve empty until the next config change retries, rather
        // than wedging diagnostics.
        this.host = undefined;
        this.fingerprint = undefined;
        this.logger.error('Failed to initialize ESLint-plugin host', err);
      }
    });
  }

  /**
   * Answer one reverse `rslint/pluginLint` request. If no host is up
   * (init pending / failed, or never configured) return empty results so Go's
   * plugin-rule diagnostics simply come back empty rather than erroring.
   */
  async lint(req: EslintPluginLintRequest): Promise<EslintPluginLintResult> {
    // Wait for any in-flight lifecycle op so we lint against the settled host,
    // not one mid-rebuild.
    await this.opChain;
    const host = this.host;
    if (!host) return { results: [] };
    return host.lint(req);
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
