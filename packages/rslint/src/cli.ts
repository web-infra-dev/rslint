import path from 'node:path';
import fs from 'node:fs';
import { execFileSync } from 'node:child_process';
import { loadConfigFile, normalizeConfig } from './config-loader.js';
import { parseArgs, classifyArgs, isJSConfigFile } from './utils/args.js';
import {
  discoverConfigs,
  filterConfigsByParentIgnores,
} from './utils/config-discovery.js';

/**
 * Pass-through execution of the Go binary with stdio inherited.
 */
function execBinary(binPath: string, argv: string[]): number {
  try {
    execFileSync(binPath, argv, { stdio: 'inherit' });
    return 0;
  } catch (error: unknown) {
    if (isExecError(error)) return error.status;
    process.stderr.write(`Failed to execute ${binPath}: ${String(error)}\n`);
    return 1;
  }
}

function isExecError(
  error: unknown,
): error is Record<string, unknown> & { status: number } {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    typeof error.status === 'number'
  );
}

/**
 * Load multiple JS configs and pipe to Go binary via stdin.
 * Tolerates individual config load failures — skips broken configs with a
 * warning and continues with the remaining configs.
 */
async function runWithJSConfigs(
  binPath: string,
  configs: Map<string, string>,
  goArgs: string[],
  cwd: string,
): Promise<number> {
  const configEntries: { configDirectory: string; entries: unknown[] }[] = [];
  const isSingleConfig = configs.size === 1;

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

    let entries: unknown[];
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
  }

  // All configs failed to load — fall back to Go binary (JSON config path)
  if (configEntries.length === 0) {
    return execBinary(binPath, goArgs);
  }

  // Filter out nested configs whose directory is covered by a parent config's
  // global ignores. This aligns with ESLint v10 behavior: when walking the
  // directory tree, global ignores prevent entering directories, so nested
  // configs in ignored directories are never discovered.
  const filteredEntries = filterConfigsByParentIgnores(configEntries);

  const payload = JSON.stringify({ configs: filteredEntries });

  try {
    execFileSync(binPath, ['--config-stdin', ...goArgs], {
      input: payload,
      stdio: ['pipe', 'inherit', 'inherit'],
      cwd,
    });
    return 0;
  } catch (error: unknown) {
    if (isExecError(error)) return error.status;
    process.stderr.write(`Failed to execute ${binPath}: ${String(error)}\n`);
    return 1;
  }
}

export async function run(
  binPath: string,
  argv: string[],
  startTime: number,
): Promise<number> {
  const cwd = process.cwd();
  const args = parseArgs(argv);

  // --init: pass through to Go
  if (args.init) {
    return execBinary(binPath, ['--init']);
  }

  // Validate explicit --config flag
  if (args.config) {
    const configPath = path.resolve(cwd, args.config);
    if (!fs.existsSync(configPath)) {
      process.stderr.write(`Error: config file not found: ${configPath}\n`);
      return 1;
    }
  }

  // Build Go args: rest (user flags, stripped of --config/--init/--start-time
  // by parseArgs) + the real start time from the Node.js entry point.
  const goArgs = [...args.rest, `--start-time=${startTime}`];

  // Classify positional arguments into files and directories
  const { files, dirs } = classifyArgs(args.positionals, cwd);

  // Discover JS/TS configs
  const configs = discoverConfigs(files, dirs, cwd, args.config);

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
    return runWithJSConfigs(binPath, jsConfigs, goArgs, cwd);
  }

  // Fall back to Go binary (handles JSON config + deprecation warning)
  const jsonGoArgs = args.config
    ? ['--config', args.config, ...goArgs]
    : goArgs;
  return execBinary(binPath, jsonGoArgs);
}
