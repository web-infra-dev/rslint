// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(
  fixtureConfig(
    'internal/plugins/import/rules/no_relative_parent_imports/testdata/modules.txtar',
  ),
).run('no-relative-parent-imports', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import foo from "./internal.js"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 1',
      code: 'import foo from "./app/index.js"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 2',
      code: 'import foo from "package"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 3',
      code: 'require("./internal.js")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      options: [
        {
          commonjs: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 4',
      code: 'require("./app/index.js")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      options: [
        {
          commonjs: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 5',
      code: 'require("package")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      options: [
        {
          commonjs: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 6',
      code: 'import("./internal.js")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 7',
      code: 'import("./app/index.js")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 8',
      code: 'import(".")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 9',
      code: 'import("path")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 10',
      code: 'import("package")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 11',
      code: 'import("@scope/package")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 12',
      code: 'export default function (numbers) { return numbers.reduce((sum, n) => sum + n, 0); }',
      filename: 'add.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 13',
      code: 'export default function three(add) { return add([1, 2]); }',
      filename: 'numbers/three.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 14',
      code: "import add from './add';\nimport three from './numbers/three';\nconsole.log(three(add));",
      filename: 'use.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 15',
      code: "import add from 'add';\nexport default function three() { return add([1,2]); }",
      filename: 'numbers/three.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 16',
      code: "import foo from 'foo';\nimport a from './lib/a';",
      filename: 'main.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 17',
      code: "import b from './b';",
      filename: 'lib/a.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 18',
      code: "import add from './add';",
      filename: 'three.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 19',
      code: "import add from './math/add';",
      filename: 'three.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
  ],
  invalid: [
    {
      name: 'run 0 invalid 0',
      code: 'import foo from "../plugin.js"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `index.js` to same directory as `../plugin.js` or consider making `../plugin.js` a package.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'require("../plugin.js")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      options: [
        {
          commonjs: true,
        },
      ],
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `index.js` to same directory as `../plugin.js` or consider making `../plugin.js` a package.",
          line: 1,
          column: 9,
          endLine: 1,
          endColumn: 23,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import("../plugin.js")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `index.js` to same directory as `../plugin.js` or consider making `../plugin.js` a package.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import foo from "./../plugin.js"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `index.js` to same directory as `./../plugin.js` or consider making `./../plugin.js` a package.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import foo from "../../api/service"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `index.js` to same directory as `../../api/service` or consider making `../../api/service` a package.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import("../../api/service")',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `index.js` to same directory as `../../api/service` or consider making `../../api/service` a package.",
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 27,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: "import add from '../add';\nexport default function three() { return add([1, 2]); }",
      filename: 'numbers/three.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `three.js` to same directory as `../add` or consider making `../add` a package.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: "import bar from '../main';",
      filename: 'lib/a.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message:
            "Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move `a.js` to same directory as `../main` or consider making `../main` a package.",
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
  ],
});
