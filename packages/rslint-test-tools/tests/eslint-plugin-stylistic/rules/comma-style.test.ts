/**
 * @fileoverview Comma style
 * @author Vignesh Anand aka vegetableman
 *
 * Ported verbatim from @stylistic/eslint-plugin v5.10.0:
 *   packages/eslint-plugin/rules/comma-style/comma-style.test.ts
 *
 * Transformations applied per the porting spec:
 *  - `run({ name, rule, valid, invalid })` -> `ruleTester.run('comma-style', null as never, { valid, invalid })`.
 *  - The `$` unindent template tag is evaluated to its real multi-line string
 *    (common leading indentation stripped, leading/trailing blank lines removed);
 *    a blank line whose only content was the common indent becomes a truly empty
 *    line (e.g. `...]: string,\n\n  f(...`). Single-line cases are kept verbatim.
 *  - `parserOptions` (ecmaVersion / sourceType) dropped — rslint resolves via
 *    tsconfig and is always `esnext`. comma-style's options (`'first'`/`'last'`
 *    plus the `exceptions` object) do not depend on ecmaVersion normalization,
 *    so dropping it changes no expectations.
 *  - comma-style has no `type` AST fields, no `filename`, no `settings`, and no
 *    `suggestions` in its upstream tests.
 *
 * The rule's `meta.messages` carry no `{{data}}` interpolation, so every error
 * pins only `messageId` (+ line/column/endLine/endColumn when upstream gives them):
 *   unexpectedLineBeforeAndAfterComma: 'Bad line breaking before and after \',\'.'
 *   expectedCommaFirst:                '\',\' should be placed first.'
 *   expectedCommaLast:                 '\',\' should be placed last.'
 *
 * Every upstream invalid case pins an explicit `errors` array (no output-only
 * cases). Import attributes use the `with { ... }` form, which ts-go accepts, so
 * no fixture is a parser-level syntax error.
 *
 * No rslint<->upstream gap surfaced for this rule: every valid case reports zero
 * diagnostics and every invalid case matches upstream's diagnostic count, rendered
 * message, pinned positions, and single-pass autofix output. There is therefore no
 * `KNOWN GAPS` block below. (Had any case diverged — a parser-level syntax error, a
 * multi-pass fix difference, etc. — it would be moved here verbatim, annotated with
 * upstream-expected vs. rslint-produced, never deleted or altered to force green.)
 */

