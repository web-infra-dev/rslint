import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('max-classes-per-file', {
  valid: [
    'class Foo {}',
    'var x = class {};',
    'var x = 5;',
    { code: 'class Foo {}', options: [1] as any },
    { code: 'class Foo {}\nclass Bar {}', options: [2] as any },
    { code: 'class Foo {}', options: [{ max: 1 }] as any },
    { code: 'class Foo {}\nclass Bar {}', options: [{ max: 2 }] as any },
    {
      code: `
                class Foo {}
                const myExpression = class {}
            `,
      options: [{ ignoreExpressions: true, max: 1 }] as any,
    },
    {
      code: `
                class Foo {}
                class Bar {}
                const myExpression = class {}
            `,
      options: [{ ignoreExpressions: true, max: 2 }] as any,
    },
  ],
  invalid: [
    {
      code: 'class Foo {}\nclass Bar {}',
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 13,
        },
      ],
    },
    {
      code: 'class Foo {}\nconst myExpression = class {}',
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
    {
      code: 'var x = class {};\nvar y = class {};',
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {}\nvar x = class {};',
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 1,
          column: 1,
          endLine: 2,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'class Foo {} class Bar {}',
      options: [1] as any,
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 26,
        },
      ],
    },
    {
      code: 'class Foo {} class Bar {} class Baz {}',
      options: [2] as any,
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 1,
          column: 1,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `
                class Foo {}
                class Bar {}
                const myExpression = class {}
            `,
      options: [{ ignoreExpressions: true, max: 1 }] as any,
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 2,
          column: 17,
          endLine: 4,
          endColumn: 46,
        },
      ],
    },
    {
      code: `
                class Foo {}
                class Bar {}
                class Baz {}
                const myExpression = class {}
            `,
      options: [{ ignoreExpressions: true, max: 2 }] as any,
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 2,
          column: 17,
          endLine: 5,
          endColumn: 46,
        },
      ],
    },
    {
      code: '/* comment */\nclass A {}\nclass B {}\n/* comment */',
      errors: [
        {
          messageId: 'maximumExceeded',
          line: 2,
          column: 1,
          endLine: 3,
          endColumn: 11,
        },
      ],
    },
  ],
});
