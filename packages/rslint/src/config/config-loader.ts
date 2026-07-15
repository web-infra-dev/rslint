/**
 * Stable public facade for `@rslint/core/config-loader`.
 *
 * ConfigModuleHost imports the config-file evaluation and normalization
 * helpers directly so this facade can expose both APIs without an ESM cycle.
 */
export {
  collectPluginMeta,
  loadConfigFile,
  loadConfigFileFresh,
  normalizeConfig,
  type PluginConfigDescriptor,
} from './config-file-loader.js';
export { ConfigModuleHost } from './config-module-host.js';
export type {
  ConfigModuleActivationPlan,
  ConfigModuleHostOptions,
  ConfigModulePluginDescriptor,
  EffectiveConfigModule,
} from './config-module-host.js';
export {
  CONFIG_DISCOVERY_PROTOCOL_VERSION,
  type ActivateConfigsRequest,
  type ActivateConfigsResponse,
  type ConfigModuleCandidate,
  type ConfigModuleEslintPluginEntry,
  type ConfigModuleLoadMode,
  type ConfigModuleLoadResult,
  type FailedConfigModuleResult,
  type LoadConfigsRequest,
  type LoadConfigsResponse,
  type LoadedConfigModuleResult,
} from './config-discovery-protocol.js';
export {
  filterConfigsByParentIgnores,
  type ConfigEntry,
} from './config-hierarchy.js';
