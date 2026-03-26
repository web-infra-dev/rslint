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
      trace: { type: 'string' },
      cpuprof: { type: 'string' },
    },
  });

  // Collect args that are not --config or --init for pass-through to Go.
  // positionals collects only true file/dir arguments.
  const rest: string[] = [];
  const positionals: string[] = [];
  for (const token of tokens) {
    if (token.kind === 'option') {
      if (token.name === 'config' || token.name === 'init') continue;
      rest.push(token.rawName);
      if (token.value != null) rest.push(token.value);
    } else if (token.kind === 'option-terminator') {
      rest.push('--');
    } else if (token.kind === 'positional') {
      rest.push(token.value);
      positionals.push(token.value);
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
