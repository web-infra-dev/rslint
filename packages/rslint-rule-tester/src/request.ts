import path from 'node:path';
import type { LintOptions } from '@rslint/core/internal';
import type { RuleTesterConfig, TestCase, ValidTestCase } from './types.ts';

/** Builds the lint request that runs one rule over one case. */
export function buildLintRequest(
  config: RuleTesterConfig,
  ruleName: string,
  root: string,
  item: ValidTestCase | TestCase,
): { request: LintOptions; filename: string } {
  const { languageOptions } = config;
  const filename = path.join(root, item.filename ?? 'input.js');
  const options =
    typeof item.options === 'function'
      ? item.options(root)
      : (item.options ?? []);
  const request: LintOptions = {
    configDirectory: root,
    workingDirectory: root,
    config: [
      {
        plugins: [config.plugin],
        languageOptions: {
          ...languageOptions,
          ...item.languageOptions,
          globals: {
            ...languageOptions?.globals,
            ...item.languageOptions?.globals,
          },
          parserOptions: {
            projectService: false,
            ...languageOptions?.parserOptions,
            ...item.languageOptions?.parserOptions,
          },
        },
        settings: item.settings,
        rules: { [ruleName]: ['error', ...options] },
      },
    ],
    fileContents: { [filename]: item.code },
  };
  return { request, filename };
}
