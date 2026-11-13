// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(fixtureConfig()).run('no-namespace', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "import { a, b } from 'foo';",
    },
    {
      name: 'run 0 valid 1',
      code: "import { a, b } from './foo';",
    },
    {
      name: 'run 0 valid 2',
      code: "import bar from 'bar';",
    },
    {
      name: 'run 0 valid 3',
      code: "import bar from './bar';",
    },
    {
      name: 'run 0 valid 4',
      code: "import * as bar from './ignored-module.ext';",
      options: [
        {
          ignore: ['*.ext'],
        },
      ],
    },
    {
      name: 'documentation examples / run 1 valid 0',
      code: "import defaultExport from './foo'",
    },
    {
      name: 'documentation examples / run 1 valid 1',
      code: "import { a, b }  from './bar'",
    },
    {
      name: 'documentation examples / run 1 valid 2',
      code: "import defaultExport, { a, b }  from './foobar'",
    },
    {
      name: 'documentation examples / run 1 valid 3',
      code: "/* eslint import/no-namespace: [\"error\", {ignore: ['*.ext']}] */\nimport * as bar from './ignored-module.ext';",
      options: [
        {
          ignore: ['*.ext'],
        },
      ],
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import * as foo from 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "import defaultExport, * as foo from 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import * as foo from './foo';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "import * as foo from './foo';\n      florp(foo.bar);\n      florp(foo['baz']);",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output:
        "import { bar, baz } from './foo';\n      florp(bar);\n      florp(baz);",
    },
    {
      name: 'run 0 invalid 4',
      code: "import * as foo from './foo';\n      const bar = 'name conflict';\n      const baz = 'name conflict';\n      const foo_baz = 'name conflict';\n      florp(foo.bar);\n      florp(foo['baz']);",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output:
        "import { bar as foo_bar, baz as foo_baz_1 } from './foo';\n      const bar = 'name conflict';\n      const baz = 'name conflict';\n      const foo_baz = 'name conflict';\n      florp(foo_bar);\n      florp(foo_baz_1);",
    },
    {
      name: 'run 0 invalid 5',
      code: "import * as foo from './foo';\n      function func(arg) {\n        florp(foo.func);\n        florp(foo['arg']);\n      }",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
      output:
        "import { func as foo_func, arg as foo_arg } from './foo';\n      function func(arg) {\n        florp(foo_func);\n        florp(foo_arg);\n      }",
    },
    {
      name: 'documentation examples / run 1 invalid 0',
      code: "import * as foo from 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      name: 'documentation examples / run 1 invalid 1',
      code: "import defaultExport, * as foo from 'foo';",
      errors: [
        {
          messageId: '',
          message: 'Unexpected namespace import.',
          line: 1,
          column: 23,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
  ],
});
