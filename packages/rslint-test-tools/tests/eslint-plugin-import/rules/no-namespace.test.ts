import { RuleTester } from '../rule-tester.js';

// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-namespace.js
// https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-namespace.md
// The wrapper checks counts/messages. Go tests and differential validation
// additionally assert full ranges, fixes, and the absence of message IDs.
const ruleTester = new RuleTester();
const rule = null as never;
const errors = (column = 8) => [
  {
    message: 'Unexpected namespace import.',
    line: 1,
    column,
    endLine: 1,
    endColumn: column + 8,
  },
];

ruleTester.run('no-namespace', rule, {
  valid: [
    { code: "import { a, b } from 'foo';" },
    { code: "import { a, b } from './foo';" },
    { code: "import bar from 'bar';" },
    { code: "import bar from './bar';" },
    {
      code: "import * as bar from './ignored-module.ext';",
      options: [{ ignore: ['*.ext'] }],
    },
  ],
  invalid: [
    {
      code: "import * as foo from 'foo';",
      output: null,
      errors: errors(),
    },
    {
      code: "import defaultExport, * as foo from 'foo';",
      output: null,
      errors: errors(23),
    },
    {
      code: "import * as foo from './foo';",
      output: null,
      errors: errors(),
    },
    // FIX_TESTS (ESLint 5+).
    {
      code: `import * as foo from './foo';
      florp(foo.bar);
      florp(foo['baz']);`,
      output: `import { bar, baz } from './foo';
      florp(bar);
      florp(baz);`,
      errors: errors(),
    },
    {
      code: `import * as foo from './foo';
      const bar = 'name conflict';
      const baz = 'name conflict';
      const foo_baz = 'name conflict';
      florp(foo.bar);
      florp(foo['baz']);`,
      output: `import { bar as foo_bar, baz as foo_baz_1 } from './foo';
      const bar = 'name conflict';
      const baz = 'name conflict';
      const foo_baz = 'name conflict';
      florp(foo_bar);
      florp(foo_baz_1);`,
      errors: errors(),
    },
    {
      code: `import * as foo from './foo';
      function func(arg) {
        florp(foo.func);
        florp(foo['arg']);
      }`,
      output: `import { func as foo_func, arg as foo_arg } from './foo';
      function func(arg) {
        florp(foo_func);
        florp(foo_arg);
      }`,
      errors: errors(),
    },
  ],
});

describe('documentation examples', () => {
  ruleTester.run('no-namespace', rule, {
    valid: [
      { code: "import defaultExport from './foo'" },
      { code: "import { a, b }  from './bar'" },
      { code: "import defaultExport, { a, b }  from './foobar'" },
      {
        code: "/* eslint import/no-namespace: [\"error\", {ignore: ['*.ext']}] */\nimport * as bar from './ignored-module.ext';",
        options: [{ ignore: ['*.ext'] }],
      },
    ],
    invalid: [
      {
        code: "import * as foo from 'foo';",
        output: null,
        errors: errors(),
      },
      {
        code: "import defaultExport, * as foo from 'foo';",
        output: null,
        errors: errors(23),
      },
    ],
  });
});
