import fs from 'node:fs';
import path from 'node:path';
import { parseArgs as nodeParseArgs } from 'node:util';

// Keep this list in sync with cmd/rslint/internal/output/format.go.
export const OUTPUT_FORMATS = [
  'default',
  'jsonline',
  'github',
  'gitlab',
] as const;

export type OutputFormat = (typeof OUTPUT_FORMATS)[number];

export function isOutputFormat(value: string): value is OutputFormat {
  return (OUTPUT_FORMATS as readonly string[]).includes(value);
}

export function isJSConfigFile(filePath: string): boolean {
  return /\.(ts|mts|cts|js|mjs|cjs)$/.test(filePath);
}

export function parseArgs(argv: string[]) {
  const { values, tokens } = nodeParseArgs({
    args: argv,
    strict: false,
    tokens: true,
    options: {
      config: { type: 'string', short: 'c' },
      init: { type: 'boolean' },
      help: { type: 'boolean', short: 'h' },
      // Detected so the JS host can size the ESLint-plugin worker pool to a
      // single worker. NOT skipped below, so it still forwards to Go in
      // `rest` (Go's native pass honors the same flag independently).
      singleThreaded: { type: 'boolean' },
      // Register known Go string-valued flags so their values are not
      // mistaken for positional file/dir arguments.
      format: { type: 'string' },
      'max-warnings': { type: 'string' },
      rule: { type: 'string', multiple: true },
      trace: { type: 'string' },
      cpuprof: { type: 'string' },
      // Consumed by the JS entry point; must not reach Go from user input.
      'start-time': { type: 'string' },
    },
  });

  // Collect args that are not --config or --init for pass-through to Go.
  // positionals collects only true file/dir arguments.
  // Flags are emitted before positional args because Go's flag.Parse stops
  // at the first positional argument. Without reordering, a flag like
  // `--rule 'no-console: off'` placed after a file path would be silently
  // ignored by Go.
  //
  // When "--" is present, positionals before it and after it are tracked
  // separately so the rebuilt rest preserves their relative position to "--".
  const flags: string[] = [];
  const positionalsBefore: string[] = [];
  const positionalsAfter: string[] = [];
  let seenTerminator = false;
  for (const token of tokens) {
    if (token.kind === 'option') {
      if (
        token.name === 'config' ||
        token.name === 'init' ||
        token.name === 'start-time'
      )
        continue;
      flags.push(token.rawName);
      if (token.value != null) flags.push(token.value);
    } else if (token.kind === 'option-terminator') {
      seenTerminator = true;
    } else if (token.kind === 'positional') {
      if (seenTerminator) {
        positionalsAfter.push(token.value);
      } else {
        positionalsBefore.push(token.value);
      }
    }
  }

  // Rebuild rest: flags first, then positionals that appeared before "--",
  // then "--" (if present), then positionals that appeared after "--".
  const positionals = [...positionalsBefore, ...positionalsAfter];
  const rest = seenTerminator
    ? [...flags, ...positionalsBefore, '--', ...positionalsAfter]
    : [...flags, ...positionalsBefore];

  return {
    // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
    config: (values.config as string) ?? null,
    // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
    init: (values.init as boolean) ?? false,
    // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
    help: (values.help as boolean) ?? false,
    // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
    singleThreaded: (values.singleThreaded as boolean) ?? false,
    // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
    format: (values.format as string) ?? null,
    rest,
    positionals,
  };
}

/**
 * Classify positional args into files and directories.
 * Keeps the caller's absolute lexical path. Config discovery and Go target
 * binding resolve physical identity separately, after lexical ownership has
 * had the first opportunity to match.
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
      if (fs.statSync(resolved).isDirectory()) {
        dirs.push(resolved);
      } else {
        files.push(resolved);
      }
    } catch {
      // Non-existent path: treat as file (Go will handle the error)
      files.push(resolved);
    }
  }
  return { files, dirs };
}
