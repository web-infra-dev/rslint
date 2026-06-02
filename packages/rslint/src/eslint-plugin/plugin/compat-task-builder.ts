/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Boundary helpers for the compat batch protocol — the shape Go sends
 * down (`lintEslintPlugin` IPC request OR `rslint/lintCompatBatch` LSP
 * request) and the shape it reads back.
 *
 * Both the CLI host (`packages/rslint/src/engine.ts`) and the LSP host
 * (`packages/vscode-extension/src/CompatPool.ts`) receive a CompatBatch
 * from Go, build per-file LintTasks against the WorkerPool, and
 * project the results back. The boundary logic is identical across the
 * two paths — only the warning sink (stderr vs logger) differs, which
 * is parameterized via {@link BuildCompatTasksOptions.onUnknownConfigKey}.
 *
 * Sharing this here keeps the wire contract single-sourced: any change
 * to the compat batch's input parsing or output projection lands in
 * exactly one place, and Go sees the same fields regardless of which
 * host produced them. Previously the CLI path passed `LintFileResult`
 * through verbatim (carrying aggregate convenience fields the LSP
 * path stripped), so Go received different supersets from each — it
 * worked only because the JSON decoder ignores unknown fields.
 */

import type { LintTask } from '../worker-pool.js';
import type { LintFileResult } from '../linter/ecma-language-plugin.js';

// ─────────────────────────────────────────────────────────────────────
// Inputs
// ─────────────────────────────────────────────────────────────────────

/**
 * Wire-format compat batch as it leaves Go and arrives at the host.
 *
 * Mirrors `internal/linter.CompatBatch` exactly. Field types are
 * permissive (`unknown` for opaque pass-through fields) so the host doesn't
 * need to re-validate Go's serialization.
 */
export interface CompatBatchInput {
  files: ReadonlyArray<{
    path: string;
    /**
     * Optional file content override. Go's `CompatLintFile`
     * (internal/linter/types.go) does NOT include a `text` field —
     * Go never sends it over the wire. The worker reads file content
     * from disk via `readFileSync(req.filePath, 'utf8')` when this is
     * absent. The field is kept here purely for in-process callers
     * (test harnesses) that want to feed synthetic source without
     * touching the filesystem.
     */
    text?: string;
    languageOptions?: unknown;
    settings?: Record<string, unknown>;
    /**
     * Filesystem-path form of the owning config's directory. Empty
     * when no JS config governs the file (JSON config / `--api`
     * mode).
     */
    configKey?: string;
  }>;
  rules?: Record<string, { options?: readonly unknown[] }>;
  collectFixes?: boolean;
  suggestionsMode?: 'off' | 'eager';
}

export interface BuildCompatTasksByConfigKeyOptions {
  /**
   * Set of `configDirectory` strings the worker pool was initialized
   * with. Used purely to detect "unknown configKey on the wire" — the
   * host's invariant is that every file's `configKey` was previously
   * declared in `WorkerPoolOptions.configs[]`. A miss is an internal
   * bug, surfaced through {@link onUnknownConfigKey} so the host can
   * log it before the worker drops the task with an internal-error
   * parseError.
   */
  configDirSet: ReadonlySet<string>;
  /**
   * Invoked once per file whose `configKey` is not in `configDirSet`.
   * The helper otherwise forwards `configKey` verbatim — the
   * worker reports the failure via `parseError`. Host can use this hook
   * to emit a clearer log message before the worker's terser one
   * arrives back.
   */
  onUnknownConfigKey?: (filePath: string, configKey: string) => void;
}

// ─────────────────────────────────────────────────────────────────────
// buildCompatTasksByConfigKey
// ─────────────────────────────────────────────────────────────────────

/**
 * Build per-file {@link LintTask}s from a CompatBatch. Each task
 * carries the file's `configKey` verbatim; the worker uses that key
 * to pick the right `LoadedPlugins` from its per-config map.
 *
 * `configDirSet` is the set of `configDirectory` strings the worker
 * pool was initialized with. If a file's `configKey` is empty OR
 * missing from the set, we still emit a task (with the empty/unknown
 * key) and let the worker report the failure via `parseError` — that
 * keeps wire-format consistency. The optional `onUnknownConfigKey`
 * hook lets the host log additional context before the worker's
 * terser error lands.
 */
export function buildCompatTasksByConfigKey(
  input: CompatBatchInput,
  options: BuildCompatTasksByConfigKeyOptions,
): LintTask[] {
  const sharedRules = Object.fromEntries(
    Object.entries(input.rules ?? {}).map(([k, v]) => [
      k,
      { options: v.options ?? [], meta: undefined },
    ]),
  );
  const collectFixes = input.collectFixes ?? false;
  const suggestionsMode: 'off' | 'eager' = input.suggestionsMode ?? 'off';

  return input.files.map((f) => {
    const configKey = f.configKey ?? '';
    if (configKey !== '' && !options.configDirSet.has(configKey)) {
      options.onUnknownConfigKey?.(f.path, configKey);
    }
    return {
      filePath: f.path,
      text: f.text,
      languageOptions: f.languageOptions as never,
      settings: f.settings,
      rules: sharedRules,
      collectFixes,
      suggestionsMode,
      configKey,
    };
  });
}

// ─────────────────────────────────────────────────────────────────────
// buildCompatBatchResult
// ─────────────────────────────────────────────────────────────────────

/**
 * Wire-format compat batch result, returned to Go. Mirrors
 * `internal/linter.CompatFileResult`. Pure projection — drops the
 * runner-internal aggregate fields (`fixes`, `suggestionsCount`) that
 * Go doesn't decode.
 */
export interface CompatBatchResult {
  results: Array<{
    filePath: string;
    diagnostics: unknown[];
    parseError?: string;
    cancelled?: boolean;
    ruleErrors?: Array<{ rule: string; message: string }>;
  }>;
}

/**
 * Project the runner's `LintFileResult[]` into the exact wire shape
 * Go decodes. Both host paths use this so Go receives a byte-stable
 * set of fields — without it the CLI path silently emitted extra
 * fields the LSP path didn't, leaving "ignore unknown fields" as the
 * only thing holding the contract together.
 */
export function buildCompatBatchResult(
  results: LintFileResult[],
): CompatBatchResult {
  return {
    results: results.map((r) => ({
      filePath: r.filePath,
      diagnostics: r.diagnostics,
      parseError: r.parseError,
      cancelled: r.cancelled,
      ruleErrors: r.ruleErrors,
    })),
  };
}
