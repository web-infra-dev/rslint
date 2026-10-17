/**
 * Native binding loader shared by the IPC host and worker byte reader.
 *
 * Loads the matching platform byte-transport addon. AST projection lives in Go.
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
import type { MemoryArena, SharedBytes } from './types.js';

export type {
  ByteInput,
  MemoryArena,
  MemoryBatch,
  MemoryConfiguration,
  NativeMemoryMapping,
  SharedBytes,
} from './types.js';

const require = createRequire(import.meta.url);

export interface NativeBinding {
  MemoryArena: new () => MemoryArena;
  readBytes(bytes: SharedBytes): Buffer;
  readBytesInto(bytes: SharedBytes, destination: Buffer): void;
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
      `@rslint/core: failed to load the native byte transport from "${pkg}". ` +
        `Ensure the matching optional dependency is installed (reinstall after ` +
        `removing node_modules and the lockfile if it is missing).`,
    );
    (err as Error & { cause?: unknown }).cause = cause;
    throw err;
  }
}
