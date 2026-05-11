/* rslint-disable @typescript-eslint/no-explicit-any */
/**
 * Severity level for a rule.
 */
export type RuleSeverity = 'off' | 'warn' | 'error';

/**
 * Plugin declaration names recognized by rslint's loader.
 */
export type KnownPlugin =
  | '@stylistic'
  | '@typescript-eslint'
  | 'import'
  | 'jest'
  | 'jsx-a11y'
  | 'promise'
  | 'react'
  | 'react-hooks'
  | 'unicorn';

/**
 * Runtime list of rslint's built-in (Go-side ported) plugin namespaces.
 * Must stay in sync with the {@link KnownPlugin} union AND with the
 * native registrations in `internal/config/config.go::RegisterAllRules`.
 *
 * Source-of-truth pairing is enforced two ways:
 *   1. The `satisfies` clause + the `_NativePluginsExhaustive` type-level
 *      check below: if a value is added to `KnownPlugin` without a
 *      matching string here, TypeScript fails compilation.
 *   2. The Go-side test
 *      `internal/config/config_native_prefixes_test.go` parses this file
 *      and asserts the array values equal the set of prefixes the
 *      Go `GlobalRuleRegistry` registers for plugin rules. A drift in
 *      either direction makes CI red.
 */
export const NATIVE_PLUGIN_PREFIXES = [
  '@stylistic',
  '@typescript-eslint',
  'import',
  'jest',
  'jsx-a11y',
  'promise',
  'react',
  'react-hooks',
  'unicorn',
] as const satisfies readonly KnownPlugin[];

// Compile-time guard: every `KnownPlugin` literal MUST appear in
// `NATIVE_PLUGIN_PREFIXES`. If the union grows without updating the
// array, `_Missing` resolves to the missing literal(s) and the
// `_NativePluginsExhaustive` constant fails to type-check.
type _Missing = Exclude<KnownPlugin, (typeof NATIVE_PLUGIN_PREFIXES)[number]>;
const _NativePluginsExhaustive: [_Missing] extends [never]
  ? true
  : ['NATIVE_PLUGIN_PREFIXES missing entries for:', _Missing] = true;
// eslint-disable-next-line @typescript-eslint/no-unused-expressions
_NativePluginsExhaustive;

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
 * Shape of an ESLint plugin instance, as accepted in `eslintPlugins`.
 *
 * Matches the ESLint plugin API: a plugin is a JS object with an optional
 * `meta` (carrying `name`/`version`) and a `rules` map. `configs` and
 * `processors` are accepted but not interpreted by rslint's runner today.
 *
 * Why a structural type and not a wider type: under the configs-flow
 * design the worker imports the user's `rslint.config.*` directly and
 * harvests plugin instances from `entry.eslintPlugins` — `meta.name`
 * is NOT used as an npm specifier. The structural minimum (`rules` and
 * optional `meta`) is required so the host's `normalizeConfig` can
 * enumerate rule names and ship a stable wire payload to Go.
 *
 * @experimental
 * The ESLint plugin compatibility surface is experimental. We support the
 * most common ESLint plugin authoring conventions but do NOT promise full
 * `SourceCode` / `RuleContext` parity — see the conformance matrix in
 * `packages/rslint-test-tools/src/eslint-conformance.ts` for the rules we
 * actively verify against upstream ESLint. APIs and runtime semantics may
 * change between minor releases without a deprecation cycle until the
 * surface stabilises.
 */
export interface ESLintPluginShape {
  meta?: {
    name?: string;
    version?: string;
  };
  /** Plugin display name. Currently accepted for type compatibility
   *  but not consumed by rslint — the worker imports plugin instances
   *  through the user config (configs-flow), not by resolving a name. */
  name?: string;
  /** Map of rule name → rule definition. Required for the plugin to contribute rules. */
  rules?: Record<string, unknown>;
  /** Plugin-provided preset configs. Accepted but not auto-applied; spread manually. */
  configs?: Record<string, unknown>;
  /** Plugin-provided processors. Not supported by rslint's plugin runtime;
   *  the field is accepted for type compatibility but silently ignored. */
  processors?: Record<string, unknown>;
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
   * Plugin names to enable for this entry. Built-in plugins are listed for
   * autocomplete; arbitrary strings are still accepted so future/third-party
   * plugins don't trip the type checker.
   *
   * Each built-in value maps to the original ESLint plugin it ports rules from:
   *
   * - `'@stylistic'`         → `@stylistic/eslint-plugin`
   * - `'@typescript-eslint'` → `@typescript-eslint/eslint-plugin`
   * - `'import'`             → `eslint-plugin-import`
   * - `'jest'`               → `eslint-plugin-jest`
   * - `'jsx-a11y'`           → `eslint-plugin-jsx-a11y`
   * - `'promise'`            → `eslint-plugin-promise`
   * - `'react'`              → `eslint-plugin-react`
   * - `'react-hooks'`        → `eslint-plugin-react-hooks`
   * - `'unicorn'`            → `eslint-plugin-unicorn`
   */
  plugins?: (KnownPlugin | (string & {}))[];
  /**
   * ESLint plugins to run inside rslint's Node-side plugin runtime.
   *
   * Orthogonal to `plugins` (which gates native rules by name). Each key is
   * the user-chosen rule namespace; the value is a loaded ESLint plugin
   * instance. When any entry is present in any matching config, rslint
   * dispatches the matching `<prefix>/<ruleName>` rules to a Node
   * WorkerPool — hosted by the CLI's Node parent (`@rslint/core`) or
   * by the LSP client (the VS Code extension) — and merges results
   * back into the diagnostic stream.
   *
   * Native rules with the same fully-qualified name take precedence; a
   * one-line stderr warning is emitted on shadow.
   *
   * @experimental
   * This API is experimental. rslint targets compatibility with the
   * mainstream ESLint plugin surface (the bulk of `SourceCode` /
   * `RuleContext` / fixer / suggestion APIs that the most-used community
   * plugins rely on), verified by the conformance harness in
   * `packages/rslint-test-tools`. We do NOT claim 1:1 parity for every
   * niche ESLint API. Type-aware rules that need a TypeScript
   * `TypeChecker` (`parserServices.program`) are out of scope. Behavior
   * may change in minor releases until the surface stabilises; if you
   * adopt this, pin the rslint version and watch CHANGELOG.
   *
   * @example
   * import unicornPlugin from 'eslint-plugin-unicorn';
   * export default defineConfig([
   *   {
   *     eslintPlugins: { uc: unicornPlugin },
   *     rules: { 'uc/no-null': 'error' },
   *   },
   * ]);
   */
  eslintPlugins?: Record<string, ESLintPluginShape>;
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
