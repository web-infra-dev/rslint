import { readFile } from 'node:fs/promises';
import path from 'node:path';
import {
  loadConfigFileFresh,
  normalizeConfig,
} from '../config/config-file-loader.js';
import { fingerprintConfigSource } from '../config/config-source.js';

const CONFIG_EXTENSIONS = new Set([
  '.js',
  '.mjs',
  '.cjs',
  '.ts',
  '.mts',
  '.cts',
]);

export interface RustConfigOptions {
  cwd?: string;
  /** An explicit module, or true to use only inline configuration. */
  overrideConfigFile?: string | true | null;
  /** Appended after the selected module, matching the public Rslint API. */
  overrideConfig?: unknown;
}

/** Can be spread directly into lintRustProjectsSync's options. */
export interface ResolvedRustConfig {
  cwd: string;
  config: Record<string, unknown>[];
  configDirectory: string;
  configArrayBase: string;
  /** Authored origin, without adding basePath's additional scope restriction. */
  configEntryDirectories: string[];
}

function record(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function normalizeOverride(value: unknown): Record<string, unknown>[] {
  if (value == null) return [];
  // Keep the public API's one-level preset expansion and error positions.
  const entries = (Array.isArray(value) ? value : [value]).flat();
  for (const [index, entry] of entries.entries()) {
    if (
      record(entry) &&
      entry.plugins !== null &&
      typeof entry.plugins === 'object' &&
      !Array.isArray(entry.plugins)
    ) {
      throw new Error(
        `[rslint] overrideConfig entry at index ${index} uses object-form "plugins". ` +
          'Community ESLint plugins in overrideConfig are not supported because ' +
          'the plugin worker cannot re-import an in-memory plugin object. Move ' +
          'the plugin declaration to rslint.config.js (or .mjs/.cjs/.ts/.mts/.cts), or use ' +
          'array-form built-in plugins.',
      );
    }
  }
  return normalizeConfig(entries);
}

function snapshot(
  config: Record<string, unknown>[],
): Record<string, unknown>[] {
  for (const entry of config) {
    if ('eslintPlugins' in entry) {
      throw new Error(
        'Rust config loading does not yet support JavaScript plugin objects',
      );
    }
  }
  // The previous host serialized normalized entries before publishing them.
  // Retain that boundary so mutating an inline option cannot alter this request.
  const serialized: unknown = JSON.parse(JSON.stringify(config));
  return normalizeConfig(serialized);
}

/**
 * Evaluate an explicitly selected JS/TS module afresh for each lint request.
 * This invokes neither a native process nor a config's exported function. Async
 * default exports are awaited by the shared loader, just as in ConfigModuleHost.
 * Imported dependencies keep the shared loader's normal module-cache semantics.
 *
 * Module evaluation observes the process cwd, as the existing public API does;
 * changing process.cwd() here would corrupt overlapping requests. In particular,
 * a Rstack invocation must start in its project directory before importing its
 * generated config module.
 *
 * Automatic hierarchical discovery is not implemented by this adapter. Reject
 * that mode instead of choosing one owner or treating a tree that has not been
 * inspected as the API's valid no-config syntax-only state.
 */
export async function resolveRustConfig(
  options: RustConfigOptions = {},
): Promise<ResolvedRustConfig> {
  if (!record(options))
    throw new Error('Rust config options must be an object');
  if (options.cwd !== undefined && typeof options.cwd !== 'string')
    throw new Error('Rust config cwd must be a string');
  const cwd = options.cwd ? path.resolve(options.cwd) : process.cwd();
  const override = normalizeOverride(options.overrideConfig);
  const selected = options.overrideConfigFile;
  if (selected === true) {
    return {
      cwd,
      config: snapshot(override),
      configDirectory: cwd,
      configArrayBase: cwd,
      configEntryDirectories: override.map(() => cwd),
    };
  }
  if (selected == null) {
    throw new Error(
      'Rust config loading does not yet support automatic configuration discovery; ' +
        'use --config, overrideConfigFile with a JS/TS module, or overrideConfigFile: true for inline-only configuration',
    );
  }
  if (typeof selected !== 'string')
    throw new Error(
      'overrideConfigFile must be a string, true, null, or undefined',
    );
  const configPath = path.resolve(cwd, selected);
  if (!CONFIG_EXTENSIONS.has(path.extname(configPath))) {
    throw new Error(
      `discover config catalog: explicit config path must name a JS/TS config module: ${JSON.stringify(configPath)}`,
    );
  }
  let entries: Record<string, unknown>[];
  let failureKind = 'load';
  try {
    const before = fingerprintConfigSource(await readFile(configPath));
    const loaded = await loadConfigFileFresh(configPath);
    failureKind = 'invalid';
    entries = snapshot(normalizeConfig(loaded));
    if (before !== fingerprintConfigSource(await readFile(configPath)))
      throw new Error(
        `config changed while it was being loaded: ${configPath}`,
      );
  } catch (error) {
    const reason = error instanceof Error ? error.message : String(error);
    if (record(error) && typeof error.code === 'string' && error.code !== '')
      failureKind = error.code;
    const context =
      failureKind === 'invalid' ? 'invalid config in' : 'failed to load config';
    throw new Error(
      `discover config catalog: all discovered JavaScript configs failed to load: ${context} ${configPath}: ${reason}`,
      { cause: error },
    );
  }
  const configDirectory = path.dirname(configPath);
  return {
    cwd,
    config: snapshot([...entries, ...override]),
    configDirectory,
    // Explicit selection anchors authored basePath at cwd, even for a module
    // in node_modules or outside cwd. Entries without basePath keep their origin.
    configArrayBase: cwd,
    configEntryDirectories: [
      ...entries.map(() => configDirectory),
      ...override.map(() => cwd),
    ],
  };
}
