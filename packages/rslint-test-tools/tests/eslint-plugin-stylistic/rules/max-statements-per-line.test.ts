/**
 * @fileoverview Tests for max-statements-per-line rule.
 * @author Kenneth Williams
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/max-statements-per-line/max-statements-per-line.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('max-statements-per-line', null as never, { valid, invalid })`
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via
 *    tsconfig and always parses at esnext; the upstream `sourceType: 'module'`
 *    cases are plain ES modules that ts-go parses identically.
 *  - The `code: [...].join('\n')` array form is kept verbatim (a string literal at
 *    module load); the one bare-array valid case is likewise its joined string.
 *  - The rule's single messageId `exceed` interpolates
 *    `{{numberOfStatementsOnThisLine}} {{statements}} ... {{maxStatementsPerLine}}`;
 *    cases that pin `data` assert the fully-rendered message, cases that pin a
 *    bare `{ messageId: 'exceed' }` assert only the diagnostic count (the template
 *    can't be resolved without data, so no message is asserted for those).
 *
 * This rule is NOT fixable, so no upstream case pins `output`; every invalid case
 * pins `errors` only. There are no Babel/Flow, no JSX, and no external-fixture
 * (`readFileSync`) cases. The `._css_` / `._json_` / `._markdown_` test files
 * don't exist for this rule.
 *
 * There is no rslint<->upstream gap for this rule: every upstream valid/invalid
 * case runs verbatim through the green `ruleTester.run` below and matches.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-statements-per-line', null as never, {
  valid: [
    { code: '{ }', options: [{ max: 1 }] },
    'var bar = 1;',
    { code: 'var bar = 1;', options: [{ max: 1 }] },
    'var bar = 1;;',
    ';(function foo() {\n})()',
    { code: 'if (condition) var bar = 1;', options: [{ max: 1 }] },
    { code: 'if (condition) { }', options: [{ max: 1 }] },
    { code: 'if (condition) { } else { }', options: [{ max: 1 }] },
    { code: 'if (condition) {\nvar bar = 1;\n} else {\nvar bar = 1;\n}', options: [{ max: 1 }] },
    { code: 'for (var i = 0; i < length; ++i) { }', options: [{ max: 1 }] },
    { code: 'for (var i = 0; i < length; ++i) {\nvar bar  = 1;\n}', options: [{ max: 1 }] },
    { code: 'switch (discriminant) { default: }', options: [{ max: 1 }] },
    { code: 'switch (discriminant) {\ndefault: break;\n}', options: [{ max: 1 }] },
    { code: 'function foo() { }', options: [{ max: 1 }] },
    { code: 'function foo() {\nif (condition) var bar = 1;\n}', options: [{ max: 1 }] },
    { code: 'function foo() {\nif (condition) {\nvar bar = 1;\n}\n}', options: [{ max: 1 }] },
    { code: '(function() { })();', options: [{ max: 1 }] },
    { code: '(function() {\nvar bar = 1;\n})();', options: [{ max: 1 }] },
    { code: 'var foo = function foo() { };', options: [{ max: 1 }] },
    { code: 'var foo = function foo() {\nvar bar = 1;\n};', options: [{ max: 1 }] },
    { code: 'var foo = { prop: () => { } };', options: [{ max: 1 }] },
    { code: 'var bar = 1; var baz = 2;', options: [{ max: 2 }] },
    { code: 'if (condition) { var bar = 1; }', options: [{ max: 2 }] },
    { code: 'if (condition) {\nvar bar = 1; var baz = 2;\n} else {\nvar bar = 1; var baz = 2;\n}', options: [{ max: 2 }] },
    { code: 'for (var i = 0; i < length; ++i) { var bar = 1; }', options: [{ max: 2 }] },
    { code: 'for (var i = 0; i < length; ++i) {\nvar bar = 1; var baz = 2;\n}', options: [{ max: 2 }] },
    { code: 'switch (discriminant) { default: break; }', options: [{ max: 2 }] },
    { code: 'switch (discriminant) {\ncase \'test\': var bar = 1; break;\ndefault: var bar = 1; break;\n}', options: [{ max: 2 }] },
    { code: 'function foo() { var bar = 1; }', options: [{ max: 2 }] },
    { code: 'function foo() {\nvar bar = 1; var baz = 2;\n}', options: [{ max: 2 }] },
    { code: 'function foo() {\nif (condition) { var bar = 1; }\n}', options: [{ max: 2 }] },
    { code: 'function foo() {\nif (condition) {\nvar bar = 1; var baz = 2;\n}\n}', options: [{ max: 2 }] },
    { code: '(function() { var bar = 1; })();', options: [{ max: 2 }] },
    { code: '(function() {\nvar bar = 1; var baz = 2;\n})();', options: [{ max: 2 }] },
    { code: 'var foo = function foo() { var bar = 1; };', options: [{ max: 2 }] },
    { code: 'var foo = function foo() {\nvar bar = 1; var baz = 2;\n};', options: [{ max: 2 }] },
    { code: 'var foo = { prop: () => { var bar = 1; } };', options: [{ max: 2 }] },
    { code: 'var bar = 1; var baz = 2; var qux = 3;', options: [{ max: 3 }] },
    { code: 'if (condition) { var bar = 1; var baz = 2; }', options: [{ max: 3 }] },
    { code: 'if (condition) { var bar = 1; } else { var bar = 1; }', options: [{ max: 3 }] },
    { code: 'switch (discriminant) { case \'test1\': ; case \'test2\': ; }', options: [{ max: 3 }] },
    { code: 'let bar = bar => { a; }, baz = baz => { b; };', options: [{ max: 3 }] },
    { code: 'function foo({[bar => { a; }]: baz = qux => { b; }}) { }', options: [{ max: 3 }] },
    { code: 'bar => { a; }, baz => { b; }, qux => { c; };', options: [{ max: 4 }] },
    { code: '[bar => { a; }, baz => { b; }, qux => { c; }];', options: [{ max: 4 }] },
    { code: 'foo(bar => { a; }, baz => { c; }, qux => { c; });', options: [{ max: 4 }] },
    { code: '({ bar: bar => { a; }, baz: baz => { c; }, qux: qux => { ; }});', options: [{ max: 4 }] },
    { code: '(bar => { a; }) ? (baz => { b; }) : (qux => { c; });', options: [{ max: 4 }] },
    {
      code: [
        'const name = \'ESLint\'',
        '',
        ';(function foo() {',
        '})()',
      ].join('\n'),
      options: [{ max: 1 }],
    },
    [
      'if (foo > 1)',
      '    foo--;',
      'else',
      '    foo++;',
    ].join('\n'),
    {
      code: 'export default foo = 0;',
      options: [{ max: 1 }],
    },
    {
      code: [
        'export default function foo() {',
        '   console.log(\'test\');',
        '}',
      ].join('\n'),
      options: [{ max: 1 }],
    },
    {
      code: 'export let foo = 0;',
      options: [{ max: 1 }],
    },
    {
      code: [
        'export function foo() {',
        '   console.log(\'test\');',
        '}',
      ].join('\n'),
      options: [{ max: 1 }],
    },
    {
      code: 'const a = 1; export const b = 2; export const c = 3;',
      options: [{
        max: 1,
        ignoredNodes: ['ExportNamedDeclaration'],
      }],
    },
    {
      code: [
        'switch (lorem) {',
        '  case ipsum: dolor(); break;',
        '}',
      ].join('\n'),
      options: [{
        max: 1,
        ignoredNodes: ['BreakStatement'],
      }],
    },
  ],
  invalid: [
    { code: 'var foo; var bar;', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'var bar = 1; var foo = 3;', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'var bar = 1; var baz = 2;', errors: [{ messageId: 'exceed' }] },
    { code: 'var bar = 1; var baz = 2;', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'if (condition) var bar = 1; if (condition) var baz = 2;', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'if (condition) var bar = 1; else var baz = 1;', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'if (condition) { } if (condition) { }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'if (condition) { var bar = 1; } else { }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'if (condition) { } else { var bar = 1; }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'if (condition) { var bar = 1; } else { var bar = 1; }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'for (var i = 0; i < length; ++i) { var bar = 1; }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'switch (discriminant) { default: break; }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'function foo() { var bar = 1; }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'function foo() { if (condition) var bar = 1; }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'function foo() { if (condition) { var bar = 1; } }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: '(function() { var bar = 1; })();', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'var foo = function foo() { var bar = 1; };', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'var foo = { prop: () => { var bar = 1; } };', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'var bar = 1; var baz = 2; var qux = 3;', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'if (condition) { var bar = 1; var baz = 2; }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'if (condition) { var bar = 1; } else { var bar = 1; }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'if (condition) { var bar = 1; var baz = 2; } else { var bar = 1; var baz = 2; }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'for (var i = 0; i < length; ++i) { var bar = 1; var baz = 2; }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'switch (discriminant) { case \'test\': break; default: break; }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'function foo() { var bar = 1; var baz = 2; }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'function foo() { if (condition) { var bar = 1; } }', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: '(function() { var bar = 1; var baz = 2; })();', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'var foo = function foo() { var bar = 1; var baz = 2; };', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'var foo = { prop: () => { var bar = 1; var baz = 2; } };', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 3, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'var bar = 1; var baz = 2; var qux = 3; var waldo = 4;', options: [{ max: 3 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 4, statements: 'statements', maxStatementsPerLine: 3.0 } }] },
    { code: 'if (condition) { var bar = 1; var baz = 2; var qux = 3; }', options: [{ max: 3 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 4, statements: 'statements', maxStatementsPerLine: 3.0 } }] },
    { code: 'if (condition) { var bar = 1; var baz = 2; } else { var bar = 1; var baz = 2; }', options: [{ max: 3 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 3.0 } }] },
    { code: 'switch (discriminant) { case \'test\': var bar = 1; break; default: var bar = 1; break; }', options: [{ max: 3 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 3.0 } }] },
    { code: 'let bar = bar => { a; }, baz = baz => { b; }, qux = qux => { c; };', options: [{ max: 3 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 4, statements: 'statements', maxStatementsPerLine: 3.0 } }] },
    { code: '(bar => { a; }) ? (baz => { b; }) : (qux => { c; });', options: [{ max: 3 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 4, statements: 'statements', maxStatementsPerLine: 3.0 } }] },
    { code: 'bar => { a; }, baz => { b; }, qux => { c; }, quux => { d; };', options: [{ max: 4 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 4.0 } }] },
    { code: '[bar => { a; }, baz => { b; }, qux => { c; }, quux => { d; }];', options: [{ max: 4 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 4.0 } }] },
    { code: 'foo(bar => { a; }, baz => { b; }, qux => { c; }, quux => { d; });', options: [{ max: 4 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 4.0 } }] },
    { code: '({ bar: bar => { a; }, baz: baz => { b; }, qux: qux => { c; }, quux: quux => { d; }});', options: [{ max: 4 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 5, statements: 'statements', maxStatementsPerLine: 4.0 } }] },
    { code: 'a; if (b) { c; d; }\nz;', options: [{ max: 2 }], errors: [{ messageId: 'exceed', data: { numberOfStatementsOnThisLine: 4, statements: 'statements', maxStatementsPerLine: 2.0 } }] },
    { code: 'export default function foo() { console.log(\'test\') }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    { code: 'export function foo() { console.log(\'test\') }', options: [{ max: 1 }], errors: [{ messageId: 'exceed' }] },
    {
      code: [
        'for (let i = 0; i < 3; i++){',
        '  if (a) foo(); else if (b) bar(); else break;',
        '}',
      ].join('\n'),
      options: [{
        max: 1,
        ignoredNodes: ['IfStatement'],
      }],
      errors: [{ messageId: 'exceed' }],
    },
  ],
});
