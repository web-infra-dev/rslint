import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-array-constructor', {
  valid: [
    'new Array(x)',
    'Array(x)',
    'new Array(9)',
    'Array(9)',
    'new foo.Array()',
    'foo.Array()',
    'new Array.foo',
    'Array.foo()',
    'new globalThis.Array',
    'const createArray = Array => new Array()',
    'var Array; new Array;',
    {
      code: 'new Array()',
      languageOptions: {
        globals: { Array: 'off' },
      },
    },
    // TypeScript
    'new Array<Foo>(1, 2, 3);',
    'new Array<Foo>();',
    'Array<Foo>(1, 2, 3);',
    'Array<Foo>();',
    'Array<Foo>(3);',
    'Array?.(x);',
    'Array?.(9);',
  ],
  invalid: [
    {
      code: 'new Array()',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'new Array',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'new Array(x, y)',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'new Array(0, 1, 2)',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'const array = Array?.();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'const array = Array(...args);',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'const array = Array(5, 6, ...args);',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'a = new (Array);',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Semicolon required before array literal to compensate for ASI
    {
      code: '\nfoo\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nnew foo\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nconst foo = function() {}\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // No semicolon required before array literal because ASI does not occur
    {
      code: '\n{}\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'foo: Array();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'while (a) Array();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'with (obj) Array();',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // No semicolon required before array literal because ASI still occurs
    {
      code: '\na++\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nfunction foo() {\n    return\n    Array();\n}\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\ndebugger\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: "\nexport { foo } from 'bar'\nArray()\n",
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nvar foo\nArray()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // TypeScript
    {
      code: 'new Array();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'Array(0, 1, 2);',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'Array?.(0, 1, 2);',
      errors: [{ messageId: 'preferLiteral' }],
    },
  ],
});
