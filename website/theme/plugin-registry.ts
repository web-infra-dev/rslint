/**
 * Metadata for a single plugin. All fields are plain strings so this file
 * can be imported by both Node (build scripts) and the browser (UI components)
 * without pulling in @rslint/core at runtime.
 */
export interface PluginMeta {
  /** Rule key prefix in rslint config, e.g. "react" → "react/jsx-key".
   *  Empty string for core ESLint rules (no prefix). */
  prefix: string;
  /** Internal group name used in rule-manifest.json, e.g. "eslint-plugin-import". */
  group: string;
  /** Named export from @rslint/core used in import statements, e.g. "importPlugin". */
  importName: string;
  /** Dot-path to the preset config object, e.g. "reactPlugin.configs.recommended".
   *  null if the plugin ships no preset. */
  presetName: string | null;
}

/**
 * Single source of truth for every supported plugin.
 * Both plugin-rule-manifest.ts (build-time) and RuleConfig.tsx (runtime UI)
 * derive their local mappings from this list — add a plugin here once and
 * both places pick it up automatically.
 */
export const PLUGIN_REGISTRY: PluginMeta[] = [
  {
    prefix: '',
    group: 'eslint',
    importName: 'js',
    presetName: 'js.configs.recommended',
  },
  {
    prefix: '@typescript-eslint',
    group: '@typescript-eslint',
    importName: 'ts',
    presetName: 'ts.configs.recommended',
  },
  {
    prefix: 'react',
    group: 'react',
    importName: 'reactPlugin',
    presetName: 'reactPlugin.configs.recommended',
  },
  {
    prefix: 'import',
    group: 'eslint-plugin-import',
    importName: 'importPlugin',
    presetName: 'importPlugin.configs.recommended',
  },
  {
    prefix: 'promise',
    group: 'eslint-plugin-promise',
    importName: 'promisePlugin',
    presetName: 'promisePlugin.configs.recommended',
  },
  {
    prefix: 'jest',
    group: 'eslint-plugin-jest',
    importName: 'jestPlugin',
    presetName: 'jestPlugin.configs.recommended',
  },
];
