/**
 * Entry point for `@rslint/core`'s internal ESLint-plugin compatibility
 * runtime — Worker pool, per-file lint pipeline, scope manager, fixer.
 * (Merged from the former `@rslint/eslint-plugin-runner` package.)
 *
 * Designed to be hosted in-process — the CLI path and the `rslint` VS Code
 * extension each host a WorkerPool here and answer plugin-lint IPC / LSP
 * requests coming back from the Go side; never a separate sidecar process.
 * The actual wiring (engine ↔ this module) is the C-line integration, which
 * is separate from this package merge.
 */

export type { ConfigDescriptor } from './types.js';

// Worker pool
export { WorkerPool } from './worker-pool.js';
export type { WorkerPoolOptions, LintTask } from './worker-pool.js';

// Compat-layer types — exported so callers can build requests and read
// responses without importing internal modules.
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

// Compat-batch boundary helpers — single-sourced wire-shape projections
// shared by every host that owns a WorkerPool (CLI engine.ts, VS Code
// extension CompatPool.ts). See compat-task-builder.ts for the
// contract rationale.
export {
  buildCompatTasksByConfigKey,
  buildCompatBatchResult,
} from './plugin/compat-task-builder.js';
export type {
  CompatBatchInput,
  CompatBatchResult,
  BuildCompatTasksByConfigKeyOptions,
} from './plugin/compat-task-builder.js';
