import { afterAll, beforeAll, describe, test } from 'rstack/test';
import { lint } from '@rslint/core/internal';
import {
  checkExpectedError,
  checkInvalidCase,
  checkSkip,
  checkValidCase,
} from './cases.ts';
import { checkOutput, checkResult } from './diagnostics.ts';
import { createFixtureRoot, removeFixtureRoot } from './fixture.ts';
import { buildLintRequest } from './request.ts';
import type { RuleTesterConfig, TestCase, ValidTestCase } from './types.ts';

export type {
  ExpectedError,
  ExpectedSuggestion,
  RuleTesterConfig,
  TestCase,
  ValidTestCase,
} from './types.ts';

export class RuleTester {
  constructor(private readonly config: RuleTesterConfig) {}

  run(
    name: string,
    cases: {
      /** A string is shorthand for `{ code }`. */
      valid: (string | ValidTestCase)[];
      invalid: TestCase[];
    },
  ): void {
    const ruleName = `${this.config.plugin}/${name}`;
    describe(ruleName, () => {
      let root = '';
      beforeAll(async () => {
        root = await createFixtureRoot(this.config.fixtureFiles ?? {});
      });
      afterAll(async () => {
        if (root) await removeFixtureRoot(root);
      });
      const groups: [string, (ValidTestCase | TestCase)[]][] = [
        [
          'valid',
          cases.valid.map((item) =>
            typeof item === 'string' ? { code: item } : item,
          ),
        ],
        ['invalid', cases.invalid],
      ];
      for (const [kind, entries] of groups) {
        entries.forEach((item, index) => {
          const run = item.skip ? test.skip : test;
          const label = item.skip ?? item.name;
          run(`${kind} ${index}${label ? `: ${label}` : ''}`, async () => {
            checkSkip(item);
            if (kind === 'valid') checkValidCase(item);
            else checkInvalidCase(item);
            const errors = item.errors ?? [];
            errors.forEach(checkExpectedError);
            const { request, filename } = buildLintRequest(
              this.config,
              ruleName,
              root,
              item,
            );
            checkResult(await lint(request), errors, {
              code: item.code,
              ruleName,
              root,
              hasOutput: item.output !== undefined,
            });
            if (item.output !== undefined) {
              const fixed = await lint({ ...request, fix: true });
              checkOutput(fixed.output, root, filename, item.output);
            }
          });
        });
      }
    });
  }
}
