import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('comma-style', null as never, {
  valid: [
    // default ("last")
    { code: 'var foo = 1, bar = 3;' },
    { code: "var foo = {'a': 1, 'b': 2};" },
    { code: 'var foo = [1, 2];' },
    { code: 'var foo = [, 2];' },
    { code: 'var foo = [1, ];' },
    { code: "var foo = ['apples', \n 'oranges'];" },
    { code: "var foo = {'a': 1, \n 'b': 2, \n'c': 3};" },
    { code: 'var foo = [1, \n2, \n3];' },
    { code: 'function foo(){var a=[1,\n 2]}' },
    { code: "function foo(){return {'a': 1,\n'b': 2}}" },
    { code: 'var foo = \n1, \nbar = \n2;' },
    { code: 'var foo = [\n(bar),\nbaz\n];' },

    // explicit "first"
    { code: 'var foo = 1, bar = 2;', options: ['first'] },
    { code: 'var foo = 1 \n ,bar = 2;', options: ['first'] },
    { code: "var foo = {'a': 1 \n ,'b': 2 \n,'c': 3};", options: ['first'] },
    { code: 'var foo = [1 \n ,2 \n, 3];', options: ['first'] },

    // exceptions
    {
      code: "var arr = ['a',\n'o'];",
      options: ['first', { exceptions: { ArrayExpression: true } }],
    },
    {
      code: 'new Foo(a\n,b);',
      options: ['last', { exceptions: { NewExpression: true } }],
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
      code: 'import { a\n, b } from "./source";',
      options: ['last', { exceptions: { ImportDeclaration: true } }],
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
      code: 'type Foo = [\n  "A",\n  "B"\n  , "C"\n];',
      options: ['first', { exceptions: { TSTupleType: true } }],
    },
  ],
  invalid: [
    // default "last" — lone comma report
    {
      code: 'var foo = 1\n,\nbar = 2;',
      output: 'var foo = 1,\nbar = 2;',
      errors: [{ messageId: 'unexpectedLineBeforeAndAfterComma' }],
    },
    {
      code: 'var foo = 1\n,bar = 2;',
      output: 'var foo = 1,\nbar = 2;',
      errors: [{ messageId: 'expectedCommaLast', column: 1, endColumn: 2 }],
    },
    {
      code: "var foo = ['apples'\n, 'oranges'];",
      output: "var foo = ['apples',\n 'oranges'];",
      errors: [{ messageId: 'expectedCommaLast' }],
    },
    {
      code: 'f(1\n, 2);',
      output: 'f(1,\n 2);',
      errors: [{ messageId: 'expectedCommaLast' }],
    },
    {
      code: 'function foo(a\n, b) { return a + b; }',
      output: 'function foo(a,\n b) { return a + b; }',
      errors: [{ messageId: 'expectedCommaLast' }],
    },
    {
      code: 'import { a\n, b } from "./source";',
      output: 'import { a,\n b } from "./source";',
      errors: [{ messageId: 'expectedCommaLast' }],
    },
    {
      code: 'var {foo\n, bar} = {foo:"apples", bar:"oranges"};',
      output: 'var {foo,\n bar} = {foo:"apples", bar:"oranges"};',
      errors: [{ messageId: 'expectedCommaLast' }],
    },

    // explicit "first"
    {
      code: 'var foo = 1,\nbar = 2;',
      output: 'var foo = 1\n,bar = 2;',
      options: ['first'],
      errors: [{ messageId: 'expectedCommaFirst', column: 12, endColumn: 13 }],
    },
    {
      code: 'new Foo(a,\nb);',
      output: 'new Foo(a\n,b);',
      options: ['first'],
      errors: [{ messageId: 'expectedCommaFirst' }],
    },

    // TS member separator
    {
      code: 'type foo = {\n  a: string,\n  b: string\n  , c: string\n}',
      output: 'type foo = {\n  a: string,\n  b: string,\n   c: string\n}',
      errors: [{ messageId: 'expectedCommaLast' }],
    },
    {
      code: 'enum MyEnum {\n  A,\n  B\n  , C\n}',
      output: 'enum MyEnum {\n  A,\n  B,\n   C\n}',
      errors: [{ messageId: 'expectedCommaLast' }],
    },
    {
      code: 'type Foo = [\n  "A",\n  "B"\n  , "C"\n];',
      output: 'type Foo = [\n  "A",\n  "B",\n   "C"\n];',
      errors: [{ messageId: 'expectedCommaLast' }],
    },
  ],
});
