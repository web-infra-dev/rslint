/**
 * @fileoverview enforce the location of single-line statements
 * @author Teddy Katz
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/nonblock-statement-body-position/nonblock-statement-body-position.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('nonblock-statement-body-position', null as never, { valid, invalid })`
 *  - The `$` unindent template tag is evaluated to its real multi-line string;
 *    plain backtick multi-line templates keep their literal leading indentation.
 *  - The local error constants (`EXPECTED_LINEBREAK` / `UNEXPECTED_LINEBREAK`)
 *    are inlined to their `{ messageId: 'expectLinebreak' | 'expectNoLinebreak' }`.
 *  - `name` / `rule` / `lang` and the type generics dropped.
 *
 * The upstream file has a single `run()` block (no `skipBabel` block) and no
 * Babel/Flow, external-fixture, or suggestion cases — nothing skipped on those
 * grounds. The `._css_` / `._json_` / `._markdown_` files don't exist for this
 * rule. There are NO `KNOWN GAPS`: every fixture is valid TypeScript and rslint
 * matches upstream exactly.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('nonblock-statement-body-position', null as never, {
  valid: [
    // 'beside' option
    'if (foo) bar;',
    'while (foo) bar;',
    'do foo; while (bar)',
    'for (;foo;) bar;',
    'for (foo in bar) baz;',
    'for (foo of bar) baz;',
    'if (foo) bar; else baz;',
    `
            if (foo) bar(
                baz
            );
        `,
    {
      code: 'if (foo) bar();',
      options: ['beside'],
    },
    {
      code: 'while (foo) bar();',
      options: ['beside'],
    },
    {
      code: 'do bar(); while (foo)',
      options: ['beside'],
    },
    {
      code: 'for (;foo;) bar();',
      options: ['beside'],
    },

    // 'below' option
    {
      code: 'if (foo)\n    bar();',
      options: ['below'],
    },
    {
      code: 'while (foo)\n    bar();',
      options: ['below'],
    },
    {
      code: 'do\n    bar();\nwhile (foo)',
      options: ['below'],
    },
    {
      code: 'for (;foo;)\n    bar();',
      options: ['below'],
    },
    {
      code: 'for (foo in bar)\n    bar();',
      options: ['below'],
    },
    {
      code: 'for (foo of bar)\n    bar();',
      options: ['below'],
    },
    {
      code: 'if (foo)\n    bar();\nelse\n    baz();',
      options: ['below'],
    },

    // 'any' option
    {
      code: 'if (foo) bar();',
      options: ['any'],
    },
    {
      code: 'if (foo)\n    bar();',
      options: ['any'],
    },

    // 'overrides' option
    {
      code: 'if (foo) bar();',
      options: ['beside', { overrides: { while: 'below' } }],
    },
    {
      code: 'while (foo)\n    bar();',
      options: ['beside', { overrides: { while: 'below' } }],
    },
    {
      code: 'while (foo)\n    bar();',
      options: ['beside', { overrides: { while: 'any' } }],
    },
    {
      code: 'while (foo) bar();',
      options: ['beside', { overrides: { while: 'any' } }],
    },
    {
      code: 'while (foo) bar();',
      options: ['any', { overrides: { while: 'beside' } }],
    },
    {
      code: ' ',
      options: ['any', { overrides: { if: 'any', else: 'any', for: 'any', while: 'any', do: 'any' } }],
    },

    // ignore 'else if'
    `
            if (foo) {
            } else if (bar) {
            }
        `,
    {
      code: 'if (foo) {\n} else if (bar) {\n}',
      options: ['below'],
    },
    `
            if (foo) {
            } else
              if (bar) {
              }
        `,
    {
      code: 'if (foo) {\n} else\n  if (bar) {\n  }',
      options: ['beside'],
    },
  ],

  invalid: [
    // 'beside' option
    {
      code: 'if (foo)\n    bar();',
      output: 'if (foo) bar();',
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'while (foo)\n    bar();',
      output: 'while (foo) bar();',
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'do\n    bar();\nwhile (foo)',
      output: 'do bar();\nwhile (foo)',
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'for (;foo;)\n    bar();',
      output: 'for (;foo;) bar();',
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'for (foo in bar)\n    baz();',
      output: 'for (foo in bar) baz();',
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'for (foo of bar)\n    baz();',
      output: 'for (foo of bar) baz();',
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'if (foo)\n    bar();\nelse\n    baz();',
      output: 'if (foo) bar();\nelse baz();',
      errors: [{ messageId: 'expectNoLinebreak' }, { messageId: 'expectNoLinebreak' }],
    },

    // 'below' option
    {
      code: 'if (foo) bar();',
      output: 'if (foo) \nbar();',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'while (foo) bar();',
      output: 'while (foo) \nbar();',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'do bar(); while (foo)',
      output: 'do \nbar(); while (foo)',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'for (;foo;) bar();',
      output: 'for (;foo;) \nbar();',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'for (foo in bar) baz();',
      output: 'for (foo in bar) \nbaz();',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'for (foo of bar) baz();',
      output: 'for (foo of bar) \nbaz();',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'if (foo) bar();\nelse baz();',
      output: 'if (foo) \nbar();\nelse \nbaz();',
      options: ['below'],
      errors: [{ messageId: 'expectLinebreak' }, { messageId: 'expectLinebreak' }],
    },

    // overrides
    {
      code: 'if (foo) bar();',
      output: 'if (foo) \nbar();',
      options: ['below', { overrides: { while: 'beside' } }],
      errors: [{ messageId: 'expectLinebreak' }],
    },
    {
      code: 'while (foo)\n    bar();',
      output: 'while (foo) bar();',
      options: ['below', { overrides: { while: 'beside' } }],
      errors: [{ messageId: 'expectNoLinebreak' }],
    },
    {
      code: 'do bar(); while (foo)',
      output: 'do \nbar(); while (foo)',
      options: ['any', { overrides: { do: 'below' } }],
      errors: [{ messageId: 'expectLinebreak' }],
    },
  ],
});
