import path from 'node:path';
import fs from 'node:fs';
import { execFileSync } from 'node:child_process';
import { parseArgs as nodeParseArgs } from 'node:util';
import {
  loadConfigFile,
  normalizeConfig,
  findJSConfig,
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

  // Collect args that are not --config or --init for pass-through to Go
  const rest: string[] = [];
  for (const token of tokens) {
    if (token.kind === 'option') {
      if (token.name === 'config' || token.name === 'init') continue;
      rest.push(token.rawName);
      if (token.value != null) rest.push(token.value);
    } else if (token.kind === 'option-terminator') {
      rest.push('--');
    } else if (token.kind === 'positional') {
      rest.push(token.value);
    }
  }

  return {
    raw: argv,
    config: (values.config as string) ?? null,
    init: (values.init as boolean) ?? false,
    rest,
  };
}

function isJSConfigFile(filePath: string): boolean {
  return /\.(ts|mts|js|mjs)$/.test(filePath);
}

/**
 * Load JS config, serialize to JSON, and pipe to Go binary via stdin.
 */
async function runWithJSConfig(
  binPath: string,
  configPath: string,
  restArgs: string[],
  cwd: string,
): Promise<number> {
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

  const configDir = path.dirname(path.resolve(cwd, configPath));
  const payload = JSON.stringify({ configDirectory: configDir, entries });

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

  // Determine config file
  let configPath: string | null = null;
  if (args.config) {
    configPath = path.resolve(cwd, args.config);
    if (!fs.existsSync(configPath)) {
      process.stderr.write(`Error: config file not found: ${configPath}\n`);
      return 1;
    }
  } else {
    configPath = findJSConfig(cwd);
  }

  // JS config file: load + stdin pipe
  if (configPath && isJSConfigFile(configPath)) {
    return runWithJSConfig(binPath, configPath, args.rest, cwd);
  }

  // Fall back to Go binary (handles JSON config + deprecation warning)
  return execBinary(binPath, args.raw);
}
