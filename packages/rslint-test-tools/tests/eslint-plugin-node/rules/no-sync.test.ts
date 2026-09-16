// Mirrors all eslint-plugin-n v18.3.0 tests and executable documentation examples.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-sync.js
import { readFileSync } from 'node:fs';
import { test } from 'rstack/test';
import { RuleTester } from '../rule-tester';
const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_sync/testdata/upstream.txtar',
    import.meta.url,
  ),
  'utf8',
);
const entries = archive.split(/^-- (.+) --\r?\n/m).slice(1);
if (entries.length === 0 || entries.length % 2 !== 0)
  throw new Error('Missing no-sync fixtures');
const fixtureFiles: Record<string, string> = {};
for (let i = 0; i < entries.length; i += 2)
  fixtureFiles[entries[i]] = entries[i + 1];
const ruleTester = new RuleTester({
  fixtureFiles,
  languageOptions: { parserOptions: { project: ['./tsconfig.json'] } },
});
ruleTester.run(
  'no-sync',
  {},
  {
    valid: [
      {
        name: 'Upstream valid 1',
        code: 'var foo = fs.foo.foo();',
        filename: 'input.js',
      },
      {
        name: 'Upstream valid 2',
        code: 'fs.fooSync;',
        filename: 'input.js',
      },
      {
        name: 'Upstream valid 3',
        code: 'fooSync;',
        filename: 'input.js',
      },
      {
        name: 'Upstream valid 4',
        code: '() => fooSync;',
        filename: 'input.js',
      },
      {
        name: 'Upstream valid 5',
        code: 'var foo = fs.fooSync;',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
      },
      {
        name: 'Upstream valid 6',
        code: 'var foo = fooSync;',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
      },
      {
        name: 'Upstream valid 7',
        code: 'if (true) {fs.fooSync();}',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
      },
      {
        name: 'Upstream valid 8',
        code: 'if (true) {fooSync();}',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
      },
      {
        name: 'Upstream valid 9',
        code: 'fooSync();',
        filename: 'input.js',
        options: [
          {
            ignores: ['fooSync'],
          },
        ],
      },
      {
        name: 'UpstreamTypes valid 1',
        code: '\ndeclare function fooSync(): void;\nfooSync();\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
              },
            ],
          },
        ],
      },
      {
        name: 'UpstreamTypes valid 2',
        code: '\ndeclare function fooSync(): void;\nfooSync();\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
                name: ['fooSync'],
              },
            ],
          },
        ],
      },
      {
        name: 'UpstreamTypes valid 3',
        code: '\nconst stylesheet = new CSSStyleSheet();\nstylesheet.replaceSync("body { font-size: 1.4em; } p { color: red; }");\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'lib',
                name: ['CSSStyleSheet.replaceSync'],
              },
            ],
          },
        ],
      },
      {
        name: 'UpstreamPackage valid 1',
        code: '\nimport { fooSync } from "aaa";\nfooSync();\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'package',
                package: 'aaa',
                name: ['fooSync'],
              },
            ],
          },
        ],
      },
      {
        name: 'Documentation valid 1',
        code: 'obj.sync();\n\nasync(function() {\n    // ...\n});',
        filename: 'input.js',
      },
      {
        name: 'Documentation valid 2',
        code: 'fs.readFileSync(somePath).toString();',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
      },
      {
        name: 'Documentation valid 3',
        code: 'fs.readFileSync(somePath);',
        filename: 'input.js',
        options: [
          {
            ignores: ['readFileSync'],
          },
        ],
      },
      {
        name: 'Documentation valid 4',
        code: 'import { fooSync } from "./foo"\nfooSync()',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
                path: './foo.ts',
              },
            ],
          },
        ],
      },
      {
        name: 'Documentation valid 5',
        code: 'import { Effect } from "effect"\nconst value = Effect.runSync(Effect.succeed(42))',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'package',
                package: 'effect',
              },
            ],
          },
        ],
      },
      {
        name: 'Documentation valid 6',
        code: 'const stylesheet = new CSSStyleSheet()\nstylesheet.replaceSync("body { font-size: 1.4em; } p { color: red; }")',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'lib',
              },
            ],
          },
        ],
      },
    ],
    invalid: [
      {
        name: 'Upstream invalid 1',
        code: 'var foo = fs.fooSync();',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'Upstream invalid 2',
        code: 'var foo = fs.fooSync.apply();',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'Upstream invalid 3',
        code: 'var foo = fooSync();',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 20,
          },
        ],
      },
      {
        name: 'Upstream invalid 4',
        code: 'var foo = fooSync.apply();',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 24,
          },
        ],
      },
      {
        name: 'Upstream invalid 5',
        code: 'var foo = fs.fooSync();',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: false,
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 11,
            endLine: 1,
            endColumn: 21,
          },
        ],
      },
      {
        name: 'Upstream invalid 6',
        code: 'if (true) {fs.fooSync();}',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 12,
            endLine: 1,
            endColumn: 22,
          },
        ],
      },
      {
        name: 'Upstream invalid 7',
        code: 'function someFunction() {fs.fooSync();}',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 8',
        code: 'function someFunction() {fs.fooSync();}',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 26,
            endLine: 1,
            endColumn: 36,
          },
        ],
      },
      {
        name: 'Upstream invalid 9',
        code: 'var a = function someFunction() {fs.fooSync();}',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 34,
            endLine: 1,
            endColumn: 44,
          },
        ],
      },
      {
        name: 'Upstream invalid 10',
        code: '() => {fs.fooSync();}',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
            ignores: ['barSync'],
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 1,
            column: 8,
            endLine: 1,
            endColumn: 18,
          },
        ],
      },
      {
        name: 'UpstreamTypes invalid 1',
        code: '\ndeclare function fooSync(): void;\nfooSync();\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
                path: '**/bar.ts',
              },
            ],
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'UpstreamTypes invalid 2',
        code: '\ndeclare function fooSync(): void;\nfooSync();\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
                name: ['barSync'],
              },
            ],
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'UpstreamTypes invalid 3',
        code: '\nconst stylesheet = new CSSStyleSheet();\nstylesheet.replaceSync("body { font-size: 1.4em; } p { color: red; }");\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
                name: ['CSSStyleSheet.replaceSync'],
              },
            ],
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'CSSStyleSheet.replaceSync'.",
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 23,
          },
        ],
      },
      {
        name: 'UpstreamPackage invalid 1',
        code: '\nimport { fooSync } from "aaa";\nfooSync();\n',
        filename: 'input.ts',
        options: [
          {
            ignores: [
              {
                from: 'file',
                name: ['fooSync'],
              },
            ],
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'fooSync'.",
            line: 3,
            column: 1,
            endLine: 3,
            endColumn: 10,
          },
        ],
      },
      {
        name: 'Documentation invalid 1',
        code: 'fs.existsSync(somePath);\n\nfunction foo() {\n  var contents = fs.readFileSync(somePath).toString();\n}',
        filename: 'input.js',
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'existsSync'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 14,
          },
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'readFileSync'.",
            line: 4,
            column: 18,
            endLine: 4,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'Documentation invalid 2',
        code: 'function foo() {\n  var contents = fs.readFileSync(somePath).toString();\n}\n\nvar bar = baz => fs.readFileSync(qux);',
        filename: 'input.js',
        options: [
          {
            allowAtRootLevel: true,
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'readFileSync'.",
            line: 2,
            column: 18,
            endLine: 2,
            endColumn: 33,
          },
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'readFileSync'.",
            line: 5,
            column: 18,
            endLine: 5,
            endColumn: 33,
          },
        ],
      },
      {
        name: 'Documentation invalid 3',
        code: 'fs.readdirSync(somePath);',
        filename: 'input.js',
        options: [
          {
            ignores: ['readFileSync'],
          },
        ],
        errors: [
          {
            messageId: 'noSync',
            message: "Unexpected sync method: 'readdirSync'.",
            line: 1,
            column: 1,
            endLine: 1,
            endColumn: 15,
          },
        ],
      },
    ],
  },
);
// Upstream mocks Node's dependency loader, which the native rule does not use.
for (const dependency of [
  'ts-declaration-location',
  'TypeScript parser services',
  'both dependencies',
]) {
  test.skip('upstream missing dependency: ' + dependency, () => {});
}
