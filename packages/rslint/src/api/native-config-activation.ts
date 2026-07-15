import {
  ConfigModuleHost,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
} from '../config/config-loader.js';

export interface PluginLintHost {
  lint(request: unknown): Promise<unknown>;
  shutdown(): Promise<void>;
}

type CreatePluginLintHost = (
  configs: Array<{ configPath: string; configDirectory: string }>,
  onLog?: (record: { level: string; source: string; text: string }) => void,
) => Promise<PluginLintHost>;

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isPluginHostFactoryModule(
  value: unknown,
): value is { createPluginLintHost: CreatePluginLintHost } {
  return isRecord(value) && typeof value.createPluginLintHost === 'function';
}

function isPluginLintHost(value: unknown): value is PluginLintHost {
  return (
    isRecord(value) &&
    typeof value.lint === 'function' &&
    typeof value.shutdown === 'function'
  );
}

let pluginHostFactoryPromise: Promise<CreatePluginLintHost> | undefined;

export async function loadPluginHostFactory(): Promise<CreatePluginLintHost> {
  pluginHostFactoryPromise ??= (async () => {
    // A package self-reference resolves to src under the test condition and to
    // dist/eslint-plugin in published builds. Keep it runtime-only: the library
    // declaration build deliberately excludes the worker implementation.
    const pluginEntry: string = '@rslint/core/eslint-plugin';
    const module: unknown = await import(/* webpackIgnore: true */ pluginEntry);
    if (!isPluginHostFactoryModule(module)) {
      throw new Error(
        'rslint ESLint-plugin entry does not export createPluginLintHost',
      );
    }
    return module.createPluginLintHost;
  })();
  const factory = await pluginHostFactoryPromise;
  return factory;
}

/** @internal Resource registry shared by activation and API close tests. */
export class PluginHostLifecycle {
  readonly #pendingBuilds = new Set<Promise<unknown>>();
  readonly #staged = new Set<PluginLintHost>();
  readonly #active = new Set<PluginLintHost>();
  readonly #shutdowns = new WeakMap<PluginLintHost, Promise<void>>();

  async trackBuild<T>(build: Promise<T>): Promise<T> {
    this.#pendingBuilds.add(build);
    void build.then(
      () => this.#pendingBuilds.delete(build),
      () => this.#pendingBuilds.delete(build),
    );
    const result = await build;
    return result;
  }

  stage(host: PluginLintHost): void {
    this.#staged.add(host);
  }

  publish(host: PluginLintHost): void {
    this.#staged.delete(host);
    this.#active.add(host);
  }

  async shutdown(host: PluginLintHost): Promise<void> {
    this.#staged.delete(host);
    this.#active.delete(host);
    let shutdown = this.#shutdowns.get(host);
    if (!shutdown) {
      shutdown = host.shutdown();
      this.#shutdowns.set(host, shutdown);
    }
    await shutdown;
  }

  async shutdownAll(): Promise<void> {
    while (this.#pendingBuilds.size > 0) {
      await Promise.allSettled([...this.#pendingBuilds]);
    }
    const shutdowns: Promise<void>[] = [];
    for (const host of new Set([...this.#staged, ...this.#active])) {
      shutdowns.push(this.shutdown(host));
    }
    await Promise.allSettled(shutdowns);
  }
}

/**
 * Stage one native-discovery activation without exposing a worker imported
 * from config bytes that differ from Go's normalized entries.
 *
 * @internal Exported from this internal module for lifecycle regression tests;
 * the module is intentionally not exposed through the package exports.
 */
export async function stageNativeConfigActivation(
  configHost: ConfigModuleHost,
  request: ActivateConfigsRequest,
  getPluginHostFactory: () => Promise<CreatePluginLintHost>,
  onLog: (record: { level: string; source: string; text: string }) => void,
  isClosing: () => boolean,
  lifecycle?: PluginHostLifecycle,
): Promise<{
  activation: ActivateConfigsResponse;
  pluginHost: PluginLintHost | null;
}> {
  let pluginHost: PluginLintHost | null = null;
  try {
    const activation = await configHost.activateConfigs(
      request,
      undefined,
      async (candidate) => {
        if (candidate.pluginConfigs.length === 0) return;
        if (isClosing()) throw new Error('rslint service is closing');
        const createPluginLintHost = await getPluginHostFactory();
        if (isClosing()) throw new Error('rslint service is closing');
        const build = (async () => {
          const host = await createPluginLintHost(
            candidate.pluginConfigs,
            onLog,
          );
          lifecycle?.stage(host);
          return host;
        })();
        pluginHost = await (lifecycle?.trackBuild(build) ?? build);
        if (isClosing()) throw new Error('rslint service is closing');
      },
    );
    // close() can start while the post-prepare fingerprint read is pending.
    if (isClosing()) throw new Error('rslint service is closing');
    return { activation, pluginHost };
  } catch (error) {
    try {
      const createdHost: unknown = pluginHost;
      if (isPluginLintHost(createdHost)) {
        await (lifecycle?.shutdown(createdHost) ?? createdHost.shutdown());
      }
    } catch {
      // Preserve the activation error: the source mismatch is what Go must
      // receive, while the host has still been asked to terminate.
    }
    throw error;
  }
}
