/* rslint-disable @typescript-eslint/no-explicit-any */
/**
 * Severity level for a rule.
 */
export type RuleSeverity = 'off' | 'warn' | 'error';

/**
 * Source of truth for the rule prefixes owned by rslint's built-in
 * (natively-ported) plugins; the `KnownPlugin` type union derives from it.
 * `NATIVE_PLUGIN_RESERVED_NAMES` unions these prefixes with the alternate
 * `eslint-plugin-*` declaration names (`NATIVE_PLUGIN_DECL_ALIASES`), so a
 * ported plugin that also has such an alias must be added to BOTH lists here —
 * kept in sync with config.go's PluginInfo.DeclNames (a Go test guards the drift).
 */
const NATIVE_PLUGINS = [
  '@typescript-eslint',
  'import',
  'jest',
  'jsx-a11y',
  'promise',
  'react',
  'react-hooks',
  'unicorn',
] as const;

// Alternate `eslint-plugin-*` declaration names that Go normalizes onto a
// native prefix (mirrors config.go's PluginInfo.DeclNames). A community plugin
// must not be mounted under one of these either: Go would normalize the key
// onto the native prefix, and the gate — which keys off the un-normalized
// `<prefix>/<rule>` — would then silently drop the community rules.
const NATIVE_PLUGIN_DECL_ALIASES = [
  'eslint-plugin-import',
  'eslint-plugin-jest',
  'eslint-plugin-jsx-a11y',
  'eslint-plugin-promise',
  'eslint-plugin-react-hooks',
  'eslint-plugin-unicorn',
] as const;

/**
 * Plugin declaration names recognized by rslint's loader.
 */
export type KnownPlugin = (typeof NATIVE_PLUGINS)[number];

/**
 * Names reserved by rslint's built-in (natively-ported) plugins: the rule
 * prefixes AND the alternate `eslint-plugin-*` declaration names Go normalizes
 * onto them. A community plugin mounted under an object-form `plugins` key may
 * not collide with any of these — native rules always win, and Go would
 * normalize an aliased key onto a native prefix and the gate would then silently
 * drop the community rules. Typed as ReadonlySet<string> so callers can probe
 * arbitrary user-supplied strings.
 */
export const NATIVE_PLUGIN_RESERVED_NAMES: ReadonlySet<string> = new Set([
  ...NATIVE_PLUGINS,
  ...NATIVE_PLUGIN_DECL_ALIASES,
]);

/**
 * Rule-specific options object. Each rule defines its own shape; until per-rule
 * types are generated, options are accepted as an open record.
 */
export type RuleOptions = Record<string, any>;

/**
 * Configuration value accepted for a single rule. Aligned with ESLint —
 * rslint's own `{ level, options }` object form has been removed.
 *
 * - `RuleSeverity` — just toggle the rule.
 * - `[RuleSeverity, ...args]` — ESLint-style array form. Most rules take a
 *   single options object (`[severity, { ... }]`); some accept positional
 *   string/object args (`[severity, "always", { ... }]`).
 */
export type RuleEntry = RuleSeverity | readonly [RuleSeverity, ...any[]];

/**
 * Map of rule name → rule configuration. Rule names are `string` (no
 * enumeration of known rules yet); the value shape is what gives editors
 * hints when typing the array or object form.
 */
export type RulesRecord = Record<string, RuleEntry>;

/**
 * TypeScript parser options. `project` may be a single tsconfig path or a list.
 */
export interface ParserOptions {
  /**
   * Enable project service for typed linting (runs the TypeScript language
   * service behind the scenes).
   */
  projectService?: boolean;
  /**
   * tsconfig.json path(s) used for typed linting. Glob patterns are supported.
   *
   * @example
   * project: './tsconfig.json'
   * @example
   * project: ['./tsconfig.app.json', './tsconfig.node.json']
   * @example
   * project: ['./tsconfig.*.json']
   */
  project?: string | string[];
}

/**
 * Access level for a declared global variable.
 */
export type GlobalAccess = boolean | 'readonly' | 'writable' | 'off';

/**
 * Map of global variable name to its access level.
 *
 * @example
 * globals: { myGlobal: 'readonly' }
 */
export type GlobalsConfig = Record<string, GlobalAccess>;

/**
 * Language-specific configuration.
 */
export interface LanguageOptions {
  parserOptions?: ParserOptions;
  /**
   * Global variables available in this file's scope, e.g. from a browser
   * or Node.js runtime. `'readonly'`/`true` allows reading; `'writable'`
   * allows reassignment. Only the string `'off'` un-declares a global
   * (undoes one a base config added) — `false` still declares it (as
   * read-only), matching ESLint's own `normalizeConfigGlobal`.
   *
   * @example
   * globals: { myGlobal: 'readonly' }
   */
  globals?: GlobalsConfig;
}

