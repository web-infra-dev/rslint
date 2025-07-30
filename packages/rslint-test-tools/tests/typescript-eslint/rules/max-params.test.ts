import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const rootPath = getFixturesRootDir();

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      projectService: false,
      tsconfigRootDir: rootPath,
    },
  },
});

ruleTester.run('max-params', {
  valid: [
    { code: 'function foo() {}' },
    { code: 'const foo = function () {};' },
    { code: 'const foo = () => {};' },
    { code: 'function foo(a) {}' },
    {
      code: `
class Foo {
  constructor(a) {}
}
    `,
    },
    {
      code: `
class Foo {
  method(this: void, a, b, c) {}
}
    `,
    },
    {
      code: `
class Foo {
  method(this: Foo, a, b) {}
}
    `,
    },
    {
      code: 'function foo(a, b, c, d) {}',
      options: [{ max: 4 }],
    },
    {
      code: 'function foo(a, b, c, d) {}',
      options: [{ maximum: 4 }],
    },
    {
      code: `
class Foo {
  method(this: void) {}
}
      `,
      options: [{ max: 0 }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ max: 1 }],
    },
    {
      code: `
class Foo {
  method(this: void, a) {}
}
      `,
      options: [{ countVoidThis: true, max: 2 }],
    },
    {
      code: `
declare function makeDate(m: number, d: number, y: number): Date;
      `,
      options: [{ max: 3 }],
    },
    {
      code: `
type sum = (a: number, b: number) => number;
      `,
      options: [{ max: 2 }],
    },
  ],
  invalid: [
    {
      code: 'function foo(a, b, c, d) {}',
      errors: [{ messageId: 'exceed' }],
    },
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
      errors: [{ messageId: 'exceed' }],
      options: [{ max: 0 }],
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
      errors: [{ messageId: 'exceed' }],
      options: [{ countVoidThis: true, max: 1 }],
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
      errors: [{ messageId: 'exceed' }],
      options: [{ max: 1 }],
    },
    {
      code: `
type sum = (a: number, b: number) => number;
      `,
      errors: [{ messageId: 'exceed' }],
      options: [{ max: 1 }],
    },
  ],
});
