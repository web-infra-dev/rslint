import path from 'node:path';
import fs from 'node:fs';
import { execFileSync } from 'node:child_process';
import { parseArgs as nodeParseArgs } from 'node:util';
import {
  loadConfigFile,
  normalizeConfig,
  findJSConfigUp,
  findJSConfigsInDir,
} from './config-loader.js';

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

function parseArgs(argv: string[]) {
  const { values, tokens } = nodeParseArgs({
    args: argv,
    strict: false,
    tokens: true,
    options: {
      config: { type: 'string' },
      init: { type: 'boolean' },
    },
  });

  // Collect args that are not --config or --init for pass-through to Go.
  // positionals collects only true file/dir arguments, excluding values
  // of unknown flags (e.g. "jsonline" after "--format").
  const rest: string[] = [];
  const positionals: string[] = [];
  let skipNextPositional = false;
  for (const token of tokens) {
    if (token.kind === 'option') {
      if (token.name === 'config' || token.name === 'init') continue;
      rest.push(token.rawName);
      if (token.value != null) {
        rest.push(token.value);
      } else {
        // Unknown option without inline value — the next token will be
        // parsed as a positional but is actually this option's value.
        skipNextPositional = true;
      }
    } else if (token.kind === 'option-terminator') {
      rest.push('--');
    } else if (token.kind === 'positional') {
      rest.push(token.value);
      if (skipNextPositional) {
        skipNextPositional = false;
      } else {
        positionals.push(token.value);
      }
    }
  }

  return {
    raw: argv,
    config: (values.config as string) ?? null,
    init: (values.init as boolean) ?? false,
    rest,
    positionals,
  };
}

export function isJSConfigFile(filePath: string): boolean {
  return /\.(ts|mts|js|mjs)$/.test(filePath);
}

/**
 * Classify positional args into files and directories.
 */
export function classifyArgs(
  positionals: string[],
  cwd: string,
): { files: string[]; dirs: string[] } {
  const files: string[] = [];
  const dirs: string[] = [];
  for (const arg of positionals) {
    const resolved = path.resolve(cwd, arg);
    try {
      // Resolve symlinks so paths are consistent with process.cwd() and
      // TypeScript's SourceFile.FileName() which both return real paths.
      const real = fs.realpathSync(resolved);
      if (fs.statSync(real).isDirectory()) {
        dirs.push(real);
      } else {
        files.push(real);
      }
    } catch {
      // Non-existent path: treat as file (Go will handle the error)
      files.push(resolved);
    }
  }
  return { files, dirs };
}

/**
 * Discover JS/TS config files for the given targets.
 *
 * For file arguments, config is searched upward from each file's directory,
 * so different files can find different configs (monorepo multi-config).
 *
 * For no-args and directory arguments, config is searched upward from the
 * starting point AND nested configs within the scope are scanned. This
 * ensures sub-package configs in a monorepo are discovered when linting
 * from the root.
 */
export function discoverConfigs(
  files: string[],
  dirs: string[],
  cwd: string,
  explicitConfig: string | null,
): Map<string, string> {
  // Map: configPath -> configDirectory
  const configs = new Map<string, string>();

  const addConfig = (configPath: string): void => {
    if (!configs.has(configPath)) {
      configs.set(configPath, path.dirname(configPath));
    }
  };

  if (explicitConfig) {
    const resolved = path.resolve(cwd, explicitConfig);
    addConfig(resolved);
    return configs;
  }

  // Collect unique start directories for upward config search
  const startDirs = new Set<string>();
  // Collect directories to scan for nested configs
  const scanDirs: string[] = [];

  if (files.length === 0 && dirs.length === 0) {
    startDirs.add(cwd);
    scanDirs.push(cwd);
  }

  // Deduplicate file directories before searching
  for (const file of files) {
    startDirs.add(path.dirname(file));
  }

  for (const dir of dirs) {
    startDirs.add(dir);
    scanDirs.push(dir);
  }

  // Upward traversal: find nearest config for each start directory
  for (const startDir of startDirs) {
    const configPath = findJSConfigUp(startDir);
    if (configPath) {
      addConfig(configPath);
    }
  }

  // Scan for nested configs within the target scope (no-args and dir-args)
  for (const dir of scanDirs) {
    for (const configPath of findJSConfigsInDir(dir)) {
      addConfig(configPath);
    }
  }

  return configs;
}

/**
 * Load multiple JS configs and pipe to Go binary via stdin.
 */
async function runWithJSConfigs(
  binPath: string,
  configs: Map<string, string>,
  restArgs: string[],
  cwd: string,
): Promise<number> {
  const configEntries: { configDirectory: string; entries: unknown[] }[] = [];

  for (const [configPath, configDir] of configs) {
    let rawConfig: unknown;
    try {
      rawConfig = await loadConfigFile(configPath);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      process.stderr.write(
        `Error: failed to load config ${configPath}: ${msg}\n`,
      );
      return 1;
    }

    let entries: unknown[];
    try {
      entries = normalizeConfig(rawConfig);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      process.stderr.write(`Error: invalid config in ${configPath}: ${msg}\n`);
      return 1;
    }

    configEntries.push({ configDirectory: configDir, entries });
  }

  const payload = JSON.stringify({ configs: configEntries });

  try {
    execFileSync(binPath, ['--config-stdin', ...restArgs], {
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

export async function run(binPath: string, argv: string[]): Promise<number> {
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

  // Classify positional arguments into files and directories
  const { files, dirs } = classifyArgs(args.positionals, cwd);

  // Discover JS/TS configs
  const configs = discoverConfigs(files, dirs, cwd, args.config);

  // Check if any discovered config is a JS/TS config
  const jsConfigs = new Map<string, string>();
  for (const [configPath, configDir] of configs) {
    if (isJSConfigFile(configPath)) {
      jsConfigs.set(configPath, configDir);
    }
  }

  if (jsConfigs.size > 0) {
    return runWithJSConfigs(binPath, jsConfigs, args.rest, cwd);
  }

  // Fall back to Go binary (handles JSON config + deprecation warning)
  return execBinary(binPath, args.raw);
}
