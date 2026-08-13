import { createRequire } from 'node:module';
import { RSLINT_GLOBAL_SETS } from './presets.js';
import {
  UPSTREAM_GLOBAL_SET_NAMES,
  type UpstreamGlobalSetName,
} from './upstream.js';
// API Extractor only handles this `export =` package through an import assignment.
// rslint-disable-next-line @typescript-eslint/no-require-imports
import type globalsCatalog = require('globals');

type UpstreamGlobalsCatalog = typeof globalsCatalog;
type GlobalsCatalog = UpstreamGlobalsCatalog & typeof RSLINT_GLOBAL_SETS;
type GlobalSet = UpstreamGlobalsCatalog[UpstreamGlobalSetName];

const loadModule = createRequire(import.meta.url);
const loadedSets = new Map<UpstreamGlobalSetName, GlobalSet>();

function loadObject<T extends object>(specifier: string): T {
  const value: unknown = loadModule(specifier);
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    throw new TypeError(
      `Expected ${JSON.stringify(specifier)} to export an object`,
    );
  }
  // The assets come from the pinned build dependency. This check is the
  // runtime boundary from Node's untyped createRequire API.
  // rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
  return value as T;
}

function loadGlobalSet(name: UpstreamGlobalSetName): GlobalSet {
  const cached = loadedSets.get(name);
  if (cached) return cached;

  const value = loadObject<GlobalSet>(`./globals/${name}.json`);
  loadedSets.set(name, value);
  return value;
}

const lazyGlobals: Record<string, unknown> = {};
for (const name of UPSTREAM_GLOBAL_SET_NAMES) {
  const getter = (): GlobalSet => {
    const value = loadGlobalSet(name);
    // Match the upstream object's ordinary data-property behavior after the
    // first access. If the container has been sealed/frozen, retain the getter
    // and serve the same cached object instead.
    const descriptor = Object.getOwnPropertyDescriptor(lazyGlobals, name);
    if (descriptor?.get === getter && descriptor.configurable) {
      Object.defineProperty(lazyGlobals, name, {
        configurable: true,
        enumerable: true,
        value,
        writable: true,
      });
    }
    return value;
  };
  Object.defineProperty(lazyGlobals, name, {
    configurable: true,
    enumerable: true,
    get: getter,
  });
}
Object.assign(lazyGlobals, RSLINT_GLOBAL_SETS);

/**
 * The complete catalog from the pinned `globals` package plus any
 * Rslint-specific sets.
 *
 * Each upstream environment is loaded synchronously and cached on first
 * property access. Rslint-specific sets are ordinary inline objects. Rslib
 * emits upstream maps under `dist/globals`, so importing the root does not
 * parse the complete catalog and consumers install no extra dependency.
 */
// The accessors and inline registry above populate every declared key.
// rslint-disable-next-line @typescript-eslint/no-unsafe-type-assertion
export const globals = lazyGlobals as GlobalsCatalog;
