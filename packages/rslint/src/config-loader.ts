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
 * - .ts/.mts: try native import() first, fall back to jiti
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
    // Try native Node.js import (Node >= 22.6 with --experimental-strip-types)
    try {
      // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment -- dynamic config import
      const mod: Record<string, unknown> = await import(
        pathToFileURL(configPath).href
      );
      return mod.default ?? mod;
    } catch {
      // native loading failed, try jiti fallback
    }

    const jiti = await loadJiti();
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
async function loadJiti(): Promise<{
  import: (path: string) => Promise<unknown>;
} | null> {
  try {
    const { createJiti } = await import('jiti');
    return createJiti(process.cwd(), { interopDefault: true });
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

  return config.map((entry: Record<string, unknown>) => ({
    files: entry.files,
    ignores: entry.ignores,
    languageOptions: entry.languageOptions,
    rules: entry.rules,
    plugins: entry.plugins,
    settings: entry.settings,
  }));
}
