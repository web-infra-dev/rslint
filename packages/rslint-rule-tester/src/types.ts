import type { LanguageOptions } from '@rslint/core';

/** Preserves upstream parser metadata; JSX parsing follows the filename. */
type TestLanguageOptions = Omit<LanguageOptions, 'parserOptions'> & {
  parserOptions?: LanguageOptions['parserOptions'] & {
    ecmaFeatures?: { jsx?: boolean };
  };
};

export interface ExpectedSuggestion {
  messageId: string;
  /** The rendered suggestion message. */
  desc: string;
  /** The source after applying only this suggestion. */
  output: string;
}

export interface ExpectedError {
  messageId: string;
  /** A function receives the fixture root with forward slashes. */
  message: string | ((root: string) => string);
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
  /** The diagnostic's only fix edit. */
  fix?: { range: [number, number]; text: string };
  /** Omitted or empty means the diagnostic must have no suggestions. */
  suggestions?: ExpectedSuggestion[];
}

export interface TestCase {
  name?: string;
  /** Reason for skipping this case. */
  skip?: string;
  code: string;
  /** Path relative to the fixture root. Defaults to the tester's `filename`. */
  filename?: string;
  /** A function receives the absolute fixture root. */
  options?: unknown[] | ((root: string) => unknown[]);
  settings?: Record<string, unknown>;
  languageOptions?: TestLanguageOptions;
  /** Required, and non-empty, for invalid cases. */
  errors: ExpectedError[];
  /** The source after applying every autofix. */
  output?: string;
}

/** A valid case expects no diagnostics, so it carries no expectations. */
export interface ValidTestCase extends Omit<TestCase, 'errors' | 'output'> {
  errors?: never;
  output?: never;
}

/** Builds a complete case after the tester has created its fixture root. */
export interface FixtureTestCase<T extends ValidTestCase | TestCase> {
  name?: string;
  /** Known at registration time; skipped cases never call create. */
  skip?: string;
  create(root: string): Omit<T, 'name' | 'skip'> & {
    name?: never;
    skip?: never;
    create?: never;
  };
}

export interface RuleTesterConfig {
  /** Plugin name and rule-name prefix. Omit or leave empty for ESLint core rules. */
  plugin?: string;
  /** Default for each case's `filename`. Defaults to `input.js`. */
  filename?: string;
  /** Files written to the fixture root before the cases run. */
  fixtureFiles?: Record<string, string>;
  languageOptions?: TestLanguageOptions;
}
