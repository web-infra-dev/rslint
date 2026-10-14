// Upstream: eslint-plugin-import v2.32.0 tests/src/rules/no-internal-modules.js.
import fs from 'node:fs';
import path from 'node:path';
import { RuleTester } from '../rule-tester.js';

const root = fs.mkdtempSync(path.join(process.cwd(), '.no-internal-modules-'));
const archive = fs.readFileSync(
  path.resolve(
    import.meta.dirname,
    '../../../../../internal/plugins/import/rules/no_internal_modules/testdata/modules.txtar',
  ),
  'utf8',
);
const chunks = archive.split(/^-- (.+) --\r?$/m);
if (chunks.length < 3) throw new Error('Missing upstream fixtures');
for (let i = 1; i < chunks.length; i += 2) {
  const file = path.join(root, chunks[i]);
  fs.mkdirSync(path.dirname(file), { recursive: true });
  fs.writeFileSync(file, chunks[i + 1].replace(/^\r?\n/, ''));
}
afterAll(() => fs.rmSync(root, { recursive: true, force: true }));
const testFilePath = (name: string) => path.join(root, name);

const ruleTester = new RuleTester();
const test = (value: any) => value;
const getTSParsers = () => ['native'];
const flatMap = (values: string[], fn: (parser: string) => any[]) =>
  values.flatMap(fn);

