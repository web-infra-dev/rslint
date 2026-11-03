// Upstream: eslint-plugin-unicorn v79.0.0, including runtime checks and docs.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/test/prefer-string-starts-ends-with.js
import fs from 'node:fs';

import { RuleTester, type InvalidTestCase } from '../rule-tester';

// Share the pinned upstream expectations with Go; rslint extras stay in Go.
type UpstreamCase = Omit<InvalidTestCase, 'filename' | 'output' | 'errors'> & {
  fileName: string;
  errors: Exclude<InvalidTestCase['errors'], number>;
  output: string[];
};

const cases: UpstreamCase[] = JSON.parse(
  fs.readFileSync(
    new URL(
      '../../../../../internal/plugins/unicorn/rules/prefer_string_starts_ends_with/testdata/upstream.json',
      import.meta.url,
    ),
    'utf8',
  ),
);

const tester = new RuleTester();
tester.run('prefer-string-starts-ends-with', {} as never, {
  valid: cases
    .filter((item) => item.errors.length === 0)
    .map((item) => ({ code: item.code, filename: item.fileName })),
  invalid: cases
    .filter((item) => item.errors.length > 0)
    .map((item) => ({
      code: item.code,
      filename: item.fileName,
      errors: item.errors,
      output: item.output[0] ?? null,
    })),
});
