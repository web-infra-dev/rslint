/* rslint-disable @typescript-eslint/no-explicit-any */
/**
 * Severity level for a rule.
 */
export type RuleSeverity = 'off' | 'warn' | 'error';

/**
 * Plugin declaration names recognized by rslint's loader.
 */
export type KnownPlugin =
  | '@typescript-eslint'
  | 'import'
  | 'jest'
  | 'promise'
  | 'react'
  | 'react-hooks'
  | 'unicorn';

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
   * - `'@typescript-eslint'` → `@typescript-eslint/eslint-plugin`
   * - `'import'`             → `eslint-plugin-import`
   * - `'jest'`               → `eslint-plugin-jest`
   * - `'promise'`            → `eslint-plugin-promise`
   * - `'react'`              → `eslint-plugin-react`
   * - `'react-hooks'`        → `eslint-plugin-react-hooks`
   * - `'unicorn'`            → `eslint-plugin-unicorn`
   */
  plugins?: (KnownPlugin | (string & {}))[];
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
