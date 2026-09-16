// All tests and JavaScript documentation examples from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-top-level-await.js
import { readFileSync } from 'node:fs';
import { RuleTester } from '../rule-tester';

const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_top_level_await/testdata/upstream.txtar',
    import.meta.url,
  ),
  'utf8',
);
const entries = archive.split(/^-- (.+) --\r?\n/m).slice(1);
if (entries.length === 0 || entries.length % 2 !== 0)
  throw new Error('Missing top-level await fixtures');
const fixtureFiles: Record<string, string> = {};
for (let i = 0; i < entries.length; i += 2)
  fixtureFiles[entries[i]] = entries[i + 1];
const ruleTester = new RuleTester({
  fixtureFiles,
  languageOptions: { sourceType: 'module' },
});
const forbiddenAt = (
  line: number,
  column: number,
  endLine: number,
  endColumn: number,
) => ({
  messageId: 'forbidden',
  message: 'Top-level `await` is forbidden in published modules.',
  line,
  column,
  endLine,
  endColumn,
});

// Unnamed input is represented by a file outside any package. tsgo parses
// await using directly, including the upstream TypeScript-parser cases.
ruleTester.run(
  'no-top-level-await',
  {},
  {
    valid: [
      // Published synchronous code.
      {
        name: 'upstream valid 1',
        code: "import * as foo from 'foo'",
        filename: 'simple-files/lib/a.js',
      },
      {
        name: 'upstream valid 2',
        code: 'for (const e of iterate()) { /* ... */ }',
        filename: 'simple-files/lib/a.js',
      },
      // Non-top-level await.
      {
        name: 'upstream valid 3',
        code: "async function fn () { const foo = await import('foo') }",
        filename: 'simple-bin/lib/a.js',
      },
      {
        name: 'upstream valid 4',
        code: 'async function fn () { for await (const e of asyncIterate()) { /* ... */ } }',
        filename: 'simple-bin/lib/a.js',
      },
      {
        name: 'upstream valid 5',
        code: "const fn = async () => await import('foo')",
        filename: 'simple-bin/lib/a.js',
      },
      {
        name: 'upstream valid 6',
        code: 'const fn = async () => { for await (const e of asyncIterate()) { /* ... */ } }',
        filename: 'simple-bin/lib/a.js',
      },
      // Unpublished files.
      {
        name: 'upstream valid 7',
        code: "const foo = await import('foo')",
        filename: 'simple-files/src/a.js',
      },
      {
        name: 'upstream valid 8',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-files/src/a.js',
      },
      {
        name: 'upstream valid 9',
        code: "const foo = await import('foo')",
        filename: 'dot-slash-files/src/a.js',
      },
      {
        name: 'upstream valid 10',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'dot-slash-files/src/a.js',
      },
      {
        name: 'upstream valid 11',
        code: "const foo = await import('foo')",
        filename: 'slash-files/src/a.js',
      },
      {
        name: 'upstream valid 12',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'slash-files/src/a.js',
      },
      {
        name: 'upstream valid 13',
        code: "const foo = await import('foo')",
        filename: 'simple-npmignore/src/a.js',
      },
      {
        name: 'upstream valid 14',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-npmignore/src/a.js',
      },
      // ignoreBin.
      {
        name: 'upstream valid 15',
        code: "const foo = await import('foo')",
        filename: 'simple-bin/a.js',
        options: [{ ignoreBin: true }],
      },
      {
        name: 'upstream valid 16',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-bin/a.js',
        options: [{ ignoreBin: true }],
      },
      {
        name: 'upstream valid 17',
        code: "#!/usr/bin/env node\nconst foo = await import('foo')",
        filename: 'simple-files/lib/a.js',
        options: [{ ignoreBin: true }],
      },
      {
        name: 'upstream valid 18',
        code: '#!/usr/bin/env node\nfor await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-files/lib/a.js',
        options: [{ ignoreBin: true }],
      },
      // await using.
      {
        name: 'upstream valid 19',
        code: 'async function f() { await using foo = x }',
        filename: 'simple-files/lib/a.js',
      },
      // convertPath.
      {
        name: 'upstream valid 20',
        code: "const foo = await import('foo')",
        filename: 'simple-files/test/a.ts',
        options: [
          { convertPath: { 'src/**/*': ['src/(.+).ts', 'lib/$1.js'] } },
        ],
      },
      // Unknown files.
      {
        name: 'upstream valid 21',
        code: "const foo = await import('foo')",
      },
      {
        name: 'upstream valid 22',
        code: "const foo = await import('foo')",
        filename: 'unknown.js',
      },
      {
        name: 'documentation',
        code: "#!/usr/bin/env node\nconst foo = await import('foo');\nfor await (const e of asyncIterate()) {\n    // ...\n}",
        filename: 'simple-files/lib/a.js',
        options: [{ ignoreBin: true }],
      },
    ],
    invalid: [
      // Published files, including executables by default.
      {
        name: 'upstream invalid 1',
        code: "const foo = await import('foo')",
        filename: 'simple-files/lib/a.js',
        errors: [forbiddenAt(1, 13, 1, 32)],
      },
      {
        name: 'upstream invalid 2',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-files/lib/a.js',
        errors: [forbiddenAt(1, 1, 1, 52)],
      },
      {
        name: 'upstream invalid 3',
        code: "const foo = await import('foo')",
        filename: 'dot-slash-files/lib/a.js',
        errors: [forbiddenAt(1, 13, 1, 32)],
      },
      {
        name: 'upstream invalid 4',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'dot-slash-files/lib/a.js',
        errors: [forbiddenAt(1, 1, 1, 52)],
      },
      {
        name: 'upstream invalid 5',
        code: "const foo = await import('foo')",
        filename: 'slash-files/lib/a.js',
        errors: [forbiddenAt(1, 13, 1, 32)],
      },
      {
        name: 'upstream invalid 6',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'slash-files/lib/a.js',
        errors: [forbiddenAt(1, 1, 1, 52)],
      },
      {
        name: 'upstream invalid 7',
        code: "const foo = await import('foo')",
        filename: 'simple-npmignore/lib/a.js',
        errors: [forbiddenAt(1, 13, 1, 32)],
      },
      {
        name: 'upstream invalid 8',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-npmignore/lib/a.js',
        errors: [forbiddenAt(1, 1, 1, 52)],
      },
      {
        name: 'upstream invalid 9',
        code: "const foo = await import('foo')",
        filename: 'simple-bin/a.js',
        errors: [forbiddenAt(1, 13, 1, 32)],
      },
      {
        name: 'upstream invalid 10',
        code: 'for await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-bin/a.js',
        errors: [forbiddenAt(1, 1, 1, 52)],
      },
      {
        name: 'upstream invalid 11',
        code: "#!/usr/bin/env node\nconst foo = await import('foo')",
        filename: 'simple-files/lib/a.js',
        errors: [forbiddenAt(2, 13, 2, 32)],
      },
      {
        name: 'upstream invalid 12',
        code: '#!/usr/bin/env node\nfor await (const e of asyncIterate()) { /* ... */ }',
        filename: 'simple-files/lib/a.js',
        errors: [forbiddenAt(2, 1, 2, 52)],
      },
      // convertPath.
      {
        name: 'upstream invalid 13',
        code: "const foo = await import('foo')",
        filename: 'simple-files/src/a.ts',
        options: [
          { convertPath: { 'src/**/*': ['src/(.+).ts', 'lib/$1.js'] } },
        ],
        errors: [forbiddenAt(1, 13, 1, 32)],
      },
      // await using.
      {
        name: 'upstream invalid 14',
        code: 'await using foo = x',
        filename: 'simple-files/lib/a.js',
        errors: [forbiddenAt(1, 1, 1, 20)],
      },
      {
        name: 'documentation',
        code: "const foo = await import('foo');\nfor await (const e of asyncIterate()) {\n    // ...\n}",
        filename: 'simple-files/lib/a.js',
        errors: [forbiddenAt(1, 13, 1, 32), forbiddenAt(2, 1, 4, 2)],
      },
    ],
  },
);
