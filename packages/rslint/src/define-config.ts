/** Glob patterns for files this config applies to */
export interface RslintConfigEntry {
  files?: string[];
  ignores?: string[];
  languageOptions?: {
    parserOptions?: {
      projectService?: boolean;
      project?: string | string[];
    };
  };
  rules?: Record<string, RuleSeverity | [RuleSeverity, ...unknown[]]>;
  plugins?: string[];
  settings?: Record<string, unknown>;
}

type RuleSeverity = 'off' | 'warn' | 'error';

/**
 * Type-safe config helper. Returns the config array as-is (identity function).
 */
export function defineConfig(config: RslintConfigEntry[]): RslintConfigEntry[] {
  return config;
}
