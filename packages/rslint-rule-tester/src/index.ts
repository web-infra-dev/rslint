import assert from 'node:assert/strict';
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
import type {
  FixtureTestCase,
  RuleTesterConfig,
  TestCase,
  ValidTestCase,
} from './types.ts';

export type {
  ExpectedError,
  FixtureTestCase,
  ExpectedSuggestion,
  RuleTesterConfig,
  TestCase,
  ValidTestCase,
} from './types.ts';

export class RuleTester {
  constructor(private readonly config: RuleTesterConfig = {}) {}

  run(
    name: string,
    cases: {
      /** A string is shorthand for `{ code }`. */
      valid: (string | ValidTestCase | FixtureTestCase<ValidTestCase>)[];
      invalid: (TestCase | FixtureTestCase<TestCase>)[];
    },
  ): void {
    const ruleName = this.config.plugin
      ? `${this.config.plugin}/${name}`
      : name;
    describe(ruleName, () => {
      let root = '';
      beforeAll(async () => {
        root = await createFixtureRoot(this.config.fixtureFiles ?? {});
      });
      afterAll(async () => {
        if (root) await removeFixtureRoot(root);
      });
      const groups: [
        string,
        (
          | ValidTestCase
          | TestCase
          | FixtureTestCase<ValidTestCase>
          | FixtureTestCase<TestCase>
        )[],
      ][] = [
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
          checkSkip(item);
          const run = item.skip ? test.skip : test;
          const label = item.skip ?? item.name;
          run(`${kind} ${index}${label ? `: ${label}` : ''}`, async () => {
            if ('create' in item) {
              assert.deepEqual(
                Object.keys(item).filter(
                  (key) => !['name', 'skip', 'create'].includes(key),
                ),
                [],
                'Fixture case fields must be returned by create(root)',
              );
            }
            const resolved = 'create' in item ? item.create(root) : item;
            if ('create' in item) {
              assert.ok(
                resolved &&
                  typeof resolved === 'object' &&
                  !Array.isArray(resolved),
                'create(root) must return a case object',
              );
              assert.equal(
                typeof resolved.code,
                'string',
                'create(root) must return string code',
              );
              for (const key of ['name', 'skip', 'create']) {
                assert.ok(
                  !Object.hasOwn(resolved, key),
                  `${key} belongs outside create(root)`,
                );
              }
            }
            if (kind === 'valid') checkValidCase(resolved);
            else checkInvalidCase(resolved);
            const errors = resolved.errors ?? [];
            errors.forEach(checkExpectedError);
            const { request, filename } = buildLintRequest(
              this.config,
              ruleName,
              root,
              resolved,
            );
            checkResult(await lint(request), errors, {
              code: resolved.code,
              ruleName,
              root,
              hasOutput: resolved.output !== undefined,
            });
            if (resolved.output !== undefined) {
              const fixed = await lint({ ...request, fix: true });
              checkOutput(fixed.output, root, filename, resolved.output);
            }
          });
        });
      }
    });
  }
}
