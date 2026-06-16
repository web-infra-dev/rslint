/**
 * @fileoverview Tests for one-var-declaration-per-line rule.
 * @author Alberto Rodríguez
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/one-var-declaration-per-line/one-var-declaration-per-line.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('one-var-declaration-per-line', null as never, { valid, invalid })`
 *  - The local `errorAt(line, column)` helper is inlined to its final
 *    `{ messageId: 'expectVarOnNewline', line, column }`.
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via
 *    tsconfig (the generated tsconfig uses `module: esnext`, so top-level
 *    `export` / `let` / `const` / `for…of` all parse).
 *  - `lang: 'js'` dropped — every fixture is valid TypeScript and runs on `.ts`.
 *
 * The single upstream `run()` block is the whole file: there is no skipBabel
 * block, no second `run()`, and no `._css_` / `._json_` / `._markdown_` file for
 * this rule. No Babel/Flow-only and no external-fixture cases exist.
 *
 * KNOWN GAPS: none. Every upstream case parses and aligns under ts-go.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('one-var-declaration-per-line', null as never, {
  valid: [
    { code: 'var a, b, c,\nd = 0;', options: ['initializations'] },
    { code: 'var a, b, c,\n\nd = 0;', options: ['initializations'] },
    { code: 'var a, b,\nc=0\nd = 0;', options: ['initializations'] },
    { code: 'let a, b;', options: ['initializations'] },
    { code: 'var a = 0; var b = 0;', options: ['initializations'] },
    'var a, b,\nc=0\nd = 0;',

    { code: 'var a,\nb,\nc,\nd = 0;', options: ['always'] },
    { code: 'var a = 0,\nb;', options: ['always'] },
    { code: 'var a = 0,\n\nb;', options: ['always'] },

    { code: 'var a; var b;', options: ['always'] },
    { code: 'for(var a = 0, b = 0;;){}', options: ['always'] },
    { code: 'for(let a = 0, b = 0;;){}', options: ['always'] },
    { code: 'for(const a = 0, b = 0;;){}', options: ['always'] },
    { code: 'for(var a in obj){}', options: ['always'] },
    { code: 'for(let a in obj){}', options: ['always'] },
    { code: 'for(const a in obj){}', options: ['always'] },
    { code: 'for(var a of arr){}', options: ['always'] },
    { code: 'for(let a of arr){}', options: ['always'] },
    { code: 'for(const a of arr){}', options: ['always'] },

    { code: 'export let a, b;', options: ['initializations'] },
    { code: 'export let a,\n b = 0;', options: ['initializations'] },
  ],

  invalid: [
    { code: 'var foo, bar;', output: 'var foo, \nbar;', options: ['always'], errors: [{ line: 1, column: 10, endLine: 1, endColumn: 13 }] },
    { code: 'var a, b;', output: 'var a, \nb;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 8 }] },
    { code: 'let a, b;', output: 'let a, \nb;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 8 }] },
    { code: 'var a, b = 0;', output: 'var a, \nb = 0;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 8 }] },
    { code: 'var a = {\n foo: bar\n}, b;', output: 'var a = {\n foo: bar\n}, \nb;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 3, column: 4 }] },
    { code: 'var a\n=0, b;', output: 'var a\n=0, \nb;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 2, column: 5 }] },
    { code: 'let a, b = 0;', output: 'let a, \nb = 0;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 8 }] },
    { code: 'const a = 0, b = 0;', output: 'const a = 0, \nb = 0;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 14 }] },

    { code: 'var foo, bar, baz = 0;', output: 'var foo, bar, \nbaz = 0;', options: ['initializations'], errors: [{ line: 1, column: 15, endLine: 1, endColumn: 22 }] },
    { code: 'var a, b, c = 0;', output: 'var a, b, \nc = 0;', options: ['initializations'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 11 }] },
    { code: 'var a, b,\nc = 0, d;', output: 'var a, b,\nc = 0, \nd;', options: ['initializations'], errors: [{ messageId: 'expectVarOnNewline', line: 2, column: 8 }] },
    { code: 'var a, b,\nc = 0, d = 0;', output: 'var a, b,\nc = 0, \nd = 0;', options: ['initializations'], errors: [{ messageId: 'expectVarOnNewline', line: 2, column: 8 }] },
    { code: 'var a\n=0, b = 0;', output: 'var a\n=0, \nb = 0;', options: ['initializations'], errors: [{ messageId: 'expectVarOnNewline', line: 2, column: 5 }] },
    { code: 'var a = {\n foo: bar\n}, b;', output: 'var a = {\n foo: bar\n}, \nb;', options: ['initializations'], errors: [{ messageId: 'expectVarOnNewline', line: 3, column: 4 }] },

    { code: 'for(var a = 0, b = 0;;){\nvar c,d;}', output: 'for(var a = 0, b = 0;;){\nvar c,\nd;}', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 2, column: 7 }] },
    { code: 'export let a, b;', output: 'export let a, \nb;', options: ['always'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 15 }] },
    { code: 'export let a, b = 0;', output: 'export let a, \nb = 0;', options: ['initializations'], errors: [{ messageId: 'expectVarOnNewline', line: 1, column: 15 }] },
  ],
});
