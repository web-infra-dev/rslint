/* rslint-disable @typescript-eslint/no-unsafe-type-assertion */
/** Shared Go-to-host task projection and host-to-Go result projection. */

import type { LintTask } from '../worker-pool.js';
import type { LintFileResult } from '../linter/ecma-language-plugin.js';
import type { SharedBytes } from '../../native/types.js';

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
  /** LSP config/worker generation. Omitted by the CLI's single-host path. */
  generation?: string;
  files: ReadonlyArray<
    {
      path: string;
      /**
       * Per-file `languageOptions`, computed by Go via `GetConfigForFile`
       * (flat-config files-glob match + deep merge). Opaque here; the
       * worker reads `sourceType`/`globals`/`parserOptions.ecmaFeatures`.
       */
      languageOptions?: unknown;
      settings?: Record<string, unknown>;
      /**
       * The owning config's absolute filesystem directory in the same form the
       * host used as its `ConfigDescriptor.configDirectory`.
       * The worker uses it to pick the right `LoadedPlugins`. Empty when
       * no JS config governs the file.
       */
      configKey?: string;
    } & (
      | {
          /** Complete frame, base64 only on LSP's JSON transport. */
          frame: string;
          frameAttachment?: never;
        }
      | {
          /** Index into the binary IPC attachments. */
          frameAttachment: number;
          frame?: never;
        }
    )
  >;
  rules?: Record<string, { options?: readonly unknown[] }>;
  /** Materialize each diagnostic's autofix payload without applying it. */
  collectFixes: boolean;
  suggestionsMode?: 'off' | 'eager';
  /** Collect per-rule execution times (driven by Go's `--timing`). */
  collectTiming?: boolean;
}

/** Host input after attachment resolution; native capabilities are local to Node. */
export interface ResolvedEslintPluginLintRequest extends Omit<
  EslintPluginLintRequest,
  'files'
> {
  files: ReadonlyArray<
    Omit<
      EslintPluginLintRequest['files'][number],
      'frameAttachment' | 'frame'
    > & {
      /** Node-owned bytes or a request-scoped native capability. */
      frame: Uint8Array | SharedBytes;
      frameAttachment?: never;
    }
  >;
}

// ─────────────────────────────────────────────────────────────────────
// buildPluginLintTasks
// ─────────────────────────────────────────────────────────────────────

/**
 * Build per-file {@link LintTask}s from a resolved plugin request. Each
 * task carries the file's `configKey` verbatim; the worker uses it to
 * pick the right `LoadedPlugins` from its per-config map.
 */
export function buildPluginLintTasks(
  input: ResolvedEslintPluginLintRequest,
): LintTask[] {
  const sharedRules = Object.fromEntries(
    Object.entries(input.rules ?? {}).map(([k, v]) => [
      k,
      { options: v.options ?? [], meta: undefined },
    ]),
  );
  const { collectFixes } = input;
  const suggestionsMode: 'off' | 'eager' = input.suggestionsMode ?? 'off';
  const collectTiming = input.collectTiming ?? false;

  return input.files.map((f) => ({
    filePath: f.path,
    frame: f.frame,
    languageOptions: f.languageOptions as never,
    settings: f.settings,
    rules: sharedRules,
    collectFixes,
    suggestionsMode,
    collectTiming,
    configKey: f.configKey ?? '',
  }));
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
    ruleTimes?: Record<string, number>;
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
      ruleTimes: r.ruleTimes,
    })),
  };
}
