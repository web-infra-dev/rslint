import { afterAll, beforeAll, describe, expect, test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import { createTempDir, cleanupTempDir } from '../../cli/js-config/helpers';
import path from 'node:path';

// All three upstream tests and the documentation example at eslint-plugin-n v18.3.0:
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/process-exit-as-throw.js
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/docs/rules/process-exit-as-throw.md
// The rule changes other rules' code paths, so these tests enable both rules.
describe('node/process-exit-as-throw', () => {
  let root: string;
  beforeAll(async () => {
    root = await createTempDir({});
  });
  afterAll(async () => {
    if (root) await cleanupTempDir(root);
  });

  const cases = [
    {
      name: 'unreachable after process.exit',
      consumer: 'no-unreachable',
      enabled: true,
      code: 'foo();\nprocess.exit(1);\nbar();',
      unreachable: true,
    },
    {
      name: 'no effect when disabled',
      consumer: 'no-unreachable',
      enabled: false,
      code: 'foo();\nprocess.exit(1);\nbar();',
    },
    {
      name: 'consistent-return accepts an exiting branch',
      consumer: 'consistent-return',
      enabled: true,
      code: `function foo() {
    if (a) {
        return 1;
    } else {
        process.exit(1);
    }
}`,
    },
    {
      name: 'documentation example',
      consumer: 'consistent-return',
      enabled: true,
      code: `function foo(a) {
    if (a) {
        return new Bar();
    } else {
        process.exit(1);
    }
}`,
    },
  ];

  for (const item of cases) {
    test(item.name, async () => {
      const result = await lint({
        configDirectory: root,
        workingDirectory: root,
        config: [
          {
            plugins: ['node'],
            languageOptions: { parserOptions: { projectService: false } },
            rules: {
              [item.consumer]: 'error',
              'node/process-exit-as-throw': item.enabled ? 'error' : 'off',
            },
          },
        ],
        fileContents: { [path.join(root, 'input.js')]: item.code },
      });
      expect(result.fileCount).toBe(1);
      expect(result.ruleCount).toBe(item.enabled ? 2 : 1);
      expect(result.diagnostics).toHaveLength(item.unreachable ? 1 : 0);
      if (item.unreachable) {
        const diagnostic = result.diagnostics[0];
        expect(diagnostic.ruleName).toBe('no-unreachable');
        expect(diagnostic.severity).toBe('error');
        expect(diagnostic.messageId).toBe('unreachableCode');
        expect(diagnostic.message).toBe('Unreachable code.');
        expect(diagnostic.range).toEqual({
          start: { line: 3, column: 1 },
          end: { line: 3, column: 7 },
        });
        expect(diagnostic.fixes).toBeUndefined();
        expect(diagnostic.suggestions).toBeUndefined();
      }
    });
  }
});
