// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig, resolveFixturePaths } from '../config';
new RuleTester(fixtureConfig()).run('no-import-module-exports', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: "const thing = require('thing')\nmodule.exports = thing",
    },
    {
      name: 'run 0 valid 1',
      code: "import thing from 'otherthing'\nconsole.log(thing.module.exports)",
    },
    {
      name: 'run 0 valid 2',
      code: "import thing from 'other-thing'\nexport default thing",
    },
    {
      name: 'run 0 valid 3',
      code: "const thing = require('thing')\nexports.foo = bar",
    },
    {
      name: 'run 0 valid 4',
      code: "import { module } from 'qunit'\nmodule.skip('A test', function () {})",
    },
    {
      name: 'run 0 valid 5',
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: 'files/no-import-module-exports/index.js',
    },
    {
      name: 'run 0 valid 6',
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: 'files/no-import-module-exports/some/other/entry-point.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              exceptions: [
                '<fixture>/files/no-import-module-exports/some/other/entry-point.js',
              ],
            },
          ],
          root,
        ),
    },
    {
      name: 'run 0 valid 7',
      code: "import * as process from 'process';\nconsole.log(process.env);",
      filename: 'files/no-import-module-exports/missing-entrypoint/cli.js',
    },
    {
      name: 'run 0 valid 8',
      code: "\n        import fs from 'fs/promises';\n\n        const subscriptions = new Map();\n        export default async (client) => {\n            const modules = await fs.readdir('./src/modules');\n            await Promise.all(\n                modules.map(async (moduleName) => {\n                    const module = await import(`./modules/${moduleName}/module.js`);\n                    if (module.enabled) {\n                        module.subscriptions.forEach((fun, event) => {\n                            if (!subscriptions.has(event)) subscriptions.set(event, []);\n                            subscriptions.get(event).push(fun);\n                        });\n                    }\n                })\n            );\n        };\n      ",
    },
    {
      name: 'run 0 valid 9',
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: 'files/no-import-module-exports/docs/lib/index.js',
    },
    {
      name: 'run 0 valid 10',
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: 'files/no-import-module-exports/docs/some-file.js',
      options: (root: string) =>
        resolveFixturePaths(
          [
            {
              exceptions: [
                '**/*/some-file.js',
                '<fixture>/files/no-import-module-exports/docs/some-file.js',
              ],
            },
          ],
          root,
        ),
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: "import { stuff } from 'starwars'\nmodule.exports = thing",
      errors: [
        {
          messageId: '',
          message:
            "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: "import thing from 'starwars'\nconst baz = module.exports = thing\nconsole.log(baz)",
      errors: [
        {
          messageId: '',
          message:
            "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: "import * as allThings from 'starwars'\nexports.bar = thing",
      errors: [
        {
          messageId: '',
          message:
            "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: "import thing from 'other-thing'\nexports.foo = bar",
      errors: [
        {
          messageId: '',
          message:
            "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: 'files/no-import-module-exports/some/other/entry-point.js',
      options: [
        {
          exceptions: ['**/*/other/file.js'],
        },
      ],
      errors: [
        {
          messageId: '',
          message:
            "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
  ],
});
