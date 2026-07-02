/**
 * Conformance: @stylistic/eslint-plugin (misc) mounted in rslint via `plugins`
 * must report identically to ESLint v10. Representative triggers from the
 * upstream suite; each verified to reproduce ESLint v10 byte-for-byte.
 */
import { runConformanceSuite } from '../conformance.js';
import type { DiffCase } from '../harness.js';

const CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'if (\na) {}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'const a = [\n1, 2, 3]',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'type Foo = [ 1, 2, 3 ]',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'const {a,b} = c',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: '1 + 1; // invalid comment',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: '// invalid comment\n1 + 1;',
    options: ['beside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: '// jscs: disable\n1 + 1;',
    options: [{ position: 'beside', applyDefaultIgnorePatterns: false }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: '1 + 1; // linter\n2 + 2; // invalid comment',
    options: [{ position: 'above', ignorePattern: 'linter' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: '\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\tvar i = 1;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: 'var x = 5, y = 2, z = 5;',
    options: [10, 4],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: '\t\t\tvar i = 1;',
    options: [15, 4],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: 'var /*this is a long non-removed inline comment*/ i = 1;',
    options: [20, 4, { ignoreComments: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'var foo; var bar;',
    options: [{ max: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'var bar = 1; var baz = 2;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'var bar = 1; var baz = 2; var qux = 3;',
    options: [{ max: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'if (condition) var bar = 1; else var baz = 1;',
    options: [{ max: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '\n                // these are\n                // line comments\n            ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '\n                //foo\n                ///bar\n            ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '\n                /* this block\n                 * is missing a newline at the start\n                 */\n            ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '\n                /*\n                 * this block\n                 * is missing a newline at the end*/\n            ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: '// invalid 1\nvar a = 5;\n\n\nvar b = 3;',
    options: [{ max: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: '// invalid 3\nvar a = 5;\n\n\n\n',
    options: [{ max: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: '// invalid 10\nvar a = 5;\n\nvar b = 3;',
    options: [{ max: 0 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: '// invalid 11\nvar a = 5;\n\n\n',
    options: [{ max: 5, maxEOF: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'function foo() {}\nfunction bar() {}',
    options: [
      {
        blankLine: 'always',
        prev: '*',
        next: { selector: 'FunctionDeclaration[id.name="bar"]' },
      },
    ],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'foo()\n\nbar()',
    options: [
      {
        blankLine: 'never',
        prev: { selector: 'ExpressionStatement[expression.callee.name="foo"]' },
        next: { selector: 'ExpressionStatement[expression.callee.name="bar"]' },
      },
    ],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'foo();\n\nfoo();',
    options: [{ blankLine: 'never', prev: '*', next: '*' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'foo();\n\n//comment\nfoo();',
    options: [{ blankLine: 'never', prev: '*', next: '*' }],
  },
];

const CLEAN_CASES: DiffCase[] = [
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'const a = [1, 2, 3]',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'exp-list-style',
    code: 'const a = {\nfoo: "bar",\nbar: 2\n}',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: '// valid comment\n1 + 1;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'line-comment-position',
    code: '1 + 1; // valid comment',
    options: ['beside'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: 'var x = 5;\nvar x = 2;',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-len',
    code: "foo('http://example.com/this/is/?a=longish&url=in#here');",
    options: [40, 4, { ignoreUrls: true }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'var bar = 1;',
    options: [{ max: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'max-statements-per-line',
    code: 'var bar = 1; var baz = 2;',
    options: [{ max: 2 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '\n            /*\n             * this is\n             * a comment\n             */\n        ',
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'multiline-comment-style',
    code: '\n                // this is a single-line comment\n            ',
    options: ['starred-block'],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: '// valid 1\nvar a = 5;\nvar b = 3;\n\n',
    options: [{ max: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'no-multiple-empty-lines',
    code: '// valid 9\nvar a = 1;\n\n',
    options: [{ max: 2, maxEOF: 1 }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'foo();bar();',
    options: [{ blankLine: 'never', prev: '*', next: '*' }],
  },
  {
    pkg: '@stylistic/eslint-plugin',
    rule: 'padding-line-between-statements',
    code: 'foo();\n\nbar();',
    options: [{ blankLine: 'always', prev: '*', next: '*' }],
  },
];

runConformanceSuite('@stylistic/eslint-plugin', CASES, CLEAN_CASES);