ruleTester.run('no-internal-modules', null as never, {
  valid: [
    // imports
    test({
      code: 'import a from "./plugin2"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [],
    }),
    test({
      code: 'const a = require("./plugin2")',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
    }),
    test({
      code: 'const a = require("./plugin2/")',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
    }),
    test({
      code: 'const dynamic = "./plugin2/"; const a = require(dynamic)',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
    }),
    test({
      code: 'import b from "./internal.js"',
      filename: testFilePath('./internal-modules/plugins/plugin2/index.js'),
    }),
    test({
      code: 'import get from "lodash.get"',
      filename: testFilePath('./internal-modules/plugins/plugin2/index.js'),
    }),
    test({
      code: 'import b from "@org/package"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
    }),
    test({
      code: 'import b from "../../api/service"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          allow: ['**/api/*'],
        },
      ],
    }),
    test({
      code: 'import "jquery/dist/jquery"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          allow: ['jquery/dist/*'],
        },
      ],
    }),
    test({
      code: 'import "./app/index.js";\nimport "./app/index"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          allow: ['**/index{.js,}'],
        },
      ],
    }),
    test({
      code: 'import a from "./plugin2/thing"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          forbid: ['**/api/*'],
        },
      ],
    }),
    test({
      code: 'const a = require("./plugin2/thing")',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          forbid: ['**/api/*'],
        },
      ],
    }),
    test({
      code: 'import b from "app/a"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['app/**/**'],
        },
      ],
    }),
    test({
      code: 'import b from "@org/package"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['@org/package/*'],
        },
      ],
    }),
    // exports
    test({
      code: 'export {a} from "./internal.js"',
      filename: testFilePath('./internal-modules/plugins/plugin2/index.js'),
    }),
    test({
      code: 'export * from "lodash.get"',
      filename: testFilePath('./internal-modules/plugins/plugin2/index.js'),
    }),
    test({
      code: 'export {b} from "@org/package"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
    }),
    test({
      code: 'export {b} from "../../api/service"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          allow: ['**/api/*'],
        },
      ],
    }),
    test({
      code: 'export * from "jquery/dist/jquery"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          allow: ['jquery/dist/*'],
        },
      ],
    }),
    test({
      code: 'export * from "./app/index.js";\nexport * from "./app/index"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          allow: ['**/index{.js,}'],
        },
      ],
    }),
    test({
      code: `
        export class AuthHelper {

          static checkAuth(auth) {
          }
        }
      `,
    }),
    ...flatMap(getTSParsers(), (parser) => [
      test({
        code: `
          export class AuthHelper {

            public static checkAuth(auth?: string): boolean {
            }
          }
        `,
        filename: testFilePath('internal-modules/plugins/plugin.ts'),
        parser,
      }),
    ]),
    test({
      code: 'export * from "./plugin2/thing"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          forbid: ['**/api/*'],
        },
      ],
    }),
    test({
      code: 'export * from "app/a"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['app/**/**'],
        },
      ],
    }),
    test({
      code: 'export { b } from "@org/package"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['@org/package/*'],
        },
      ],
    }),
    test({
      code: 'export * from "./app/index.js";\nexport * from "./app/index"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['**/index.ts'],
        },
      ],
    }),
  ],

  invalid: [
    // imports
    test({
      code: 'import "./plugin2/index.js";\nimport "./plugin2/app/index"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          allow: ['*/index.js'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "./plugin2/app/index" is not allowed.',
          line: 2,
          column: 8,
        },
      ],
    }),
    test({
      code: 'import "./app/index.js"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      errors: [
        {
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 8,
        },
      ],
    }),
    test({
      code: 'import b from "./plugin2/internal"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      errors: [
        {
          message: 'Reaching to "./plugin2/internal" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import a from "../api/service/index"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          allow: ['**/internal-modules/*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "../api/service/index" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "@org/package/internal"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      errors: [
        {
          message: 'Reaching to "@org/package/internal" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import get from "debug/node"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      errors: [
        {
          message: 'Reaching to "debug/node" is not allowed.',
          line: 1,
          column: 17,
        },
      ],
    }),
    test({
      code: 'import "./app/index.js"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['*/app/*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 8,
        },
      ],
    }),
    test({
      code: 'import b from "@org/package"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['@org/**'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "@org/package" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import b from "app/a/b"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['app/**/**'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "app/a/b" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'import get from "lodash.get"',
      filename: testFilePath('./internal-modules/plugins/plugin2/index.js'),
      options: [
        {
          forbid: ['lodash.*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "lodash.get" is not allowed.',
          line: 1,
          column: 17,
        },
      ],
    }),
    test({
      code: 'import "./app/index.js";\nimport "./app/index"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['**/index{.js,}'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 8,
        },
        {
          message: 'Reaching to "./app/index" is not allowed.',
          line: 2,
          column: 8,
        },
      ],
    }),
    // Unsupported upstream webpack resolver case: native rules cannot execute JS resolvers.
    // test({
    // code: 'import "@/api/service";',
    // options: [{
    // forbid: ['**/api/*'],
    // }],
    // errors: [{
    // message: 'Reaching to "@/api/service" is not allowed.',
    // line: 1,
    // column: 8,
    // }],
    // settings: {
    // 'import/resolver': {
    // webpack: {
    // config: {
    // resolve: {
    // alias: {
    // '@': testFilePath('internal-modules'),
    // },
    // },
    // },
    // },
    // },
    // },
    // }),
    // exports
    test({
      code: 'export * from "./plugin2/index.js";\nexport * from "./plugin2/app/index"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          allow: ['*/index.js'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "./plugin2/app/index" is not allowed.',
          line: 2,
          column: 15,
        },
      ],
    }),
    test({
      code: 'export * from "./app/index.js"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      errors: [
        {
          message: 'Reaching to "./app/index.js" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'export {b} from "./plugin2/internal"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      errors: [
        {
          message: 'Reaching to "./plugin2/internal" is not allowed.',
          line: 1,
          column: 17,
        },
      ],
    }),
    test({
      code: 'export {a} from "../api/service/index"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          allow: ['**/internal-modules/*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "../api/service/index" is not allowed.',
          line: 1,
          column: 17,
        },
      ],
    }),
    test({
      code: 'export {b} from "@org/package/internal"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      errors: [
        {
          message: 'Reaching to "@org/package/internal" is not allowed.',
          line: 1,
          column: 17,
        },
      ],
    }),
    test({
      code: 'export {get} from "debug/node"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      errors: [
        {
          message: 'Reaching to "debug/node" is not allowed.',
          line: 1,
          column: 19,
        },
      ],
    }),
    test({
      code: 'export * from "./plugin2/thing"',
      filename: testFilePath('./internal-modules/plugins/plugin.js'),
      options: [
        {
          forbid: ['**/plugin2/*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "./plugin2/thing" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
    test({
      code: 'export * from "app/a"',
      filename: testFilePath('./internal-modules/plugins/plugin2/internal.js'),
      options: [
        {
          forbid: ['**'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "app/a" is not allowed.',
          line: 1,
          column: 15,
        },
      ],
    }),
  ],
});

// Pinned documentation examples. Babel's export-default-from example is
// unsupported; use export { default as getUser } from '../actions/getUser'.
new RuleTester().run('no-internal-modules', null as never, {
  valid: [
    {
      filename: path.join(root, 'my-project/entry.js'),
      code: "import 'source-map-support/register';\nimport { settings } from '../app';\nimport getUser from '../actions/getUser';\nexport * from 'source-map-support/register';\nexport { settings } from '../app';",
      options: [
        {
          allow: ['**/actions/*', 'source-map-support/*'],
        },
      ],
    },
    {
      filename: path.join(root, 'my-project/entry.js'),
      code: "import 'source-map-support';\nimport { getUser } from '../actions';\nexport * from 'source-map-support';\nexport { getUser } from '../actions';",
      options: [
        {
          forbid: ['**/actions/*', 'source-map-support/*'],
        },
      ],
    },
  ],
  invalid: [
    {
      filename: path.join(root, 'my-project/entry.js'),
      code: "import { settings } from './app/index';\nimport userReducer from './reducer/user';\nimport configureStore from './redux/configureStore';\nexport { settings } from './app/index';\nexport * from './reducer/user';",
      options: [
        {
          allow: ['**/actions/*', 'source-map-support/*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "./app/index" is not allowed.',
          line: 1,
          column: 26,
          endLine: 1,
          endColumn: 39,
        },
        {
          message: 'Reaching to "./reducer/user" is not allowed.',
          line: 2,
          column: 25,
          endLine: 2,
          endColumn: 41,
        },
        {
          message: 'Reaching to "./redux/configureStore" is not allowed.',
          line: 3,
          column: 28,
          endLine: 3,
          endColumn: 52,
        },
        {
          message: 'Reaching to "./app/index" is not allowed.',
          line: 4,
          column: 26,
          endLine: 4,
          endColumn: 39,
        },
        {
          message: 'Reaching to "./reducer/user" is not allowed.',
          line: 5,
          column: 15,
          endLine: 5,
          endColumn: 31,
        },
      ],
    },
    {
      filename: path.join(root, 'my-project/entry.js'),
      code: "import 'source-map-support/register';\nimport getUser from '../actions/getUser';\nexport * from 'source-map-support/register';",
      options: [
        {
          forbid: ['**/actions/*', 'source-map-support/*'],
        },
      ],
      errors: [
        {
          message: 'Reaching to "source-map-support/register" is not allowed.',
          line: 1,
          column: 8,
          endLine: 1,
          endColumn: 37,
        },
        {
          message: 'Reaching to "../actions/getUser" is not allowed.',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 41,
        },
        {
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
