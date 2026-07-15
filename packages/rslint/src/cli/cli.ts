import path from 'node:path';
import {
  OUTPUT_FORMATS,
  parseArgs,
  isJSConfigFile,
  isOutputFormat,
} from '../utils/args.js';
import { resolveRslintBinary } from '../internal/resolve-binary.js';

export type RunCLIOptions = {
  /**
   * The command-line arguments to parse, matching the shape of Node.js `process.argv`
   * @default process.argv
   */
  argv?: string[];
};

export async function run(
  binPath: string,
  argv: string[],
  startTime: number,
): Promise<number> {
  const cwd = process.cwd();
  const args = parseArgs(argv);

  // --init: pass through to Go (no config payload — Go writes the default
  // config to disk and prints the "Created …" line, forwarded via `output`).
  // It intentionally takes priority over unrelated lint flags, matching the
  // existing fast-path contract.
  if (args.init) {
    const { runEngine } = await import('./engine.js');
    return runEngine({ binPath, goArgs: ['--init'], cwd });
  }

  // Reject an invalid stdout protocol before config discovery/evaluation.
  // Help retains its existing priority and is forwarded to Go, which owns the
  // usage text. Go validates format again after the IPC init payload is merged
  // so every CLI entry path receives the same single-error behavior.
  if (!args.help && args.format !== null && !isOutputFormat(args.format)) {
    process.stderr.write(
      `error: invalid output format ${JSON.stringify(args.format)} (expected ${OUTPUT_FORMATS.slice(0, -1).join(', ')}, or ${OUTPUT_FORMATS.at(-1)})\n`,
    );
    return 2;
  }

  // Build Go args: start-time flag BEFORE positional args, because Go's
  // flag.Parse stops at the first positional argument.
  const goArgs = [`--start-time=${startTime}`, ...args.rest];

  // JSON configuration remains native-Go. JS/TS configuration discovery is
  // initiated by Go after the IPC channel is live; Node only evaluates the
  // exact frontier candidates Go sends back through reverse RPC.
  const explicitConfigPath = args.config
    ? path.resolve(cwd, args.config)
    : null;
  const usesExplicitJSConfig =
    explicitConfigPath != null && isJSConfigFile(explicitConfigPath);
  const jsonGoArgs =
    args.config && !usesExplicitJSConfig
      ? ['--config', args.config, ...goArgs]
      : goArgs;
  const { runEngine } = await import('./engine.js');
  return runEngine({
    binPath,
    goArgs: jsonGoArgs,
    cwd,
    runtime: { singleThreaded: args.singleThreaded },
    extraInit: {
      configDiscovery: {
        mode: usesExplicitJSConfig
          ? 'explicit'
          : args.config
            ? 'disabled'
            : 'auto',
        explicitConfigPath: usesExplicitJSConfig
          ? explicitConfigPath
          : undefined,
        inputs: args.positionals.length === 0 ? ['.'] : args.positionals,
      },
    },
  });
}

export async function runCLI({
  argv = process.argv,
}: RunCLIOptions = {}): Promise<void> {
  const startTime = Date.now();
  const exitCode = await run(resolveRslintBinary(), argv.slice(2), startTime);
  // Let stdout/stderr flush naturally instead of terminating the process.
  process.exitCode = exitCode;
}
