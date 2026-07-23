import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('prefer-object-spread', {
  valid: [
    'Object.assign()',
    'let a = Object.assign(a, b)',
    'Object.assign(a, b)',
    'let a = Object.assign(b, { c: 1 })',
    'const bar = { ...foo }',
    'Object.assign(...foo)',
    'Object.assign(foo, { bar: baz })',
    'Object.assign({}, ...objects)',
    "foo({ foo: 'bar' })",
    `
        const Object = {};
        Object.assign({}, foo);
        `,
    `
        const Object = {};
        Object.assign({ foo: 'bar' });
        `,
    `
        const Object = require('foo');
        Object.assign({ foo: 'bar' });
        `,
    `
        import Object from 'foo';
        Object.assign({ foo: 'bar' });
        `,
    `
        import { Something as Object } from 'foo';
        Object.assign({ foo: 'bar' });
        `,
    `
        import { Object, Array } from 'globals';
        Object.assign({ foo: 'bar' });
        `,
    // Note: unlike ESLint (which only recognizes `globalThis` under a
    // sufficiently recent ecmaVersion), rslint always recognizes
    // `globalThis.Object.assign(...)` — see the Go upstream test's
    // globalThis cases and the rule doc's "Differences from ESLint".
    `
                var globalThis = foo;
                globalThis.Object.assign({}, foo)
                `,
    'class C { #assign; foo() { Object.#assign({}, foo); } }',

    // ignore Object.assign() with > 1 arguments if any of the arguments is an object expression with a getter/setter
    'Object.assign({ get a() {} }, {})',
    'Object.assign({ set a(val) {} }, {})',
    'Object.assign({ get a() {} }, foo)',
    'Object.assign({ set a(val) {} }, foo)',
    "Object.assign({ foo: 'bar', get a() {}, baz: 'quux' }, quuux)",
    "Object.assign({ foo: 'bar', set a(val) {} }, { baz: 'quux' })",
    'Object.assign({}, { get a() {} })',
    'Object.assign({}, { set a(val) {} })',
    "Object.assign({}, { foo: 'bar', get a() {} }, {})",
    "Object.assign({ foo }, bar, {}, { baz: 'quux', set a(val) {}, quuux }, {})",
  ],
  invalid: [
    {
      code: 'Object.assign({}, foo)',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign  ({}, foo)',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({}, { foo: 'bar' })",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({}, baz, { foo: 'bar' })",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({}, { foo: 'bar', baz: 'foo' })",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({ foo: 'bar' }, baz)",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({ foo: 'bar' }, cats, dogs, trees, birds)",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({ foo: 'bar' }, Object.assign({ bar: 'foo' }, baz))",
      errors: [
        { messageId: 'useSpreadMessage', line: 1, column: 1 },
        { messageId: 'useSpreadMessage', line: 1, column: 31 },
      ],
    },
    {
      code: "Object.assign({foo: 'bar', ...bar}, baz)",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign({}, { foo, bar, baz })',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: "Object.assign({}, { [bar]: 'foo' })",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign({ ...bar }, { ...baz })',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign({})',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign({ foo: bar })',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'let a = Object.assign({})',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 9 }],
    },
    {
      code: 'let a = Object.assign({}, a)',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 9 }],
    },
    {
      code: 'let a = Object.assign({ a: 1 }, b)',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 9 }],
    },
    {
      code: 'Object.assign({}, a ? b : {}, b => c, a = 2)',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
    {
      code: '[1, 2, Object.assign({}, a)]',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 8 }],
    },
    {
      code: 'const foo = Object.assign({}, a)',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 13 }],
    },
    {
      code: 'function foo() { return Object.assign({}, a) }',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 25 }],
    },
    {
      code: 'foo(Object.assign({}, a));',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 5 }],
    },
    {
      code: "const x = { foo: 'bar', baz: Object.assign({}, a) }",
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 30 }],
    },
    {
      code: 'Object.assign({ });',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign({\n});',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'globalThis.Object.assign({ });',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'globalThis.Object.assign({\n});',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'Object.assign({ get a() {}, set b(val) {} })',
      errors: [{ messageId: 'useLiteralMessage', line: 1, column: 1 }],
    },
    {
      code: 'const obj = Object.assign<{}, Record<string, string[]>>({}, getObject());',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 13 }],
    },
    {
      code: 'Object.assign<{}, A>({}, foo);',
      errors: [{ messageId: 'useSpreadMessage', line: 1, column: 1 }],
    },
  ],
});
