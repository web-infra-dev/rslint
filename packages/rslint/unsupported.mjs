/** Compatibility names for interfaces outside the Rust+JS rsbuild migration. */
function unavailable(name) {
  const error = new Error(
    `${name} is not supported by the Rust+JS rsbuild migration. ` +
      'Use Rslint or runCLI from @rslint/core with the supported rules. ' +
      'No Go service, IPC bridge, or plugin worker fallback is available.',
  );
  error.code = 'ERR_RSLINT_UNSUPPORTED_ENTRYPOINT';
  throw error;
}
export class RSLintService {
  constructor() {
    unavailable('RSLintService');
  }
}
export class NodeRslintService {
  constructor() {
    unavailable('NodeRslintService');
  }
}
export class WorkerPool {
  constructor() {
    unavailable('WorkerPool');
  }
}
export async function lint() {
  unavailable('The low-level lint service');
}
export function readAttachmentBytes() {
  unavailable('IPC attachment access');
}
export function createPluginLintHost() {
  unavailable('Community plugin hosts');
}
export function lintFile() {
  unavailable('Community plugin linting');
}
export function loadPluginsFromConfigs() {
  unavailable('Community plugin loading');
}
export function loadPluginsFromConfigFile() {
  unavailable('Community plugin loading');
}
export function buildPluginLintTasks() {
  unavailable('Community plugin tasks');
}
export function buildPluginLintResult() {
  unavailable('Community plugin results');
}
