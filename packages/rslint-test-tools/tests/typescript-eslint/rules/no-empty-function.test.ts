import { RuleTester, getFixturesRootDir } from '../RuleTester.ts';

const ruleTester = new RuleTester({
  languageOptions: {
    parserOptions: {
      project: './tsconfig.json',
      tsconfigRootDir: getFixturesRootDir(),
    },
  },
});

ruleTester.run('no-empty-function', {
  valid: [
    {
      code: `
class Person {
  private name: string;
  constructor(name: string) {
    this.name = name;
  }
}
      `,
    },
    {
      code: `
class Person {
  constructor(private name: string) {}
}
      `,
    },
    {
      code: `
class Person {
  constructor(name: string) {}
}
      `,
      options: [{ allow: ['constructors'] }],
    },
    {
      code: `
class Person {
  otherMethod(name: string) {}
}
      `,
      options: [{ allow: ['methods'] }],
    },
    {
      code: `
class Foo {
  private constructor() {}
}
      `,
      options: [{ allow: ['private-constructors'] }],
    },
    {
      code: `
class Foo {
  protected constructor() {}
}
      `,
      options: [{ allow: ['protected-constructors'] }],
    },
    {
      code: `
function foo() {
  const a = null;
}
      `,
    },
    {
      code: `
class Foo {
  @decorator()
  foo() {}
}
      `,
      options: [{ allow: ['decoratedFunctions'] }],
    },
    {
      code: `
class Foo extends Base {
  override foo() {}
}
      `,
      options: [{ allow: ['overrideMethods'] }],
    },
    // Additional tests for various allowed types
    {
      code: `const foo = () => {};`,
      options: [{ allow: ['arrowFunctions'] }],
    },
    {
      code: `function* foo() {}`,
      options: [{ allow: ['generatorFunctions'] }],
    },
    {
      code: `
class Foo {
  *method() {}
}
      `,
      options: [{ allow: ['generatorMethods'] }],
    },
    {
      code: `
class Foo {
  get foo() {}
}
      `,
      options: [{ allow: ['getters'] }],
    },
    {
      code: `
class Foo {
  set foo(value) {}
}
      `,
      options: [{ allow: ['setters'] }],
    },
    {
      code: `async function foo() {}`,
      options: [{ allow: ['asyncFunctions'] }],
    },
    {
      code: `
class Foo {
  async method() {}
}
      `,
      options: [{ allow: ['asyncMethods'] }],
    },
    {
      code: `function foo() {}`,
      options: [{ allow: ['functions'] }],
    },
  ],

  invalid: [
    {
      code: `
class Person {
  constructor(name: string) {}
}
      `,
      errors: [
        {
          column: 29,
          data: {
            name: 'constructor',
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Person {
  otherMethod(name: string) {}
}
      `,
      errors: [
        {
          column: 29,
          data: {
            name: "method 'otherMethod'",
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  private constructor() {}
}
      `,
      errors: [
        {
          column: 25,
          data: {
            name: 'constructor',
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  protected constructor() {}
}
      `,
      errors: [
        {
          column: 27,
          data: {
            name: 'constructor',
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
function foo() {}
      `,
      errors: [
        {
          column: 16,
          data: {
            name: "function 'foo'",
          },
          line: 2,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  @decorator()
  foo() {}
}
      `,
      errors: [
        {
          column: 9,
          data: {
            name: "method 'foo'",
          },
          line: 4,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo extends Base {
  override foo() {}
}
      `,
      errors: [
        {
          column: 18,
          data: {
            name: "method 'foo'",
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    // Additional invalid test cases
    {
      code: `const foo = () => {};`,
      errors: [
        {
          column: 19,
          data: {
            name: "arrow function 'foo'",
          },
          line: 1,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `function* foo() {}`,
      errors: [
        {
          column: 17,
          data: {
            name: "function 'foo'",
          },
          line: 1,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  *method() {}
}
      `,
      errors: [
        {
          column: 13,
          data: {
            name: "method 'method'",
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  get foo() {}
}
      `,
      errors: [
        {
          column: 13,
          data: {
            name: "getter 'foo'",
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  set foo(value) {}
}
      `,
      errors: [
        {
          column: 18,
          data: {
            name: "setter 'foo'",
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `async function foo() {}`,
      errors: [
        {
          column: 22,
          data: {
            name: "function 'foo'",
          },
          line: 1,
          messageId: 'unexpected',
        },
      ],
    },
    {
      code: `
class Foo {
  async method() {}
}
      `,
      errors: [
        {
          column: 18,
          data: {
            name: "method 'method'",
          },
          line: 3,
          messageId: 'unexpected',
        },
      ],
    },
  ],
});