import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('comma-style', null as never, {
  valid: [
    'var foo = 1, bar = 3;',
    'var foo = {\'a\': 1, \'b\': 2};',
    'var foo = [1, 2];',
    'var foo = [, 2];',
    'var foo = [1, ];',
    'var foo = [\'apples\', \n \'oranges\'];',
    'var foo = {\'a\': 1, \n \'b\': 2, \n\'c\': 3};',
    'var foo = {\'a\': 1, \n \'b\': 2, \'c\':\n 3};',
    'var foo = {\'a\': 1, \n \'b\': 2, \'c\': [{\'d\': 1}, \n {\'e\': 2}, \n {\'f\': 3}]};',
    'var foo = [1, \n2, \n3];',
    'function foo(){var a=[1,\n 2]}',
    'function foo(){return {\'a\': 1,\n\'b\': 2}}',
    'var foo = \n1, \nbar = \n2;',
    'var foo = [\n(bar),\nbaz\n];',
    'var foo = [\n(bar\n),\nbaz\n];',
    'var foo = [\n(\nbar\n),\nbaz\n];',
    {
      code: 'new Foo(a\n,b);',
      options: ['last', { exceptions: { NewExpression: true } }],
    },
    {
      code: 'var foo = [\n(bar\n)\n,baz\n];',
      options: ['first'],
    },
    'var foo = \n1, \nbar = [1,\n2,\n3]',
    {
      code: 'var foo = [\'apples\'\n,\'oranges\'];',
      options: ['first'],
    },
    {
      code: 'var foo = 1, bar = 2;',
      options: ['first'],
    },
    {
      code: 'var foo = 1 \n ,bar = 2;',
      options: ['first'],
    },
    {
      code: 'var foo = {\'a\': 1 \n ,\'b\': 2 \n,\'c\': 3};',
      options: ['first'],
    },
    {
      code: 'var foo = [1 \n ,2 \n, 3];',
      options: ['first'],
    },
    {
      code: 'function foo(){return {\'a\': 1\n,\'b\': 2}}',
      options: ['first'],
    },
    {
      code: 'function foo(){var a=[1\n, 2]}',
      options: ['first'],
    },
    {
      code: 'new Foo(a,\nb);',
      options: ['first', { exceptions: { NewExpression: true } }],
    },
    {
      code: 'f(1\n, 2);',
      options: ['last', { exceptions: { CallExpression: true } }],
    },
    {
      code: 'function foo(a\n, b) { return a + b; }',
      options: ['last', { exceptions: { FunctionDeclaration: true } }],
    },
    {
      code: 'var a = \'a\',\no = \'o\';',
      options: ['first', { exceptions: { VariableDeclaration: true } }],
    },
    {
      code: 'var arr = [\'a\',\n\'o\'];',
      options: ['first', { exceptions: { ArrayExpression: true } }],
    },
    {
      code: 'var obj = {a: \'a\',\nb: \'b\'};',
      options: ['first', { exceptions: { ObjectExpression: true } }],
    },
    {
      code: 'var a = \'a\',\no = \'o\',\narr = [1,\n2];',
      options: ['first', { exceptions: { VariableDeclaration: true, ArrayExpression: true } }],
    },
    {
      code: 'var ar ={fst:1,\nsnd: [1,\n2]};',
      options: ['first', { exceptions: { ArrayExpression: true, ObjectExpression: true } }],
    },
    {
      code: 'var a = \'a\',\nar ={fst:1,\nsnd: [1,\n2]};',
      options: ['first', { exceptions: { ArrayExpression: true, ObjectExpression: true, VariableDeclaration: true } }],
    },
    {
      code: 'const foo = (a\n, b) => { return a + b; }',
      options: ['last', { exceptions: { ArrowFunctionExpression: true } }],
    },
    {
      code: 'function foo([a\n, b]) { return a + b; }',
      options: ['last', { exceptions: { ArrayPattern: true } }],
    },
    {
      code: 'const foo = ([a\n, b]) => { return a + b; }',
      options: ['last', { exceptions: { ArrayPattern: true } }],
    },
    {
      code: 'import { a\n, b } from \'./source\';',
      options: ['last', { exceptions: { ImportDeclaration: true } }],
    },
    {
      code: 'const foo = function (a\n, b) { return a + b; }',
      options: ['last', { exceptions: { FunctionExpression: true } }],
    },
    {
      code: 'var {foo\n, bar} = {foo:\'apples\', bar:\'oranges\'};',
      options: ['last', { exceptions: { ObjectPattern: true } }],
    },
    {
      code: 'var {foo\n, bar} = {foo:\'apples\', bar:\'oranges\'};',
      options: ['first', { exceptions: { ObjectPattern: true } }],
    },
    {
      code: 'new Foo(a,\nb);',
      options: ['first', { exceptions: { NewExpression: true } }],
    },
    {
      code: 'f(1\n, 2);',
      options: ['last', { exceptions: { CallExpression: true } }],
    },
    {
      code: 'function foo(a\n, b) { return a + b; }',
      options: ['last', { exceptions: { FunctionDeclaration: true } }],
    },
    {
      code: 'const foo = function (a\n, b) { return a + b; }',
      options: ['last', { exceptions: { FunctionExpression: true } }],
    },
    {
      code: 'function foo([a\n, b]) { return a + b; }',
      options: ['last', { exceptions: { ArrayPattern: true } }],
    },
    {
      code: 'const foo = (a\n, b) => { return a + b; }',
      options: ['last', { exceptions: { ArrowFunctionExpression: true } }],
    },
    {
      code: 'const foo = ([a\n, b]) => { return a + b; }',
      options: ['last', { exceptions: { ArrayPattern: true } }],
    },
    {
      code: 'import { a\n, b } from \'./source\';',
      options: ['last', { exceptions: { ImportDeclaration: true } }],
    },
    {
      code: 'var {foo\n, bar} = {foo:\'apples\', bar:\'oranges\'};',
      options: ['last', { exceptions: { ObjectPattern: true } }],
    },
    {
      code: 'new Foo(a,\nb);',
      options: ['last', { exceptions: { NewExpression: false } }],
    },
    {
      code: 'new Foo(a\n,b);',
      options: ['last', { exceptions: { NewExpression: true } }],
    },
    'var foo = [\n , \n 1, \n 2 \n];',
    {
      code: 'const [\n , \n , \n a, \n b, \n] = arr;',
      options: ['last', { exceptions: { ArrayPattern: false } }],
    },
    {
      code: 'const [\n ,, \n a, \n b, \n] = arr;',
      options: ['last', { exceptions: { ArrayPattern: false } }],
    },
    {
      code: 'const arr = [\n 1 \n , \n ,2 \n]',
      options: ['first'],
    },
    {
      code: 'const arr = [\n ,\'fifi\' \n]',
      options: ['first'],
    },
    {
      code: 'import {\n  A,\n  B\n  , C\n} from \'module3\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};\nimport \'module4\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};',
      options: ['last', { exceptions: { ImportDeclaration: true } }],
    },
    {
      code: 'let a, b, c;\nexport {\n  a,\n  b\n  , c\n};\nexport {\n  A,\n  B\n  , C\n} from \'module1\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};\nexport * from \'module2\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};',
      options: ['last', { exceptions: { ExportAllDeclaration: true, ExportNamedDeclaration: true } }],
    },
    {
      code: 'import(\n  a,\n  b\n);\nimport(\n  c\n  , d\n);',
      options: ['first', { exceptions: { ImportExpression: true } }],
    },
    {
      code: 'import(\n  a,\n);\nimport(\n  b, c,\n);',
      options: ['last', { exceptions: { ImportExpression: false } }],
    },
    {
      code: 'import(\n  a\n,);\nimport(\n  b, c\n,);',
      options: ['first', { exceptions: { ImportExpression: false } }],
    },
    {
      code: 'const x = (\n  a,\n  b\n  , c\n);',
      options: ['first', { exceptions: { SequenceExpression: true } }],
    },
    {
      code: 'class MyClass implements\n  A,\n  B\n, C {\n}\nconst a = class implements\n  A,\n  B\n, C {\n}',
      options: ['first', { exceptions: { ClassDeclaration: true, ClassExpression: true } }],
    },
    {
      code: 'function f(\n  a,\n  b\n  , c\n)\ntype a = (\n  a,\n  b\n  , c\n) => r\ntype a = new (\n  a,\n  b\n  , c\n) => r\nabstract class Base {\n  f(\n    a,\n    b\n    , c\n  );\n}',
      options: ['first', { exceptions: { TSDeclareFunction: true, TSFunctionType: true, TSConstructorType: true, TSEmptyBodyFunctionExpression: true } }],
    },
    {
      code: 'enum MyEnum {\n  A,\n  B\n  , C\n}',
      options: ['first', { exceptions: { TSEnumBody: true } }],
    },
    {
      code: 'type foo = {\n  a: string,\n  b: string\n  , c: string\n}',
      options: ['first', { exceptions: { TSTypeLiteral: true } }],
    },
    {
      code: 'type foo = {\n  new (\n    a,\n    b\n    , c\n  ): any,\n  (\n    a,\n    b\n    , c\n  ): any,\n  [\n    a: string,\n    b: string\n    , c: string\n  ]: string,\n\n  f(\n    a: string,\n    b: string\n    , c: string\n  ): number,\n}',
      options: ['first', { exceptions: { TSTypeLiteral: true, TSCallSignatureDeclaration: true, TSConstructSignatureDeclaration: true, TSIndexSignature: true, TSMethodSignature: true } }],
    },
    {
      code: 'interface Foo extends\n  A,\n  B\n  , C\n{\n  a: string,\n  b: string\n  , c: string\n}',
      options: ['first', { exceptions: { TSInterfaceBody: true, TSInterfaceDeclaration: true } }],
    },
    {
      code: 'type Foo = [\n  "A",\n  "B"\n  , "C"\n];',
      options: ['first', { exceptions: { TSTupleType: true } }],
    },
    {
      code: 'type Foo<\n  A,\n  B\n  , C\n> = Bar<\n  A,\n  B\n  , C\n>;',
      options: ['first', { exceptions: { TSTypeParameterDeclaration: true, TSTypeParameterInstantiation: true } }],
    },
  ],

  invalid: [
    {
      code: 'var foo = { a: 1. //comment \n, b: 2\n}',
      output: 'var foo = { a: 1., //comment \n b: 2\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'var foo = { a: 1. //comment \n //comment1 \n //comment2 \n, b: 2\n}',
      output: 'var foo = { a: 1., //comment \n //comment1 \n //comment2 \n b: 2\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'var foo = 1\n,\nbar = 2;',
      output: 'var foo = 1,\nbar = 2;',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: 'var foo = 1 //comment\n,\nbar = 2;',
      output: 'var foo = 1, //comment\nbar = 2;',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: 'var foo = 1 //comment\n, // comment 2\nbar = 2;',
      output: 'var foo = 1, //comment // comment 2\nbar = 2;',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: 'new Foo(a\n,\nb);',
      output: 'new Foo(a,\nb);',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: 'var foo = 1\n,bar = 2;',
      output: 'var foo = 1,\nbar = 2;',
      errors: [
        {
          messageId: 'expectedCommaLast',
          column: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: 'f([1,2\n,3]);',
      output: 'f([1,2,\n3]);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'f([1,2\n,]);',
      output: 'f([1,2,\n]);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'f([,2\n,3]);',
      output: 'f([,2,\n3]);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'var foo = [\'apples\'\n, \'oranges\'];',
      output: 'var foo = [\'apples\',\n \'oranges\'];',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'var [foo\n, bar] = [\'apples\', \'oranges\'];',
      output: 'var [foo,\n bar] = [\'apples\', \'oranges\'];',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'f(1\n, 2);',
      output: 'f(1,\n 2);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'function foo(a\n, b) { return a + b; }',
      output: 'function foo(a,\n b) { return a + b; }',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'const foo = function (a\n, b) { return a + b; }',
      output: 'const foo = function (a,\n b) { return a + b; }',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'function foo([a\n, b]) { return a + b; }',
      output: 'function foo([a,\n b]) { return a + b; }',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'const foo = (a\n, b) => { return a + b; }',
      output: 'const foo = (a,\n b) => { return a + b; }',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'const foo = ([a\n, b]) => { return a + b; }',
      output: 'const foo = ([a,\n b]) => { return a + b; }',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'import { a\n, b } from \'./source\';',
      output: 'import { a,\n b } from \'./source\';',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'var {foo\n, bar} = {foo:\'apples\', bar:\'oranges\'};',
      output: 'var {foo,\n bar} = {foo:\'apples\', bar:\'oranges\'};',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'var foo = 1,\nbar = 2;',
      output: 'var foo = 1\n,bar = 2;',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
          column: 12,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'f([1,\n2,3]);',
      output: 'f([1\n,2,3]);',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var foo = [\'apples\', \n \'oranges\'];',
      output: 'var foo = [\'apples\' \n ,\'oranges\'];',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var foo = {\'a\': 1, \n \'b\': 2\n ,\'c\': 3};',
      output: 'var foo = {\'a\': 1 \n ,\'b\': 2\n ,\'c\': 3};',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var a = \'a\',\no = \'o\',\narr = [1,\n2];',
      output: 'var a = \'a\',\no = \'o\',\narr = [1\n,2];',
      options: ['first', { exceptions: { VariableDeclaration: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var a = \'a\',\nobj = {a: \'a\',\nb: \'b\'};',
      output: 'var a = \'a\',\nobj = {a: \'a\'\n,b: \'b\'};',
      options: ['first', { exceptions: { VariableDeclaration: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var a = \'a\',\nobj = {a: \'a\',\nb: \'b\'};',
      output: 'var a = \'a\'\n,obj = {a: \'a\',\nb: \'b\'};',
      options: ['first', { exceptions: { ObjectExpression: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var a = \'a\',\narr = [1,\n2];',
      output: 'var a = \'a\'\n,arr = [1,\n2];',
      options: ['first', { exceptions: { ArrayExpression: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var ar =[1,\n{a: \'a\',\nb: \'b\'}];',
      output: 'var ar =[1,\n{a: \'a\'\n,b: \'b\'}];',
      options: ['first', { exceptions: { ArrayExpression: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var ar =[1,\n{a: \'a\',\nb: \'b\'}];',
      output: 'var ar =[1\n,{a: \'a\',\nb: \'b\'}];',
      options: ['first', { exceptions: { ObjectExpression: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var ar ={fst:1,\nsnd: [1,\n2]};',
      output: 'var ar ={fst:1,\nsnd: [1\n,2]};',
      options: ['first', { exceptions: { ObjectExpression: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var ar ={fst:1,\nsnd: [1,\n2]};',
      output: 'var ar ={fst:1\n,snd: [1,\n2]};',
      options: ['first', { exceptions: { ArrayExpression: true } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'new Foo(a,\nb);',
      output: 'new Foo(a\n,b);',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'var foo = [\n(bar\n)\n,\nbaz\n];',
      output: 'var foo = [\n(bar\n),\nbaz\n];',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
          column: 1,
          endColumn: 2,
        },
      ],
    },
    {
      code: '[(foo),\n,\nbar]',
      output: '[(foo),,\nbar]',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: 'new Foo(a\n,b);',
      output: 'new Foo(a,\nb);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: '[\n[foo(3)],\n,\nbar\n];',
      output: '[\n[foo(3)],,\nbar\n];',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: '[foo//\n,/*block\ncomment*/];',
      output: '[foo,//\n/*block\ncomment*/];',
      errors: [
        {
          messageId: 'unexpectedLineBeforeAndAfterComma',
        },
      ],
    },
    {
      code: 'import {\n  A,\n  B\n  , C\n} from \'module3\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};\nimport \'module4\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};',
      output: 'import {\n  A,\n  B,\n   C\n} from \'module3\' with {\n  a: \'v\',\n  b: \'v\',\n   c: \'v\'\n};\nimport \'module4\' with {\n  a: \'v\',\n  b: \'v\',\n   c: \'v\'\n};',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'import {\n  A,\n  B\n  , C\n} from \'module3\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};\nimport \'module4\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};',
      output: 'import {\n  A\n  ,B\n  , C\n} from \'module3\' with {\n  a: \'v\'\n  ,b: \'v\'\n  , c: \'v\'\n};\nimport \'module4\' with {\n  a: \'v\'\n  ,b: \'v\'\n  , c: \'v\'\n};',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'let a, b, c;\nexport {\n  a,\n  b\n  , c\n};\nexport {\n  A,\n  B\n  , C\n} from \'module1\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};\nexport * from \'module2\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};',
      output: 'let a, b, c;\nexport {\n  a,\n  b,\n   c\n};\nexport {\n  A,\n  B,\n   C\n} from \'module1\' with {\n  a: \'v\',\n  b: \'v\',\n   c: \'v\'\n};\nexport * from \'module2\' with {\n  a: \'v\',\n  b: \'v\',\n   c: \'v\'\n};',
      errors: [
        {
          messageId: 'expectedCommaLast',
          line: 5,
        },
        {
          messageId: 'expectedCommaLast',
          line: 10,
        },
        {
          messageId: 'expectedCommaLast',
          line: 14,
        },
        {
          messageId: 'expectedCommaLast',
          line: 19,
        },
      ],
    },
    {
      code: 'let a, b, c;\nexport {\n  a,\n  b\n  , c\n};\nexport {\n  A,\n  B\n  , C\n} from \'module1\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};\nexport * from \'module2\' with {\n  a: \'v\',\n  b: \'v\'\n  , c: \'v\'\n};',
      output: 'let a, b, c;\nexport {\n  a\n  ,b\n  , c\n};\nexport {\n  A\n  ,B\n  , C\n} from \'module1\' with {\n  a: \'v\'\n  ,b: \'v\'\n  , c: \'v\'\n};\nexport * from \'module2\' with {\n  a: \'v\'\n  ,b: \'v\'\n  , c: \'v\'\n};',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'const x = (\n  a,\n  b\n  , c\n);',
      output: 'const x = (\n  a,\n  b,\n   c\n);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'const x = (\n  a,\n  b\n  , c\n);',
      output: 'const x = (\n  a\n  ,b\n  , c\n);',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'import(\n  a,\n  b\n);\nimport(\n  c\n  , d\n);',
      output: 'import(\n  a,\n  b\n);\nimport(\n  c,\n   d\n);',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'import(\n  a,\n  b\n);\nimport(\n  c\n  , d\n);',
      output: 'import(\n  a\n  ,b\n);\nimport(\n  c\n  , d\n);',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'class MyClass implements\n  A,\n  B\n, C {\n}\nconst a = class implements\n  A,\n  B\n, C {\n}',
      output: 'class MyClass implements\n  A,\n  B,\n C {\n}\nconst a = class implements\n  A,\n  B,\n C {\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'class MyClass implements\n  A,\n  B\n, C {\n}\nconst a = class implements\n  A,\n  B\n, C {\n}',
      output: 'class MyClass implements\n  A\n  ,B\n, C {\n}\nconst a = class implements\n  A\n  ,B\n, C {\n}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'function f(\n  a,\n  b\n  , c\n)\ntype a = (\n  a,\n  b\n  , c\n) => r\ntype a = new (\n  a,\n  b\n  , c\n) => r\nabstract class Base {\n  f(\n    a,\n    b\n    , c\n  );\n}',
      output: 'function f(\n  a,\n  b,\n   c\n)\ntype a = (\n  a,\n  b,\n   c\n) => r\ntype a = new (\n  a,\n  b,\n   c\n) => r\nabstract class Base {\n  f(\n    a,\n    b,\n     c\n  );\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'function f(\n  a,\n  b\n  , c\n)\ntype a = (\n  a,\n  b\n  , c\n) => r\ntype a = new (\n  a,\n  b\n  , c\n) => r\nabstract class Base {\n  f(\n    a,\n    b\n    , c\n  );\n}',
      output: 'function f(\n  a\n  ,b\n  , c\n)\ntype a = (\n  a\n  ,b\n  , c\n) => r\ntype a = new (\n  a\n  ,b\n  , c\n) => r\nabstract class Base {\n  f(\n    a\n    ,b\n    , c\n  );\n}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'enum MyEnum {\n  A,\n  B\n  , C\n}',
      output: 'enum MyEnum {\n  A,\n  B,\n   C\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'enum MyEnum {\n  A,\n  B\n  , C\n}',
      output: 'enum MyEnum {\n  A\n  ,B\n  , C\n}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'type foo = {\n  a: string,\n  b: string\n  , c: string\n}',
      output: 'type foo = {\n  a: string,\n  b: string,\n   c: string\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'type foo = {\n  a: string,\n  b: string\n  , c: string\n}',
      output: 'type foo = {\n  a: string\n  ,b: string\n  , c: string\n}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'type foo = {\n  new (\n    a,\n    b\n    , c\n  ): any,\n  (\n    a,\n    b\n    , c\n  ): any,\n  [\n    a: string,\n    b: string\n    , c: string\n  ]: string,\n\n  f(\n    a: string,\n    b: string\n    , c: string\n  ): number,\n}',
      output: 'type foo = {\n  new (\n    a,\n    b,\n     c\n  ): any,\n  (\n    a,\n    b,\n     c\n  ): any,\n  [\n    a: string,\n    b: string,\n     c: string\n  ]: string,\n\n  f(\n    a: string,\n    b: string,\n     c: string\n  ): number,\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'type foo = {\n  new (\n    a,\n    b\n    , c\n  ): any,\n  (\n    a,\n    b\n    , c\n  ): any,\n  [\n    a: string,\n    b: string\n    , c: string\n  ]: string,\n\n  f(\n    a: string,\n    b: string\n    , c: string\n  ): number,\n}',
      output: 'type foo = {\n  new (\n    a\n    ,b\n    , c\n  ): any\n  ,(\n    a\n    ,b\n    , c\n  ): any\n  ,[\n    a: string\n    ,b: string\n    , c: string\n  ]: string\n\n  ,f(\n    a: string\n    ,b: string\n    , c: string\n  ): number\n,}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'interface Foo extends\n  A,\n  B\n  , C\n{\n  a: string,\n  b: string\n  , c: string\n}',
      output: 'interface Foo extends\n  A,\n  B,\n   C\n{\n  a: string,\n  b: string,\n   c: string\n}',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'interface Foo extends\n  A,\n  B\n  , C\n{\n  a: string,\n  b: string\n  , c: string\n}',
      output: 'interface Foo extends\n  A\n  ,B\n  , C\n{\n  a: string\n  ,b: string\n  , c: string\n}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'type Foo = [\n  "A",\n  "B"\n  , "C"\n];',
      output: 'type Foo = [\n  "A",\n  "B",\n   "C"\n];',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'type Foo = [\n  "A",\n  "B"\n  , "C"\n];',
      output: 'type Foo = [\n  "A"\n  ,"B"\n  , "C"\n];',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'type Foo<\n  A,\n  B\n  , C\n> = Bar<\n  A,\n  B\n  , C\n>;',
      output: 'type Foo<\n  A,\n  B,\n   C\n> = Bar<\n  A,\n  B,\n   C\n>;',
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'type Foo<\n  A,\n  B\n  , C\n> = Bar<\n  A,\n  B\n  , C\n>;',
      output: 'type Foo<\n  A\n  ,B\n  , C\n> = Bar<\n  A\n  ,B\n  , C\n>;',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'import a\n  , {Foo} from \'module\'\nimport b\n  , {} from \'module\'\nimport c,\n  {Bar} from \'module\'\nimport d,\n  {} from \'module\'',
      output: 'import a,\n   {Foo} from \'module\'\nimport b,\n   {} from \'module\'\nimport c,\n  {Bar} from \'module\'\nimport d,\n  {} from \'module\'',
      options: ['last', { exceptions: { ImportDeclaration: false } }],
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'import a\n  , {Foo} from \'module\'\nimport b\n  , {} from \'module\'\nimport c,\n  {Bar} from \'module\'\nimport d,\n  {} from \'module\'',
      output: 'import a\n  , {Foo} from \'module\'\nimport b\n  , {} from \'module\'\nimport c\n  ,{Bar} from \'module\'\nimport d\n  ,{} from \'module\'',
      options: ['first', { exceptions: { ImportDeclaration: false } }],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'const x = {a,b\n,}',
      output: 'const x = {a,b,\n}',
      options: ['last'],
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'const x = {a,b,\n}',
      output: 'const x = {a,b\n,}',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
    {
      code: 'const x = [,\n  (a),\n  (b),\n  (c),\n]\nconst y = [\n  ,(a)\n  ,(b)\n  ,(c)\n  ,]',
      output: 'const x = [,\n  (a),\n  (b),\n  (c),\n]\nconst y = [\n  ,(a),\n  (b),\n  (c),\n  ]',
      options: ['last'],
      errors: [
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
        {
          messageId: 'expectedCommaLast',
        },
      ],
    },
    {
      code: 'const x = [,\n  (a),\n  (b),\n  (c),\n]\nconst y = [\n  ,(a)\n  ,(b)\n  ,(c)\n  ,]',
      output: 'const x = [,\n  (a)\n  ,(b)\n  ,(c)\n,]\nconst y = [\n  ,(a)\n  ,(b)\n  ,(c)\n  ,]',
      options: ['first'],
      errors: [
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
        {
          messageId: 'expectedCommaFirst',
        },
      ],
    },
  ],
});
