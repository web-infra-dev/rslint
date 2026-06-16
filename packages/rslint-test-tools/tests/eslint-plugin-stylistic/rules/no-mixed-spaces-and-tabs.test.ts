/**
 * @fileoverview Disallow mixed spaces and tabs for indentation
 * @author Jary Niebur
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/no-mixed-spaces-and-tabs/no-mixed-spaces-and-tabs.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('no-mixed-spaces-and-tabs', null as never, { valid, invalid })`
 *  - `parserOptions` (`ecmaVersion: 6`) dropped — rslint resolves via tsconfig;
 *    template literals already parse under the default ESNext target.
 *  - The `mixedSpacesAndTabs` message takes no `data`, so each error keeps just
 *    `{ messageId, line, column, endLine, endColumn }` (all positions pinned
 *    upstream are asserted).
 *
 * Every `code` string is a single-line JS string literal whose `\t`/space/`\n`
 * escapes are preserved byte-for-byte from upstream — there is no `$` unindent
 * tag and no plain-backtick multi-line template in this file.
 *
 * This rule is NOT fixable (`meta.fixable` is undefined upstream and in the
 * installed plugin), so no case pins `output`.
 *
 * The upstream file contains NO `readFileSync` external-fixture cases, NO
 * `suggestions`, NO Babel/Flow cases, and NO second (skipBabel) `run()` block.
 * The `._css_` / `._json_` / `._markdown_` test files don't exist for this rule.
 *
 * No case surfaces a rslint<->upstream gap, so the KNOWN GAPS block below is
 * intentionally empty.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-mixed-spaces-and-tabs', null as never, {
  valid: [
    'foo',
    'foo \t',
    'foo\t ',
    '\tvar x = 5;',
    '\t\tvar x = 5;',
    ' var x = 5;',
    '    var x = 5;',
    ' foo\t',
    '\tfoo ',
    '\t/*\n\t * Hello\n\t */',
    '// foo\n\t/**\n\t * Hello\n\t */',
    '/*\n\n \t \n*/',
    '/*\t */ //',
    '/*\n \t*/ //',
    '/*\n\t *//*\n \t*/',
    '// \t',
    '/*\n*/\t ',
    '/* \t\n\t \n \t\n\t */ \t',
    {
      code: '\tvar x = 5,\n\t    y = 2;',
      options: [true],
    },
    {
      code: '/*\n\t */`\n\t   `;',
    },
    {
      code: '/*\n\t */var a = `\n\t   `, b = `\n\t   `/*\t \n\t \n*/;',
    },
    {
      code: '/*\t `template inside comment` */',
    },
    {
      code: 'var foo = `\t /* comment inside template\t */`;',
    },
    {
      code: '`\n\t   `;',
    },
    {
      code: '`\n\t   \n`;',
    },
    {
      code: '`\t   `;',
    },
    {
      code: 'const foo = `${console}\n\t foo`;',
    },
    {
      code: '`\t   `;`   \t`',
    },
    {
      code: '`foo${ 5 }\t    `;',
    },
    '\' \t\\\n\t multiline string\';',
    '\'\t \\\n \tmultiline string\';',
    {
      code: '\tvar x = 5,\n\t    y = 2;',
      options: ['smart-tabs'],
    },
    {
      code: '\t\t\t   foo',
      options: ['smart-tabs'],
    },
    {
      code: 'foo',
      options: ['smart-tabs'],
    },
    {
      code: 'foo \t',
      options: ['smart-tabs'],
    },
    {
      code: 'foo\t ',
      options: ['smart-tabs'],
    },
    {
      code: '\tfoo \t',
      options: ['smart-tabs'],
    },
    {
      code: '\tvar x = 5;',
      options: ['smart-tabs'],
    },
    {
      code: '\t\tvar x = 5;',
      options: ['smart-tabs'],
    },
    {
      code: ' var x = 5;',
      options: ['smart-tabs'],
    },
    {
      code: '    var x = 5;',
      options: ['smart-tabs'],
    },
    {
      code: ' foo\t',
      options: ['smart-tabs'],
    },
    {
      code: '\tfoo ',
      options: ['smart-tabs'],
    },
  ],

  invalid: [
    {
      code: 'function add(x, y) {\n\t return x + y;\n}',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: '\t ;\n/*\n\t * Hello\n\t */',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: ' \t/* comment */',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '\t // comment',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '\t var a /* comment */ = 1;',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: ' \tvar b = 1; // comment',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
      ],
    },
    {
      code: '/**/\n \t/*\n \t*/',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: '\t var x = 5, y = 2, z = 5;\n\n\t \tvar j =\t x + y;\nz *= j;',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 3,
        },
        {
          messageId: 'mixedSpacesAndTabs',
          line: 3,
          column: 1,
          endLine: 3,
          endColumn: 3,
        },
      ],
    },
    {
      code: '\tvar x = 5,\n  \t  y = 2;',
      options: [true],
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 4,
        },
      ],
    },
    {
      code: '\tvar x = 5,\n  \t  y = 2;',
      options: ['smart-tabs'],
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 2,
          endLine: 2,
          endColumn: 4,
        },
      ],
    },
    {
      code: '`foo${\n \t  5 }bar`;',
      options: ['smart-tabs'],
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: '`foo${\n\t  5 }bar`;',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: '  \t\'\';',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '\'\'\n\t ',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 2,
          column: 1,
          endLine: 2,
          endColumn: 3,
        },
      ],
    },
    {
      code: '   \tfoo',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: '\t\t\t foo',
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 3,
          endLine: 1,
          endColumn: 5,
        },
      ],
    },
    {
      code: '\t \tfoo',
      options: ['smart-tabs'],
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 2,
          endLine: 1,
          endColumn: 4,
        },
      ],
    },
    {
      code: '\t\t\t   \tfoo',
      options: ['smart-tabs'],
      errors: [
        {
          messageId: 'mixedSpacesAndTabs',
          line: 1,
          column: 6,
          endLine: 1,
          endColumn: 8,
        },
      ],
    },
  ],
});
