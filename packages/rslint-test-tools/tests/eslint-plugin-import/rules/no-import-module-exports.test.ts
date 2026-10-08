// Upstream: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/tests/src/rules/no-import-module-exports.js
// Documentation: https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/docs/rules/no-import-module-exports.md
import path from 'node:path';

import { RuleTester } from '../rule-tester.js';

const ruleTester = new RuleTester();
const exceptionFile = path.resolve(
  import.meta.dirname,
  '../files/no-import-module-exports/some/other/entry-point.js',
);
const docsExceptionFile = path.resolve(
  import.meta.dirname,
  '../files/no-import-module-exports/docs/some-file.js',
);
const error = {
  message:
    "Cannot use import declarations in modules that export using CommonJS (module.exports = 'foo' or exports.bar = 'hi')",
};

ruleTester.run('no-import-module-exports', null as never, {
  valid: [
    { code: "const thing = require('thing')\nmodule.exports = thing" },
    {
      code: "import thing from 'otherthing'\nconsole.log(thing.module.exports)",
    },
    { code: "import thing from 'other-thing'\nexport default thing" },
    { code: "const thing = require('thing')\nexports.foo = bar" },
    {
      code: "import { module } from 'qunit'\nmodule.skip('A test', function () {})",
    },
    {
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: './files/no-import-module-exports/index.js',
    },
    {
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: exceptionFile,
      // The exact path keeps this valid inside hidden worktree directories;
      // the original upstream glob is covered independently by the Go suite.
      options: [{ exceptions: ['**/*/other/entry-point.js', exceptionFile] }],
    },
    {
      code: "import * as process from 'process';\nconsole.log(process.env);",
      filename: './files/no-import-module-exports/missing-entrypoint/cli.js',
    },
    {
      code: `
        import fs from 'fs/promises';

        const subscriptions = new Map();
        export default async (client) => {
            const modules = await fs.readdir('./src/modules');
            await Promise.all(
                modules.map(async (moduleName) => {
                    const module = await import(\`./modules/\${moduleName}/module.js\`);
                    if (module.enabled) {
                        module.subscriptions.forEach((fun, event) => {
                            if (!subscriptions.has(event)) subscriptions.set(event, []);
                            subscriptions.get(event).push(fun);
                        });
                    }
                })
            );
        };
      `,
    },
    // Documentation pass examples.
    {
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: './files/no-import-module-exports/docs/lib/index.js',
    },
    {
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: docsExceptionFile,
      options: [{ exceptions: ['**/*/some-file.js', docsExceptionFile] }],
    },
  ],
  invalid: [
    {
      code: "import { stuff } from 'starwars'\nmodule.exports = thing",
      errors: [error],
    },
    {
      code: "import thing from 'starwars'\nconst baz = module.exports = thing\nconsole.log(baz)",
      errors: [error],
    },
    {
      code: "import * as allThings from 'starwars'\nexports.bar = thing",
      errors: [error],
    },
    {
      code: "import thing from 'other-thing'\nexports.foo = bar",
      errors: [error],
    },
    {
      code: "import foo from 'path';\nmodule.exports = foo;",
      filename: exceptionFile,
      options: [{ exceptions: ['**/*/other/file.js'] }],
      errors: [error],
    },
  ],
});
