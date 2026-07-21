/**
 * Compute the `@rslint/native-<tuple>` platform package for the host.
 *
 * Shared by the native loader (`load-binding.ts`) and the packaged-isolation
 * test, so they always agree on which platform package to load vs. stage.
 */
import { execSync } from 'node:child_process';
import { readFileSync } from 'node:fs';

// musl detection (Linux only): a glibc `.node` and a musl `.node` are not
// interchangeable. Probe order mirrors napi-rs: ldd file → process.report →
// `ldd --version`.

function isMuslFromFilesystem(): boolean | null {
  try {
    return readFileSync('/usr/bin/ldd', 'utf-8').includes('musl');
  } catch {
    return null;
  }
}

function isMuslFromReport(): boolean | null {
  const report = process.report as unknown as
    { getReport(): unknown; excludeNetwork?: boolean } | undefined;
  if (typeof report?.getReport !== 'function') {
    return null;
  }
  report.excludeNetwork = true;
  const parsed = report.getReport() as {
    header?: { glibcVersionRuntime?: string };
    sharedObjects?: string[];
  };
  if (parsed.header?.glibcVersionRuntime) {
    return false;
  }
  if (Array.isArray(parsed.sharedObjects)) {
    return parsed.sharedObjects.some(
      (f) => f.includes('libc.musl-') || f.includes('ld-musl-'),
    );
  }
  return null;
}

function isMuslFromChildProcess(): boolean {
  try {
    return execSync('ldd --version', { encoding: 'utf8' }).includes('musl');
  } catch {
    return false;
  }
}

function isMusl(): boolean {
  let musl = isMuslFromFilesystem();
  if (musl === null) {
    musl = isMuslFromReport();
  }
  if (musl === null) {
    musl = isMuslFromChildProcess();
  }
  return musl ?? false;
}

/** The platform tuple, e.g. `darwin-arm64`, `linux-x64-musl`, `win32-x64-msvc`. */
export function platformTuple(): string {
  const { platform, arch } = process;
  switch (platform) {
    case 'darwin':
      return `darwin-${arch}`;
    case 'win32':
      // Only msvc targets are built; skip napi's gnu/msvc probe, which
      // mis-detects under Electron's shared-library Node.
      return `win32-${arch}-msvc`;
    case 'linux':
      return `linux-${arch}-${isMusl() ? 'musl' : 'gnu'}`;
    default:
      throw new Error(
        `@rslint/core: unsupported platform ${platform}-${arch} (no native parser binary)`,
      );
  }
}

/** The `@rslint/native-<tuple>` platform package name for the host. */
export function platformPackageName(): string {
  return `@rslint/native-${platformTuple()}`;
}
