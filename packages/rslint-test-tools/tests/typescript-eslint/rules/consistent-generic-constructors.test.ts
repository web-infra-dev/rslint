// Upstream: @typescript-eslint/eslint-plugin v8.46.3.
// All 38 valid and 34 invalid cases preserve their input, options and output.
// Full diagnostic positions/messages were verified with that exact version.
import { RuleTester } from '@rslint/rule-tester';

const noFormat = (strings: TemplateStringsArray, ...values: unknown[]) =>
  String.raw({ raw: strings }, ...values);

const ruleTester = new RuleTester({
  plugin: '@typescript-eslint',
  filename: 'input.ts',
});

ruleTester.run('consistent-generic-constructors', {
  valid: [
    // default: constructor
    'const a = new Foo();',
    'const a = new Foo<string>();',
    'const a: Foo<string> = new Foo<string>();',
    'const a: Foo = new Foo();',
    'const a: Bar<string> = new Foo();',
    'const a: Foo = new Foo<string>();',
    'const a: Bar = new Foo<string>();',
    'const a: Bar<string> = new Foo<string>();',
    'const a: Foo<string> = Foo<string>();',
    'const a: Foo<string> = Foo();',
    'const a: Foo = Foo<string>();',
    `
class Foo {
  a = new Foo<string>();
}
    `,
    `
class Foo {
  accessor a = new Foo<string>();
}
    `,
    `
function foo(a: Foo = new Foo<string>()) {}
    `,
    `
function foo({ a }: Foo = new Foo<string>()) {}
    `,
    `
function foo([a]: Foo = new Foo<string>()) {}
    `,
    `
class A {
  constructor(a: Foo = new Foo<string>()) {}
}
    `,
    `
const a = function (a: Foo = new Foo<string>()) {};
    `,
    {
      code: `
const foo: Foo<string> = new Foo();
      `,
      languageOptions: {
        parserOptions: {
          isolatedDeclarations: true,
        },
      },
    },
    // type-annotation
    {
      code: 'const a = new Foo();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Foo<string> = new Foo();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Foo<string> = new Foo<string>();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Foo = new Foo();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Bar = new Foo<string>();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Bar<string> = new Foo<string>();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Foo<string> = Foo<string>();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Foo<string> = Foo();',
      options: ['type-annotation'],
    },
    {
      code: 'const a: Foo = Foo<string>();',
      options: ['type-annotation'],
    },
    {
      code: 'const a = new (class C<T> {})<string>();',
      options: ['type-annotation'],
    },
    {
      code: `
class Foo {
  a: Foo<string> = new Foo();
}
      `,
      options: ['type-annotation'],
    },
    {
      code: `
class Foo {
  accessor a: Foo<string> = new Foo();
}
      `,
      options: ['type-annotation'],
    },
    {
      code: `
function foo(a: Foo<string> = new Foo()) {}
      `,
      options: ['type-annotation'],
    },
    {
      code: `
function foo({ a }: Foo<string> = new Foo()) {}
      `,
      options: ['type-annotation'],
    },
    {
      code: `
function foo([a]: Foo<string> = new Foo()) {}
      `,
      options: ['type-annotation'],
    },
    {
      code: `
class A {
  constructor(a: Foo<string> = new Foo()) {}
}
      `,
      options: ['type-annotation'],
    },
    {
      code: `
const a = function (a: Foo<string> = new Foo()) {};
      `,
      options: ['type-annotation'],
    },
    {
      code: `
const [a = new Foo<string>()] = [];
      `,
      options: ['type-annotation'],
    },
    {
      code: `
function a([a = new Foo<string>()]) {}
      `,
      options: ['type-annotation'],
    },
  ],
  invalid: [
    {
      code: 'const a: Foo<string> = new Foo();',
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 33,
          suggestions: [],
        },
      ],
      output: 'const a = new Foo<string>();',
    },
    {
      code: 'const a: Map<string, number> = new Map();',
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 41,
          suggestions: [],
        },
      ],
      output: 'const a = new Map<string, number>();',
    },
    {
      code: noFormat`const a: Map <string, number> = new Map();`,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: `const a = new Map<string, number>();`,
    },
    {
      code: noFormat`const a: Map< string, number > = new Map();`,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 43,
          suggestions: [],
        },
      ],
      output: `const a = new Map< string, number >();`,
    },
    {
      code: noFormat`const a: Map<string, number> = new Map ();`,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: `const a = new Map<string, number> ();`,
    },
    {
      code: noFormat`const a: Foo<number> = new Foo;`,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 31,
          suggestions: [],
        },
      ],
      output: `const a = new Foo<number>();`,
    },
    {
      code: 'const a: /* comment */ Foo/* another */ <string> = new Foo();',
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 61,
          suggestions: [],
        },
      ],
      output: `const a = new Foo/* comment *//* another */<string>();`,
    },
    {
      code: 'const a: Foo/* comment */ <string> = new Foo /* another */();',
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 61,
          suggestions: [],
        },
      ],
      output: `const a = new Foo/* comment */<string> /* another */();`,
    },
    {
      code: noFormat`const a: Foo<string> = new \n Foo \n ();`,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 1,
          column: 7,
          endLine: 3,
          endColumn: 4,
          suggestions: [],
        },
      ],
      output: `const a = new \n Foo<string> \n ();`,
    },
    {
      code: `
class Foo {
  a: Foo<string> = new Foo();
}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 30,
          suggestions: [],
        },
      ],
      output: `
class Foo {
  a = new Foo<string>();
}
      `,
    },
    {
      code: `
class Foo {
  [a]: Foo<string> = new Foo();
}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 32,
          suggestions: [],
        },
      ],
      output: `
class Foo {
  [a] = new Foo<string>();
}
      `,
    },
    {
      code: `
class Foo {
  accessor a: Foo<string> = new Foo();
}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 39,
          suggestions: [],
        },
      ],
      output: `
class Foo {
  accessor a = new Foo<string>();
}
      `,
    },
    {
      code: `
class Foo {
  accessor a = new Foo<string>();
}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 34,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
class Foo {
  accessor a: Foo<string> = new Foo();
}
      `,
    },
    {
      code: `
class Foo {
  accessor [a]: Foo<string> = new Foo();
}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 41,
          suggestions: [],
        },
      ],
      output: `
class Foo {
  accessor [a] = new Foo<string>();
}
      `,
    },
    {
      code: `
function foo(a: Foo<string> = new Foo()) {}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 40,
          suggestions: [],
        },
      ],
      output: `
function foo(a = new Foo<string>()) {}
      `,
    },
    {
      code: `
function foo({ a }: Foo<string> = new Foo()) {}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 44,
          suggestions: [],
        },
      ],
      output: `
function foo({ a } = new Foo<string>()) {}
      `,
    },
    {
      code: `
function foo([a]: Foo<string> = new Foo()) {}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 42,
          suggestions: [],
        },
      ],
      output: `
function foo([a] = new Foo<string>()) {}
      `,
    },
    {
      code: `
class A {
  constructor(a: Foo<string> = new Foo()) {}
}
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 41,
          suggestions: [],
        },
      ],
      output: `
class A {
  constructor(a = new Foo<string>()) {}
}
      `,
    },
    {
      code: `
const a = function (a: Foo<string> = new Foo()) {};
      `,
      errors: [
        {
          messageId: 'preferConstructor',
          message:
            'The generic type arguments should be specified as part of the constructor type arguments.',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 47,
          suggestions: [],
        },
      ],
      output: `
const a = function (a = new Foo<string>()) {};
      `,
    },
    {
      code: 'const a = new Foo<string>();',
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 28,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: 'const a: Foo<string> = new Foo();',
    },
    {
      code: 'const a = new Map<string, number>();',
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 36,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: 'const a: Map<string, number> = new Map();',
    },
    {
      code: noFormat`const a = new Map <string, number> ();`,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `const a: Map<string, number> = new Map  ();`,
    },
    {
      code: noFormat`const a = new Map< string, number >();`,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 38,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `const a: Map< string, number > = new Map();`,
    },
    {
      code: noFormat`const a = new \n Foo<string> \n ();`,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 3,
          endColumn: 4,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `const a: Foo<string> = new \n Foo \n ();`,
    },
    {
      code: 'const a = new Foo/* comment */ <string> /* another */();',
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 56,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `const a: Foo<string> = new Foo/* comment */  /* another */();`,
    },
    {
      code: 'const a = new Foo</* comment */ string, /* another */ number>();',
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 1,
          column: 7,
          endLine: 1,
          endColumn: 64,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `const a: Foo</* comment */ string, /* another */ number> = new Foo();`,
    },
    {
      code: `
class Foo {
  a = new Foo<string>();
}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 25,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
class Foo {
  a: Foo<string> = new Foo();
}
      `,
    },
    {
      code: `
class Foo {
  [a] = new Foo<string>();
}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 27,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
class Foo {
  [a]: Foo<string> = new Foo();
}
      `,
    },
    {
      code: `
class Foo {
  [a + b] = new Foo<string>();
}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 3,
          column: 3,
          endLine: 3,
          endColumn: 31,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
class Foo {
  [a + b]: Foo<string> = new Foo();
}
      `,
    },
    {
      code: `
function foo(a = new Foo<string>()) {}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 35,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
function foo(a: Foo<string> = new Foo()) {}
      `,
    },
    {
      code: `
function foo({ a } = new Foo<string>()) {}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 39,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
function foo({ a }: Foo<string> = new Foo()) {}
      `,
    },
    {
      code: `
function foo([a] = new Foo<string>()) {}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 37,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
function foo([a]: Foo<string> = new Foo()) {}
      `,
    },
    {
      code: `
class A {
  constructor(a = new Foo<string>()) {}
}
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 3,
          column: 15,
          endLine: 3,
          endColumn: 36,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
class A {
  constructor(a: Foo<string> = new Foo()) {}
}
      `,
    },
    {
      code: `
const a = function (a = new Foo<string>()) {};
      `,
      errors: [
        {
          messageId: 'preferTypeAnnotation',
          message:
            'The generic type arguments should be specified as part of the type annotation.',
          line: 2,
          column: 21,
          endLine: 2,
          endColumn: 42,
          suggestions: [],
        },
      ],
      options: ['type-annotation'],
      output: `
const a = function (a: Foo<string> = new Foo()) {};
      `,
    },
  ],
});
