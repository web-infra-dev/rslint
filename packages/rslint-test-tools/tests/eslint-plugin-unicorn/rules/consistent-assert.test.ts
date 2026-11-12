// Ported from eslint-plugin-unicorn v75.0.0; see LICENSE.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v75.0.0/test/consistent-assert.js
import {
  RuleTester,
  type InvalidTestCase,
  type ValidTestCase,
} from '../rule-tester';

const defaults = {
  filename: 'src/virtual.js',
  languageOptions: { sourceType: 'module' as const },
};

const valid = (code: string, filename = defaults.filename): ValidTestCase => ({
  ...defaults,
  code,
  filename,
});

const error = (name: string) => ({
  messageId: 'consistent-assert/error',
  message: `Prefer \`${name}.ok(…)\` over \`${name}(…)\`.`,
});

const invalid = (
  code: string,
  output: string,
  name = 'assert',
  filename = defaults.filename,
): InvalidTestCase => ({
  ...defaults,
  code,
  output,
  filename,
  errors: [error(name)],
});

const validCases: ValidTestCase[] = [
  valid('assert(foo)'),
  valid('import assert from "assert";'),
  valid("import assert from 'node:assert';\nassert;"),
  valid("import customAssert from 'node:assert';\nassert(foo);"),
  valid("function foo (assert) {\n\tassert(bar);\n}"),
  valid("import assert from 'node:assert';\n\nfunction foo (assert) {\n\tassert(bar);\n}"),
  valid("import {strict} from 'node:assert/strict';\n\nstrict(foo);"),
  valid("import * as assert from 'node:assert';\nassert(foo);"),
  valid("export * as assert from 'node:assert';\nassert(foo);"),
  valid("export {default as assert} from 'node:assert';\nexport {assert as strict} from 'node:assert';\nassert(foo);"),
  valid("import assert from 'node:assert/strict';\nconsole.log(assert)"),
  valid("import {'strict' as assert} from 'assert';\nassert(foo)"),
  valid('import type assert from "node:assert/strict";', 'src/virtual.ts'),
  valid('import type assert from "node:assert/strict";\nassert();', 'src/virtual.ts'),
  valid('import {type strict as assert} from "node:assert/strict";', 'src/virtual.ts'),
  valid('import {type strict as assert} from "node:assert/strict";\nassert();', 'src/virtual.ts'),
  valid('import type {strict as assert} from "node:assert/strict";', 'src/virtual.ts'),
  valid('import type {strict as assert} from "node:assert/strict";\nassert();', 'src/virtual.ts'),
];

const multipleCode = "import assert from 'assert';\nassert(foo)\nassert(bar)\nassert(baz)";
const multipleOutput = "import assert from 'assert';\nassert.ok(foo)\nassert.ok(bar)\nassert.ok(baz)";

const allCode =
  "import a, {strict as b, default as c} from 'node:assert';\n" +
  "import d, {strict as e, default as f} from 'assert';\n" +
  "import g, {default as h} from 'node:assert/strict';\n" +
  "import i, {default as j} from 'assert/strict';\n" +
  'a(foo);\nb(foo);\nc(foo);\nd(foo);\ne(foo);\nf(foo);\ng(foo);\nh(foo);\ni(foo);\nj(foo);';

const allOutput =
  "import a, {strict as b, default as c} from 'node:assert';\n" +
  "import d, {strict as e, default as f} from 'assert';\n" +
  "import g, {default as h} from 'node:assert/strict';\n" +
  "import i, {default as j} from 'assert/strict';\n" +
  'a.ok(foo);\nb.ok(foo);\nc.ok(foo);\nd.ok(foo);\ne.ok(foo);\nf.ok(foo);\ng.ok(foo);\nh.ok(foo);\ni.ok(foo);\nj.ok(foo);';

const invalidCases: InvalidTestCase[] = [
  invalid("import assert from 'assert';\nassert(foo)", "import assert from 'assert';\nassert.ok(foo)"),
  invalid("import assert from 'node:assert';\nassert(foo)", "import assert from 'node:assert';\nassert.ok(foo)"),
  invalid("import assert from 'assert/strict';\nassert(foo)", "import assert from 'assert/strict';\nassert.ok(foo)"),
  invalid("import assert from 'node:assert/strict';\nassert(foo)", "import assert from 'node:assert/strict';\nassert.ok(foo)"),
  invalid("import customAssert from 'assert';\ncustomAssert(foo)", "import customAssert from 'assert';\ncustomAssert.ok(foo)", 'customAssert'),
  invalid("import customAssert from 'node:assert';\ncustomAssert(foo)", "import customAssert from 'node:assert';\ncustomAssert.ok(foo)", 'customAssert'),
  {
    ...defaults,
    code: multipleCode,
    output: multipleOutput,
    errors: [error('assert'), error('assert'), error('assert')],
  },
  invalid("import {strict} from 'assert';\nstrict(foo)", "import {strict} from 'assert';\nstrict.ok(foo)", 'strict'),
  invalid("import {strict as assert} from 'assert';\nassert(foo)", "import {strict as assert} from 'assert';\nassert.ok(foo)"),
  {
    ...defaults,
    code: allCode,
    output: allOutput,
    errors: 'abcdefghij'.split('').map(error),
  },
  invalid("import assert from 'node:assert';\nassert?.(foo)", "import assert from 'node:assert';\nassert.ok?.(foo)"),
  invalid(
    "import assert from 'assert';\n\n((\n\t/* comment */ ((\n\t\t/* comment */\n\t\tassert\n\t\t/* comment */\n\t\t)) /* comment */\n\t\t(/* comment */ typeof foo === 'string', 'foo must be a string' /** after comment */)\n));",
    "import assert from 'assert';\n\n((\n\t/* comment */ ((\n\t\t/* comment */\n\t\tassert.ok\n\t\t/* comment */\n\t\t)) /* comment */\n\t\t(/* comment */ typeof foo === 'string', 'foo must be a string' /** after comment */)\n));",
  ),
];

new RuleTester().run('consistent-assert', {} as never, {
  valid: validCases,
  invalid: invalidCases,
});
