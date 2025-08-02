// Forked and modified from https://github.com/typescript-eslint/typescript-eslint/blob/16c344ec7d274ea542157e0f19682dd1930ab838/packages/rule-tester/src/RuleTester.ts#L4

import path from 'node:path';
import { test, describe } from 'node:test';
import fs from 'node:fs';
import { fileURLToPath } from 'node:url';
import { lint, type Diagnostic } from '@rslint/core';
import assert from 'node:assert';

interface TsDiagnostic {
  line: number;
  column: number;
  endLine: number;
  endColumn: number;
  messageId: string;
  suggestions: any[];
}
function toCamelCase(name: string): string {
  return name.replace(/-([a-z])/g, g => g[1].toUpperCase());
}

function toMatchSnapshot(received: any, testName: string, caseIndex: number) {
  const serialized = `// Rstest Snapshot v1

exports[\`${testName} > invalid ${caseIndex}\`] = \`
${JSON.stringify(received, null, 2)}
\`;
`;

  // For now, just log the snapshot - in a real implementation you'd save it to a file
  console.log(`Snapshot for ${testName} case ${caseIndex}:`);
  console.log(serialized);
}
// Rules that use kebab-case messageIds and should not be converted to camelCase
const KEBAB_CASE_MESSAGE_ID_RULES = ['consistent-type-assertions'];

// messageIds that should remain in kebab-case for consistent-type-assertions rule
const KEBAB_CASE_MESSAGE_IDS = new Set([
  'angle-bracket',
  'as',
  'never',
  'replaceArrayTypeAssertionWithAnnotation',
  'replaceArrayTypeAssertionWithSatisfies',
  'replaceObjectTypeAssertionWithAnnotation',
  'replaceObjectTypeAssertionWithSatisfies',
  'unexpectedArrayTypeAssertion',
  'unexpectedObjectTypeAssertion',
]);

// check whether rslint diagnostics and typescript-eslint diagnostics are semantic equal
function checkDiagnosticEqual(
  rslintDiagnostic: Diagnostic[],
  tsDiagnostic: TsDiagnostic[],
  ruleName: string,
) {
  assert(
    rslintDiagnostic.length === tsDiagnostic.length,
    `Length mismatch: ${rslintDiagnostic.length} !== ${tsDiagnostic.length}`,
  );
  for (let i = 0; i < rslintDiagnostic.length; i++) {
    const rslintDiag = rslintDiagnostic[i];
    const tsDiag = tsDiagnostic[i];
    // check rule match
    const expectedMessageId =
      KEBAB_CASE_MESSAGE_ID_RULES.includes(ruleName) &&
      KEBAB_CASE_MESSAGE_IDS.has(rslintDiag.messageId)
        ? rslintDiag.messageId // Keep as-is for kebab-case messageIds
        : toCamelCase(rslintDiag.messageId); // Convert to camelCase for normal rules
    assert(
      expectedMessageId === tsDiag.messageId,
      `Message mismatch: ${expectedMessageId} !== ${tsDiag.messageId} (rule: ${ruleName})`,
    );

    // check range match
    // tsDiag sometimes doesn't have line and column, so we need to check that
    if (tsDiag.line) {
      assert(
        rslintDiag.range.start.line === tsDiag.line,
        `Start line mismatch: ${rslintDiag.range.start.line} !== ${tsDiag.line}`,
      );
    }
    if (tsDiag.endLine) {
      assert(
        rslintDiag.range.end.line === tsDiag.endLine,
        `End line mismatch: ${rslintDiag.range.end.line} !== ${tsDiag.endLine}`,
      );
    }
    if (tsDiag.column) {
      // Handle both 0-based and 1-based column indexing differences
      const isMatch =
        rslintDiag.range.start.column === tsDiag.column ||
        rslintDiag.range.start.column === tsDiag.column - 1;
      assert(
        isMatch,
        `Start column mismatch: ${rslintDiag.range.start.column} !== ${tsDiag.column} (±1 for indexing)`,
      );
    }
    if (tsDiag.endColumn) {
      // Handle both 0-based and 1-based column indexing differences
      const isMatch =
        rslintDiag.range.end.column === tsDiag.endColumn ||
        rslintDiag.range.end.column === tsDiag.endColumn - 1;
      assert(
        isMatch,
        `End column mismatch: ${rslintDiag.range.end.column} !== ${tsDiag.endColumn} (±1 for indexing)`,
      );
    }
  }
}

