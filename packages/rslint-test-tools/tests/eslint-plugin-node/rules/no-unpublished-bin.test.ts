// Mirrors all cases and the package example from eslint-plugin-n v18.3.0.
// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/tests/lib/rules/no-unpublished-bin.js
import { readFileSync } from 'node:fs';
import { RuleTester } from '../rule-tester';

const archive = readFileSync(
  new URL(
    '../../../../../internal/plugins/node/rules/no_unpublished_bin/testdata/upstream.txtar',
    import.meta.url,
  ),
  'utf8',
);
const entries = archive.split(/^-- (.+) --\r?\n/m).slice(1);
if (entries.length === 0 || entries.length % 2 !== 0)
  throw new Error('Missing package fixtures');
const fixtureFiles: Record<string, string> = {};
for (let i = 0; i < entries.length; i += 2)
  fixtureFiles[entries[i]] = entries[i + 1];

new RuleTester({
  fixtureFiles,
  languageOptions: { sourceType: 'commonjs' },
}).run('no-unpublished-bin', null, {
  valid: [
    {
      filename: 'simple-ok/a.js',
      code: "'simple-ok/a.js'",
      name: 'simple-ok/a.js',
    },
    {
      filename: 'multi-ok/a.js',
      code: "'multi-ok/a.js'",
      name: 'multi-ok/a.js',
    },
    {
      filename: 'multi-ok/b.js',
      code: "'multi-ok/b.js'",
      name: 'multi-ok/b.js',
    },
    {
      filename: 'simple-files/x.js',
      code: "'simple-files/x.js'",
      name: 'simple-files/x.js',
    },
    {
      filename: 'multi-files/x.js',
      code: "'multi-files/x.js'",
      name: 'multi-files/x.js',
    },
    {
      filename: 'simple-files/lib/a.js',
      code: "'simple-files/lib/a.js'",
      name: 'simple-files/lib/a.js',
    },
    {
      filename: 'multi-files/lib/a.js',
      code: "'multi-files/lib/a.js'",
      name: 'multi-files/lib/a.js',
    },
    {
      filename: 'simple-npmignore/x.js',
      code: "'simple-npmignore/x.js'",
      name: 'simple-npmignore/x.js',
    },
    {
      filename: 'multi-npmignore/x.js',
      code: "'multi-npmignore/x.js'",
      name: 'multi-npmignore/x.js',
    },
    {
      filename: 'simple-npmignore/lib/a.js',
      code: "'simple-npmignore/lib/a.js'",
      name: 'simple-npmignore/lib/a.js',
    },
    {
      filename: 'multi-npmignore/lib/a.js',
      code: "'multi-npmignore/lib/a.js'",
      name: 'multi-npmignore/lib/a.js',
    },
    {
      filename: 'issue115/lib/a.js',
      code: "'issue115/lib/a.js'",
      name: 'issue115/lib/a.js',
    },
    {
      filename: 'brace-extglob/bin/foo.js',
      code: "'brace-extglob/bin/foo.js'",
      name: 'brace-extglob/bin/foo.js',
    },
    {
      code: "'stdin'",
      name: 'unnamed input',
    },
    {
      filename: 'simple-files/a.js',
      code: "'simple-files/a.js'",
      options: [
        {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      ],
      name: 'simple-files/a.js',
    },
    {
      filename: 'multi-files/a.js',
      code: "'multi-files/a.js'",
      options: [
        {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      ],
      name: 'multi-files/a.js',
    },
    {
      filename: 'simple-npmignore/a.js',
      code: "'simple-npmignore/a.js'",
      options: [
        {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      ],
      name: 'simple-npmignore/a.js',
    },
    {
      filename: 'multi-npmignore/a.js',
      code: "'multi-npmignore/a.js'",
      options: [
        {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      ],
      name: 'multi-npmignore/a.js',
    },
    {
      filename: 'simple-files/a.js',
      code: "'simple-files/a.js'",
      settings: {
        node: {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      },
      name: 'simple-files/a.js',
    },
    {
      filename: 'multi-files/a.js',
      code: "'multi-files/a.js'",
      settings: {
        node: {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      },
      name: 'multi-files/a.js',
    },
    {
      filename: 'simple-npmignore/a.js',
      code: "'simple-npmignore/a.js'",
      settings: {
        node: {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      },
      name: 'simple-npmignore/a.js',
    },
    {
      filename: 'multi-npmignore/a.js',
      code: "'multi-npmignore/a.js'",
      settings: {
        node: {
          convertPath: {
            'a.js': ['a.js', 'lib/a.js'],
          },
        },
      },
      name: 'multi-npmignore/a.js',
    },
    {
      name: 'documentation package example',
      code: '',
      filename: 'documentation/bin/index.js',
    },
  ],
  invalid: [
    {
      filename: 'simple-files/a.js',
      code: "'simple-files/a.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      name: 'simple-files/a.js',
    },
    {
      filename: 'multi-files/a.js',
      code: "'multi-files/a.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      name: 'multi-files/a.js',
    },
    {
      filename: 'multi-files/b.js',
      code: "'multi-files/b.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'b.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      name: 'multi-files/b.js',
    },
    {
      filename: 'simple-npmignore/a.js',
      code: "'simple-npmignore/a.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
      name: 'simple-npmignore/a.js',
    },
    {
      filename: 'multi-npmignore/a.js',
      code: "'multi-npmignore/a.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      name: 'multi-npmignore/a.js',
    },
    {
      filename: 'simple-files/x.js',
      code: "'simple-files/x.js'",
      options: [
        {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      name: 'simple-files/x.js',
    },
    {
      filename: 'multi-files/x.js',
      code: "'multi-files/x.js'",
      options: [
        {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      name: 'multi-files/x.js',
    },
    {
      filename: 'multi-files/x.js',
      code: "'multi-files/x.js'",
      options: [
        {
          convertPath: {
            'x.js': ['x.js', 'b.js'],
          },
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'b.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      name: 'multi-files/x.js',
    },
    {
      filename: 'simple-npmignore/x.js',
      code: "'simple-npmignore/x.js'",
      options: [
        {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
      name: 'simple-npmignore/x.js',
    },
    {
      filename: 'multi-npmignore/x.js',
      code: "'multi-npmignore/x.js'",
      options: [
        {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      name: 'multi-npmignore/x.js',
    },
    {
      filename: 'simple-npmignore/x.js',
      code: "'simple-npmignore/x.js'",
      options: [
        {
          convertPath: [
            {
              include: ['x.js'],
              replace: ['x.js', 'a.js'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
      name: 'simple-npmignore/x.js',
    },
    {
      filename: 'multi-npmignore/x.js',
      code: "'multi-npmignore/x.js'",
      options: [
        {
          convertPath: [
            {
              include: ['x.js'],
              replace: ['x.js', 'a.js'],
            },
          ],
        },
      ],
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      name: 'multi-npmignore/x.js',
    },
    {
      filename: 'simple-files/x.js',
      code: "'simple-files/x.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 20,
        },
      ],
      settings: {
        node: {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      },
      name: 'simple-files/x.js',
    },
    {
      filename: 'multi-files/x.js',
      code: "'multi-files/x.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      settings: {
        node: {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      },
      name: 'multi-files/x.js',
    },
    {
      filename: 'multi-files/x.js',
      code: "'multi-files/x.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'b.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 19,
        },
      ],
      settings: {
        node: {
          convertPath: {
            'x.js': ['x.js', 'b.js'],
          },
        },
      },
      name: 'multi-files/x.js',
    },
    {
      filename: 'simple-npmignore/x.js',
      code: "'simple-npmignore/x.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 24,
        },
      ],
      settings: {
        node: {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      },
      name: 'simple-npmignore/x.js',
    },
    {
      filename: 'multi-npmignore/x.js',
      code: "'multi-npmignore/x.js'",
      errors: [
        {
          messageId: 'invalidIgnored',
          message:
            "npm ignores 'a.js'. Check 'files' field of 'package.json' or '.npmignore'.",
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 23,
        },
      ],
      settings: {
        node: {
          convertPath: {
            'x.js': ['x.js', 'a.js'],
          },
        },
      },
      name: 'multi-npmignore/x.js',
    },
  ],
});
