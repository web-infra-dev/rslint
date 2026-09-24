import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-empty-named-blocks.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-empty-named-blocks.md
// This wrapper checks diagnostic counts and messages. Go tests additionally
// assert complete ranges, fixes, suggestions, and the absence of message IDs.
const ruleTester = new RuleTester();
const rule = null as never;
const message = 'Unexpected empty named import block';

function suggestionCases(codes: string[], output = "import 'mod';") {
  return codes.map((code) => ({
    code,
    output: null,
    errors: [
      {
        message,
        line: 1,
        column: 1,
        endLine: 1,
        endColumn: code.length + 1,
        suggestions: [
          { desc: 'Remove unused import', output: '' },
          { desc: 'Remove empty import block', output },
        ],
      },
    ],
  }));
}

ruleTester.run('no-empty-named-blocks', rule, {
  valid: [
    { code: "import 'mod';" },
    { code: "import Default from 'mod';" },
    { code: "import { Named } from 'mod';" },
    { code: "import Default, { Named } from 'mod';" },
    { code: "import * as Namespace from 'mod';" },
    // TypeScript.
    { code: "import type Default from 'mod';" },
    { code: "import type { Named } from 'mod';" },
    { code: "import type Default, { Named } from 'mod';" },
    { code: "import type * as Namespace from 'mod';" },
    {
      code: `
        module.exports = {
          rules: {
            'keyword-spacing': ['error', {overrides: {}}],
          }
        };
      `,
    },
    {
      code: `
        import { DESCRIPTORS, NODE } from '../helpers/constants';
        // ...
        import { timeLimitedPromise } from '../helpers/helpers';
        // ...
        import { DESCRIPTORS2 } from '../helpers/constants';
      `,
    },
  ],
  invalid: [
    {
      code: "import Default, {} from 'mod';",
      output: "import Default from 'mod';",
      errors: [{ message }],
    },
    ...suggestionCases([
      "import {} from 'mod';",
      "import{}from'mod';",
      "import {} from'mod';",
      "import {}from 'mod';",
    ]),
    // TypeScript.
    ...suggestionCases([
      "import type {} from 'mod';",
      "import type {}from 'mod';",
      "import type{}from 'mod';",
      "import type {}from'mod';",
    ]),
    {
      code: "import type Default, {} from 'mod';",
      output: "import type Default from 'mod';",
      errors: [{ message }],
    },
  ],
});

describe('documentation examples', () => {
  ruleTester.run('no-empty-named-blocks', rule, {
    valid: [
      { code: "import { mod } from 'mod'" },
      { code: "import Default, { mod } from 'mod'" },
      { code: "import type { mod } from 'mod'" },
    ],
    invalid: [
      ...suggestionCases(
        ["import {} from 'mod'", "import type {} from 'mod'"],
        "import 'mod'",
      ),
      {
        code: "import Default, {} from 'mod'",
        output: "import Default from 'mod'",
        errors: [{ message }],
      },
      {
        code: "import type Default, {} from 'mod'",
        output: "import type Default from 'mod'",
        errors: [{ message }],
      },
    ],
  });
});

// Retain every Flow test/documentation example. The TypeScript parser does
// not support Flow's `import typeof` syntax.
describe.skip('Flow typeof imports are unsupported', () => {
  ruleTester.run('no-empty-named-blocks', rule, {
    valid: [
      { code: "import typeof Default from 'mod'; // babel old" },
      { code: "import typeof { Named } from 'mod'; // babel old" },
      { code: "import typeof Default, { Named } from 'mod'; // babel old" },
      { code: "import typeof { mod } from 'mod'" },
    ],
    invalid: [
      ...suggestionCases([
        "import typeof {} from 'mod';",
        "import typeof {}from 'mod';",
        "import typeof {} from'mod';",
        "import typeof{}from'mod';",
      ]),
      {
        code: "import typeof Default, {} from 'mod';",
        output: "import typeof Default from 'mod';",
        errors: [{ message }],
      },
      ...suggestionCases(["import typeof {} from 'mod'"], "import 'mod'"),
      {
        code: "import typeof Default, {} from 'mod'",
        output: "import typeof Default from 'mod'",
        errors: [{ message }],
      },
    ],
  });
});
