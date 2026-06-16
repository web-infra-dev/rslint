/**
 * @fileoverview Tests for space-before-block rule.
 * @author Mathias Schreck <https://github.com/lo1tuma>
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/space-before-blocks/space-before-blocks._js_.test.ts
 *   packages/eslint-plugin/rules/space-before-blocks/space-before-blocks._ts_.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, lang, ... })` -> `ruleTester.run('space-before-blocks', null as never, { valid, invalid })`
 *  - The local option helpers (`alwaysArgs` / `neverArgs` / `functionsOnlyArgs` /
 *    `keywordOnlyArgs` / `classesOnlyArgs` / `*OthersOffArgs`) and error helpers
 *    (`expectedSpacingError` = `{ messageId: 'missingSpace' }`,
 *    `expectedNoSpacingError` = `{ messageId: 'unexpectedSpace' }`) are inlined to
 *    their final literal values.
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via
 *    tsconfig (target `esnext`, module `esnext`), so the ES module `export` forms,
 *    class fields, and `class { static {} }` blocks parse natively.
 *  - `parser: tsParser` dropped — those two cases are already valid `.ts` (a method
 *    return-type annotation `): void{}` and `function foo(): null {}`); rslint
 *    always parses with ts-go, so they need no parser override.
 *  - The `$` unindent template tag (in the `_ts_` file) is evaluated to its real
 *    multi-line string (strip the common 8-space lead, drop leading/trailing blank
 *    lines).
 *
 * The `space-before-blocks` messages take no interpolation data, so the rendered
 * `message` text is fixed (`Missing space before opening brace.` /
 * `Unexpected space before opening brace.`). Each file has exactly one `run()`
 * block (no trailing skipBabel block); the rule has no suggestions and no
 * output-only invalid cases (every invalid pins `errors`).
 *
 * KNOWN GAPS: none. The TS-only fixtures (return-type annotations, `enum` /
 * `interface` / `namespace` / `declare module` bodies, `class { static {} }`) were
 * verified through the rslint CLI to parse cleanly under ts-go and to produce the
 * upstream-expected diagnostics, columns, and fixes — nothing is isolated.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('space-before-blocks', null as never, {
  valid: [
    // ---- from space-before-blocks._js_.test.ts ----
    'if(a) {}',
    'if(a)  {}',
    { code: 'if(a){}', options: ['never'] },
    { code: 'if(a){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'if(a) {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'if(a){ function b() {} }', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'if(a) { function b(){} }', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    'if(a)\n{}',
    { code: 'if(a)\n{}', options: ['never'] },
    'if(a) {}else {}',
    { code: 'if(a){}else{}', options: ['never'] },
    { code: 'if(a){}else{}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'if(a) {} else {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'if(a){ function b() {} }else{}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'if(a) { function b(){} } else {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    'function a() {}',
    { code: 'function a(){}', options: ['never'] },
    {
      code: 'export default class{}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
    },
    {
      code: 'export default class {}',
      options: [{ functions: 'never', keywords: 'never', classes: 'always' }],
    },
    {
      code: 'export default function a() {}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
    },
    {
      code: 'export default function a(){}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
    },
    { code: 'export function a(){}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'export function a() {}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'function a(){}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'function a() {}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'function a(){ if(b) {} }', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'function a() { if(b){} }', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    'switch(a.b(c < d)) { case \'foo\': foo(); break; default: if (a) { bar(); } }',
    'switch(a) { }',
    'switch(a)  {}',
    { code: 'switch(a.b(c < d)){ case \'foo\': foo(); break; default: if (a){ bar(); } }', options: ['never'] },
    { code: 'switch(a.b(c < d)){ case \'foo\': foo(); break; default: if (a){ bar(); } }', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'switch(a){}', options: ['never'] },
    { code: 'switch(a){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'switch(a) {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    'try {}catch(a) {}',
    { code: 'try{}catch(a){}', options: ['never'] },
    { code: 'try{}catch(a){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'try {} catch(a) {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'try{ function b() {} }catch(a){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'try { function b(){} } catch(a) {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    'for(;;) {}',
    { code: 'for(;;){}', options: ['never'] },
    { code: 'for(;;){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'for(;;) {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'for(;;){ function a() {} }', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'for(;;) { function a(){} }', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    'while(a) {}',
    { code: 'while(a){}', options: ['never'] },
    { code: 'while(a){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'while(a) {}', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    { code: 'while(a){ function b() {} }', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'while(a) { function b(){} }', options: [{ functions: 'never', keywords: 'always', classes: 'never' }] },
    {
      code: 'class test { constructor() {} }',
      options: [{ functions: 'always', keywords: 'never' }],
    },
    {
      code: 'class test { constructor(){} }',
      options: [{ functions: 'never', keywords: 'never', classes: 'always' }],
    },
    {
      code: 'class test{ constructor() {} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
    },
    {
      code: 'class test {}',
      options: [{ functions: 'never', keywords: 'never', classes: 'always' }],
    },
    {
      code: 'class test{}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
    },
    {
      code: 'class test{}',
      options: ['never'],
    },
    {
      code: 'class test {}',
    },
    { code: 'function a(){if(b) {}}', options: [{ functions: 'off', keywords: 'always', classes: 'off' }] },
    { code: 'function a() {if(b) {}}', options: [{ functions: 'off', keywords: 'always', classes: 'off' }] },
    { code: 'function a() {if(b){}}', options: [{ functions: 'always', keywords: 'off', classes: 'off' }] },
    { code: 'function a() {if(b) {}}', options: [{ functions: 'always', keywords: 'off', classes: 'off' }] },
    {
      code: 'class test { constructor(){if(a){}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'always' }],
    },
    {
      code: 'class test { constructor() {if(a){}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'always' }],
    },
    {
      code: 'class test { constructor(){if(a) {}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'always' }],
    },
    {
      code: 'class test { constructor() {if(a) {}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'always' }],
    },
    { code: 'function a(){if(b){}}', options: [{ functions: 'off', keywords: 'never', classes: 'off' }] },
    { code: 'function a() {if(b){}}', options: [{ functions: 'off', keywords: 'never', classes: 'off' }] },
    { code: 'function a(){if(b){}}', options: [{ functions: 'never', keywords: 'off', classes: 'off' }] },
    { code: 'function a(){if(b) {}}', options: [{ functions: 'never', keywords: 'off', classes: 'off' }] },
    {
      code: 'class test{ constructor(){if(a){}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'never' }],
    },
    {
      code: 'class test{ constructor() {if(a){}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'never' }],
    },
    {
      code: 'class test{ constructor(){if(a) {}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'never' }],
    },
    {
      code: 'class test{ constructor() {if(a) {}} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'never' }],
    },

    // https://github.com/eslint/eslint/issues/3769
    { code: '()=>{};', options: ['always'] },
    { code: '() => {};', options: ['never'] },

    // https://github.com/eslint/eslint/issues/1338
    'if(a) {}else{}',
    { code: 'if(a){}else {}', options: ['never'] },
    { code: 'try {}catch(a){}', options: [{ functions: 'always', keywords: 'never', classes: 'never' }] },
    { code: 'export default class{}', options: [{ functions: 'never', keywords: 'never', classes: 'always' }] },

    // https://github.com/eslint/eslint/issues/15082
    { code: 'switch(x) { case 9:{ break; } }', options: ['always'] },
    { code: 'switch(x){ case 9: { break; } }', options: ['never'] },
    { code: 'switch(x) { case (9):{ break; } }', options: ['always'] },
    { code: 'switch(x){ case (9): { break; } }', options: ['never'] },
    { code: 'switch(x) { default:{ break; } }', options: ['always'] },
    { code: 'switch(x){ default: { break; } }', options: ['never'] },

    // not conflict with `keyword-spacing`
    {
      code: '(class{ static{} })',
      options: ['always'],
    },
    {
      code: '(class { static {} })',
      options: ['never'],
    },

    // ---- from space-before-blocks._ts_.test.ts ----
    {
      code: 'enum Test{\n  KEY1 = 2,\n}',
      options: ['never'],
    },
    {
      code: 'interface Test{\n  prop1: number;\n}',
      options: ['never'],
    },
    {
      code: 'enum Test {\n  KEY1 = 2,\n}',
      options: ['always'],
    },
    {
      code: 'interface Test {\n  prop1: number;\n}',
      options: ['always'],
    },
    {
      code: 'enum Test{\n  KEY1 = 2,\n}',
      options: [{ classes: 'never' }],
    },
    {
      code: 'interface Test{\n  prop1: number;\n}',
      options: [{ classes: 'never' }],
    },
    {
      code: 'enum Test {\n  KEY1 = 2,\n}',
      options: [{ classes: 'always' }],
    },
    {
      code: 'interface Test {\n  prop1: number;\n}',
      options: [{ classes: 'always' }],
    },
    {
      code: 'interface Test{\n  prop1: number;\n}',
      options: [{ classes: 'off' }],
    },
    {
      code: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      options: ['never'],
    },
    {
      code: 'namespace Test {\n  type foo = number;\n}\ndeclare module \'foo\' {\n  type foo = number;\n}',
      options: ['always'],
    },
    {
      code: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      options: [{ modules: 'never' }],
    },
    {
      code: 'namespace Test {\n  type foo = number;\n}\ndeclare module \'foo\' {\n  type foo = number;\n}',
      options: [{ modules: 'always' }],
    },
    {
      code: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      options: [{ modules: 'off' }],
    },
  ],
  invalid: [
    // ---- from space-before-blocks._js_.test.ts ----
    {
      code: 'if(a){}',
      output: 'if(a) {}',
      errors: [{ messageId: 'missingSpace', line: 1, column: 6 }],
    },
    {
      code: 'if(a){}',
      output: 'if(a) {}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'if(a) {}',
      output: 'if(a){}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'if(a) { function a() {} }',
      output: 'if(a){ function a() {} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace', line: 1, column: 7 }],
    },
    {
      code: 'if(a) { function a() {} }',
      output: 'if(a) { function a(){} }',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace', line: 1, column: 22 }],
    },
    {
      code: 'if(a) {}',
      output: 'if(a){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'function a(){}',
      output: 'function a() {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'function a() {}',
      output: 'function a(){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'function a()    {}',
      output: 'function a(){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'function a(){ if (a){} }',
      output: 'function a() { if (a){} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace', line: 1, column: 13 }],
    },
    {
      code: 'function a() { if (a) {} }',
      output: 'function a(){ if (a) {} }',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace', line: 1, column: 14 }],
    },
    {
      code: 'function a(){}',
      output: 'function a() {}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'function a() {}',
      output: 'function a(){}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(a){}',
      output: 'switch(a) {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(a) {}',
      output: 'switch(a){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(a){}',
      output: 'switch(a) {}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(a) {}',
      output: 'switch(a){}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(a.b()){ case \'foo\': foo(); break; default: if (a) { bar(); } }',
      output: 'switch(a.b()) { case \'foo\': foo(); break; default: if (a) { bar(); } }',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(a.b()) { case \'foo\': foo(); break; default: if (a){ bar(); } }',
      output: 'switch(a.b()){ case \'foo\': foo(); break; default: if (a){ bar(); } }',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'try{}catch(a){}',
      output: 'try{}catch(a) {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'try {}catch(a) {}',
      output: 'try {}catch(a){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'try {} catch(a){}',
      output: 'try {} catch(a) {}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'try { function b() {} } catch(a) {}',
      output: 'try { function b(){} } catch(a) {}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace', line: 1, column: 20 }],
    },
    {
      code: 'try{ function b(){} }catch(a){}',
      output: 'try{ function b() {} }catch(a){}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace', line: 1, column: 18 }],
    },
    {
      code: 'for(;;){}',
      output: 'for(;;) {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'for(;;) {}',
      output: 'for(;;){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'for(;;){}',
      output: 'for(;;) {}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'for(;;) {}',
      output: 'for(;;){}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'for(;;){ function a(){} }',
      output: 'for(;;){ function a() {} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'for(;;) { function a() {} }',
      output: 'for(;;) { function a(){} }',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'while(a){}',
      output: 'while(a) {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'while(a) {}',
      output: 'while(a){}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'while(a){}',
      output: 'while(a) {}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'while(a) {}',
      output: 'while(a){}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'while(a){ function a(){} }',
      output: 'while(a){ function a() {} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'while(a) { function a() {} }',
      output: 'while(a) { function a(){} }',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'export function a() { if(b) {} }',
      output: 'export function a() { if(b){} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'export function a(){ if(b){} }',
      output: 'export function a(){ if(b) {} }',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'export function a(){}',
      output: 'export function a() {}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'export default function (a) {}',
      output: 'export default function (a){}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'export function a() {}',
      output: 'export function a(){}',
      options: [{ functions: 'never', keywords: 'always', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'class test{}',
      output: 'class test {}',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'class test{}',
      output: 'class test {}',
      options: [{ functions: 'never', keywords: 'never', classes: 'always' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'class test{ constructor(){} }',
      output: 'class test{ constructor() {} }',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'class test { constructor() {} }',
      output: 'class test { constructor(){} }',
      options: [{ functions: 'never', keywords: 'never', classes: 'always' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'class test {}',
      output: 'class test{}',
      options: [{ functions: 'always', keywords: 'never', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'class test {}',
      output: 'class test{}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'if(a){ function a(){} }',
      output: 'if(a){ function a() {} }',
      options: [{ functions: 'always', keywords: 'off', classes: 'off' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'if(a) { function a(){} }',
      output: 'if(a) { function a() {} }',
      options: [{ functions: 'always', keywords: 'off', classes: 'off' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'if(a){ function a(){} }',
      output: 'if(a) { function a(){} }',
      options: [{ functions: 'off', keywords: 'always', classes: 'off' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'if(a){ function a() {} }',
      output: 'if(a) { function a() {} }',
      options: [{ functions: 'off', keywords: 'always', classes: 'off' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'class test{ constructor(){} }',
      output: 'class test { constructor(){} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'always' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'class test{ constructor() {} }',
      output: 'class test { constructor() {} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'always' }],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'if(a){ function a() {} }',
      output: 'if(a){ function a(){} }',
      options: [{ functions: 'never', keywords: 'off', classes: 'off' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'if(a) { function a() {} }',
      output: 'if(a) { function a(){} }',
      options: [{ functions: 'never', keywords: 'off', classes: 'off' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'if(a) { function a(){} }',
      output: 'if(a){ function a(){} }',
      options: [{ functions: 'off', keywords: 'never', classes: 'off' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'if(a) { function a() {} }',
      output: 'if(a){ function a() {} }',
      options: [{ functions: 'off', keywords: 'never', classes: 'off' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'class test { constructor(){} }',
      output: 'class test{ constructor(){} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'class test { constructor() {} }',
      output: 'class test{ constructor() {} }',
      options: [{ functions: 'off', keywords: 'off', classes: 'never' }],
      errors: [{ messageId: 'unexpectedSpace' }],
    },

    // https://github.com/eslint/eslint/issues/13553
    {
      code: 'class A { foo(bar: string): void{} }',
      output: 'class A { foo(bar: string): void {} }',
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'function foo(): null {}',
      output: 'function foo(): null{}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },

    // https://github.com/eslint/eslint/issues/15082 regression tests (only blocks after switch case colons should be excluded)
    {
      code: 'label:{}',
      output: 'label: {}',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'label: {}',
      output: 'label:{}',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(x) { case 9: label:{ break; } }',
      output: 'switch(x) { case 9: label: { break; } }',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(x){ case 9: label: { break; } }',
      output: 'switch(x){ case 9: label:{ break; } }',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(x) { case 9: if(y){ break; } }',
      output: 'switch(x) { case 9: if(y) { break; } }',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(x){ case 9: if(y) { break; } }',
      output: 'switch(x){ case 9: if(y){ break; } }',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(x) { case 9: y;{ break; } }',
      output: 'switch(x) { case 9: y; { break; } }',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(x){ case 9: y; { break; } }',
      output: 'switch(x){ case 9: y;{ break; } }',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },
    {
      code: 'switch(x) { case 9: switch(y){} }',
      output: 'switch(x) { case 9: switch(y) {} }',
      options: ['always'],
      errors: [{ messageId: 'missingSpace' }],
    },
    {
      code: 'switch(x){ case 9: switch(y) {} }',
      output: 'switch(x){ case 9: switch(y){} }',
      options: ['never'],
      errors: [{ messageId: 'unexpectedSpace' }],
    },

    // ---- from space-before-blocks._ts_.test.ts ----
    {
      code: 'enum Test{\n  A = 2,\n  B = 1,\n}',
      output: 'enum Test {\n  A = 2,\n  B = 1,\n}',
      errors: [
        {
          messageId: 'missingSpace',
        },
      ],
      options: ['always'],
    },
    {
      code: 'interface Test{\n  prop1: number;\n}',
      output: 'interface Test {\n  prop1: number;\n}',
      errors: [
        {
          messageId: 'missingSpace',
        },
      ],
      options: ['always'],
    },
    {
      code: 'enum Test{\n  A = 2,\n  B = 1,\n}',
      output: 'enum Test {\n  A = 2,\n  B = 1,\n}',
      errors: [
        {
          messageId: 'missingSpace',
        },
      ],
      options: [{ classes: 'always' }],
    },
    {
      code: 'interface Test{\n  prop1: number;\n}',
      output: 'interface Test {\n  prop1: number;\n}',
      errors: [
        {
          messageId: 'missingSpace',
        },
      ],
      options: [{ classes: 'always' }],
    },
    {
      code: 'enum Test {\n  A = 2,\n  B = 1,\n}',
      output: 'enum Test{\n  A = 2,\n  B = 1,\n}',
      errors: [
        {
          messageId: 'unexpectedSpace',
        },
      ],
      options: ['never'],
    },
    {
      code: 'interface Test {\n  prop1: number;\n}',
      output: 'interface Test{\n  prop1: number;\n}',
      errors: [
        {
          messageId: 'unexpectedSpace',
        },
      ],
      options: ['never'],
    },
    {
      code: 'enum Test {\n  A = 2,\n  B = 1,\n}',
      output: 'enum Test{\n  A = 2,\n  B = 1,\n}',
      errors: [
        {
          messageId: 'unexpectedSpace',
        },
      ],
      options: [{ classes: 'never' }],
    },
    {
      code: 'interface Test {\n  prop1: number;\n}',
      output: 'interface Test{\n  prop1: number;\n}',
      errors: [
        {
          messageId: 'unexpectedSpace',
        },
      ],
      options: [{ classes: 'never' }],
    },
    {
      code: 'namespace Test {\n  type foo = number;\n}\ndeclare module \'foo\' {\n  type foo = number;\n}',
      output: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      errors: [
        {
          messageId: 'unexpectedSpace',
        },
        {
          messageId: 'unexpectedSpace',
        },
      ],
      options: ['never'],
    },
    {
      code: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      output: 'namespace Test {\n  type foo = number;\n}\ndeclare module \'foo\' {\n  type foo = number;\n}',
      errors: [
        {
          messageId: 'missingSpace',
        },
        {
          messageId: 'missingSpace',
        },
      ],
      options: ['always'],
    },
    {
      code: 'namespace Test {\n  type foo = number;\n}\ndeclare module \'foo\' {\n  type foo = number;\n}',
      output: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      errors: [
        {
          messageId: 'unexpectedSpace',
        },
        {
          messageId: 'unexpectedSpace',
        },
      ],
      options: [{ modules: 'never' }],
    },
    {
      code: 'namespace Test{\n  type foo = number;\n}\ndeclare module \'foo\'{\n  type foo = number;\n}',
      output: 'namespace Test {\n  type foo = number;\n}\ndeclare module \'foo\' {\n  type foo = number;\n}',
      errors: [
        {
          messageId: 'missingSpace',
        },
        {
          messageId: 'missingSpace',
        },
      ],
      options: [{ modules: 'always' }],
    },
  ],
});
