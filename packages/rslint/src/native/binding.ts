/**
 * Native binding loader shared by IPC memory transport and worker parser.
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
import type { MemoryArena, ParseResult, SharedBytes } from './types.js';

export type {
  ByteInput,
  CommentObj,
  MemoryArena,
  MemoryBatch,
  MemoryConfiguration,
  NativeMemoryMapping,
  ParseResult,
  SharedBytes,
} from './types.js';

const require = createRequire(import.meta.url);

export interface NativeBinding {
  MemoryArena: new () => MemoryArena;
  readBytes(bytes: SharedBytes): Buffer;
  parseSharedBytes(
    filename: string,
    source: SharedBytes,
    sourceType: string,
    jsx: boolean,
  ): { parsed: ParseResult; sourceText: string; hadBom: boolean };
  parse(
    filename: string,
    source: string,
    sourceType: string,
    jsx: boolean,
  ): ParseResult;
}

let binding: NativeBinding | undefined;

export function getNativeBinding(): NativeBinding {
  if (binding) return binding;
  const pkg = platformPackageName();
  try {
    // Platform package exports `.` -> `./rslint.<tuple>.node`. A `.node` addon
    // is CommonJS-only, so loading it through this dynamic require is the point.
    // The platform package implements the N-API ABI described by NativeBinding.
    // rslint-disable-next-line @typescript-eslint/no-var-requires, @typescript-eslint/no-require-imports, @typescript-eslint/no-unsafe-type-assertion
    binding = require(pkg) as NativeBinding;
    return binding;
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
