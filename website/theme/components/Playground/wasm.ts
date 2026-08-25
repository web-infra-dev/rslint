import type { RSLintService } from '@rslint/core/service';

export const MINIMUM_WASM_VERSION = '0.8.0';

const NPM_REGISTRY_URL = 'https://registry.npmjs.org/@rslint%2Fwasm';
const UNPKG_BASE_URL = 'https://unpkg.com/@rslint/wasm';

interface WasmModule {
  initialize(options: { wasmURL: string }): Promise<RSLintService>;
}

interface NpmPackageMetadata {
  versions?: Record<string, unknown>;
}

let activeService:
  { version: string; promise: Promise<RSLintService> } | undefined;

function parseStableVersion(version: string): [number, number, number] | null {
  const match = /^(\d+)\.(\d+)\.(\d+)$/.exec(version);
  if (!match) return null;
  return [Number(match[1]), Number(match[2]), Number(match[3])];
}

function compareVersions(left: string, right: string): number {
  const leftParts = parseStableVersion(left)!;
  const rightParts = parseStableVersion(right)!;
  for (let index = 0; index < leftParts.length; index++) {
    const difference = leftParts[index] - rightParts[index];
    if (difference !== 0) return difference;
  }
  return 0;
}

export async function fetchWasmVersions(
  signal?: AbortSignal,
): Promise<string[]> {
  const response = await fetch(NPM_REGISTRY_URL, { signal });
  if (!response.ok) {
    throw new Error(`npm registry returned ${response.status}`);
  }

  const metadata = (await response.json()) as NpmPackageMetadata;
  const versions = Object.keys(metadata.versions ?? {})
    .filter((version) => parseStableVersion(version) !== null)
    .filter((version) => compareVersions(version, MINIMUM_WASM_VERSION) >= 0)
    .sort((left, right) => compareVersions(right, left));

  if (versions.length === 0) {
    throw new Error(
      `no stable version at or above ${MINIMUM_WASM_VERSION} was found`,
    );
  }
  return versions;
}

export async function ensureWasmService(
  version: string,
): Promise<RSLintService> {
  if (activeService?.version === version) return activeService.promise;

  const previousService = activeService?.promise;
  const packageUrl = `${UNPKG_BASE_URL}@${version}`;
  const promise = (async () => {
    const previous = await previousService?.catch(() => undefined);
    await previous?.close();
    const wasmModule = (await import(
      /* rspackIgnore: true */ `${packageUrl}/dist/index.mjs`
    )) as WasmModule;
    return wasmModule.initialize({ wasmURL: `${packageUrl}/rslint.wasm` });
  })();

  activeService = { version, promise };
  try {
    return await promise;
  } catch (error) {
    if (activeService?.promise === promise) activeService = undefined;
    throw error;
  }
}
