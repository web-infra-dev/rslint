import fs from 'node:fs';
import path from 'node:path';

import { RuleTester } from '../rule-tester';

interface FixtureCase {
  code: string;
  options?: unknown;
  globals?: Record<string, boolean>;
  skip?: boolean;
  tsx?: boolean;
}

interface FixtureInvalidCase extends FixtureCase {
  errors: Array<{
    messageId?: string;
    line?: number;
    column?: number;
    endLine?: number;
    endColumn?: number;
  }>;
}

interface FixtureSuite {
  valid: FixtureCase[];
  invalid: FixtureInvalidCase[];
}

const fixturePath = path.resolve(
  import.meta.dirname,
  '../../../../../internal/rules/no_unused_vars/no_unused_vars_upstream.json',
);
const fixture = JSON.parse(
  fs.readFileSync(fixturePath, 'utf8'),
) as FixtureSuite;

function languageOptions(globals?: Record<string, boolean>) {
  if (!globals) return undefined;
  return {
    globals: Object.fromEntries(
      Object.entries(globals).map(([name, enabled]) => [
        name,
        enabled ? 'readonly' : 'off',
      ]),
    ),
  };
}

function integrationCode(code: string) {
  return code.replace(/[ \t]+$/gm, '');
}

const ruleTester = new RuleTester();

ruleTester.run('no-unused-vars', {
  valid: fixture.valid.map((testCase) => ({
    code: integrationCode(testCase.code),
    options: testCase.options,
    languageOptions: languageOptions(testCase.globals),
    filename: testCase.tsx ? 'src/virtual.tsx' : undefined,
    skip: testCase.skip,
  })) as any,
  invalid: fixture.invalid.map((testCase) => ({
    code: integrationCode(testCase.code),
    errors: testCase.errors,
    options: testCase.options,
    languageOptions: languageOptions(testCase.globals),
    filename: testCase.tsx ? 'src/virtual.tsx' : undefined,
    skip: testCase.skip,
  })) as any,
});
