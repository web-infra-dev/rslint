import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('comma-spacing', null as never, {
  valid: [
    // comment-around-comma shapes (default)
    { code: `myfunc(404, true/* bla bla bla */, 'hello');` },
    { code: `myfunc(404, true /* bla bla bla */, 'hello');` },
    { code: `myfunc(404, true/* bla bla bla *//* hi */, 'hello');` },
    { code: `myfunc(404, true/* bla bla bla */ /* hi */, 'hello');` },
    { code: `myfunc(404, true, /* bla bla bla */ 'hello');` },
    { code: "myfunc(404, // comment\n true, /* bla bla bla */ 'hello');" },
    {
      code: "myfunc(404, // comment\n true,/* bla bla bla */ 'hello');",
      options: [{ before: false, after: false }],
    },

    // arrays and array holes
    { code: `var a = 1, b = 2;` },
    { code: `var arr = [,];` },
    { code: `var arr = [, ];` },
    { code: `var arr = [ ,];` },
    { code: `var arr = [ , ];` },
    { code: `var arr = [1,];` },
    { code: `var arr = [1, ];` },
    { code: `var arr = [, 2];` },
    { code: `var arr = [ , 2];` },
    { code: `var arr = [1, 2];` },
    { code: `var arr = [,,];` },
    { code: `var arr = [ ,,];` },
    { code: `var arr = [, ,];` },
    { code: `var arr = [,, ];` },
    { code: `var arr = [ , ,];` },
    { code: `var arr = [ ,, ];` },
    { code: `var arr = [, , ];` },
    { code: `var arr = [ , , ];` },
    { code: `var arr = [1, , ];` },
    { code: `var arr = [, 2, ];` },
    { code: `var arr = [, , 3];` },
    { code: `var arr = [,, 3];` },
    { code: `var arr = [1, 2, ];` },
    { code: `var arr = [, 2, 3];` },
    { code: `var arr = [1, , 3];` },
    { code: `var arr = [1, 2, 3];` },
    { code: `var arr = [1, 2, 3,];` },
    { code: `var arr = [1, 2, 3, ];` },

    // objects
    { code: `var obj = {'foo':'bar', 'baz':'qur'};` },
    { code: `var obj = {'foo':'bar', 'baz':'qur', };` },
    { code: `var obj = {'foo':'bar', 'baz':'qur',};` },
    { code: "var obj = {'foo':'bar', 'baz':\n'qur'};" },
    { code: "var obj = {'foo':\n'bar', 'baz':\n'qur'};" },

    // functions / arrows
    { code: `function foo(a, b){}` },
    { code: `function foo(a, b = 1){}` },
    { code: `function foo(a = 1, b, c){}` },
    { code: `var foo = (a, b) => {}` },
    { code: `var foo = (a=1, b) => {}` },
    { code: `var foo = a => a + 2` },

    // sequence expressions and parens
    { code: `a, b` },
    { code: `var a = (1 + 2, 2);` },
    { code: `a(b, c)` },
    { code: `new A(b, c)` },
    { code: `foo((a), b)` },
    { code: `var b = ((1 + 2), 2);` },
    { code: `parseInt((a + b), 10)` },
    { code: `go.boom((a + b), 10)` },
    { code: `go.boom((a + b), 10, (4))` },
    { code: `var x = [ (a + c), (b + b) ]` },
    { code: `['  ,  ']` },
    { code: '[`  ,  `]' },
    { code: '`${[1, 2]}`' },
    { code: `fn(a, b,)` },
    { code: `const fn = (a, b,) => {}` },
    { code: `const fn = function (a, b,) {}` },
    { code: `function fn(a, b,) {}` },
    { code: `foo(/,/, 'a')` },
    { code: `var x = ',,,,,';` },
    { code: `var code = 'var foo = 1, bar = 3;'` },
    { code: "['apples', \n 'oranges'];" },
    { code: `({x: 'var x,y,z'})` },

    // { before:true, after:false }
    {
      code: "var obj = {'foo':\n'bar' ,'baz':\n'qur'};",
      options: [{ before: true, after: false }],
    },
    { code: `var a = 1 ,b = 2;`, options: [{ before: true, after: false }] },
    { code: `function foo(a ,b){}`, options: [{ before: true, after: false }] },
    { code: `var arr = [,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [, ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ , ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 , ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ ,2];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 ,2];`, options: [{ before: true, after: false }] },
    { code: `var arr = [,,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ ,,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [, ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [,, ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ , ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ ,, ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [, , ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ , , ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 , ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ ,2 ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [,2 , ];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ , ,3];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 ,2 ,];`, options: [{ before: true, after: false }] },
    { code: `var arr = [ ,2 ,3];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 , ,3];`, options: [{ before: true, after: false }] },
    { code: `var arr = [1 ,2 ,3];`, options: [{ before: true, after: false }] },

    // { before:true, after:true }
    {
      code: `var obj = {'foo':'bar' , 'baz':'qur'};`,
      options: [{ before: true, after: true }],
    },
    {
      code: `var obj = {'foo':'bar' ,'baz':'qur' , };`,
      options: [{ before: true, after: false }],
    },
    { code: `var a = 1 , b = 2;`, options: [{ before: true, after: true }] },
    { code: `var arr = [, ];`, options: [{ before: true, after: true }] },
    { code: `var arr = [,,];`, options: [{ before: true, after: true }] },
    { code: `var arr = [1 , ];`, options: [{ before: true, after: true }] },
    { code: `var arr = [ , 2];`, options: [{ before: true, after: true }] },
    { code: `var arr = [1 , 2];`, options: [{ before: true, after: true }] },
    { code: `var arr = [, , ];`, options: [{ before: true, after: true }] },
    { code: `var arr = [1 , , ];`, options: [{ before: true, after: true }] },
    { code: `var arr = [ , 2 , ];`, options: [{ before: true, after: true }] },
    { code: `var arr = [ , , 3];`, options: [{ before: true, after: true }] },
    { code: `var arr = [1 , 2 , ];`, options: [{ before: true, after: true }] },
    { code: `var arr = [, 2 , 3];`, options: [{ before: true, after: true }] },
    { code: `var arr = [1 , , 3];`, options: [{ before: true, after: true }] },
    {
      code: `var arr = [1 , 2 , 3];`,
      options: [{ before: true, after: true }],
    },
    { code: `a , b`, options: [{ before: true, after: true }] },

    // { before:false, after:false }
    { code: `var arr = [,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [ ,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [1,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [,2];`, options: [{ before: false, after: false }] },
    { code: `var arr = [ ,2];`, options: [{ before: false, after: false }] },
    { code: `var arr = [1,2];`, options: [{ before: false, after: false }] },
    { code: `var arr = [,,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [ , , ];`, options: [{ before: false, after: false }] },
    { code: `var arr = [ ,,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [1,,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [,2,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [ ,2,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [,,3];`, options: [{ before: false, after: false }] },
    { code: `var arr = [1,2,];`, options: [{ before: false, after: false }] },
    { code: `var arr = [,2,3];`, options: [{ before: false, after: false }] },
    { code: `var arr = [1,,3];`, options: [{ before: false, after: false }] },
    { code: `var arr = [1,2,3];`, options: [{ before: false, after: false }] },
    { code: `var a = (1 + 2,2)`, options: [{ before: false, after: false }] },

    // templates and destructuring
    { code: 'var a; console.log(`${a}`, "a");' },
    { code: `var [a, b] = [1, 2];` },
    { code: `var [a, b, ] = [1, 2];` },
    { code: `var [a, b,] = [1, 2];` },
    { code: `var [a, , b] = [1, 2, 3];` },
    { code: `var [a,, b] = [1, 2, 3];` },
    { code: `var [ , b] = a;` },
    { code: `var [, b] = a;` },
    { code: `var { a,} = a;` },
    { code: `import { a,} from 'mod';` },

    // JSX
    { code: `<a>,</a>` },
    { code: `<a>  ,  </a>` },
    {
      code: `<a>Hello, world</a>`,
      options: [{ before: true, after: false }],
    },

    // Backwards-compat null-element + comment shapes
    { code: `[a, /**/ , ]`, options: [{ before: false, after: true }] },
    { code: `[a , /**/, ]`, options: [{ before: true, after: true }] },
    { code: `[a, /**/ , ] = foo`, options: [{ before: false, after: true }] },
    { code: `[a , /**/, ] = foo`, options: [{ before: true, after: true }] },

    // TS
    { code: `const Foo = <T,>(foo: T) => {}` },
    { code: `function foo<T,>() {}` },
    { code: `class Foo<T, T1> {}` },
    { code: `type Foo<T, P,> = Bar<T, P>` },
    { code: `interface Foo<T, T1,>{}` },
    { code: `let foo,` },
  ],

  invalid: [
    {
      code: `a(b,c)`,
      output: `a(b , c)`,
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `new A(b,c)`,
      output: `new A(b , c)`,
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `var a = 1 ,b = 2;`,
      output: `var a = 1, b = 2;`,
      errors: [
        { message: `There should be no space before ','.` },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `var arr = [1 , 2];`,
      output: `var arr = [1, 2];`,
      errors: [{ message: `There should be no space before ','.` }],
    },
    {
      code: `var arr = [1 , ];`,
      output: `var arr = [1, ];`,
      errors: [{ message: `There should be no space before ','.` }],
    },
    {
      code: `var arr = [1 ,2];`,
      output: `var arr = [1, 2];`,
      errors: [
        { message: `There should be no space before ','.` },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `var arr = [(1) , 2];`,
      output: `var arr = [(1), 2];`,
      errors: [{ message: `There should be no space before ','.` }],
    },
    {
      code: `var arr = [1, 2];`,
      output: `var arr = [1 ,2];`,
      options: [{ before: true, after: false }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { message: `There should be no space after ','.` },
      ],
    },
    {
      code: 'var arr = [1\n  , 2];',
      output: 'var arr = [1\n  ,2];',
      options: [{ before: false, after: false }],
      errors: [{ message: `There should be no space after ','.` }],
    },
    {
      code: 'var arr = [1,\n  2];',
      output: 'var arr = [1 ,\n  2];',
      options: [{ before: true, after: false }],
      errors: [{ messageId: 'missing', data: { loc: 'before' } }],
    },
    {
      code: "var obj = {'foo':\n'bar', 'baz':\n'qur'};",
      output: "var obj = {'foo':\n'bar' ,'baz':\n'qur'};",
      options: [{ before: true, after: false }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { message: `There should be no space after ','.` },
      ],
    },
    {
      code: 'var obj = {a: 1\n  ,b: 2};',
      output: 'var obj = {a: 1\n  , b: 2};',
      errors: [{ messageId: 'missing', data: { loc: 'after' } }],
    },
    {
      code: 'var obj = {a: 1 ,\n  b: 2};',
      output: 'var obj = {a: 1,\n  b: 2};',
      options: [{ before: false, after: false }],
      errors: [{ message: `There should be no space before ','.` }],
    },
    {
      code: `var arr = [1 ,2];`,
      output: `var arr = [1 , 2];`,
      options: [{ before: true, after: true }],
      errors: [{ messageId: 'missing', data: { loc: 'after' } }],
    },
    {
      code: `var arr = [1,2];`,
      output: `var arr = [1 , 2];`,
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: "var obj = {'foo':\n'bar','baz':\n'qur'};",
      output: "var obj = {'foo':\n'bar' , 'baz':\n'qur'};",
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `var arr = [1 , 2];`,
      output: `var arr = [1,2];`,
      options: [{ before: false, after: false }],
      errors: [
        { message: `There should be no space before ','.` },
        { message: `There should be no space after ','.` },
      ],
    },
    {
      code: `a ,b`,
      output: `a, b`,
      options: [{ before: false, after: true }],
      errors: [
        { message: `There should be no space before ','.` },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `function foo(a,b){}`,
      output: `function foo(a , b){}`,
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `var foo = (a,b) => {}`,
      output: `var foo = (a , b) => {}`,
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `var foo = (a = 1,b) => {}`,
      output: `var foo = (a = 1 , b) => {}`,
      options: [{ before: true, after: true }],
      errors: [
        { messageId: 'missing', data: { loc: 'before' } },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `function foo(a = 1 ,b = 2) {}`,
      output: `function foo(a = 1, b = 2) {}`,
      options: [{ before: false, after: true }],
      errors: [
        { message: `There should be no space before ','.` },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `<a>{foo(1 ,2)}</a>`,
      output: `<a>{foo(1, 2)}</a>`,
      errors: [
        { message: `There should be no space before ','.` },
        { messageId: 'missing', data: { loc: 'after' } },
      ],
    },
    {
      code: `myfunc(404, true/* bla bla bla */ , 'hello');`,
      output: `myfunc(404, true/* bla bla bla */, 'hello');`,
      errors: [{ message: `There should be no space before ','.` }],
    },
    {
      code: `myfunc(404, true,/* bla bla bla */ 'hello');`,
      output: `myfunc(404, true, /* bla bla bla */ 'hello');`,
      errors: [{ messageId: 'missing', data: { loc: 'after' } }],
    },
    {
      code: "myfunc(404,// comment\n true, 'hello');",
      output: "myfunc(404, // comment\n true, 'hello');",
      errors: [{ messageId: 'missing', data: { loc: 'after' } }],
    },

    // TS
    {
      code: `function Foo<T,T1>() {}`,
      output: `function Foo<T, T1>() {}`,
      errors: [
        { messageId: 'missing', column: 15, line: 1, data: { loc: 'after' } },
      ],
    },
    {
      code: `function Foo<T , T1>() {}`,
      output: `function Foo<T, T1>() {}`,
      errors: [
        {
          messageId: 'unexpected',
          column: 16,
          line: 1,
          data: { loc: 'before' },
        },
      ],
    },
    {
      code: `function Foo<T ,T1>() {}`,
      output: `function Foo<T, T1>() {}`,
      errors: [
        {
          messageId: 'unexpected',
          column: 16,
          line: 1,
          data: { loc: 'before' },
        },
        { messageId: 'missing', column: 16, line: 1, data: { loc: 'after' } },
      ],
    },
    {
      code: `function Foo<T, T1>() {}`,
      output: `function Foo<T,T1>() {}`,
      options: [{ before: false, after: false }],
      errors: [
        {
          messageId: 'unexpected',
          column: 15,
          line: 1,
          data: { loc: 'after' },
        },
      ],
    },
    {
      code: `function Foo<T,T1>() {}`,
      output: `function Foo<T ,T1>() {}`,
      options: [{ before: true, after: false }],
      errors: [
        { messageId: 'missing', column: 15, line: 1, data: { loc: 'before' } },
      ],
    },
    {
      code: `let foo ,`,
      output: `let foo,`,
      errors: [
        {
          messageId: 'unexpected',
          column: 9,
          line: 1,
          data: { loc: 'before' },
        },
      ],
    },
    {
      code: `type Foo<T,P,> = Bar<T,P>`,
      output: `type Foo<T, P,> = Bar<T, P>`,
      errors: [
        { messageId: 'missing', column: 11, line: 1, data: { loc: 'after' } },
        { messageId: 'missing', column: 23, line: 1, data: { loc: 'after' } },
      ],
    },
  ],
});
