/**
 * Native napi parser loader for `@rslint/core`'s worker runtime.
 *
 * Resolves and loads the platform-specific `.node` — the oxc-based JS/TS/JSX
 * parser exposed as `parse` — from the matching `@rslint/native-<tuple>`
 * platform package. This replaces the former standalone `@rslint/native`
 * wrapper package whose napi-generated loader did the same dispatch.
 *
 * Resolution is IDENTICAL in dev and prod — the host only ever has one
 * platform, so there is no dev-only branch:
 *   - prod: `npm install` ships only the platform package matching the host
 *     (filtered by each package's `os`/`cpu`/`libc`), which carries the `.node`.
 *   - dev:  `pnpm build` drops the freshly built host `.node` into
 *     `npm/rslint/<tuple>/`, the same workspace package this require resolves to.
 *
 * A `.node` addon is CommonJS-only — ESM cannot `import` it — so we load it via
 * `createRequire(import.meta.url)`. The package name is computed at runtime, so
 * rspack can't statically follow the `require`; the binary stays external
 * (intended — a `.node` can't be bundled). Same `createRequire` pattern the
 * worker already uses for `os.cpus()` in worker-pool.ts.
 */
import { createRequire } from 'node:module';

import { platformPackageName } from './platform-tuple.js';

const require = createRequire(import.meta.url);

/** ESLint-shape comment (`{ type, value, start, end }`); start/end are UTF-16 offsets. */
export interface CommentObj {
  /** "Line" | "Block" */
  type: string;
  /** Comment body with the `//` or block delimiters stripped. */
  value: string;
  start: number;
  end: number;
}

/** Parser output: ESTree JSON + comments + columnar token arrays (all UTF-16 offsets). */
export interface ParseResult {
  /** ESTree as a JSON string (no `range`; normalize-ast derives it from start/end). */
  program: string;
  comments: Array<CommentObj>;
  /** Parser-driven token stream in columnar form. */
  tokenTypes: Uint8Array;
  tokenStarts: Uint32Array;
  tokenEnds: Uint32Array;
}

interface NativeBinding {
  parse(
    filename: string,
    source: string,
    sourceType: string,
    jsx: boolean,
  ): ParseResult;
}

function loadBinding(): NativeBinding {
  const pkg = platformPackageName();
  try {
    // Platform package exports `.` -> `./rslint.<tuple>.node`. A `.node` addon
    // is CommonJS-only, so loading it through this dynamic require is the point.
    // rslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports
    return require(pkg) as NativeBinding;
  } catch (cause) {
    const err = new Error(
      `@rslint/core: failed to load the native parser from "${pkg}". ` +
        `Ensure the matching optional dependency is installed (reinstall after ` +
        `removing node_modules and the lockfile if it is missing).`,
    );
    (err as Error & { cause?: unknown }).cause = cause;
    throw err;
  }
}

const binding = loadBinding();

/**
 * Parse JS/TS/JSX source -> ESTree JSON + ESLint-shape comments + token stream
 * (all UTF-16 offsets). Throws only on the parser's source-size guard / a Rust
 * panic; recoverable syntax errors still return a best-effort AST.
 */
export const parse = binding.parse;
