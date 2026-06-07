/**
 * Entry point for `@rslint/core`'s internal ESLint-plugin compatibility
 * runtime — Worker pool, per-file lint pipeline, scope manager, fixer.
 * (Merged from the former `@rslint/eslint-plugin-runner` package.)
 *
 * Designed to be hosted in-process — the CLI path and the `rslint` VS Code
 * extension each host a WorkerPool here and answer plugin-lint IPC / LSP
 * requests coming back from the Go side; never a separate sidecar process.
 * The actual wiring (engine/extension ↔ this module) lives in the CLI
 * engine and the VS Code PluginLintPool host.
 */

export type { ConfigDescriptor } from './types.js';

// Worker pool
export { WorkerPool } from './worker-pool.js';
export type { WorkerPoolOptions, LintTask } from './worker-pool.js';

// CLI / LSP plugin-lint host — owns a WorkerPool and answers reverse
// `pluginLint` requests. Shared by the CLI engine and the VS Code
// extension's PluginLintPool so the request→tasks→result boundary is
// single-sourced.
export { createPluginLintHost } from './host.js';
export type { PluginLintHost } from './host.js';

// Plugin-lint runtime types — exported so callers can build requests and
// read responses without importing internal modules.
export type {
  Diagnostic,
  ESTreeNode,
  SuggestionDescriptor,
  SuggestionsMode,
  RuleContext,
} from './linter/context.js';
export type {
  LintFileRequest,
  LintFileResult,
  RuleConfig,
} from './linter/ecma-language-plugin.js';

// Direct single-file lint entry — bypasses the WorkerPool, useful for
// in-process testing / audit harnesses that want deterministic
// behavior without the worker spawn cost. Production consumers should
// continue to go through `WorkerPool` for parallelism + cancel/timeout
// guarantees.
export { lintFile } from './linter/ecma-language-plugin.js';
export type { LoadedPlugins } from './plugin/plugin-loader.js';
export {
  loadPluginsFromConfigs,
  loadPluginsFromConfigFile,
} from './plugin/plugin-loader.js';

// eslintPlugins lint boundary helpers — single-sourced wire-shape
// projections shared by every host that owns a WorkerPool (CLI
// engine.ts, VS Code PluginLintPool.ts). See plugin-lint-protocol.ts
// for the contract rationale.
export {
  buildPluginLintTasks,
  buildPluginLintResult,
} from './plugin/plugin-lint-protocol.js';
export type {
  EslintPluginLintRequest,
  EslintPluginLintResult,
  BuildPluginLintTasksOptions,
} from './plugin/plugin-lint-protocol.js';
