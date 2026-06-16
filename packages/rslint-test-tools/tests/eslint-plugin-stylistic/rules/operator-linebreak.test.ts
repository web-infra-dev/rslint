/**
 * @fileoverview Operator linebreak rule tests
 * @author Benoît Zugmeyer
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/operator-linebreak/operator-linebreak.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, ... })` -> `ruleTester.run('operator-linebreak', null as never, { valid, invalid })`
 *  - `name` / `rule` / `import` lines dropped.
 *  - The `$` unindent template tag is evaluated to its real multi-line string
 *    (63 TS-syntax fixtures — 21 valid, 42 invalid: import-equals / type-alias /
 *    conditional / union / intersection / type-parameter / enum-member). Plain
 *    single-line `\n` string literals are preserved verbatim.
 *  - `parserOptions` (ecmaVersion 2020/2021/2022) dropped — rslint resolves via
 *    tsconfig and ts-go parses logical-assignment / class-field / accessor syntax
 *    without an ecmaVersion knob.
 *  - The `data: { operator }` on each error is kept; the RuleTester renders the
 *    plugin's own `meta.messages` template (operatorAtBeginning / operatorAtEnd /
 *    badLinebreak / noLinebreak) with that data and asserts the full message text,
 *    plus line/column/endLine/endColumn where upstream pins them.
 *
 * The upstream file is a single `run()` block (no skipBabel second block), has NO
 * Babel/Flow cases, NO spread/helper error builders, NO `readFileSync` external
 * fixtures, and NO `suggestions`. The `._css_` / `._json_` / `._markdown_` test
 * files don't exist for this rule.
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('operator-linebreak', null as never, {
  valid: [
    '1 + 1',
    '1 + 1 + 1',
    '1 +\n1',
    '1 + (1 +\n1)',
    'f(1 +\n1)',
    '1 || 1',
    '1 || \n1',
    'a += 1',
    'var a;',
    'var o = \nsomething',
    'o = \nsomething',
    '\'a\\\n\' +\n \'c\'',
    '\'a\' +\n \'b\\\n\'',
    '(a\n) + b',
    'answer = everything \n?  42 \n:  foo;',
    { code: 'answer = everything ?\n  42 :\n  foo;', options: ['after'] },

    { code: 'a ? 1 + 1\n:2', options: [null, { overrides: { '?': 'after' } }] },
    { code: 'a ?\n1 +\n 1\n:2', options: [null, { overrides: { '?': 'after' } }] },
    { code: 'o = 1 \n+ 1 - foo', options: [null, { overrides: { '+': 'before' } }] },

    { code: '1\n+ 1', options: ['before'] },
    { code: '1 + 1\n+ 1', options: ['before'] },
    { code: 'f(1\n+ 1)', options: ['before'] },
    { code: '1 \n|| 1', options: ['before'] },
    { code: 'a += 1', options: ['before'] },
    { code: 'answer = everything \n?  42 \n:  foo;', options: ['before'] },

    { code: '1 + 1', options: ['none'] },
    { code: '1 + 1 + 1', options: ['none'] },
    { code: '1 || 1', options: ['none'] },
    { code: 'a += 1', options: ['none'] },
    { code: 'var a;', options: ['none'] },
    { code: '\n1 + 1', options: ['none'] },
    { code: '1 + 1\n', options: ['none'] },
    { code: 'answer = everything ? 42 : foo;', options: ['none'] },
    { code: '(a\n) + (\nb)', options: ['none'] },
    { code: 'answer = everything \n?\n 42 : foo;', options: [null, { overrides: { '?': 'ignore' } }] },
    { code: 'answer = everything ? 42 \n:\n foo;', options: [null, { overrides: { ':': 'ignore' } }] },

    {
      code: 'a \n &&= b',
      options: ['after', { overrides: { '&&=': 'ignore' } }],
    },
    {
      code: 'a ??= \n b',
      options: ['before', { overrides: { '??=': 'ignore' } }],
    },
    {
      code: 'a ||= \n b',
      options: ['after', { overrides: { '=': 'before' } }],
    },
    {
      code: 'a \n &&= b',
      options: ['before', { overrides: { '&=': 'after' } }],
    },
    {
      code: 'a \n ||= b',
      options: ['before', { overrides: { '|=': 'after' } }],
    },
    {
      code: 'a &&= \n b',
      options: ['after', { overrides: { '&&': 'before' } }],
    },
    {
      code: 'a ||= \n b',
      options: ['after', { overrides: { '||': 'before' } }],
    },
    {
      code: 'a ??= \n b',
      options: ['after', { overrides: { '??': 'before' } }],
    },

    // class fields
    {
      code: 'class C { foo =\n0 }',
    },
    {
      code: 'class C { foo\n= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { [foo\n]= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { [foo]\n= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { [foo\n]\n= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { [foo\n]= 0 }',
      options: ['after'],
    },
    {
      code: 'class C { [foo\n]=\n0 }',
      options: ['after'],
    },
    {
      code: 'class C { [foo\n]= 0 }',
      options: ['none'],
    },
    {
      code: 'class C { foo\n=\n0 }',
      options: ['none', { overrides: { '=': 'ignore' } }],
    },
    {
      code: 'class C { accessor foo =\n0 }',
    },
    {
      code: 'class C { accessor foo\n= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { accessor [foo\n]= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { accessor [foo]\n= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { accessor [foo\n]\n= 0 }',
      options: ['before'],
    },
    {
      code: 'class C { accessor [foo\n]= 0 }',
      options: ['after'],
    },
    {
      code: 'class C { accessor [foo\n]=\n0 }',
      options: ['after'],
    },
    {
      code: 'class C { accessor [foo\n]= 0 }',
      options: ['none'],
    },
    {
      code: 'class C { accessor foo\n=\n0 }',
      options: ['none', { overrides: { '=': 'ignore' } }],
    },
    // TSImportEqualsDeclaration
    {
      code: "import F1\n  = A;\nimport F2\n  = A.B.C;\nimport F3\n  = require('mod');\nimport F1 = A;\nimport F2 = A.B.C;\nimport F3 = require('mod');",
      options: ['before'],
    },
    {
      code: "import F1 =\n  A;\nimport F2 =\n  A.B.C;\nimport F3 =\n  require('mod');\nimport F1 = A;\nimport F2 = A.B.C;\nimport F3 = require('mod');",
      options: ['after'],
    },
    {
      code: "import F1 = A;\nimport F2 = A.B.C;\nimport F3 = require('mod');",
      options: ['none'],
    },
    // TSTypeAliasDeclaration
    {
      code: "type A\n  = string;\ntype A = string;",
      options: ['before'],
    },
    {
      code: "type A =\n  string;\ntype A = string;",
      options: ['after'],
    },
    {
      code: "type A = string;",
      options: ['none'],
    },
    // TSConditionalType
    {
      code: "type A = Foo extends Bar\n  ? true\n  : false;\ntype A = Foo extends Bar ? true : false;",
      options: ['before'],
    },
    {
      code: "type A = Foo extends Bar ?\n  true :\n  false;\ntype A = Foo extends Bar ? true : false;",
      options: ['after'],
    },
    {
      code: "type A = Foo extends Bar ? true : false;",
      options: ['none'],
    },
    // TSIntersectionType
    {
      code: "type A = Foo\n  & Bar\n  & {};\ntype A = Foo & {};",
      options: ['before'],
    },
    {
      code: "type A = Foo &\n  Bar &\n  {};\ntype A = Foo & {};",
      options: ['after'],
    },
    {
      code: "type A = Foo & {};",
      options: ['none'],
    },
    // TSUnionType
    {
      code: "type A = Foo\n  | Bar\n  | {};\ntype A = Foo | {};",
      options: ['before'],
    },
    {
      code: "type A = Foo |\n  Bar |\n  {};\ntype A = Foo | {};",
      options: ['after'],
    },
    {
      code: "type A = Foo | {};",
      options: ['none'],
    },
    // TSTypeParameter
    {
      code: "type Foo<T\n  = number> = {\n  a: T;\n};\ntype Foo<T = number> = {\n  a: T;\n};",
      options: ['before'],
    },
    {
      code: "type Foo<T =\n  number> = {\n  a: T;\n};\ntype Foo<T = number> = {\n  a: T;\n};",
      options: ['after'],
    },
    {
      code: "type Foo<T = number> = {\n  a: T;\n};",
      options: ['none'],
    },
    // TSEnumMember
    {
      code: "enum Foo {\n  A,\n  B = 2,\n  C\n    = 4,\n}",
      options: ['before'],
    },
    {
      code: "enum Foo {\n  A,\n  B = 2,\n  C =\n    4,\n}",
      options: ['after'],
    },
    {
      code: "enum Foo {\n  A,\n  B = 2,\n}",
      options: ['none'],
    },
  ],

  invalid: [
    {
      code: '1\n+ 1',
      output: '1 +\n1',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: '1 + 2 \n + 3',
      output: '1 + 2 + \n 3',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '+' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: '1\n+\n1',
      output: '1+\n1',
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: '1 + (1\n+ 1)',
      output: '1 + (1 +\n1)',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'f(1\n+ 1);',
      output: 'f(1 +\n1);',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: '1 \n || 1',
      output: '1 || \n 1',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '||' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'a\n += 1',
      output: 'a +=\n 1',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '+=' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'var a\n = 1',
      output: 'var a =\n 1',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '=' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: '(b)\n*\n(c)',
      output: '(b)*\n(c)',
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: '*' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'answer = everything ?\n  42 :\n  foo;',
      output: 'answer = everything\n  ? 42\n  : foo;',
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '?' },
        line: 1,
        column: 21,
        endLine: 1,
        endColumn: 22,
      }, {
        messageId: 'operatorAtBeginning',
        data: { operator: ':' },
        line: 2,
        column: 6,
        endLine: 2,
        endColumn: 7,
      }],
    },

    {
      code: 'answer = everything \n?  42 \n:  foo;',
      output: 'answer = everything  ? \n42  : \nfoo;',
      options: ['after'],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '?' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'operatorAtEnd',
        data: { operator: ':' },
        line: 3,
        column: 1,
        endLine: 3,
        endColumn: 2,
      }],
    },

    {
      code: '1 +\n1',
      output: '1\n+ 1',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '+' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 4,
      }],
    },
    {
      code: 'f(1 +\n1);',
      output: 'f(1\n+ 1);',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '+' },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: '1 || \n 1',
      output: '1 \n || 1',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '||' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'a += \n1',
      output: 'a \n+= 1',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '+=' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'var a = \n1',
      output: 'var a \n= 1',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '=' },
        line: 1,
        column: 7,
        endLine: 1,
        endColumn: 8,
      }],
    },
    {
      code: 'answer = everything ?\n  42 :\n  foo;',
      output: 'answer = everything\n  ? 42\n  : foo;',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '?' },
        line: 1,
        column: 21,
        endLine: 1,
        endColumn: 22,
      }, {
        messageId: 'operatorAtBeginning',
        data: { operator: ':' },
        line: 2,
        column: 6,
        endLine: 2,
        endColumn: 7,
      }],
    },

    {
      code: '1 +\n1',
      output: '1 +1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 4,
      }],
    },
    {
      code: '1\n+1',
      output: '1+1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'f(1 +\n1);',
      output: 'f(1 +1);',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+' },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'f(1\n+ 1);',
      output: 'f(1+ 1);',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: '1 || \n 1',
      output: '1 ||  1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '||' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: '1 \n || 1',
      output: '1  || 1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '||' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 4,
      }],
    },
    {
      code: 'a += \n1',
      output: 'a += 1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+=' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'a \n+= 1',
      output: 'a += 1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+=' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'var a = \n1',
      output: 'var a = 1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '=' },
        line: 1,
        column: 7,
        endLine: 1,
        endColumn: 8,
      }],
    },
    {
      code: 'var a \n = 1',
      output: 'var a  = 1',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '=' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'answer = everything ?\n  42 \n:  foo;',
      output: 'answer = everything ?  42 :  foo;',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '?' },
        line: 1,
        column: 21,
        endLine: 1,
        endColumn: 22,
      }, {
        messageId: 'noLinebreak',
        data: { operator: ':' },
        line: 3,
        column: 1,
        endLine: 3,
        endColumn: 2,
      }],
    },
    {
      code: 'answer = everything\n?\n42 + 43\n:\nfoo;',
      output: 'answer = everything?42 + 43:foo;',
      options: ['none'],
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: '?' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }, {
        messageId: 'badLinebreak',
        data: { operator: ':' },
        line: 4,
        column: 1,
        endLine: 4,
        endColumn: 2,
      }],
    },
    {
      code: 'a = b \n  >>> \n c;',
      output: 'a = b   >>> \n c;',
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: '>>>' },
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 6,
      }],
    },
    {
      code: 'foo +=\n42;\nbar -=\n12\n+ 5;',
      output: 'foo +=42;\nbar -=\n12\n+ 5;',
      options: ['after', { overrides: { '+=': 'none', '+': 'before' } }],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '+=' },
        line: 1,
        column: 5,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: 'answer = everything\n?\n42\n:\nfoo;',
      output: 'answer = everything\n?\n42\n:foo;',
      options: ['after', { overrides: { '?': 'ignore', ':': 'before' } }],
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: ':' },
        line: 4,
        column: 1,
        endLine: 4,
        endColumn: 2,
      }],
    },
    {

      // Insert an additional space to avoid changing the operator to ++ or --.
      code: 'foo+\n+bar',
      output: 'foo\n+ +bar',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '+' },
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 5,
      }],
    },
    {
      code: 'foo //comment\n&& bar',
      output: 'foo && //comment\nbar',
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '&&' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'foo//comment\n+\nbar',
      output: null,
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'foo\n+//comment\nbar',
      output: null,
      options: ['before'],
      errors: [{
        messageId: 'badLinebreak',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'foo /* a */ \n+ /* b */ bar',
      output: null, // Not fixed because there is a comment on both sides
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '+' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'foo /* a */ +\n /* b */ bar',
      output: null, // Not fixed because there is a comment on both sides
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '+' },
        line: 1,
        column: 13,
        endLine: 1,
        endColumn: 14,
      }],
    },
    {
      code: 'foo ??\n bar',
      output: 'foo\n ?? bar',
      options: ['after', { overrides: { '??': 'before' } }],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '??' },
      }],
    },

    {
      code: 'a \n  &&= b',
      output: 'a &&= \n  b',
      options: ['after'],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '&&=' },
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 6,
      }],
    },
    {
      code: 'a ||=\n b',
      output: 'a\n ||= b',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '||=' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'a  ??=\n b',
      output: 'a  ??= b',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '??=' },
        line: 1,
        column: 4,
        endLine: 1,
        endColumn: 7,
      }],
    },
    {
      code: 'a \n  &&= b',
      output: 'a   &&= b',
      options: ['before', { overrides: { '&&=': 'none' } }],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '&&=' },
        line: 2,
        column: 3,
        endLine: 2,
        endColumn: 6,
      }],
    },
    {
      code: 'a ||=\nb',
      output: 'a\n||= b',
      options: ['after', { overrides: { '||=': 'before' } }],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '||=' },
        line: 1,
        column: 3,
        endLine: 1,
        endColumn: 6,
      }],
    },
    {
      code: 'a\n??=b',
      output: 'a??=\nb',
      options: ['none', { overrides: { '??=': 'after' } }],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '??=' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 4,
      }],
    },

    // class fields
    {
      code: 'class C { a\n= 0; }',
      output: 'class C { a =\n0; }',
      options: ['after'],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '=' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'class C { a =\n0; }',
      output: 'class C { a\n= 0; }',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '=' },
        line: 1,
        column: 13,
        endLine: 1,
        endColumn: 14,
      }],
    },
    {
      code: 'class C { a =\n0; }',
      output: 'class C { a =0; }',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '=' },
        line: 1,
        column: 13,
        endLine: 1,
        endColumn: 14,
      }],
    },
    {
      code: 'class C { [a]\n= 0; }',
      output: 'class C { [a] =\n0; }',
      options: ['after'],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '=' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'class C { [a] =\n0; }',
      output: 'class C { [a]\n= 0; }',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '=' },
        line: 1,
        column: 15,
        endLine: 1,
        endColumn: 16,
      }],
    },
    {
      code: 'class C { [a]\n =0; }',
      output: 'class C { [a] =0; }',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '=' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 3,
      }],
    },
    {
      code: 'class C { accessor a\n= 0; }',
      output: 'class C { accessor a =\n0; }',
      options: ['after'],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '=' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'class C { accessor a =\n0; }',
      output: 'class C { accessor a\n= 0; }',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '=' },
        line: 1,
        column: 22,
        endLine: 1,
        endColumn: 23,
      }],
    },
    {
      code: 'class C { accessor a =\n0; }',
      output: 'class C { accessor a =0; }',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '=' },
        line: 1,
        column: 22,
        endLine: 1,
        endColumn: 23,
      }],
    },
    {
      code: 'class C { accessor [a]\n= 0; }',
      output: 'class C { accessor [a] =\n0; }',
      options: ['after'],
      errors: [{
        messageId: 'operatorAtEnd',
        data: { operator: '=' },
        line: 2,
        column: 1,
        endLine: 2,
        endColumn: 2,
      }],
    },
    {
      code: 'class C { accessor [a] =\n0; }',
      output: 'class C { accessor [a]\n= 0; }',
      options: ['before'],
      errors: [{
        messageId: 'operatorAtBeginning',
        data: { operator: '=' },
        line: 1,
        column: 24,
        endLine: 1,
        endColumn: 25,
      }],
    },
    {
      code: 'class C { accessor [a]\n =0; }',
      output: 'class C { accessor [a] =0; }',
      options: ['none'],
      errors: [{
        messageId: 'noLinebreak',
        data: { operator: '=' },
        line: 2,
        column: 2,
        endLine: 2,
        endColumn: 3,
      }],
    },
    // TSImportEqualsDeclaration
    {
      code: "import F1 =\n  A;\nimport F2 =\n  A.B.C;\nimport F3 =\n  require('mod');",
      output: "import F1\n  = A;\nimport F2\n  = A.B.C;\nimport F3\n  = require('mod');",
      options: ['before'],
      errors: [
        { messageId: 'operatorAtBeginning' },
        { messageId: 'operatorAtBeginning' },
        { messageId: 'operatorAtBeginning' },
      ],
    },
    {
      code: "import F1\n  = A;\nimport F2\n  = A.B.C;\nimport F3\n  = require('mod');",
      output: "import F1 =\n  A;\nimport F2 =\n  A.B.C;\nimport F3 =\n  require('mod');",
      options: ['after'],
      errors: [
        { messageId: 'operatorAtEnd' },
        { messageId: 'operatorAtEnd' },
        { messageId: 'operatorAtEnd' },
      ],
    },
    {
      code: "import F1\n  = A;\nimport F2 =\n  A.B.C;\nimport F3\n  = require('mod');",
      output: "import F1  = A;\nimport F2 =  A.B.C;\nimport F3  = require('mod');",
      options: ['none'],
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
    },
    // TSTypeAliasDeclaration
    {
      code: "type A =\n  string;",
      output: "type A\n  = string;",
      options: ['before'],
      errors: [
        { messageId: 'operatorAtBeginning' },
      ],
    },
    {
      code: "type A\n  = string;",
      output: "type A =\n  string;",
      errors: [
        { messageId: 'operatorAtEnd' },
      ],
      options: ['after'],
    },
    {
      code: "type A\n  = string;\ntype A =\n  string;",
      output: "type A  = string;\ntype A =  string;",
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
      options: ['none'],
    },
    // TSConditionalType
    {
      code: "type A = Foo extends Bar ?\n  true :\n  false;",
      output: "type A = Foo extends Bar\n  ? true\n  : false;",
      errors: [
        { messageId: 'operatorAtBeginning' },
        { messageId: 'operatorAtBeginning' },
      ],
      options: ['before'],
    },
    {
      code: "type A = Foo extends Bar\n  ? true\n  : false;",
      output: "type A = Foo extends Bar ?\n  true :\n  false;",
      errors: [
        { messageId: 'operatorAtEnd' },
        { messageId: 'operatorAtEnd' },
      ],
      options: ['after'],
    },
    {
      code: "type A = Foo extends Bar ?\n  true :\n  false;",
      output: "type A = Foo extends Bar ?  true :  false;",
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
      options: ['none'],
    },
    // TSIntersectionType
    {
      code: "type A = Foo &\n  Bar &\n  {};",
      output: "type A = Foo\n  & Bar\n  & {};",
      options: ['before'],
      errors: [
        { messageId: 'operatorAtBeginning' },
        { messageId: 'operatorAtBeginning' },
      ],
    },
    {
      code: "type A = Foo\n  & Bar\n  & {};",
      output: "type A = Foo &\n  Bar &\n  {};",
      options: ['after'],
      errors: [
        { messageId: 'operatorAtEnd' },
        { messageId: 'operatorAtEnd' },
      ],
    },
    {
      code: "type A = Foo &\n  Bar\n  & {};",
      output: "type A = Foo &  Bar  & {};",
      options: ['none'],
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
    },
    // TSUnionType
    {
      code: "type A = Foo |\n  Bar |\n  {};",
      output: "type A = Foo\n  | Bar\n  | {};",
      options: ['before'],
      errors: [
        { messageId: 'operatorAtBeginning' },
        { messageId: 'operatorAtBeginning' },
      ],
    },
    {
      code: "type A = Foo\n  | Bar\n  | {};",
      output: "type A = Foo |\n  Bar |\n  {};",
      options: ['after'],
      errors: [
        { messageId: 'operatorAtEnd' },
        { messageId: 'operatorAtEnd' },
      ],
    },
    {
      code: "type A = Foo |\n  Bar\n  | {};",
      output: "type A = Foo |  Bar  | {};",
      options: ['none'],
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
    },
    // TSTypeParameter
    {
      code: "type Foo<T =\n  number> = {\n  a: T;\n};",
      output: "type Foo<T\n  = number> = {\n  a: T;\n};",
      options: ['before'],
      errors: [
        { messageId: 'operatorAtBeginning' },
      ],
    },
    {
      code: "type Foo<T\n  = number> = {\n  a: T;\n};",
      output: "type Foo<T =\n  number> = {\n  a: T;\n};",
      options: ['after'],
      errors: [
        { messageId: 'operatorAtEnd' },
      ],
    },
    {
      code: "type Foo<T\n  = number> = {\n  a: T;\n};\ntype Foo<T =\n  number> = {\n  a: T;\n};",
      output: "type Foo<T  = number> = {\n  a: T;\n};\ntype Foo<T =  number> = {\n  a: T;\n};",
      options: ['none'],
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
    },
    // TSEnumMember
    {
      code: "enum Foo {\n  A,\n  B = 2,\n  C =\n    4,\n}",
      output: "enum Foo {\n  A,\n  B = 2,\n  C\n    = 4,\n}",
      options: ['before'],
      errors: [
        { messageId: 'operatorAtBeginning' },
      ],
    },
    {
      code: "enum Foo {\n  A,\n  B = 2,\n  C\n    = 4,\n}",
      output: "enum Foo {\n  A,\n  B = 2,\n  C =\n    4,\n}",
      errors: [
        { messageId: 'operatorAtEnd' },
      ],
      options: ['after'],
    },
    {
      code: "enum Foo {\n  A,\n  B = 2,\n  C\n    = 4,\n  D =\n    6,\n}",
      output: "enum Foo {\n  A,\n  B = 2,\n  C    = 4,\n  D =    6,\n}",
      errors: [
        { messageId: 'noLinebreak' },
        { messageId: 'noLinebreak' },
      ],
      options: ['none'],
    },
  ],
});
