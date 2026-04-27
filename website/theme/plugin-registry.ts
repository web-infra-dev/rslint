/**
 * Metadata for a single plugin. The file declares only string identifiers
 * (no @rslint/core import), so it can be loaded both at build time by Node
 * scripts and at runtime by browser UI components.
 */
export interface PluginMeta {
  /**
   * Rule key prefix users write in their rslint config (e.g. `import` for
   * `import/no-self-import`). Empty string for the eslint core group, whose
   * rules carry no prefix. After `@` stripping this also serves as the doc
   * route slug — see {@link groupToRouteSlug}.
   */
  prefix: string;
  /**
   * Join key against `rule.group` in `website/generated/rule-manifest.json`,
   * which mirrors the `PLUGIN_NAME` constant from
   * `internal/plugins/<plugin>/plugin.go` (or the literal `"eslint"` for
   * core rules). Used to look up the prefix / preset / import name that
   * belongs to a manifest entry.
   */
  group: string;
  /** Named export from `@rslint/core` (e.g. `importPlugin`). */
  importName: string;
  /**
   * Dot-path to the preset config object, e.g. `reactPlugin.configs.recommended`.
   * `null` if the plugin ships no preset.
   */
  presetName: string | null;
  /** Human-readable preset description, shown in the docs preset table. */
  description: string;
}

/**
 * Single source of truth for every supported plugin. Build-time
 * (`plugin-rule-manifest.ts`) and runtime UI (`RuleConfig.tsx`,
 * `PresetTable.tsx`, `PluginPrefixTable.tsx`, `RuleStates/rule.tsx`)
 * all derive their local mappings from this list — add a plugin here
 * once and every consumer picks it up automatically.
 */
export const PLUGIN_REGISTRY: PluginMeta[] = [
  {
    prefix: '',
    group: 'eslint',
    importName: 'js',
    presetName: 'js.configs.recommended',
    description: 'JavaScript recommended rules',
  },
  {
    prefix: '@typescript-eslint',
    group: '@typescript-eslint',
    importName: 'ts',
    presetName: 'ts.configs.recommended',
    description: 'TypeScript recommended rules (includes ESLint core rules)',
  },
  {
    prefix: 'react',
    group: 'react',
    importName: 'reactPlugin',
    presetName: 'reactPlugin.configs.recommended',
    description: 'React rules',
  },
  {
    prefix: 'react-hooks',
    group: 'eslint-plugin-react-hooks',
    importName: 'reactHooksPlugin',
    presetName: 'reactHooksPlugin.configs.recommended',
    description: 'React Hooks rules',
  },
  {
    prefix: 'import',
    group: 'eslint-plugin-import',
    importName: 'importPlugin',
    presetName: 'importPlugin.configs.recommended',
    description: 'Import/export rules',
  },
  {
    prefix: 'promise',
    group: 'eslint-plugin-promise',
    importName: 'promisePlugin',
    presetName: 'promisePlugin.configs.recommended',
    description: 'Promise rules',
  },
  {
    prefix: 'jest',
    group: 'eslint-plugin-jest',
    importName: 'jestPlugin',
    presetName: 'jestPlugin.configs.recommended',
    description: 'Jest rules',
  },
  {
    prefix: 'unicorn',
    group: 'eslint-plugin-unicorn',
    importName: 'unicornPlugin',
    presetName: 'unicornPlugin.configs.recommended',
    description: 'Unicorn rules',
  },
];

/**
 * Convert a manifest plugin group (the value of `PLUGIN_NAME` in
 * `internal/plugins/<plugin>/plugin.go`) to the slug used as the
 * `/rules/<slug>/` URL segment, the sidebar directory name, and the
 * sidebar label.
 *
 * Resolution order:
 *   1. Look up the matching {@link PLUGIN_REGISTRY} entry and reuse its
 *      `prefix` (the same identifier users write in their rslint config,
 *      e.g. `import` / `jest` / `react-hooks`). This keeps doc URLs,
 *      sidebar labels, and rule keys in sync.
 *   2. Fall back to the raw `group` for unregistered plugins or for
 *      entries whose `prefix` is empty (the eslint core group).
 *   3. Strip a leading `@` so scoped names like `@typescript-eslint`
 *      produce `typescript-eslint`.
 */
export function groupToRouteSlug(group: string): string {
  const meta = PLUGIN_REGISTRY.find((p) => p.group === group);
  return (meta?.prefix || group).replace(/^@/, '');
}
