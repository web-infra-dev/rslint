import path from 'node:path';
import fs from 'node:fs';
import {
  loadConfigFile,
  normalizeConfig,
  collectPluginMeta,
} from '../config/config-loader.js';
import { parseArgs, classifyArgs, isJSConfigFile } from '../utils/args.js';
import {
  coalesceCaseAliasedConfigs,
  discoverConfigs,
  filterConfigsByParentIgnores,
  findNativeCaseAliasConfigPath,
  findJSConfigsForDirectories,
  findJSConfigsForFiles,
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
 * Broken candidates are skipped when another selected config remains usable.
 * For an explicit file, ancestor configs are loaded lazily only after its
 * nearer candidate fails.
 */
async function runWithJSConfigs(
  binPath: string,
  configs: Map<string, string>,
  goArgs: string[],
  cwd: string,
  singleThreaded: boolean,
  explicitFileTargetsByConfigPath = new Map<string, string[]>(),
  explicitDirectoryTargetsByConfigPath = new Map<string, string[]>(),
): Promise<number> {
  ({
    configs,
    explicitFileTargetsByConfigPath,
    explicitDirectoryTargetsByConfigPath,
  } = coalesceCaseAliasedConfigs(
    configs,
    explicitFileTargetsByConfigPath,
    explicitDirectoryTargetsByConfigPath,
  ));
  const configEntries: ConfigEntry[] = [];
  const dirToPath = new Map<string, string>();
  const initialConfigCount = configs.size;
  const pendingConfigs = [...configs];
  const knownConfigPaths = new Set(
    pendingConfigs.map(([configPath]) => path.normalize(configPath)),
  );
  const loadedConfigPaths = new Set<string>();
  const failedConfigPaths = new Set<string>();
  const failures: Array<{
    configPath: string;
    kind: 'load' | 'invalid';
    message: string;
  }> = [];

  const addTargets = (
    targetMap: Map<string, string[]>,
    configPath: string,
    targets: readonly string[],
  ): void => {
    if (targets.length === 0) return;
    const existing = targetMap.get(configPath) ?? [];
    targetMap.set(configPath, [...new Set([...existing, ...targets])]);
  };

  const scheduleAncestorFallback = (
    failedConfigDirectory: string,
    files: readonly string[],
    directories: readonly string[],
  ): void => {
    if (files.length === 0 && directories.length === 0) return;
    const ancestorPath = findJSConfigUp(path.dirname(failedConfigDirectory));
    if (!ancestorPath) return;

    const normalizedPath = path.normalize(ancestorPath);
    const configDirectory = path.dirname(normalizedPath);
    const selectedPath =
      findNativeCaseAliasConfigPath(normalizedPath, configDirectory, configs) ??
      normalizedPath;
    addTargets(explicitFileTargetsByConfigPath, selectedPath, files);
    addTargets(explicitDirectoryTargetsByConfigPath, selectedPath, directories);
    if (loadedConfigPaths.has(selectedPath)) return;
    if (failedConfigPaths.has(selectedPath)) {
      scheduleAncestorFallback(
        configs.get(selectedPath) ?? path.dirname(selectedPath),
        files,
        directories,
      );
      return;
    }
    if (!knownConfigPaths.has(selectedPath)) {
      configs.set(selectedPath, configDirectory);
      knownConfigPaths.add(selectedPath);
      pendingConfigs.push([selectedPath, configDirectory]);
    }
  };
  for (let index = 0; index < pendingConfigs.length; index++) {
    const [configPath, configDir] = pendingConfigs[index];
    const normalizedConfigPath = path.normalize(configPath);
    let rawConfig: unknown;
    try {
      rawConfig = await loadConfigFile(normalizedConfigPath);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      failures.push({
        configPath: normalizedConfigPath,
        kind: 'load',
        message: msg,
      });
      failedConfigPaths.add(normalizedConfigPath);
      scheduleAncestorFallback(
        configDir,
        explicitFileTargetsByConfigPath.get(normalizedConfigPath) ?? [],
        explicitDirectoryTargetsByConfigPath.get(normalizedConfigPath) ?? [],
      );
      continue;
    }

    let entries: ConfigEntry['entries'];
    try {
      entries = normalizeConfig(rawConfig);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      failures.push({
        configPath: normalizedConfigPath,
        kind: 'invalid',
        message: msg,
      });
      failedConfigPaths.add(normalizedConfigPath);
      scheduleAncestorFallback(
        configDir,
        explicitFileTargetsByConfigPath.get(normalizedConfigPath) ?? [],
        explicitDirectoryTargetsByConfigPath.get(normalizedConfigPath) ?? [],
      );
      continue;
    }

    configEntries.push({ configDirectory: configDir, entries });
    dirToPath.set(configDir, normalizedConfigPath);
    loadedConfigPaths.add(normalizedConfigPath);
  }

  // Lazy import — keeps engine.ts (and its node:child_process dependency) out
  // of the browser/wasm bundle, loaded only on the real CLI path.
  const { runEngine } = await import('./engine.js');

  // JS configs were discovered, but none survived loading/normalization. Do
  // not fall back to JSON/default config; that would lint with unrelated rules.
  if (configEntries.length === 0) {
    if (initialConfigCount === 1 && failures.length > 0) {
      const nearest = failures[0];
      const description =
        nearest.kind === 'load' ? 'failed to load config' : 'invalid config in';
      process.stderr.write(
        `Error: ${description} ${nearest.configPath}: ${nearest.message}\n`,
      );
    } else {
      for (const failure of failures) {
        process.stderr.write(
          `Warning: skipping config ${failure.configPath}: ${failure.message}\n`,
        );
      }
    }
    return 1;
  }

  for (const failure of failures) {
    process.stderr.write(
      `Warning: skipping config ${failure.configPath}: ${failure.message}\n`,
    );
  }

  // Exclude loaded nested config candidates hidden behind a parent config's
  // global ignores from the directory target set. Explicit file args are
  // different: ESLint resolves the nearest config for each file even if
  // directory traversal would not enter that subtree. Keep those explicit-file
  // configs, but scope them to the explicit files so they do not reopen full
  // directory discovery.
  const explicitFileConfigDirs = new Set(
    [...explicitFileTargetsByConfigPath.keys()].map((configPath) =>
      path.dirname(configPath),
    ),
  );
  const directoryFilteredEntries = filterConfigsByParentIgnores(configEntries);
  const directoryFilteredDirs = new Set(
    directoryFilteredEntries.map((entry) => entry.configDirectory),
  );
  const filteredEntries = filterConfigsByParentIgnores(
    configEntries,
    explicitFileConfigDirs,
  );
  const wireConfigEntries = filteredEntries.map((ce) => {
    const configPath = dirToPath.get(ce.configDirectory) ?? '';
    const explicitTargets = explicitFileTargetsByConfigPath.get(configPath);
    const explicitOnly = !directoryFilteredDirs.has(ce.configDirectory);
    return {
      configPath,
      configDirectory: ce.configDirectory,
      entries: ce.entries,
      targetFiles:
        explicitTargets && explicitTargets.length > 0
          ? explicitTargets
          : undefined,
      explicitOnly: explicitOnly || undefined,
    };
  });
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
  const fileConfigs = findJSConfigsForFiles(files);
  const directoryConfigs = findJSConfigsForDirectories(dirs);
  const configs = await discoverConfigs(
    files,
    dirs,
    cwd,
    args.config,
    fileConfigs,
    directoryConfigs,
  );
  const explicitFileTargetsByConfigPath = new Map<string, string[]>();
  const explicitDirectoryTargetsByConfigPath = new Map<string, string[]>();
  if (!args.config) {
    for (const file of files) {
      const nearestConfig = fileConfigs.get(path.normalize(file));
      if (nearestConfig) {
        const explicitTargets =
          explicitFileTargetsByConfigPath.get(nearestConfig);
        if (explicitTargets) {
          explicitTargets.push(file);
        } else {
          explicitFileTargetsByConfigPath.set(nearestConfig, [file]);
        }
      }
    }
    for (const dir of dirs) {
      const nearestConfig = directoryConfigs.get(path.normalize(dir));
      if (!nearestConfig) continue;
      const explicitTargets =
        explicitDirectoryTargetsByConfigPath.get(nearestConfig);
      if (explicitTargets) {
        explicitTargets.push(dir);
      } else {
        explicitDirectoryTargetsByConfigPath.set(nearestConfig, [dir]);
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
      explicitFileTargetsByConfigPath,
      explicitDirectoryTargetsByConfigPath,
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
