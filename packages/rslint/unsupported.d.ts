/** These legacy entrypoints reject; use Rslint or runCLI from @rslint/core. */
export declare class RSLintService {
  constructor(...args: unknown[]);
}
export declare class NodeRslintService {
  constructor(...args: unknown[]);
}
export declare class WorkerPool {
  constructor(...args: unknown[]);
}
export declare function lint(...args: unknown[]): Promise<never>;
export declare function readAttachmentBytes(...args: unknown[]): never;
export declare function createPluginLintHost(...args: unknown[]): never;
export declare function lintFile(...args: unknown[]): never;
export declare function loadPluginsFromConfigs(...args: unknown[]): never;
export declare function loadPluginsFromConfigFile(...args: unknown[]): never;
export declare function buildPluginLintTasks(...args: unknown[]): never;
export declare function buildPluginLintResult(...args: unknown[]): never;
