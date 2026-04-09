import fs from 'node:fs';
import path from 'node:path';
import { parseArgs as nodeParseArgs } from 'node:util';

export function isJSConfigFile(filePath: string): boolean {
  return /\.(ts|mts|js|mjs)$/.test(filePath);
}

export function parseArgs(argv: string[]) {
  const { values, tokens } = nodeParseArgs({
    args: argv,
    strict: false,
    tokens: true,
    options: {
      config: { type: 'string' },
      init: { type: 'boolean' },
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
    config: (values.config as string) ?? null,
    init: (values.init as boolean) ?? false,
    rest,
    positionals,
  };
}

/**
 * Classify positional args into files and directories.
 * Resolves symlinks so paths are consistent with process.cwd() and
 * TypeScript's SourceFile.FileName() which both return real paths.
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
