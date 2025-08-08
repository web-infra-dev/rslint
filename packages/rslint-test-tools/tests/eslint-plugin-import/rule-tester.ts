import path from 'node:path';
import util from 'node:util';

import type { RuleTester as ESLintRuleTester } from 'eslint';

import { lint } from '@rslint/core';

// Port from https://github.com/eslint/eslint/blob/34f0723e2d0faf8ac8dc95ec56e6d181bd6b67f2/lib/rule-tester/rule-tester.js#L1145

export class RuleTester {
  run(
    ruleName: string,
    rule: never,
    cases: {
      valid: ESLintRuleTester.ValidTestCase[];
      invalid: ESLintRuleTester.InvalidTestCase[];
    },
  ) {
    ruleName = 'import/' + ruleName;
    describe(ruleName, () => {
      const cwd = process.cwd();
      const config = path.resolve(import.meta.dirname, './rslint.json');

      // test whether case has only
      let hasOnly =
        cases.valid.some(x => {
          if (typeof x === 'object' && x.only) {
            return true;
          } else {
            return false;
          }
        }) || cases.invalid.some(x => x.only);

      test('valid', async () => {
        for (const validCase of cases.valid) {
          if (hasOnly) {
            if (typeof validCase === 'string') {
              continue;
            }
            if (!validCase.only) {
              continue;
            }
          }
          const code =
            typeof validCase === 'string' ? validCase : validCase.code;

          const options =
            typeof validCase === 'string' ? [] : validCase.options || [];
          const defaultFilename = 'src/virtual.ts';
          const filename =
            typeof validCase === 'string'
              ? defaultFilename
              : (validCase.filename ?? defaultFilename);

          const diags = await lint({
            config,
            workingDirectory: cwd,
            fileContents: {
              [filename]: code,
            },
            ruleOptions: {
              [ruleName]: options,
            },
          });

          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)}`,
          );
        }
      });
      test('invalid', async t => {
        for (const item of cases.invalid) {
          assert.ok(
            item.errors || item.errors === 0,
            `Did not specify errors for an invalid test of ${ruleName}`,
          );
          if (Array.isArray(item.errors) && item.errors.length === 0) {
            assert.fail('Invalid cases must have at least one error');
          }

          const { code, only = false, options = [] } = item;
          if (hasOnly && !only) {
            continue;
          }
          const defaultFilename = 'src/virtual.ts';
          const filename =
            typeof item === 'string'
              ? defaultFilename
              : (item.filename ?? defaultFilename);

          const diags = await lint({
            config,
            workingDirectory: cwd,
            fileContents: {
              [filename]: code,
            },
            ruleOptions: {
              [ruleName]: options,
            },
          });

          if (typeof item.errors === 'number') {
            if (item.errors === 0) {
              assert.fail(
                "Invalid cases must have 'error' value greater than 0",
              );
            }

            assert.strictEqual(
              diags.diagnostics.length,
              item.errors,
              util.format(
                'Should have %d error%s but had %d: %s',
                item.errors,
                item.errors === 1 ? '' : 's',
                diags.diagnostics.length,
                util.inspect(diags.diagnostics),
              ),
            );
          } else {
            assert.strictEqual(
              diags.diagnostics.length,
              item.errors.length,
              util.format(
                'Should have %d error%s but had %d: %s',
                item.errors.length,
                item.errors.length === 1 ? '' : 's',
                diags.diagnostics.length,
                util.inspect(diags.diagnostics),
              ),
            );

            const hasMessageOfThisRule = diags.diagnostics.some(
              d => d.ruleName === ruleName,
            );

            for (let i = 0, l = item.errors.length; i < l; i++) {
              const error = item.errors[i];
              const message = diags.diagnostics[i];

              assert(
                hasMessageOfThisRule,
                'Error rule name should be the same as the name of the rule being tested',
              );
              if (typeof error === 'string' || error instanceof RegExp) {
                // Just an error message.
                assertMessageMatches(message.message, error);
                assert.ok(
                  message.suggestions === void 0,
                  `Error at index ${i} has suggestions. Please convert the test error into an object and specify 'suggestions' property on it to test suggestions.`,
                );
              } else if (typeof error === 'object' && error !== null) {
                // TODO: handle object error
                // https://github.com/eslint/eslint/blob/34f0723e2d0faf8ac8dc95ec56e6d181bd6b67f2/lib/rule-tester/rule-tester.js#L1145
              }
            }
          }
        }
      });
    });
  }
}

/**
 * Asserts that the message matches its expected value. If the expected
 * value is a regular expression, it is checked against the actual
 * value.
 */
function assertMessageMatches(actual: string, expected: string | RegExp): void {
  if (expected instanceof RegExp) {
    // assert.js doesn't have a built-in RegExp match function
    assert.ok(
      expected.test(actual),
      `Expected '${actual}' to match ${expected}`,
    );
  } else {
    assert.strictEqual(actual, expected);
  }
}
