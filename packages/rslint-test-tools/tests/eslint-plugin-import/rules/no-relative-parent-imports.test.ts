// Upstream: eslint-plugin-import v2.32.0 tests/src/rules/no-relative-parent-imports.js
// and docs/rules/no-relative-parent-imports.md. Babel's standard syntax uses
// the native parser; all upstream cases run without skips.
import fs from 'node:fs';
import path from 'node:path';

import { RuleTester } from '../rule-tester.js';

const root = fs.mkdtempSync(
  path.join(process.cwd(), '.no-relative-parent-imports-'),
);
const archive = fs.readFileSync(
  path.resolve(
    import.meta.dirname,
    '../../../../../internal/plugins/import/rules/no_relative_parent_imports/testdata/modules.txtar',
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

const filename = path.join(root, 'internal-modules/plugins/plugin2/index.js');
const message = (dependency: string, file = 'index.js') =>
  `Relative imports from parent directories are not allowed. Please either pass what you're importing through at runtime (dependency injection), move \`${file}\` to same directory as \`${dependency}\` or consider making \`${dependency}\` a package.`;

new RuleTester().run('no-relative-parent-imports', null as never, {
  valid: [
    { code: 'import foo from "./internal.js"', filename },
    { code: 'import foo from "./app/index.js"', filename },
    { code: 'import foo from "package"', filename },
    {
      code: 'require("./internal.js")',
      filename,
      options: [{ commonjs: true }],
    },
    {
      code: 'require("./app/index.js")',
      filename,
      options: [{ commonjs: true }],
    },
    {
      code: 'require("package")',
      filename,
      options: [{ commonjs: true }],
    },
    { code: 'import("./internal.js")', filename },
    { code: 'import("./app/index.js")', filename },
    { code: 'import(".")', filename },
    { code: 'import("path")', filename },
    { code: 'import("package")', filename },
    { code: 'import("@scope/package")', filename },
    // Documentation examples and both relocation strategies.
    {
      code: 'export default function (numbers) { return numbers.reduce((sum, n) => sum + n, 0); }',
      filename: path.join(root, 'add.js'),
    },
    {
      code: 'export default function three(add) { return add([1, 2]); }',
      filename: path.join(root, 'numbers/three.js'),
    },
    {
      code: "import add from './add';\nimport three from './numbers/three';\nconsole.log(three(add));",
      filename: path.join(root, 'use.js'),
    },
    {
      code: "import add from 'add';\nexport default function three() { return add([1,2]); }",
      filename: path.join(root, 'numbers/three.js'),
    },
    {
      code: "import foo from 'foo';\nimport a from './lib/a';",
      filename: path.join(root, 'main.js'),
    },
    {
      code: "import b from './b';",
      filename: path.join(root, 'lib/a.js'),
    },
    {
      code: "import add from './add';",
      filename: path.join(root, 'three.js'),
    },
    {
      code: "import add from './math/add';",
      filename: path.join(root, 'three.js'),
    },
  ],
  invalid: [
    {
      code: 'import foo from "../plugin.js"',
      filename,
      errors: [message('../plugin.js')],
    },
    {
      code: 'require("../plugin.js")',
      filename,
      options: [{ commonjs: true }],
      errors: [message('../plugin.js')],
    },
    {
      code: 'import("../plugin.js")',
      filename,
      errors: [message('../plugin.js')],
    },
    {
      code: 'import foo from "./../plugin.js"',
      filename,
      errors: [message('./../plugin.js')],
    },
    {
      code: 'import foo from "../../api/service"',
      filename,
      errors: [message('../../api/service')],
    },
    {
      code: 'import("../../api/service")',
      filename,
      errors: [message('../../api/service')],
    },
    {
      code: "import add from '../add';\nexport default function three() { return add([1, 2]); }",
      filename: path.join(root, 'numbers/three.js'),
      errors: [message('../add', 'three.js')],
    },
    {
      code: "import bar from '../main';",
      filename: path.join(root, 'lib/a.js'),
      errors: [message('../main', 'a.js')],
    },
  ],
});
