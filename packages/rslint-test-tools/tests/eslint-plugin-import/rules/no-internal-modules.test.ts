// Migrated from eslint-plugin-import@2.32.0 and the existing Rslint regressions.
// Assertions record current native output; skipped inputs retain their upstream assertions.
import { RuleTester } from '@rslint/rule-tester';
import { fixtureConfig } from '../config';
new RuleTester(
  fixtureConfig(
    'internal/plugins/import/rules/no_internal_modules/testdata/modules.txtar',
  ),
).run('no-internal-modules', {
  valid: [
    {
      name: 'run 0 valid 0',
      code: 'import a from "./plugin2"',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 1',
      code: 'const a = require("./plugin2")',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 2',
      code: 'const a = require("./plugin2/")',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 3',
      code: 'const dynamic = "./plugin2/"; const a = require(dynamic)',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 4',
      code: 'import b from "./internal.js"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 5',
      code: 'import get from "lodash.get"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 6',
      code: 'import b from "@org/package"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 7',
      code: 'import b from "../../api/service"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          allow: ['**/api/*'],
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
      name: 'run 0 valid 8',
      code: 'import "jquery/dist/jquery"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          allow: ['jquery/dist/*'],
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
      name: 'run 0 valid 9',
      code: 'import "./app/index.js";\nimport "./app/index"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          allow: ['**/index{.js,}'],
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
      name: 'run 0 valid 10',
      code: 'import a from "./plugin2/thing"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          forbid: ['**/api/*'],
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
      name: 'run 0 valid 11',
      code: 'const a = require("./plugin2/thing")',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          forbid: ['**/api/*'],
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
      name: 'run 0 valid 12',
      code: 'import b from "app/a"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['app/**/**'],
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
      name: 'run 0 valid 13',
      code: 'import b from "@org/package"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['@org/package/*'],
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
      name: 'run 0 valid 14',
      code: 'export {a} from "./internal.js"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 15',
      code: 'export * from "lodash.get"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 16',
      code: 'export {b} from "@org/package"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 17',
      code: 'export {b} from "../../api/service"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          allow: ['**/api/*'],
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
      name: 'run 0 valid 18',
      code: 'export * from "jquery/dist/jquery"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          allow: ['jquery/dist/*'],
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
      name: 'run 0 valid 19',
      code: 'export * from "./app/index.js";\nexport * from "./app/index"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          allow: ['**/index{.js,}'],
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
      name: 'run 0 valid 20',
      code: '\n        export class AuthHelper {\n\n          static checkAuth(auth) {\n          }\n        }\n      ',
    },
    {
      name: 'run 0 valid 21 [upstream parser metadata: {"parser":"native"}]',
      code: '\n          export class AuthHelper {\n\n            public static checkAuth(auth?: string): boolean {\n            }\n          }\n        ',
      filename: 'internal-modules/plugins/plugin.ts',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
    },
    {
      name: 'run 0 valid 22',
      code: 'export * from "./plugin2/thing"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          forbid: ['**/api/*'],
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
      name: 'run 0 valid 23',
      code: 'export * from "app/a"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['app/**/**'],
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
      name: 'run 0 valid 24',
      code: 'export { b } from "@org/package"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['@org/package/*'],
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
      name: 'run 0 valid 25',
      code: 'export * from "./app/index.js";\nexport * from "./app/index"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['**/index.ts'],
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
      name: 'run 1 valid 0',
      code: "import 'source-map-support/register';\nimport { settings } from '../app';\nimport getUser from '../actions/getUser';\nexport * from 'source-map-support/register';\nexport { settings } from '../app';",
      filename: 'my-project/entry.js',
      options: [
        {
          allow: ['**/actions/*', 'source-map-support/*'],
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
      name: 'run 1 valid 1',
      code: "import 'source-map-support';\nimport { getUser } from '../actions';\nexport * from 'source-map-support';\nexport { getUser } from '../actions';",
      filename: 'my-project/entry.js',
      options: [
        {
          forbid: ['**/actions/*', 'source-map-support/*'],
        },
      ],
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
      code: 'import "./plugin2/index.js";\nimport "./plugin2/app/index"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          allow: ['*/index.js'],
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
          message: 'Reaching to "./plugin2/app/index" is not allowed.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 1',
      code: 'import "./app/index.js"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 2',
      code: 'import b from "./plugin2/internal"',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "./plugin2/internal" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      name: 'run 0 invalid 3',
      code: 'import a from "../api/service/index"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          allow: ['**/internal-modules/*'],
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
          message: 'Reaching to "../api/service/index" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 4',
      code: 'import b from "@org/package/internal"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "@org/package/internal" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      name: 'run 0 invalid 5',
      code: 'import get from "debug/node"',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "debug/node" is not allowed.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 6',
      code: 'import "./app/index.js"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['*/app/*'],
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
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 7',
      code: 'import b from "@org/package"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['@org/**'],
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
          message: 'Reaching to "@org/package" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 8',
      code: 'import b from "app/a/b"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['app/**/**'],
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
          message: 'Reaching to "app/a/b" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      name: 'run 0 invalid 9',
      code: 'import get from "lodash.get"',
      filename: 'internal-modules/plugins/plugin2/index.js',
      options: [
        {
          forbid: ['lodash.*'],
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
          message: 'Reaching to "lodash.get" is not allowed.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 29,
        },
      ],
    },
    {
      name: 'run 0 invalid 10',
      code: 'import "./app/index.js";\nimport "./app/index"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['**/index{.js,}'],
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
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 24,
        },
        {
          messageId: '',
          message: 'Reaching to "./app/index" is not allowed.',
          line: 2,
          column: 8,
          endLine: 2,
          endColumn: 21,
        },
      ],
    },
    {
      name: 'run 0 invalid 11',
      code: 'export * from "./plugin2/index.js";\nexport * from "./plugin2/app/index"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          allow: ['*/index.js'],
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
          message: 'Reaching to "./plugin2/app/index" is not allowed.',
          line: 2,
          column: 15,
          endLine: 2,
          endColumn: 36,
        },
      ],
    },
    {
      name: 'run 0 invalid 12',
      code: 'export * from "./app/index.js"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 13',
      code: 'export {b} from "./plugin2/internal"',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "./plugin2/internal" is not allowed.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 37,
        },
      ],
    },
    {
      name: 'run 0 invalid 14',
      code: 'export {a} from "../api/service/index"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          allow: ['**/internal-modules/*'],
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
          message: 'Reaching to "../api/service/index" is not allowed.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      name: 'run 0 invalid 15',
      code: 'export {b} from "@org/package/internal"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "@org/package/internal" is not allowed.',
          line: 1,
          column: 17,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      name: 'run 0 invalid 16',
      code: 'export {get} from "debug/node"',
      filename: 'internal-modules/plugins/plugin.js',
      languageOptions: {
        parserOptions: {
          project: false,
          projectService: false,
        },
      },
      errors: [
        {
          messageId: '',
          message: 'Reaching to "debug/node" is not allowed.',
          line: 1,
          column: 19,
          endLine: 1,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 0 invalid 17',
      code: 'export * from "./plugin2/thing"',
      filename: 'internal-modules/plugins/plugin.js',
      options: [
        {
          forbid: ['**/plugin2/*'],
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
          message: 'Reaching to "./plugin2/thing" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 32,
        },
      ],
    },
    {
      name: 'run 0 invalid 18',
      code: 'export * from "app/a"',
      filename: 'internal-modules/plugins/plugin2/internal.js',
      options: [
        {
          forbid: ['**'],
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
          message: 'Reaching to "app/a" is not allowed.',
          line: 1,
          column: 15,
          endLine: 1,
          endColumn: 22,
        },
      ],
    },
    {
      name: 'run 1 invalid 0',
      code: "import { settings } from './app/index';\nimport userReducer from './reducer/user';\nimport configureStore from './redux/configureStore';\nexport { settings } from './app/index';\nexport * from './reducer/user';",
      filename: 'my-project/entry.js',
      options: [
        {
          allow: ['**/actions/*', 'source-map-support/*'],
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
          message: 'Reaching to "./app/index" is not allowed.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
        },
        {
          messageId: '',
          message: 'Reaching to "./reducer/user" is not allowed.',
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 41,
        },
        {
          messageId: '',
          message: 'Reaching to "./redux/configureStore" is not allowed.',
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 52,
        },
        {
          messageId: '',
          message: 'Reaching to "./app/index" is not allowed.',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 39,
        },
        {
          messageId: '',
          message: 'Reaching to "./reducer/user" is not allowed.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 31,
        },
      ],
    },
    {
      name: 'run 1 invalid 1',
      code: "import 'source-map-support/register';\nimport getUser from '../actions/getUser';\nexport * from 'source-map-support/register';",
      filename: 'my-project/entry.js',
      options: [
        {
          forbid: ['**/actions/*', 'source-map-support/*'],
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
          message: 'Reaching to "source-map-support/register" is not allowed.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 37,
        },
        {
          messageId: '',
          message: 'Reaching to "../actions/getUser" is not allowed.',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 41,
        },
        {
          messageId: '',
          message: 'Reaching to "source-map-support/register" is not allowed.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 44,
        },
      ],
    },
  ],
});
