import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();
const error = { messageId: 'shouldBeLast' };

ruleTester.run('default-param-last', {
  valid: [
    'function f() {}',
    'function f(a) {}',
    'function f(a = 5) {}',
    'function f(a, b) {}',
    'function f(a, b = 5) {}',
    'function f(a, b = 5, c = 5) {}',
    'function f(a, b = 5, ...c) {}',
    'const f = () => {}',
    'const f = (a) => {}',
    'const f = (a = 5) => {}',
    'const f = function f() {}',
    'const f = function f(a) {}',
    'const f = function f(a = 5) {}',

    'function foo() {}',
    'function foo(a: number) {}',
    'function foo(a = 1) {}',
    'function foo(a?: number) {}',
    'function foo(a: number, b: number) {}',
    'function foo(a: number, b: number, c?: number) {}',
    'function foo(a: number, b = 1) {}',
    'function foo(a: number, b = 1, c = 1) {}',
    'function foo(a: number, b = 1, c?: number) {}',
    'function foo(a: number, b?: number, c = 1) {}',
    'function foo(a: number, b = 1, ...c) {}',

    'const foo = function () {};',
    'const foo = function (a: number) {};',
    'const foo = function (a = 1) {};',
    'const foo = function (a?: number) {};',
    'const foo = function (a: number, b: number) {};',
    'const foo = function (a: number, b: number, c?: number) {};',
    'const foo = function (a: number, b = 1) {};',
    'const foo = function (a: number, b = 1, c = 1) {};',
    'const foo = function (a: number, b = 1, c?: number) {};',
    'const foo = function (a: number, b?: number, c = 1) {};',
    'const foo = function (a: number, b = 1, ...c) {};',

    'const foo = () => {};',
    'const foo = (a: number) => {};',
    'const foo = (a = 1) => {};',
    'const foo = (a?: number) => {};',
    'const foo = (a: number, b: number) => {};',
    'const foo = (a: number, b: number, c?: number) => {};',
    'const foo = (a: number, b = 1) => {};',
    'const foo = (a: number, b = 1, c = 1) => {};',
    'const foo = (a: number, b = 1, c?: number) => {};',
    'const foo = (a: number, b?: number, c = 1) => {};',
    'const foo = (a: number, b = 1, ...c) => {};',

    'class Foo { constructor(a: number, b: number, c: number) {} }',
    'class Foo { constructor(a: number, b?: number, c = 1) {} }',
    'class Foo { constructor(a: number, b = 1, c?: number) {} }',
    'class Foo { constructor(public a: number, protected b: number, private c: number) {} }',
    'class Foo { constructor(public a: number, protected b?: number, private c = 10) {} }',
    'class Foo { constructor(public a: number, protected b = 10, private c?: number) {} }',
    'class Foo { constructor(a: number, protected b?: number, private c = 0) {} }',
    'class Foo { constructor(a: number, b?: number, private c = 0) {} }',
    'class Foo { constructor(a: number, private b?: number, c = 0) {} }',
  ],
  invalid: [
    { code: 'function f(a = 5, b) {}', errors: [error] },
    { code: 'function f(a = 5, b = 6, c) {}', errors: [error, error] },
    { code: 'function f (a = 5, b, c = 6, d) {}', errors: [error, error] },
    { code: 'function f(a = 5, b, c = 5) {}', errors: [error] },
    { code: 'const f = (a = 5, b, ...c) => {}', errors: [error] },
    { code: 'const f = function f (a, b = 5, c) {}', errors: [error] },
    { code: 'const f = (a = 5, { b }) => {}', errors: [error] },
    { code: 'const f = ({ a } = {}, b) => {}', errors: [error] },
    {
      code: 'const f = ({ a, b } = { a: 1, b: 2 }, c) => {}',
      errors: [error],
    },
    { code: 'const f = ([a] = [], b) => {}', errors: [error] },
    { code: 'const f = ([a, b] = [1, 2], c) => {}', errors: [error] },

    { code: 'function foo(a = 1, b: number) {}', errors: [error] },
    {
      code: 'function foo(a = 1, b = 2, c: number) {}',
      errors: [error, error],
    },
    {
      code: 'function foo(a = 1, b: number, c = 2, d: number) {}',
      errors: [error, error],
    },
    { code: 'function foo(a = 1, b: number, c = 2) {}', errors: [error] },
    { code: 'function foo(a = 1, b: number, ...c) {}', errors: [error] },
    { code: 'function foo(a?: number, b: number) {}', errors: [error] },
    {
      code: 'function foo(a: number, b?: number, c: number) {}',
      errors: [error],
    },
    {
      code: 'function foo(a = 1, b?: number, c: number) {}',
      errors: [error, error],
    },

    { code: 'const foo = function (a = 1, b: number) {};', errors: [error] },
    {
      code: 'const foo = function (a = 1, b = 2, c: number) {};',
      errors: [error, error],
    },
    {
      code: 'const foo = function (a = 1, b: number, c = 2, d: number) {};',
      errors: [error, error],
    },
    {
      code: 'const foo = function (a = 1, b: number, c = 2) {};',
      errors: [error],
    },
    {
      code: 'const foo = function (a = 1, b: number, ...c) {};',
      errors: [error],
    },
    {
      code: 'const foo = function (a?: number, b: number) {};',
      errors: [error],
    },
    {
      code: 'const foo = function (a: number, b?: number, c: number) {};',
      errors: [error],
    },
    {
      code: 'const foo = function (a = 1, b?: number, c: number) {};',
      errors: [error, error],
    },

    { code: 'const foo = (a = 1, b: number) => {};', errors: [error] },
    {
      code: 'const foo = (a = 1, b = 2, c: number) => {};',
      errors: [error, error],
    },
    {
      code: 'const foo = (a = 1, b: number, c = 2, d: number) => {};',
      errors: [error, error],
    },
    { code: 'const foo = (a = 1, b: number, c = 2) => {};', errors: [error] },
    { code: 'const foo = (a = 1, b: number, ...c) => {};', errors: [error] },
    { code: 'const foo = (a?: number, b: number) => {};', errors: [error] },
    {
      code: 'const foo = (a: number, b?: number, c: number) => {};',
      errors: [error],
    },
    {
      code: 'const foo = (a = 1, b?: number, c: number) => {};',
      errors: [error, error],
    },

    {
      code: 'class Foo { constructor(public a: number, protected b?: number, private c: number) {} }',
      errors: [error],
    },
    {
      code: 'class Foo { constructor(public a: number, protected b = 0, private c: number) {} }',
      errors: [error],
    },
    {
      code: 'class Foo { constructor(public a?: number, private b: number) {} }',
      errors: [error],
    },
    {
      code: 'class Foo { constructor(public a = 0, private b: number) {} }',
      errors: [error],
    },
    { code: 'class Foo { constructor(a = 0, b: number) {} }', errors: [error] },
    {
      code: 'class Foo { constructor(a?: number, b: number) {} }',
      errors: [error],
    },
  ],
});
