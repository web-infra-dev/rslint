/**
 * Pure helper functions used by `CompatPool`. Kept in a separate
 * module so they can be unit-tested without VS Code mocks — CompatPool
 * itself imports `vscode` (for `Disposable` / `CancellationToken`),
 * which can't be loaded outside the extension host.
 *
 * Anything here MUST be:
 *   - synchronous,
 *   - free of `vscode` imports,
 *   - free of dynamic `import()` (we test these as plain modules).
 *
 * The actual WorkerPool wiring stays in CompatPool.ts.
 */

import { statSync } from 'node:fs';
import { posix as posixPath } from 'node:path';
import { fileURLToPath } from 'node:url';

import type { ConfigDescriptor } from '@rslint/eslint-plugin-runner';
import { configEntryHasEslintPlugins } from '@rslint/eslint-plugin-runner/plugin-source';

/**
 * One config the extension discovered. Mirrors the wire shape sent to
 * Go via `rslint/configUpdate`. The lean `entries[]` carry only what
 * Go reads for placeholder rule registration; the configs-flow worker
 * imports the actual config file directly to recover live plugin
 * instances.
 */
export interface NormalizedConfig {
  /** URI form of the config's directory, e.g. "file:///proj". */
  configDirectory: string;
  /** Absolute filesystem path of the config file itself, e.g.
   *  "/proj/rslint.config.mjs". The worker imports this. */
  configPath: string;
  entries: Array<{
    eslintPlugins?: Array<{ prefix: string; ruleNames: string[] }>;
  }>;
}

/**
 * Identity hash for a set of configs. Two equivalent config sets
 * produce equal fingerprints regardless of array order. The pool
 * restarts whenever this changes — that covers:
 *
 *   - Config file added or removed from the set
 *   - Config file edited (mtime / size shifts)
 *   - File-system identity changes (rare; covered by `mtimeMs`+`size`
 *     together to keep collisions essentially impossible even at 1-sec
 *     mtime granularity on FAT-family filesystems)
 *   - Whether ANY config declares any plugin: the boolean
 *     `hasAnyPlugin` is baked into the fingerprint, so tier-3
 *     transitions (all plugins removed) flow through naturally
 *
 * When a config file is missing on disk at fingerprint time we record
 * a sentinel `[path, -1, -1, false]` — a config that disappeared is a
 * different state from a config that's present and unchanged, and we
 * want the next reconfigure to drain the pool.
 *
 * JSON encoding is byte-stable and trivially diff-able in tests.
 */
export function fingerprintConfigs(configs: NormalizedConfig[]): string {
  const rows = configs
    .map((c): readonly [string, number, number, boolean] => {
      const hasPlugin = c.entries.some(
        (e) => Array.isArray(e.eslintPlugins) && e.eslintPlugins.length > 0,
      );
      let mtimeMs = -1;
      let size = -1;
      if (c.configPath) {
        try {
          const st = statSync(c.configPath);
          mtimeMs = st.mtimeMs;
          size = st.size;
        } catch {
          // Missing / unreadable — sentinel values above stay.
        }
      }
      return [c.configPath, mtimeMs, size, hasPlugin];
    })
    .sort((a, b) => {
      if (a[0] < b[0]) return -1;
      if (a[0] > b[0]) return 1;
      return 0;
    });
  return JSON.stringify(rows);
}

/**
 * Convert a `file://...` URL string to a filesystem path. Returns
 * empty string when the input isn't a parseable URL.
 *
 * MUST byte-equal what the Go LSP server's `internal/lsp.uriToPath`
 * produces for the same input URI, since CompatPool's per-file
 * routing keys lintBatch files by the value Go wrote into
 * `CompatLintFile.ConfigKey`. Drift between the two implementations
 * causes silent "rule not found" on every file in workspaces whose
 * path contains a space, non-ASCII character, or (on Windows) a drive
 * letter — the lookup misses and CompatPool treats the configKey as
 * unknown, returning empty pluginEntries.
 *
 * Specifically Go does:
 *   - `url.ParseRequestURI` (percent-decodes the path automatically)
 *   - Windows drive: strip the leading `/` from `/C:/...` → `C:/...`
 *
 * We mirror that here with Node's `fileURLToPath`, which also does
 * the percent-decode and the Windows drive-letter handling. Then we
 * normalize separators to forward slashes (Go uses `/` everywhere;
 * `fileURLToPath` on Windows returns backslashes).
 *
 * Examples:
 *   file:///proj                  → /proj
 *   file:///Users/John%20Doe/p    → /Users/John Doe/p
 *   file:///%E6%B5%8B%E8%AF%95    → /测试
 *   file:///C:/Users/proj  (Win)  → C:/Users/proj
 *
 * Bad input (not a parseable file URL) returns empty string so the
 * caller skips that config rather than crashing the extension.
 */
export function uriToPath(uri: string): string {
  if (!uri) return '';
  try {
    //   1. fileURLToPath → percent-decode + Windows drive-letter handling
    //   2. backslash → forward slash
    //   3. path.posix.normalize → collapse `.`/`..` + dedupe `//`
    //
    // Steps 1-2 mirror Go `internal/lsp.uriToPath`. Step 3 is EXTRA: Go's
    // uriToPath returns the URL path verbatim — it does NOT call
    // `tspath.NormalizePath`, so no `.`/`..` collapse and no `//` dedupe on
    // the Go side. We run posix.normalize here defensively for non-canonical
    // input URIs. In practice both sides land on identical bytes because the
    // extension only ever feeds canonical URIs (`Uri.file().toString()`), so
    // the `ConfigKey` lookup matches; the normalize is belt-and-suspenders,
    // not a Go-mirroring step.
    return posixPath.normalize(fileURLToPath(uri).replaceAll('\\', '/'));
  } catch {
    return '';
  }
}

/**
 * Extract `ConfigDescriptor[]` from the extension's `NormalizedConfig[]`.
 * Only configs that declare at least one plugin are included — workers
 * pay a config-import cost per descriptor, so passing in a plugin-less
 * config would just waste an import.
 *
 * `configDirectory` is the URI form on input; we convert to a
 * filesystem path to byte-match what Go writes into each file's
 * `configKey` during compat dispatch. A URI we can't parse is skipped.
 */
export function extractConfigDescriptors(
  configs: NormalizedConfig[],
): ConfigDescriptor[] {
  const out: ConfigDescriptor[] = [];
  for (const cfg of configs) {
    if (!cfg.entries.some(configEntryHasEslintPlugins)) continue;
    if (!cfg.configPath) continue;
    const dirPath = uriToPath(cfg.configDirectory);
    if (!dirPath) continue;
    out.push({ configPath: cfg.configPath, configDirectory: dirPath });
  }
  return out;
}