interface RuleTesterOptions {
  languageOptions?: {
    parserOptions?: {
      project?: string;
      tsconfigRootDir?: string;
      projectService?: boolean;
    };
  };
}
export class RuleTester {
  options: RuleTesterOptions;
  constructor(options: RuleTesterOptions) {
    this.options = options;
  }
  public run(
    ruleName: string,
    cases: {
      valid: (
        | string
        | {
            code: string;
            options?: any;
            only?: boolean;
            skip?: boolean;
            filename?: string;
          }
      )[];
      invalid: {
        code: string;
        errors: any[];
        options?: any;
        only?: boolean;
        skip?: boolean;
        output?: string | string[];
        filename?: string;
        languageOptions?: RuleTesterOptions['languageOptions'];
      }[];
    },
  ) {
    // Extract the base rule name from descriptive test names like 'ban-ts-comment (ts-expect-error)'
    const baseRuleName = ruleName.split(' ')[0];
    describe(ruleName, () => {
      let cwd =
        this.options.languageOptions?.parserOptions?.tsconfigRootDir ||
        process.cwd();
      // Use the fixtures directory as the working directory
      // Use the fixtures-specific config file
      const config = path.resolve(cwd, 'rslint.json');
      let virtual_entry = path.resolve(cwd, 'src/virtual.ts');
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
          if (typeof validCase === 'object' && validCase.skip) {
            continue;
          }
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

          // Use custom filename if provided, otherwise use default virtual entry
          const filename =
            typeof validCase === 'string'
              ? virtual_entry
              : validCase.filename
                ? path.resolve(cwd, validCase.filename)
                : virtual_entry;

          // workaround for this hardcoded path https://github.com/typescript-eslint/typescript-eslint/blob/main/packages/eslint-plugin/tests/rules/no-floating-promises.test.ts#L712
          if (Array.isArray(options)) {
            for (const opt of options) {
              if (Array.isArray(opt.allowForKnownSafeCalls)) {
                for (const item of opt.allowForKnownSafeCalls) {
                  if (item.path) {
                    item.path = filename;
                  }
                }
              }
            }
          }
          // Use the existing project config with virtual file content
          const diags = await lint({
            config,
            workingDirectory: cwd,
            files: [filename], // Enable consistent file isolation for valid cases
            fileContents: {
              [filename]: code,
            },
            ruleOptions: {
              [baseRuleName]: options,
            },
          });

          assert(
            diags.diagnostics?.length === 0,
            `Expected no diagnostics for valid case, but got: ${JSON.stringify(diags)}`,
          );
        }
      });
      test('invalid', async t => {
        let caseIndex = 1;
        for (const item of cases.invalid) {
          const {
            code,
            errors,
            only = false,
            skip = false,
            options = [],
            filename,
          } = item;
          if (skip) {
            continue;
          }
          if (hasOnly && !only) {
            continue;
          }

          // Use custom filename if provided, otherwise use default virtual entry
          const testFilename = filename
            ? path.resolve(cwd, filename)
            : virtual_entry;

          // Use the existing project config with virtual file content
          const diags = await lint({
            config,
            workingDirectory: cwd,
            files: [testFilename], // Fix file isolation to include virtual files
            fileContents: {
              [testFilename]: code,
            },
            ruleOptions: {
              [baseRuleName]: options,
            },
          });

          toMatchSnapshot(diags, ruleName, caseIndex);

          assert(
            diags.diagnostics?.length > 0,
            `Expected diagnostics for invalid case`,
          );
          checkDiagnosticEqual(diags.diagnostics, errors, ruleName);
          caseIndex++;
        }
      });
    });
  }
}

/**
 * Simple no-op tag to mark code samples as "should not format with prettier"
 *   for the plugin-test-formatting lint rule
 */
export function noFormat(raw: TemplateStringsArray, ...keys: string[]): string {
  return String.raw({ raw }, ...keys);
}
