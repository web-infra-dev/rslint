/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/**
 * Boundary helpers for the eslintPlugins lint protocol — the shape Go
 * sends down (`pluginLint` IPC request OR `rslint/pluginLint`
 * LSP request) and the shape it reads back.
 *
 * Both the CLI host (`packages/rslint/src/cli/engine.ts`) and the LSP host
 * (`packages/vscode-extension/src/PluginLintPool.ts`) receive an
 * EslintPluginLintRequest from Go, build per-file LintTasks against the
 * WorkerPool, and project the results back. The boundary logic is
 * identical across the two paths — only the warning sink (stderr vs
 * logger) differs, parameterized via
 * {@link BuildPluginLintTasksOptions.onUnknownConfigKey}.
 *
 * Sharing this here keeps the wire contract single-sourced: any change
 * to the request parsing or result projection lands in exactly one
 * place, and Go sees the same fields regardless of which host produced
 * them.
 */

import type { LintTask } from '../worker-pool.js';
import type { LintFileResult } from '../linter/ecma-language-plugin.js';

// ─────────────────────────────────────────────────────────────────────
// Inputs
// ─────────────────────────────────────────────────────────────────────

/**
 * Wire-format lint request as it leaves Go and arrives at the host.
 *
 * Mirrors Go's `internal/linter.EslintPluginLintRequest`. Field types
 * are permissive (`unknown` for opaque pass-through fields) so the host
 * doesn't need to re-validate Go's serialization.
 */
export interface EslintPluginLintRequest {
  files: ReadonlyArray<{
    path: string;
    /**
     * Optional file content override. The CLI host leaves it absent —
     * the worker reads from disk via `readFileSync` (and re-reads
     * post-fix content across `--fix` passes). The LSP host sends it so
     * an unsaved editor buffer's overlay text is linted instead of the
     * stale on-disk copy. Also used by in-process test harnesses.
     */
    text?: string;
    /**
     * Per-file `languageOptions`, computed by Go via `GetConfigForFile`
     * (flat-config files-glob match + deep merge). Opaque here; the
     * worker reads `sourceType`/`globals`/`parserOptions.ecmaFeatures`.
     */
    languageOptions?: unknown;
    settings?: Record<string, unknown>;
    /**
     * The owning config's directory in the SAME form the host used as
     * its `ConfigDescriptor.configDirectory` (CLI: fs path; LSP: URI).
     * The worker uses it to pick the right `LoadedPlugins`. Empty when
     * no JS config governs the file.
     */
    configKey?: string;
  }>;
  rules?: Record<string, { options?: readonly unknown[] }>;
  /** Collect autofixes (driven by Go's `--fix`). */
  fix?: boolean;
  suggestionsMode?: 'off' | 'eager';
}

export interface BuildPluginLintTasksOptions {
  /**
   * Set of `configDirectory` strings the worker pool was initialized
   * with. Used purely to detect "unknown configKey on the wire" — the
   * host's invariant is that every file's `configKey` was previously
   * declared in `WorkerPoolOptions.configs[]`. A miss is an internal
   * bug surfaced through {@link onUnknownConfigKey}.
   */
  configDirSet: ReadonlySet<string>;
  /**
   * Invoked once per file whose `configKey` is not in `configDirSet`.
   * The helper otherwise forwards `configKey` verbatim — the worker
   * reports the failure via `parseError`.
   */
  onUnknownConfigKey?: (filePath: string, configKey: string) => void;
}

// ─────────────────────────────────────────────────────────────────────
// buildPluginLintTasks
// ─────────────────────────────────────────────────────────────────────

/**
 * Build per-file {@link LintTask}s from an EslintPluginLintRequest. Each
 * task carries the file's `configKey` verbatim; the worker uses it to
 * pick the right `LoadedPlugins` from its per-config map.
 *
 * If a file's `configKey` is empty OR missing from `configDirSet`, we
 * still emit a task (with the empty/unknown key) and let the worker
 * report the failure via `parseError` — keeping wire-format consistency.
 */
export function buildPluginLintTasks(
  input: EslintPluginLintRequest,
  options: BuildPluginLintTasksOptions,
): LintTask[] {
  const sharedRules = Object.fromEntries(
    Object.entries(input.rules ?? {}).map(([k, v]) => [
      k,
      { options: v.options ?? [], meta: undefined },
    ]),
  );
  // `fix` is the wire-level name (mirrors Go's `EslintPluginLintRequest.Fix`);
  // the worker's per-task field stays `collectFixes`.
  const collectFixes = input.fix ?? false;
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
// buildPluginLintResult
// ─────────────────────────────────────────────────────────────────────

/**
 * Wire-format lint result, returned to Go. Mirrors Go's
 * `internal/linter.EslintPluginFileResult`. Pure projection — drops the
 * runner-internal aggregate fields (`fixes`, `suggestionsCount`) that
 * Go doesn't decode.
 */
export interface EslintPluginLintResult {
  results: Array<{
    filePath: string;
    diagnostics: unknown[];
    parseError?: string;
    cancelled?: boolean;
    ruleErrors?: Array<{ rule: string; message: string }>;
  }>;
}

/**
 * Project the runner's `LintFileResult[]` into the exact wire shape Go
 * decodes. Both host paths use this so Go receives a byte-stable set of
 * fields.
 */
export function buildPluginLintResult(
  results: LintFileResult[],
): EslintPluginLintResult {
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
