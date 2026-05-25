import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-obj-calls', {
  valid: [
    'var x = Math.random();',
    'var x = JSON.parse(foo);',
    'Reflect.get(foo, "x");',
    'new Intl.Segmenter();',
    'var x = Math;',
    'var x = Math.PI;',
    'var x = foo.Math();',
    'var x = new foo.Math();',
    'JSON.parse(foo)',
    'new JSON.parse',
    // globalThis property access (not calling the global itself)
    'var x = new globalThis.Math.foo;',
    'new globalThis.Object()',
    // Shadowed variable — should not be flagged
    'function f() { var Math = 1; Math(); }',
    'function f(JSON: any) { JSON(); }',
    'function f() { var globalThis = { Math: () => {} }; globalThis.Math(); }',
  ],
  invalid: [
    {
      code: 'Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'var x = JSON();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'var x = Reflect();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'Atomics();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'Intl();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new JSON();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new Reflect();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new Atomics();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'new Intl();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    // globalThis access
    {
      code: 'var x = globalThis.Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'var x = new globalThis.Math();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'var x = globalThis.JSON();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'var x = globalThis.Reflect();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'globalThis.Atomics();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'globalThis.Intl();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    // globalThis with optional chaining
    {
      code: 'var x = globalThis?.Reflect();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    {
      code: 'var x = (globalThis?.Reflect)();',
      errors: [{ messageId: 'unexpectedCall' }],
    },
    // multiple errors in one expression
    {
      code: 'Math( JSON() );',
      errors: [
        { messageId: 'unexpectedCall' },
        { messageId: 'unexpectedCall' },
      ],
    },
    {
      code: 'globalThis.Math( globalThis.JSON() );',
      errors: [
        { messageId: 'unexpectedCall' },
        { messageId: 'unexpectedCall' },
      ],
    },
    // indirect references via variable assignment
    {
      code: 'var foo = JSON; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = Math; new foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = bar ? baz : JSON; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = globalThis.JSON; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    // indirect via logical operators
    {
      code: 'var foo = undefined || JSON; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = undefined ?? Reflect; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    // TS type assertions as pass-through in initializer
    {
      code: 'var foo = JSON as any; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = JSON satisfies any; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = <any>JSON; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    {
      code: 'var foo = JSON!; foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    // comma operator
    {
      code: 'var foo = (0, JSON); foo();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    // multi-hop indirect references
    {
      code: 'var a = JSON; var b = a; b();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
    // direct call with TS assertion as callee
    {
      code: '(JSON as any)();',
      errors: [{ messageId: 'unexpectedRefCall' }],
    },
  ],
});
