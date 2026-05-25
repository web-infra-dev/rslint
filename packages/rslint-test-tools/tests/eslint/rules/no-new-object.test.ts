import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-new-object', {
  valid: [
    // Non-matching AST shapes
    'var myObject = {};',
    'var myObject = new CustomObject();',
    'var foo = Object("foo");',
    'var foo = new foo.Object()',

    // Shadowing — key declaration types
    'var Object = function Object() {}; new Object();',
    'function Object() {} new Object();',
    'class Object { constructor(){} } new Object();',

    // Shadowing — parameter / destructuring / catch
    'function bar(Object) { var baz = new Object(); }',
    'var { Object } = obj; new Object();',
    'try {} catch(Object) { new Object(); }',

    // Scope propagation & hoisting
    'var Object = 1; function f() { new Object(); }',
    'new Object(); var Object = 1;',
  ],
  invalid: [
    // Basic forms
    {
      code: 'var foo = new Object()',
      errors: [{ messageId: 'preferLiteral' }],
    },
    { code: 'new Object();', errors: [{ messageId: 'preferLiteral' }] },
    { code: 'new Object', errors: [{ messageId: 'preferLiteral' }] },
    {
      code: 'var foo = new Object("foo")',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Parenthesized callee
    { code: 'new (Object)()', errors: [{ messageId: 'preferLiteral' }] },
    // TS type assertion on callee
    { code: 'new (Object as any)()', errors: [{ messageId: 'preferLiteral' }] },
    // TS generic type argument
    { code: 'new Object<any>()', errors: [{ messageId: 'preferLiteral' }] },

    // Multiple errors
    {
      code: 'new Object(); new Object();',
      errors: [{ messageId: 'preferLiteral' }, { messageId: 'preferLiteral' }],
    },

    // Non-shadowing scope boundaries
    {
      code: 'function bar() { return function Object() {}; } var baz = new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '{ let Object = 1; } new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'function foo() { var Object = 1; } new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '(function(Object) {})(1); new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'try {} catch(Object) {} new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Inside scopes without shadow
    {
      code: 'function f() { new Object(); }',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'const f = () => new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'class C { m() { new Object(); } }',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'function a() { function b() { function c() { new Object(); } } }',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Expression contexts
    {
      code: '({ m() { new Object() } })',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'function f(x = new Object()) {}',
      errors: [{ messageId: 'preferLiteral' }],
    },
    { code: 'foo(new Object())', errors: [{ messageId: 'preferLiteral' }] },
    {
      code: 'function f() { return new Object(); }',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Class body
    {
      code: 'class C { x = new Object() }',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'class C { static x = new Object() }',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // TypeScript type-level declarations do NOT shadow
    {
      code: 'type Object = string; new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: 'interface Object { x: number } new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },

    // Mixed scopes
    {
      code: 'function f(Object) { new Object(); } new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
    {
      code: '{ let Object = 1; new Object(); } new Object();',
      errors: [{ messageId: 'preferLiteral' }],
    },
  ],
});
