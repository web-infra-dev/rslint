import path from 'node:path';
import fs from 'node:fs';
import {
  loadConfigFile,
  normalizeConfig,
  collectPluginMeta,
} from '../config/config-loader.js';
import { parseArgs, classifyArgs, isJSConfigFile } from '../utils/args.js';
import {
  discoverConfigs,
  filterConfigsByParentIgnores,
  findJSConfigUp,
  type ConfigEntry,
} from '../utils/config-discovery.js';
import { resolveRslintBinary } from '../internal/resolve-binary.js';

export type RunCLIOptions = {
  /**
   * The command-line arguments to parse, matching the shape of Node.js `process.argv`
   * @default process.argv
   */
  argv?: string[];
};

/**
 * Load multiple JS/TS configs and run them through the Go binary over IPC.
 * Tolerates individual config load failures — skips broken configs with a
 * warning and continues with the remaining configs.
 */
async function runWithJSConfigs(
  binPath: string,
  configs: Map<string, string>,
  goArgs: string[],
  cwd: string,
  singleThreaded: boolean,
  protectedConfigFiles = new Map<string, string[]>(),
): Promise<number> {
  const configEntries: ConfigEntry[] = [];
  const dirToPath = new Map<string, string>();
  const isSingleConfig = configs.size === 1;
  const protectedConfigDirs = new Set(
    [...protectedConfigFiles.keys()].map((configPath) =>
      path.dirname(configPath),
    ),
  );

  for (const [configPath, configDir] of configs) {
    let rawConfig: unknown;
    try {
      rawConfig = await loadConfigFile(configPath);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      if (isSingleConfig) {
        process.stderr.write(
          `Error: failed to load config ${configPath}: ${msg}\n`,
        );
        return 1;
      }
      process.stderr.write(`Warning: skipping config ${configPath}: ${msg}\n`);
      continue;
    }

    let entries: ConfigEntry['entries'];
    try {
      entries = normalizeConfig(rawConfig);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      if (isSingleConfig) {
        process.stderr.write(
          `Error: invalid config in ${configPath}: ${msg}\n`,
        );
        return 1;
      }
      process.stderr.write(`Warning: skipping config ${configPath}: ${msg}\n`);
      continue;
    }

    configEntries.push({ configDirectory: configDir, entries });
    dirToPath.set(configDir, configPath);
  }

  // Lazy import — keeps engine.ts (and its node:child_process dependency) out
  // of the browser/wasm bundle, loaded only on the real CLI path.
  const { runEngine } = await import('./engine.js');

  // JS configs were discovered, but none survived loading/normalization. Do
  // not fall back to JSON/default config; that would lint with unrelated rules.
  if (configEntries.length === 0) {
    return 1;
  }

  // Filter out nested configs whose directory is covered by a parent config's
  // global ignores (ESLint v10 alignment). Then hand the configs to Go in the
  // IPC `init` payload (no more `--config-stdin` stdin pipe).
  const unprotectedEntries = filterConfigsByParentIgnores(configEntries);
  const unprotectedDirs = new Set(
    unprotectedEntries.map((entry) => entry.configDirectory),
  );
  const filteredEntries = filterConfigsByParentIgnores(
    configEntries,
    protectedConfigDirs,
  );
  const wireConfigEntries = filteredEntries.map((ce) => ({
    configPath: dirToPath.get(ce.configDirectory) ?? '',
    configDirectory: ce.configDirectory,
    entries: ce.entries,
    targetFiles: unprotectedDirs.has(ce.configDirectory)
      ? undefined
      : protectedConfigFiles.get(dirToPath.get(ce.configDirectory) ?? ''),
  }));
  const { eslintPluginEntries, pluginConfigs } =
    collectPluginMeta(wireConfigEntries);
  return runEngine({
    binPath,
    goArgs,
    configs: wireConfigEntries,
    cwd,
    eslintPluginEntries,
    pluginConfigs,
    runtime: { singleThreaded },
  });
}

export async function run(
  binPath: string,
  argv: string[],
  startTime: number,
): Promise<number> {
  const cwd = process.cwd();
  const args = parseArgs(argv);

  // --init: pass through to Go (no config payload — Go writes the default
  // config to disk and prints the "Created …" line, forwarded via `output`).
  if (args.init) {
    const { runEngine } = await import('./engine.js');
    return runEngine({ binPath, goArgs: ['--init'], configs: [], cwd });
  }

  // Validate explicit --config flag
  if (args.config) {
    const configPath = path.resolve(cwd, args.config);
    if (!fs.existsSync(configPath)) {
      process.stderr.write(`Error: config file not found: ${configPath}\n`);
      return 1;
    }
  }

  // Build Go args: start-time flag BEFORE positional args, because Go's
  // flag.Parse stops at the first positional argument.
  const goArgs = [`--start-time=${startTime}`, ...args.rest];

  // Classify positional arguments into files and directories
  const { files, dirs } = classifyArgs(args.positionals, cwd);

  // Discover JS/TS configs
  const configs = await discoverConfigs(files, dirs, cwd, args.config);
  const protectedConfigFiles = new Map<string, string[]>();
  if (!args.config) {
    for (const file of files) {
      const nearestConfig = findJSConfigUp(path.dirname(file));
      if (nearestConfig) {
        const protectedFiles = protectedConfigFiles.get(nearestConfig);
        if (protectedFiles) {
          protectedFiles.push(file);
        } else {
          protectedConfigFiles.set(nearestConfig, [file]);
        }
      }
    }
  }

  // Check if any discovered config is a JS/TS config.
  // NOTE: If any JS config is found (even in subdirectories), the entire flow
  // switches to the JS config path. A root JSON config (rslint.json) will be
  // bypassed in this case. This is a known limitation of mixing JSON and JS
  // config formats. JSON config is deprecated — projects should migrate to JS.
  const jsConfigs = new Map<string, string>();
  for (const [configPath, configDir] of configs) {
    if (isJSConfigFile(configPath)) {
      jsConfigs.set(configPath, configDir);
    }
  }

  if (jsConfigs.size > 0) {
    return runWithJSConfigs(
      binPath,
      jsConfigs,
      goArgs,
      cwd,
      args.singleThreaded,
      protectedConfigFiles,
    );
  }

  // Fall back to Go binary (handles JSON config + deprecation warning); no
  // config payload, so Go loads JSON config from disk itself.
  const jsonGoArgs = args.config
    ? ['--config', args.config, ...goArgs]
    : goArgs;
  const { runEngine } = await import('./engine.js');
  return runEngine({ binPath, goArgs: jsonGoArgs, configs: [], cwd });
}

export async function runCLI({
  argv = process.argv,
}: RunCLIOptions = {}): Promise<void> {
  const startTime = Date.now();
  const exitCode = await run(resolveRslintBinary(), argv.slice(2), startTime);
  // Let stdout/stderr flush naturally instead of terminating the process.
  process.exitCode = exitCode;
}
