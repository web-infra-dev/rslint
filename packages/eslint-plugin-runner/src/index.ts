/**
 * Public entry point for `@rslint/eslint-plugin-runner`.
 *
 * This package owns the ESLint-plugin compatibility runtime — Worker
 * pool, per-file lint pipeline, scope manager, fixer. Two consumers
 * embed it directly:
 *
 *   - `@rslint/core/engine.ts` — CLI parent process. It spawns the
 *     Go binary as a child, then hosts a WorkerPool here and answers
 *     `lintEslintPlugin` IPC requests coming back from Go.
 *   - `rslint` VS Code extension — LSP client. It hosts a WorkerPool
 *     here and answers `rslint/lintCompatBatch` LSP custom requests
 *     coming from the Go LSP server.
 *
 * Neither consumer spawns a separate "sidecar" process — there's no
 * stdin/stdout sidecar entry in this package anymore. Plugin execution
 * always happens inside the Node process that already owns the user's
 * rslint config (CLI: shim process; LSP: extension host).
 */

// IPC layer (still used by the CLI path: cli.ts ↔ engine.ts ↔ Go binary)
export { IpcClient, encodeFrame, decodeFrame } from './ipc-client.js';
export type {
  MessageKind,
  IpcMessage,
  ErrorResponseData,
  InboundRequestHandler,
  NotificationHandler,
  ConfigDescriptor,
} from './types.js';

// Worker pool
export { WorkerPool } from './worker-pool.js';
export {
  WorkerClosedError,
  isWorkerClosedError,
  WORKER_CLOSED_CODE,
} from './errors.js';
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
export {
  selectPluginSource,
  configEntryHasEslintPlugins,
} from './plugin-source.js';

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
