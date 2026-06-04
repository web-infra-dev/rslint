/* rslint-disable @typescript-eslint/no-explicit-any */
/**
 * Severity level for a rule.
 */
export type RuleSeverity = 'off' | 'warn' | 'error';

/**
 * Single source of truth for the prefixes owned by rslint's built-in
 * (natively-ported) plugins. The `KnownPlugin` type union and the
 * `NATIVE_PLUGIN_PREFIXES` runtime Set both derive from this, so adding a ported
 * plugin is a one-line change (no second list to keep in sync).
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

/**
 * Plugin declaration names recognized by rslint's loader.
 */
export type KnownPlugin = (typeof NATIVE_PLUGINS)[number];

/**
 * Prefixes owned by rslint's built-in (natively-ported) plugins. A community
 * plugin mounted under an object-form `plugins` key may not collide with these
 * — native rules always win, so a collision would silently shadow the mounted
 * plugin. Typed as ReadonlySet<string> so callers can probe arbitrary
 * user-supplied strings.
 */
export const NATIVE_PLUGIN_PREFIXES: ReadonlySet<string> = new Set(
  NATIVE_PLUGINS,
);

/**
 * Rule-specific options object. Each rule defines its own shape; until per-rule
 * types are generated, options are accepted as an open record.
 */
export type RuleOptions = Record<string, any>;

/**
 * Configuration value accepted for a single rule.
 *
 * - `RuleSeverity` — just toggle the rule.
 * - `[RuleSeverity, ...args]` — ESLint-style array form. Most rules take a
 *   single options object (`[severity, { ... }]`); some accept positional
 *   string/object args (`[severity, "always", { ... }]`).
 * - `{ level, options }` — object form supported by the loader.
 */
export type RuleEntry =
  | RuleSeverity
  | readonly [RuleSeverity, ...any[]]
  | { level: RuleSeverity; options?: RuleOptions };

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
 * Language-specific configuration.
 */
export interface LanguageOptions {
  parserOptions?: ParserOptions;
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
  /**
   * Glob patterns for files this entry applies to.
   *
   * @example
   * files: ['src/**', 'tests/**']
   */
  files?: string[];
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
