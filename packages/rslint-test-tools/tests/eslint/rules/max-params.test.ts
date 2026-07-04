import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-params', {
  valid: [
    'function test(d, e, f) {}',
    { code: 'var test = function(a, b, c) {};', options: [3] as any },
    { code: 'var test = (a, b, c) => {};', options: [3] as any },
    { code: 'var test = function test(a, b, c) {};', options: [3] as any },
    {
      code: 'var test = function(a, b, c) {};',
      options: [{ max: 3 }] as any,
    },

    'function foo() {}',
    'const foo = function () {};',
    'const foo = () => {};',
    'function foo(a) {}',
    `
class Foo {
  constructor(a) {}
}
    `,
    `
class Foo {
  method(this: void, a, b, c) {}
}
    `,
    `
class Foo {
  method(this: Foo, a, b) {}
}
    `,
    {
      code: 'function foo(a, b, c, d) {}',
      options: [{ max: 4 }] as any,
    },
    {
      code: 'function foo(a, b, c, d) {}',
      options: [{ maximum: 4 }] as any,
    },
    {
      code: `
class Foo {
  method(this: void) {}
}
      `,
      options: [{ max: 0 }] as any,
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ max: 1 }] as any,
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ countVoidThis: true, max: 2 }] as any,
    },
    {
      code: 'function testD(this: void, a) {}',
      options: [{ max: 1 }] as any,
    },
    {
      code: 'function testD(this: void, a) {}',
      options: [{ countVoidThis: true, max: 2 }] as any,
    },
    {
      code: 'const testE = function (this: void, a) {}',
      options: [{ max: 1 }] as any,
    },
    {
      code: 'const testE = function (this: void, a) {}',
      options: [{ countVoidThis: true, max: 2 }] as any,
    },
    {
      code: `
declare function makeDate(m: number, d: number, y: number): Date;
      `,
      options: [{ max: 3 }] as any,
    },
    {
      code: `
type sum = (a: number, b: number) => number;
      `,
      options: [{ max: 2 }] as any,
    },
    {
      code: 'function foo(this: unknown[], a, b, c) {}',
      options: [{ max: 3, countThis: 'never' }] as any,
    },
    {
      code: 'function foo(this: void, a, b, c) {}',
      options: [{ max: 3, countThis: 'except-void' }] as any,
    },
  ],
  invalid: [
    {
      code: 'function test(a, b, c) {}',
      options: [2] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function test(a, b, c, d) {}',
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'var test = function(a, b, c, d) {};',
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 12 }],
    },
    {
      code: 'var test = (a, b, c, d) => {};',
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 25 }],
    },
    {
      code: '(function(a, b, c, d) {});',
      options: [3] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 2 }],
    },
    {
      code: 'var test = function test(a, b, c) {};',
      options: [1] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 12 }],
    },
    {
      code: 'function test(a, b, c) {}',
      options: [{ max: 2 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function test(a, b, c, d) {}',
      options: [{}] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: 'function test(a) {}',
      options: [{ max: 0 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },
    {
      code: `function test(a, b, c) {
              // Just to make it longer
            }`,
      options: [{ max: 2 }] as any,
      errors: [{ messageId: 'exceed', line: 1, column: 1 }],
    },

    { code: 'function foo(a, b, c, d) {}', errors: [{ messageId: 'exceed' }] },
    {
      code: 'const foo = function (a, b, c, d) {};',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'const foo = (a, b, c, d) => {};',
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'const foo = a => {};',
      options: [{ max: 0 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
class Foo {
  method(this: void, a, b, c, d) {}
}
      `,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ countVoidThis: true, max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ countThis: 'always', max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'function testD(this: void, a) {}',
      options: [{ countVoidThis: true, max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'function testD(this: void, a) {}',
      options: [{ countThis: 'always', max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'const testE = function (this: void, a) {}',
      options: [{ countThis: 'always', max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'function testFunction(test: void, a: number) {}',
      options: [{ countThis: 'except-void', max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'const testE = function (this: void, a) {}',
      options: [{ countVoidThis: true, max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'function testFunction(test: void, a: number) {}',
      options: [{ countVoidThis: false, max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
class Foo {
  method(this: Foo, a, b, c) {}
}
      `,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
declare function makeDate(m: number, d: number, y: number): Date;
      `,
      options: [{ max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: `
type sum = (a: number, b: number) => number;
      `,
      options: [{ max: 1 }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'function foo(this: unknown[], a, b, c) {}',
      options: [{ max: 3, countThis: 'always' }] as any,
      errors: [{ messageId: 'exceed' }],
    },
    {
      code: 'function foo(this: unknown[], a, b, c) {}',
      options: [{ max: 3, countThis: 'except-void' }] as any,
      errors: [{ messageId: 'exceed' }],
    },
  ],
});
