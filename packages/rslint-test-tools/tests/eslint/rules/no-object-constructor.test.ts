import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-object-constructor', {
  valid: [
    'new Object(x)',
    'Object(x)',
    'new globalThis.Object',
    'const createObject = Object => new Object()',
    'var Object; new Object;',
    {
      code: 'new Object()',
      languageOptions: {
        globals: { Object: 'off' },
      },
    },
  ],
  invalid: [
    {
      code: 'new Object',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'Object()',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'const fn = () => Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'Object() instanceof Object;',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'const obj = Object?.();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '(new Object() instanceof Object);',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Semicolon required before `({})` to compensate for ASI
    {
      code: '\nfoo\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nfoo()\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nnew foo\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\n(a++)\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\n++a\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nconst foo = function() {}\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nconst foo = class {}\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nvar foo = { bar: baz }\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // No semicolon required before `({})` because ASI does not occur
    {
      code: '\n{}\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nfunction foo() {}\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nclass Foo {}\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'foo: Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'while (a) Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\ndo Object();\nwhile (a);\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'for (const prop in obj) Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'with (obj) Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // No semicolon required before `({})` because ASI still occurs
    {
      code: '\na++\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nfunction foo() {\n    return\n    Object();\n}\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\ndo {}\nwhile (a)\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\ndebugger\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nfoo: break foo\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: "\nexport { foo } from 'bar'\nObject()\n",
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '\nvar foo\nObject()\n',
      errors: [{ messageId: 'preferLiteral' }],
    },
  ],
});
