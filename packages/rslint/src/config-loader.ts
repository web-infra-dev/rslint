import fs from 'node:fs';
import path from 'node:path';
import { pathToFileURL } from 'node:url';

export const JS_CONFIG_FILES = [
  'rslint.config.js',
  'rslint.config.mjs',
  'rslint.config.ts',
  'rslint.config.mts',
];

export function findJSConfig(cwd: string): string | null {
  for (const name of JS_CONFIG_FILES) {
    const p = path.join(cwd, name);
    if (fs.existsSync(p)) return p;
  }
  return null;
}

/**
 * Load a JS/TS config file.
 * - .js/.mjs: native import()
 * - .ts/.mts: native import() when Node.js has TypeScript support (>= 22.6),
 *             otherwise fall back to jiti
 */
export async function loadConfigFile(configPath: string): Promise<unknown> {
  const ext = path.extname(configPath);

  if (ext === '.js' || ext === '.mjs') {
    const mod: Record<string, unknown> = await import(
      pathToFileURL(configPath).href
    );
    return mod.default ?? mod;
  }

  if (ext === '.ts' || ext === '.mts') {
    // Use feature detection to decide the loading strategy (same as rsbuild).
    // process.features.typescript is available in Node.js >= 22.6.
    const useNative = Boolean(process.features.typescript);

    if (useNative) {
      const mod: Record<string, unknown> = await import(
        pathToFileURL(configPath).href
      );
      return mod.default ?? mod;
    }

    const jiti = await loadJiti(configPath);
    if (jiti) {
      const resolved = await jiti.import(configPath);
      return extractDefault(resolved);
    }

    throw new Error(
      `Failed to load TypeScript config file: ${configPath}\n` +
        `To load .ts config files, either:\n` +
        `  1. Use Node.js >= 22.6 (with native TypeScript support)\n` +
        `  2. Install jiti as a dependency: npm install -D jiti`,
    );
  }

  throw new Error(`Unsupported config file extension: ${ext}`);
}

/**
 * Try to load jiti (optional peer dependency).
 */
async function loadJiti(configPath: string): Promise<{
  import: (path: string) => Promise<unknown>;
} | null> {
  try {
    const { createJiti } = await import('jiti');
    return createJiti(path.dirname(configPath), { interopDefault: true });
  } catch {
    return null;
  }
}

function extractDefault(mod: unknown): unknown {
  if (typeof mod === 'object' && mod !== null && 'default' in mod) {
    return mod.default;
  }
  return mod;
}

/**
 * Validate and strip non-serializable fields from the config.
 */
export function normalizeConfig(config: unknown): Record<string, unknown>[] {
  if (!Array.isArray(config)) {
    throw new Error(
      `rslint config must export an array (flat config format), got ${typeof config}`,
    );
  }

  return config
    .filter((entry: unknown, index: number) => {
      if (entry == null || typeof entry !== 'object') {
        console.warn(
          `[rslint] Config entry at index ${index} is not an object (got ${entry === null ? 'null' : typeof entry}), skipping.`,
        );
        return false;
      }
      return true;
    })
    .map((entry: Record<string, unknown>, index: number) => {
      if (entry.files != null && !Array.isArray(entry.files)) {
        throw new Error(
          `[rslint] Config entry at index ${index}: "files" must be an array, got ${typeof entry.files}`,
        );
      }
      if (entry.ignores != null && !Array.isArray(entry.ignores)) {
        throw new Error(
          `[rslint] Config entry at index ${index}: "ignores" must be an array, got ${typeof entry.ignores}`,
        );
      }
      return {
        files: entry.files,
        ignores: entry.ignores,
        languageOptions: entry.languageOptions,
        rules: entry.rules,
        plugins: entry.plugins,
        settings: entry.settings,
      };
    });
}
