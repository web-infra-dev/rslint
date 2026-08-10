import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('consistent-return', {
  valid: [
    'function foo() { return; }',
    'function foo() { if (true) return; }',
    'function foo() { if (true) return; else return; }',
    'function foo() { if (true) return true; else return false; }',
    'f(function() { return; })',
    'f(function() { if (true) return; })',
    'f(function() { if (true) return; else return; })',
    'f(function() { if (true) return true; else return false; })',
    'function foo() { function bar() { return true; } return; }',
    'function foo() { function bar() { return; } return false; }',
    'function Foo() { if (!(this instanceof Foo)) return new Foo(); }',
    'function foo() { if (true) return 5; else return undefined; }',
    'function foo() { if (true) return 5; else return void 0; }',
    {
      code: 'function foo() { if (true) return; else return undefined; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
    },
    {
      code: 'function foo() { if (true) return; else return void 0; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
    },
    {
      code: 'function foo() { if (true) return undefined; else return; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
    },
    {
      code: 'function foo() { if (true) return undefined; else return void 0; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
    },
    {
      code: 'function foo() { if (true) return void 0; else return; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
    },
    {
      code: 'function foo() { if (true) return void 0; else return undefined; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
    },
    'var x = () => {  return {}; };',
    'if (true) { return 1; } return 0;',

    // https://github.com/eslint/eslint/issues/7790
    'class Foo { constructor() { if (true) return foo; } }',
    'var Foo = class { constructor() { if (true) return foo; } }',
  ],
  invalid: [
    {
      code: 'function foo() { if (true) return true; else return; }',
      errors: [{ messageId: 'missingReturnValue', column: 46 }],
    },
    {
      code: 'var foo = () => { if (true) return true; else return; }',
      errors: [{ messageId: 'missingReturnValue', column: 47 }],
    },
    {
      code: 'function foo() { if (true) return; else return false; }',
      errors: [{ messageId: 'unexpectedReturnValue', column: 41 }],
    },
    {
      code: 'f(function() { if (true) return true; else return; })',
      errors: [{ messageId: 'missingReturnValue', column: 44 }],
    },
    {
      code: 'f(function() { if (true) return; else return false; })',
      errors: [{ messageId: 'unexpectedReturnValue', column: 39 }],
    },
    {
      code: 'f(a => { if (true) return; else return false; })',
      errors: [{ messageId: 'unexpectedReturnValue', column: 33 }],
    },
    {
      code: 'function foo() { if (true) return true; return undefined; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
      errors: [{ messageId: 'missingReturnValue', column: 41 }],
    },
    {
      code: 'function foo() { if (true) return true; return void 0; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
      errors: [{ messageId: 'missingReturnValue', column: 41 }],
    },
    {
      code: 'function foo() { if (true) return undefined; return true; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
      errors: [{ messageId: 'unexpectedReturnValue', column: 46 }],
    },
    {
      code: 'function foo() { if (true) return void 0; return true; }',
      options: [{ treatUndefinedAsUnspecified: true }] as any,
      errors: [{ messageId: 'unexpectedReturnValue', column: 43 }],
    },
    {
      code: 'if (true) { return 1; } return;',
      errors: [{ messageId: 'missingReturnValue', column: 25 }],
    },
    {
      code: 'function foo() { if (a) return true; }',
      errors: [{ messageId: 'missingReturn', column: 10 }],
    },
    {
      code: 'function _foo() { if (a) return true; }',
      errors: [{ messageId: 'missingReturn', column: 10 }],
    },
    {
      code: 'f(function foo() { if (a) return true; });',
      errors: [{ messageId: 'missingReturn', column: 12 }],
    },
    {
      code: 'f(function() { if (a) return true; });',
      errors: [{ messageId: 'missingReturn', column: 3 }],
    },
    {
      code: 'f(() => { if (a) return true; });',
      errors: [{ messageId: 'missingReturn', column: 6 }],
    },
    {
      code: 'var obj = {foo() { if (a) return true; }};',
      errors: [{ messageId: 'missingReturn', column: 12 }],
    },
    {
      code: 'class A {foo() { if (a) return true; }};',
      errors: [{ messageId: 'missingReturn', column: 10 }],
    },
    {
      code: 'if (a) return true;',
      errors: [{ messageId: 'missingReturn', column: 1 }],
    },
    {
      code: 'class A { CapitalizedFunction() { if (a) return true; } }',
      errors: [{ messageId: 'missingReturn', column: 11 }],
    },
    {
      code: '({ constructor() { if (a) return true; } });',
      errors: [{ messageId: 'missingReturn', column: 4 }],
    },
  ],
});
