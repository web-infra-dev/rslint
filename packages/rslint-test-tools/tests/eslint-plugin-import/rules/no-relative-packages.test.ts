// Upstream: eslint-plugin-import v2.32.0 tests and documentation.
// The original outputs remain as the upstream coverage record; Rslint
// intentionally preserves diagnostics/ranges without those automatic edits.
import fs from 'node:fs';
import path from 'node:path';
import { lint } from '@rslint/core/internal';

import { RuleTester } from '../rule-tester.js';
const root = fs.mkdtempSync(path.join(process.cwd(), '.no-relative-packages-'));
const archive = fs.readFileSync(
  path.resolve(
    import.meta.dirname,
    '../../../../../internal/plugins/import/rules/no_relative_packages/testdata/modules.txtar',
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
const valid = [
  {
    filename: 'tests/files/package/index.js',
    code: 'import foo from "./index.js"',
  },
  {
    filename: 'tests/files/package/index.js',
    code: 'import bar from "../bar"',
  },
  {
    filename: 'tests/files/package-named/index.js',
    code: 'import {foo} from "a"',
  },
  {
    filename: 'tests/files/package/index.js',
    code: 'const bar = require("../bar.js")',
  },
  {
    filename: 'tests/files/package/index.js',
    code: 'const bar = require("../not/a/file/path.js")',
  },
  { filename: 'tests/files/package/index.js', code: 'import "package"' },
  { filename: 'tests/files/package/index.js', code: 'require("../bar.js")' },
  {
    filename: 'my-project/packages/foo/index.js',
    code: "import bar from 'bar';",
  },
  { filename: 'my-project/entry.js', code: "import bar from 'bar';" },
];
const invalid = [
  {
    filename: 'tests/files/bar.js',
    code: 'import foo from "./package-named"',
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `./package-named`',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 34,
      },
    ],
    output: 'import foo from "package-named"',
  },
  {
    filename: 'tests/files/package/index.js',
    code: 'import foo from "../package-named"',
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `../package-named`',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 35,
      },
    ],
    output: 'import foo from "package-named"',
  },
  {
    filename: 'tests/files/package/index.js',
    code: 'import foo from "../package-scoped"',
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `../package-scoped`',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 36,
      },
    ],
    output: 'import foo from "@scope/package-named"',
  },
  {
    filename: 'tests/files/package-named/index.js',
    code: 'import bar from "../bar"',
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `../bar`',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 25,
      },
    ],
    output: 'import bar from "eslint-plugin-import/tests/files/bar"',
  },
  {
    filename: 'my-project/packages/foo/index.js',
    code: "import bar from '../bar';",
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `../bar`',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 25,
      },
    ],
    output: 'import bar from "bar";',
  },
  {
    filename: 'my-project/packages/foo/index.js',
    code: "import entry from '../../entry.js';",
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `../../entry.js`',
        line: 1,
        column: 19,
        endLine: 1,
        endColumn: 35,
      },
    ],
    output: 'import entry from "my-project/entry.js";',
  },
  {
    filename: 'my-project/entry.js',
    code: "import bar from './packages/bar';",
    errors: [
      {
        message:
          'Relative import from another package is not allowed. Use a package import instead of `./packages/bar`',
        line: 1,
        column: 17,
        endLine: 1,
        endColumn: 33,
      },
    ],
    output: 'import bar from "bar";',
  },
];
for (const item of [...valid, ...invalid])
  item.filename = path.join(root, item.filename);
const languageOptions = {
  parserOptions: { project: false, projectService: false },
};

new RuleTester().run('no-relative-packages', null as never, {
  valid: valid.map((item) => ({ ...item, languageOptions })),
  invalid: invalid.map((item) => ({ ...item, output: null, languageOptions })),
});

// The suite wrapper checks messages; assert ranges and edits independently.
test('upstream diagnostic ranges without unsafe autofixes', async () => {
  for (const item of invalid) {
    const result = await lint({
      config: [
        {
          plugins: ['import'],
          languageOptions,
          rules: { 'import/no-relative-packages': 'error' },
        },
      ],
      configDirectory: root,
      workingDirectory: root,
      fileContents: { [item.filename]: item.code },
    });
    assert.strictEqual(result.fileCount, 1);
    assert.strictEqual(result.ruleCount, 1);
    assert.strictEqual(result.diagnostics.length, 1);
    const diagnostic = result.diagnostics[0];
    const expected = item.errors[0];
    assert.strictEqual(diagnostic.ruleName, 'import/no-relative-packages');
    assert.strictEqual(diagnostic.messageId, '');
    assert.strictEqual(diagnostic.message, expected.message);
    assert.deepStrictEqual(diagnostic.range, {
      start: { line: expected.line, column: expected.column },
      end: { line: expected.endLine, column: expected.endColumn },
    });
    assert.ok(!diagnostic.suggestions?.length);
    assert.ok(!diagnostic.fixes?.length);
    const fixed = await lint({
      config: [
        {
          plugins: ['import'],
          languageOptions,
          rules: { 'import/no-relative-packages': 'error' },
        },
      ],
      configDirectory: root,
      workingDirectory: root,
      fileContents: { [item.filename]: item.code },
      fix: true,
    });
    assert.ok(!fixed.output || Object.keys(fixed.output).length === 0);
    assert.strictEqual(fixed.diagnostics.length, 1);
  }
});