/**
 * A real ESLint plugin object, as exported by community packages
 * (`eslint-plugin-unicorn`, etc.). Only the fields rslint consumes are
 * typed; the open index keeps arbitrary plugin shapes assignable.
 */
export interface ESLintPlugin {
  meta?: { name?: string; version?: string };
  name?: string;
  rules?: Record<string, unknown>;
  configs?: Record<string, unknown>;
  [key: string]: unknown;
}

/**
 * A single entry in an rslint config array. Multiple entries may target
 * different file globs and are merged at lint time.
 */
export interface RslintConfigEntry {
  /** Optional human-readable name for this config entry. */
  name?: string;
  /**
   * Glob selectors for files this entry applies to. Top-level selectors are
   * ORed; strings inside one nested array are ANDed, matching ESLint flat
   * config semantics.
   *
   * @example
   * files: ['src/**', 'tests/**']
   */
  files?: Array<string | string[]>;
  /**
   * Glob patterns excluded from this entry.
   *
   * @example
   * ignores: ['node_modules/**', 'dist/**']
   */
  ignores?: string[];
  /** Language-level configuration (parser, etc.). */
  languageOptions?: LanguageOptions;
  /**
   * Plugins enabled for this entry. Two forms:
   *
   * - **Array of names** — built-in (natively-ported) plugins, e.g.
   *   `plugins: ['@typescript-eslint', 'unicorn']`. Built-in names are listed
   *   for autocomplete; arbitrary strings are still accepted. Each built-in
   *   maps to the ESLint plugin it ports rules from:
   *   `'@typescript-eslint'` → `@typescript-eslint/eslint-plugin`,
   *   `'import'` → `eslint-plugin-import`, `'jest'` → `eslint-plugin-jest`,
   *   `'jsx-a11y'` → `eslint-plugin-jsx-a11y`, `'promise'` → `eslint-plugin-promise`,
   *   `'react'` → `eslint-plugin-react`, `'react-hooks'` → `eslint-plugin-react-hooks`,
   *   `'unicorn'` → `eslint-plugin-unicorn`.
   *
   * - **Object of plugin instances** — community ESLint plugins mounted by
   *   prefix, e.g. `{ unicorn }` after `import unicorn from 'eslint-plugin-unicorn'`.
   *   Their JS rule functions run in a Node worker; only `{prefix, ruleNames}`
   *   metadata reaches the Go core. The live objects never cross the wire — the
   *   worker re-imports this config file to obtain them, so local-path and
   *   monorepo-versioned plugins resolve correctly. A prefix may not collide
   *   with a built-in plugin name.
   *
   * A single entry uses one form. To combine built-in and community plugins,
   * declare them in separate config entries (merged at lint time).
   *
   * @example
   * plugins: ['@typescript-eslint', 'unicorn']
   * @example
   * import unicorn from 'eslint-plugin-unicorn';
   * export default [{ plugins: { unicorn }, rules: { 'unicorn/no-null': 'error' } }];
   */
  plugins?: (KnownPlugin | (string & {}))[] | Record<string, ESLintPlugin>;
  /** Shared settings accessible to rules. */
  settings?: Record<string, any>;
  /** Rule configuration map. */
  rules?: RulesRecord;
}

/** Top-level rslint config: an array of entries. */
export type RslintConfig = RslintConfigEntry[];

/**
 * Type-safe config helper. Returns the config array as-is (identity function).
 */
export function defineConfig(config: RslintConfig): RslintConfig {
  return config;
}

/**
 * Define a global-ignores config entry.
 *
 * Mirrors ESLint's `globalIgnores` helper: returns a config entry that contains
 * only `ignores`. Because the entry has no `files` (and no rules/plugins/etc.),
 * the patterns are treated as *global* ignores — applied across every other
 * config entry — instead of being scoped to a single entry's `files`.
 *
 * @example
 * export default defineConfig([
 *   globalIgnores(['dist/**', 'coverage/**']),
 *   { files: ['src/**'], rules: { 'no-console': 'error' } },
 * ]);
 */
export function globalIgnores(ignorePatterns: string[]): RslintConfigEntry {
  if (!Array.isArray(ignorePatterns)) {
    throw new TypeError('ignorePatterns must be an array');
  }
  if (ignorePatterns.length === 0) {
    throw new TypeError('ignorePatterns must contain at least one pattern');
  }
  return { ignores: ignorePatterns };
}
