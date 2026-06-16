/**
 * @fileoverview Tests for padded-blocks rule.
 * @author Mathias Schreck <https://github.com/lo1tuma>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/padded-blocks/padded-blocks.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('padded-blocks', null as never, { valid, invalid })`
 *  - `name` / `rule` / `import` lines dropped.
 *  - `parserOptions` (ecmaVersion 6 / 2022) dropped — rslint resolves via tsconfig
 *    and ts-go parses class fields / class static blocks without an ecmaVersion knob.
 *  - Errors pinning `messageId` (missingPadBlock / extraPadBlock) + line/column/
 *    endLine/endColumn are kept; the RuleTester renders the plugin's own
 *    `meta.messages` text and asserts message + positions where upstream pins them.
 *    (Both messages are static strings with no `data` interpolation.)
 *
 * The upstream file is a single `run()` block (no skipBabel second block), uses
 * only plain single-line `\n`/`\r\n` string literals (NO `$` unindent tags), has
 * NO Babel/Flow cases, NO spread/helper error builders, NO `readFileSync` external
 * fixtures, and NO `suggestions`. The `._css_` / `._json_` / `._markdown_` test
 * files don't exist for this rule.
 *
 * Five invalid cases surface a real rslint<->upstream divergence — the autofix
 * `output` only. For all five, rslint's DIAGNOSTICS match upstream exactly (same
 * count, message, and line/column/endLine/endColumn); the difference is the fix
 * result: upstream's RuleTester runs a single fix pass (`verifyAfterFix: false`),
 * whereas rslint fixes to a stable point (multi-pass), so a first-pass fix that
 * opens a single-line block onto its own lines then triggers a second-pass blank-
 * line padding. They are moved to the KNOWN GAPS block at the bottom of this file.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('padded-blocks', null as never, {
  valid: [
    '{\n\na();\n\n}',
    '{\n\n\na();\n\n\n}',
    '{\n\n//comment\na();\n\n}',
    '{\n\na();\n//comment\n\n}',
    '{\n\na()\n//comment\n\n}',
    '{\n\na = 1\n\n}',
    '{//comment\n\na();\n\n}',
    '{ /* comment */\n\na();\n\n}',
    '{ /* comment \n */\n\na();\n\n}',
    '{ /* comment \n */ /* another comment \n */\n\na();\n\n}',
    '{ /* comment \n */ /* another comment \n */\n\na();\n\n/* comment \n */ /* another comment \n */}',

    '{\n\na();\n\n/* comment */ }',
    { code: '{\n\na();\n\n/* comment */ }', options: ['always'] },
    { code: '{\n\na();\n\n/* comment */ }', options: [{ blocks: 'always' }] },

    { code: 'switch (a) {}', options: [{ switches: 'always' }] },
    { code: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n\n}', options: ['always'] },
    { code: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n\n}', options: [{ switches: 'always' }] },
    { code: 'switch (a) {\n\n//comment\ncase 0: foo();//comment\n\n}', options: [{ switches: 'always' }] },
    { code: 'switch (a) {//comment\n\ncase 0: foo();\ncase 1: bar();\n\n/* comment */}', options: [{ switches: 'always' }] },

    { code: 'class A{\n\nfoo(){}\n\n}' },
    { code: 'class A{\n\nfoo(){}\n\n}', options: ['always'] },
    { code: 'class A{}', options: [{ classes: 'always' }] },
    { code: 'class A{\n\n}', options: [{ classes: 'always' }] },
    { code: 'class A{\n\nfoo(){}\n\n}', options: [{ classes: 'always' }] },

    { code: '{\na();\n}', options: ['never'] },
    { code: '{\na();}', options: ['never'] },
    { code: '{a();\n}', options: ['never'] },
    { code: '{a();}', options: ['never'] },
    { code: '{a();}', options: ['always', { allowSingleLineBlocks: true }] },
    { code: '{\n\na();\n\n}', options: ['always', { allowSingleLineBlocks: true }] },
    { code: '{//comment\na();}', options: ['never'] },
    { code: '{\n//comment\na()\n}', options: ['never'] },
    { code: '{a();//comment\n}', options: ['never'] },
    { code: '{\na();\n//comment\n}', options: ['never'] },
    { code: '{\na()\n//comment\n}', options: ['never'] },
    { code: '{\na()\n//comment\nb()\n}', options: ['never'] },
    { code: 'function a() {\n/* comment */\nreturn;\n/* comment*/\n}', options: ['never'] },
    { code: '{\n// comment\ndebugger;\n// comment\n}', options: ['never'] },
    { code: '{\n\n// comment\nif (\n// comment\n a) {}\n\n }', options: ['always'] },
    { code: '{\n// comment\nif (\n// comment\n a) {}\n }', options: ['never'] },
    { code: '{\n// comment\nif (\n// comment\n a) {}\n }', options: [{ blocks: 'never' }] },

    { code: 'switch (a) {\ncase 0: foo();\n}', options: ['never'] },
    { code: 'switch (a) {\ncase 0: foo();\n}', options: [{ switches: 'never' }] },

    { code: 'class A{\nfoo(){}\n}', options: ['never'] },
    { code: 'class A{\nfoo(){}\n}', options: [{ classes: 'never' }] },

    { code: 'class A{\n\nfoo;\n\n}' },
    { code: 'class A{\nfoo;\n}', options: ['never'] },

    { code: '{\n\na();\n/* comment */ }', options: ['start'] },
    { code: '{\n\na();\n/* comment */ }', options: [{ blocks: 'start' }] },
    { code: 'switch (a) {}', options: [{ switches: 'start' }] },
    { code: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n}', options: ['start'] },
    { code: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n}', options: [{ switches: 'start' }] },
    { code: 'switch (a) {\n\n//comment\ncase 0: foo();//comment\n}', options: [{ switches: 'start' }] },
    { code: 'switch (a) {//comment\n\ncase 0: foo();\ncase 1: bar();\n/* comment */}', options: [{ switches: 'start' }] },
    { code: 'class A{\n\nfoo(){}\n}', options: ['start'] },
    { code: 'class A{}', options: [{ classes: 'start' }] },
    { code: 'class A{\n}', options: [{ classes: 'start' }] },
    { code: 'class A{\n\nfoo(){}\n}', options: [{ classes: 'start' }] },

    { code: '{\na();\n\n/* comment */ }', options: ['end'] },
    { code: '{\na();\n\n/* comment */ }', options: [{ blocks: 'end' }] },
    { code: 'switch (a) {}', options: [{ switches: 'end' }] },
    { code: 'switch (a) {\ncase 0: foo();\ncase 1: bar();\n\n}', options: ['end'] },
    { code: 'switch (a) {\ncase 0: foo();\ncase 1: bar();\n\n}', options: [{ switches: 'end' }] },
    { code: 'switch (a) {\n//comment\ncase 0: foo();//comment\n\n}', options: [{ switches: 'end' }] },
    { code: 'switch (a) {//comment\ncase 0: foo();\ncase 1: bar();\n\n/* comment */}', options: [{ switches: 'end' }] },
    { code: 'class A{\nfoo(){}\n\n}', options: ['end'] },
    { code: 'class A{}', options: [{ classes: 'end' }] },
    { code: 'class A{\n}', options: [{ classes: 'end' }] },
    { code: 'class A{\nfoo(){}\n\n}', options: [{ classes: 'end' }] },

    // Ignore block statements if not configured
    { code: '{\na();\n}', options: [{ switches: 'always' }] },
    { code: '{\n\na();\n\n}', options: [{ switches: 'never' }] },

    // Ignore switch statements if not configured
    { code: 'switch (a) {\ncase 0: foo();\ncase 1: bar();\n}', options: [{ blocks: 'always', classes: 'always' }] },
    { code: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n\n}', options: [{ blocks: 'never', classes: 'never' }] },

    // Ignore class statements if not configured
    { code: 'class A{\nfoo(){}\n}', options: [{ blocks: 'always' }] },
    { code: 'class A{\n\nfoo(){}\n\n}', options: [{ blocks: 'never' }] },

    // class static blocks
    {
      code: 'class C {\n\n static {\n\nfoo;\n\n} \n\n}',
      options: ['always'],
    },
    {
      code: 'class C {\n\n static {// comment\n\nfoo;\n\n/* comment */} \n\n}',
      options: ['always'],
    },
    {
      code: 'class C {\n\n static {\n\n// comment\nfoo;\n// comment\n\n} \n\n}',
      options: ['always'],
    },
    {
      code: 'class C {\n\n static {\n\n// comment\n\nfoo;\n\n// comment\n\n} \n\n}',
      options: ['always'],
    },
    {
      code: 'class C {\n\n static { foo; } \n\n}',
      options: ['always', { allowSingleLineBlocks: true }],
    },
    {
      code: 'class C {\n\n static\n { foo; } \n\n}',
      options: ['always', { allowSingleLineBlocks: true }],
    },
    {
      code: 'class C {\n\n static {} static {\n} static {\n\n} \n\n}', // empty blocks are ignored
      options: ['always'],
    },
    {
      code: 'class C {\n\n static\n\n { foo; } \n\n}',
      options: ['always', { allowSingleLineBlocks: true }],
    },
    {
      code: 'class C {\n static {\n\nfoo;\n\n} \n}',
      options: [{ blocks: 'always', classes: 'never' }], // "blocks" applies to static blocks
    },
    {
      code: 'class C {\n static {\nfoo;\n} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static {foo;} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static\n {\nfoo;\n} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static\n\n {\nfoo;\n} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static\n\n {foo;} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static {// comment\nfoo;\n/* comment */} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static {\n// comment\nfoo;\n// comment\n} \n}',
      options: ['never'],
    },
    {
      code: 'class C {\n static {} static {\n} static {\n\n} \n}', // empty blocks are ignored
      options: ['never'],
    },
    {
      code: 'class C {\n\n static {\nfoo;\n} \n\n}',
      options: [{ blocks: 'never', classes: 'always' }], // "blocks" applies to static blocks
    },
    {
      code: 'class C {\n\n static {\nfoo;\n} static {\n\nfoo;\n\n} \n\n}',
      options: [{ classes: 'always' }], // if there's no "blocks" in the object option, static blocks are ignored
    },
    {
      code: 'class C {\n static {\nfoo;\n} static {\n\nfoo;\n\n} \n}',
      options: [{ classes: 'never' }], // if there's no "blocks" in the object option, static blocks are ignored
    },

  ],
  invalid: [
    {
      code: '{\n//comment\na();\n\n}',
      output: '{\n\n//comment\na();\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{ //comment\na();\n\n}',
      output: '{ //comment\n\na();\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 3,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n//comment\n}',
      output: '{\n\na();\n//comment\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 4,
          column: 10,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na()\n//comment\n}',
      output: '{\n\na()\n//comment\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 4,
          column: 10,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\na();\n\n}',
      output: '{\n\na();\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n}',
      output: '{\n\na();\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 3,
          column: 5,
          endLine: 4,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\na();\n}',
      output: '{\n\na();\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 2,
          column: 5,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\r\na();\r\n}',
      output: '{\n\r\na();\r\n\n}',
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 2,
          column: 5,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\ncase 0: foo();\ncase 1: bar();\n}',
      output: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 12,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 3,
          column: 15,
          endLine: 4,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\ncase 0: foo();\ncase 1: bar();\n}',
      output: 'switch (a) {\n\ncase 0: foo();\ncase 1: bar();\n\n}',
      options: [{ switches: 'always' }],
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 12,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 3,
          column: 15,
          endLine: 4,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n//comment\ncase 0: foo();//comment\n}',
      output: 'switch (a) {\n\n//comment\ncase 0: foo();//comment\n\n}',
      options: [{ switches: 'always' }],
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 12,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 3,
          column: 24,
          endLine: 4,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\nconstructor(){}\n}',
      output: 'class A {\n\nconstructor(){}\n\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 9,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 2,
          column: 16,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\nconstructor(){}\n}',
      output: 'class A {\n\nconstructor(){}\n\n}',
      options: [{ classes: 'always' }],
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 9,
          endLine: 2,
          endColumn: 1,
        },
        {
          messageId: 'missingPadBlock',
          line: 2,
          column: 16,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\na()\n//comment\n\n}',
      output: '{\na()\n//comment\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 10,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n\n}',
      output: '{\na();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 5,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\r\n\r\na();\r\n\r\n}',
      output: '{\na();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 5,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\n\n  a();\n\n\n}',
      output: '{\n  a();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 4,
          endColumn: 3,
        },
        {
          messageId: 'extraPadBlock',
          line: 4,
          column: 7,
          endLine: 7,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n}',
      output: '{\na();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\n\ta();\n}',
      output: '{\n\ta();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 2,
        },
      ],
    },
    {
      code: '{\na();\n\n}',
      output: '{\na();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 2,
          column: 5,
          endLine: 4,
          endColumn: 1,
        },
      ],
    },
    {
      code: '  {\n    a();\n\n  }',
      output: '  {\n    a();\n  }',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 2,
          column: 9,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: '{\n// comment\nif (\n// comment\n a) {}\n\n}',
      output: '{\n\n// comment\nif (\n// comment\n a) {}\n\n}',
      options: ['always'],
      errors: [
        {
          messageId: 'missingPadBlock',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\n// comment\nif (\n// comment\n a) {}\n}',
      output: '{\n// comment\nif (\n// comment\n a) {}\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\n// comment\nif (\n// comment\n a) {}\n}',
      output: '{\n// comment\nif (\n// comment\n a) {}\n}',
      options: [{ blocks: 'never' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n\ncase 0: foo();\n\n}',
      output: 'switch (a) {\ncase 0: foo();\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 12,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 15,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n\ncase 0: foo();\n}',
      output: 'switch (a) {\ncase 0: foo();\n}',
      options: [{ switches: 'never' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 12,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\ncase 0: foo();\n\n  }',
      output: 'switch (a) {\ncase 0: foo();\n  }',
      options: [{ switches: 'never' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 2,
          column: 15,
          endLine: 4,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\nconstructor(){\nfoo();\n}\n}',
      options: ['never'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 9,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 14,
          endLine: 5,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 5,
          column: 7,
          endLine: 7,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 7,
          column: 2,
          endLine: 9,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\nconstructor(){\n\nfoo();\n\n}\n}',
      options: [{ classes: 'never' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 9,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 7,
          column: 2,
          endLine: 9,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\nconstructor(){\nfoo();\n}\n}',
      options: [{ blocks: 'never', classes: 'never' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 9,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 14,
          endLine: 5,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 5,
          column: 7,
          endLine: 7,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 7,
          column: 2,
          endLine: 9,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n\n}',
      output: '{\n\na();\n}',
      options: ['start'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 5,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n\n}',
      output: '{\n\na();\n}',
      options: [{ blocks: 'start' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 5,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n\ncase 0: foo();\n\n}',
      output: 'switch (a) {\n\ncase 0: foo();\n}',
      options: ['start'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 15,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n\ncase 0: foo();\n\n  }',
      output: 'switch (a) {\n\ncase 0: foo();\n  }',
      options: [{ switches: 'start' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 15,
          endLine: 5,
          endColumn: 3,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\n\nconstructor(){\n\nfoo();\n}\n}',
      options: ['start'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 5,
          column: 7,
          endLine: 7,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 7,
          column: 2,
          endLine: 9,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n}',
      options: [{ classes: 'start' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 7,
          column: 2,
          endLine: 9,
          endColumn: 1,
        },
      ],
    },

    {
      code: '{\n\na();\n\n}',
      output: '{\na();\n\n}',
      options: ['end'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: '{\n\na();\n\n}',
      output: '{\na();\n\n}',
      options: [{ blocks: 'end' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 1,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n\ncase 0: foo();\n\n}',
      output: 'switch (a) {\ncase 0: foo();\n\n}',
      options: ['end'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 12,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'switch (a) {\n\ncase 0: foo();\n\n  }',
      output: 'switch (a) {\ncase 0: foo();\n\n  }',
      options: [{ switches: 'end' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 12,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\nconstructor(){\nfoo();\n\n}\n\n}',
      options: ['end'],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 9,
          endLine: 3,
          endColumn: 1,
        },
        {
          messageId: 'extraPadBlock',
          line: 3,
          column: 14,
          endLine: 5,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'class A {\n\nconstructor(){\n\nfoo();\n\n}\n\n}',
      output: 'class A {\nconstructor(){\n\nfoo();\n\n}\n\n}',
      options: [{ classes: 'end' }],
      errors: [
        {
          messageId: 'extraPadBlock',
          line: 1,
          column: 9,
          endLine: 3,
          endColumn: 1,
        },
      ],
    },
    {
      code: 'function foo() { // a\n\n  b;\n}',
      output: 'function foo() { // a\n  b;\n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }],
    },
    {
      code: 'function foo() { /* a\n */\n\n  bar;\n}',
      output: 'function foo() { /* a\n */\n  bar;\n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }],
    },
    {
      code: 'function foo() {\n\n  bar;\n/* a\n */}',
      output: 'function foo() {\n\n  bar;\n\n/* a\n */}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }],
    },
    {
      code: 'function foo() { /* a\n */\n/* b\n */\n  bar;\n}',
      output: 'function foo() { /* a\n */\n\n/* b\n */\n  bar;\n\n}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }, { messageId: 'missingPadBlock' }],
    },
    {
      code: 'function foo() { /* a\n */ /* b\n */\n  bar;\n}',
      output: 'function foo() { /* a\n */ /* b\n */\n\n  bar;\n\n}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }, { messageId: 'missingPadBlock' }],
    },
    {
      code: 'function foo() { /* a\n */ /* b\n */\n  bar;\n/* c\n *//* d\n */}',
      output: 'function foo() { /* a\n */ /* b\n */\n\n  bar;\n\n/* c\n *//* d\n */}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }, { messageId: 'missingPadBlock' }],
    },
    {
      code: 'class A{\nfoo;\n}',
      output: 'class A{\n\nfoo;\n\n}',
      errors: [{ messageId: 'missingPadBlock' }, { messageId: 'missingPadBlock' }],
    },
    {
      code: 'class A{\n\nfoo;\n\n}',
      output: 'class A{\nfoo;\n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }, { messageId: 'extraPadBlock' }],
    },

    // class static blocks
    {
      code: 'class C {\n\n static {\nfoo;\n\n} \n\n}',
      output: 'class C {\n\n static {\n\nfoo;\n\n} \n\n}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }],
    },
    {
      code: 'class C {\n\n static\n {\nfoo;\n\n} \n\n}',
      output: 'class C {\n\n static\n {\n\nfoo;\n\n} \n\n}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }],
    },
    {
      code: 'class C {\n\n static\n\n {\nfoo;\n\n} \n\n}',
      output: 'class C {\n\n static\n\n {\n\nfoo;\n\n} \n\n}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }],
    },
    {
      code: 'class C {\n\n static {\n\nfoo;\n} \n\n}',
      output: 'class C {\n\n static {\n\nfoo;\n\n} \n\n}',
      options: ['always'],
      errors: [{ messageId: 'missingPadBlock' }],
    },
    {
      code: 'class C {\n\n static {\nfoo;\n} \n\n}',
      output: 'class C {\n\n static {\n\nfoo;\n\n} \n\n}',
      options: ['always'],
      errors: [
        { messageId: 'missingPadBlock' },
        { messageId: 'missingPadBlock' },
      ],
    },
    {
      code: 'class C {\n\n static {// comment\nfoo;\n/* comment */} \n\n}',
      output: 'class C {\n\n static {// comment\n\nfoo;\n\n/* comment */} \n\n}',
      options: ['always'],
      errors: [
        { messageId: 'missingPadBlock' },
        { messageId: 'missingPadBlock' },
      ],
    },
    {
      code: 'class C {\n\n static {\n// comment\nfoo;\n// comment\n} \n\n}',
      output: 'class C {\n\n static {\n\n// comment\nfoo;\n// comment\n\n} \n\n}',
      options: ['always'],
      errors: [
        { messageId: 'missingPadBlock' },
        { messageId: 'missingPadBlock' },
      ],
    },
    {
      code: 'class C {\n\n static {\n// comment\n\nfoo;\n\n// comment\n} \n\n}',
      output: 'class C {\n\n static {\n\n// comment\n\nfoo;\n\n// comment\n\n} \n\n}',
      options: ['always'],
      errors: [
        { messageId: 'missingPadBlock' },
        { messageId: 'missingPadBlock' },
      ],
    },
    {
      code: 'class C {\n static {\nfoo;\n} \n}',
      output: 'class C {\n static {\n\nfoo;\n\n} \n}',
      options: [{ blocks: 'always', classes: 'never' }], // "blocks" applies to static blocks
      errors: [
        { messageId: 'missingPadBlock' },
        { messageId: 'missingPadBlock' },
      ],
    },
    {
      code: 'class C {\n static {\n\nfoo;\n} \n}',
      output: 'class C {\n static {\nfoo;\n} \n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }],
    },
    {
      code: 'class C {\n static\n {\n\nfoo;\n} \n}',
      output: 'class C {\n static\n {\nfoo;\n} \n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }],
    },
    {
      code: 'class C {\n static\n\n {\n\nfoo;\n} \n}',
      output: 'class C {\n static\n\n {\nfoo;\n} \n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }],
    },
    {
      code: 'class C {\n static {\nfoo;\n\n} \n}',
      output: 'class C {\n static {\nfoo;\n} \n}',
      options: ['never'],
      errors: [{ messageId: 'extraPadBlock' }],
    },
    {
      code: 'class C {\n static {\n\nfoo;\n\n} \n}',
      output: 'class C {\n static {\nfoo;\n} \n}',
      options: ['never'],
      errors: [
        { messageId: 'extraPadBlock' },
        { messageId: 'extraPadBlock' },
      ],
    },
    {
      code: 'class C {\n static {// comment\n\nfoo;\n\n/* comment */} \n}',
      output: 'class C {\n static {// comment\nfoo;\n/* comment */} \n}',
      options: ['never'],
      errors: [
        { messageId: 'extraPadBlock' },
        { messageId: 'extraPadBlock' },
      ],
    },
    {
      code: 'class C {\n static {\n\n// comment\nfoo;\n// comment\n\n} \n}',
      output: 'class C {\n static {\n// comment\nfoo;\n// comment\n} \n}',
      options: ['never'],
      errors: [
        { messageId: 'extraPadBlock' },
        { messageId: 'extraPadBlock' },
      ],
    },
    {
      code: 'class C {\n\n static {\n\nfoo;\n\n} \n\n}',
      output: 'class C {\n\n static {\nfoo;\n} \n\n}',
      options: [{ blocks: 'never', classes: 'always' }], // "blocks" applies to static blocks
      errors: [
        { messageId: 'extraPadBlock' },
        { messageId: 'extraPadBlock' },
      ],
    },
  ],
});

/**
 * ============================ padded-blocks — KNOWN GAPS ============================
 *
 * These five invalid cases are ported verbatim from upstream but are NOT run in the
 * green `ruleTester.run` above, because rslint's autofix `output` diverges. This is
 * NOT a diagnostic miss: for every case rslint reports the SAME diagnostics as
 * upstream — same count, same `missingPadBlock` message, and (where upstream pins
 * them) the same line/column/endLine/endColumn (verified directly against the CLI).
 *
 * The sole divergence is the fix RESULT. Upstream's RuleTester runs a SINGLE fix
 * pass (the runner sets `verifyAfterFix: false`); rslint applies fixes to a STABLE
 * point (multi-pass). When the first pass turns a single-line block into a multi-
 * line one (e.g. `{a();}` -> `{\na();\n}`), a second pass then adds the missing
 * blank-line padding (`{\n\na();\n\n}`). Upstream records only that first pass —
 * upstream itself even annotates case (5) "this is still not padded, the subsequent
 * fix below will add another pair of `\n`". These are real, documented gaps, never
 * silenced.
 *
 *   (1) {
 *         code: '{\na();}',
 *         output: '{\n\na();\n}',            // rslint: '{\n\na();\n\n}'
 *         errors: [
 *           { messageId: 'missingPadBlock', line: 1, column: 1, endLine: 2, endColumn: 1 },
 *           { messageId: 'missingPadBlock', line: 2, column: 5, endLine: 2, endColumn: 5 },
 *         ],
 *       }
 *
 *   (2) {
 *         code: '{a();\n}',
 *         output: '{\na();\n\n}',            // rslint: '{\n\na();\n\n}'
 *         errors: [
 *           { messageId: 'missingPadBlock', line: 1, column: 1, endLine: 1, endColumn: 2 },
 *           { messageId: 'missingPadBlock', line: 1, column: 6, endLine: 2, endColumn: 1 },
 *         ],
 *       }
 *
 *   (3) {
 *         code: '{a();\n}',
 *         output: '{\na();\n\n}',            // rslint: '{\n\na();\n\n}'
 *         options: [{ blocks: 'always' }],
 *         errors: [
 *           { messageId: 'missingPadBlock', line: 1, column: 1, endLine: 1, endColumn: 2 },
 *           { messageId: 'missingPadBlock', line: 1, column: 6, endLine: 2, endColumn: 1 },
 *         ],
 *       }
 *
 *   (4) {
 *         code: '{a();}',
 *         output: '{\na();\n}',              // rslint: '{\n\na();\n\n}'
 *         errors: [
 *           { messageId: 'missingPadBlock', line: 1, column: 1, endLine: 1, endColumn: 2 },
 *           { messageId: 'missingPadBlock', line: 1, column: 6, endLine: 1, endColumn: 6 },
 *         ],
 *       }
 *
 *   (5) {
 *         code: 'class C {\n\n static {foo;} \n\n}',
 *         // upstream comment: still not padded; the subsequent fix adds another pair of `\n`.
 *         output: 'class C {\n\n static {\nfoo;\n} \n\n}',  // rslint: 'class C {\n\n static {\n\nfoo;\n\n} \n\n}'
 *         options: ['always'],
 *         errors: [
 *           { messageId: 'missingPadBlock' },
 *           { messageId: 'missingPadBlock' },
 *         ],
 *       }
 *
 * ===================================================================================
 */
