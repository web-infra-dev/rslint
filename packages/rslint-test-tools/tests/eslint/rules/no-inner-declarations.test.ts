import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('no-inner-declarations', {
  valid: [
    // Default mode ("functions") with blockScopedFunctions "allow" (default)
    'function doSomething() { }',
    'function doSomething() { function somethingElse() { } }',
    '(function() { function doSomething() { } }());',
    'if (test) { var fn = function() { }; }',
    'if (test) { var fn = function expr() { }; }',
    'function decl() { var fn = function expr() { }; }',
    'if (test) var x = 42;',
    'var x = 1;',
    'var fn = function() { };',
    'function foo() { if (test) { var x = 1; } }',
    'if (test) { var foo; }',
    'function doSomething() { while (test) { var foo; } }',

    // Block-scoped functions in MODULE files (have import/export → strict → allowed)
    'export {}; if (foo) function f(){}',
    'export {}; function bar() { if (foo) function f(){}; }',
    'export {}; while (test) { function doSomething() { } }',
    'export {}; do { function foo() {} } while (test)',

    // Block-scoped functions with "use strict" directive → allowed
    '"use strict"; if (foo) function f(){}',
    '"use strict"; function bar() { if (foo) function f(){}; }',

    // Block-scoped functions inside function with "use strict" → allowed
    'function outer() { "use strict"; if (foo) function f(){}; }',
    'function outer() { "use strict"; { function inner() {} } }',

    // Block-scoped functions inside class body (implicit strict) → allowed
    'class C { method() { if(test) { function somethingElse() { } } } }',
    'class C { method() { { function bar() { } } } }',

    // Export declarations
    'export function foo() {}',
    'export default function() {}',

    // "both" mode - valid placements
    { code: 'function doSomething() { }', options: ['both'] as any },
    { code: 'function doSomething() { var x = 1; }', options: ['both'] as any },
    { code: 'var x = 1;', options: ['both'] as any },
    { code: 'var foo = 42;', options: ['both'] as any },
    { code: 'var fn = function() { };', options: ['both'] as any },
    { code: 'function foo() { var x = 1; }', options: ['both'] as any },
    { code: '(function() { var foo; }());', options: ['both'] as any },

    // Arrow functions
    'export {}; foo(() => { function bar() { } });',
    { code: 'var fn = () => {var foo;}', options: ['both'] as any },
    {
      code: 'const doSomething = () => { var foo = 42; }',
      options: ['both'] as any,
    },

    // Class methods
    'var x = {doSomething() {function doSomethingElse() {}}}',
    { code: 'var x = {doSomething() {var foo;}}', options: ['both'] as any },
    {
      code: 'class C { method() { function foo() {} } }',
      options: ['both'] as any,
    },
    { code: 'class C { method() { var x; } }', options: ['both'] as any },

    // Class static blocks
    {
      code: 'class C { static { function foo() {} } }',
      options: ['both'] as any,
    },
    { code: 'class C { static { var x; } }', options: ['both'] as any },

    // let/const never flagged in "both" mode
    { code: 'if (test) { let x = 1; }', options: ['both'] as any },
    { code: 'if (test) { const x = 1; }', options: ['both'] as any },

    // Export with "both" mode
    { code: 'export var foo: any;', options: ['both'] as any },
    { code: 'export function bar() {}', options: ['both'] as any },
    { code: 'export default function baz() {}', options: ['both'] as any },

    // blockScopedFunctions "allow" in module context
    {
      code: 'export {}; function foo() { { function bar() { } } }',
      options: ['functions', { blockScopedFunctions: 'allow' }] as any,
    },

    // TypeScript-specific: namespace/module blocks
    'namespace Foo { function bar() {} }',
    { code: 'namespace Foo { var x = 1; }', options: ['both'] as any },
    "declare module 'foo' { function bar(): void; }",
  ],
  invalid: [
    // blockScopedFunctions "allow" (default) in NON-STRICT script files
    // No import/export, no "use strict" → script → non-strict → reported
    {
      code: 'if (foo) function f(){}',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'function bar() { if (foo) function f(){}; }',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'while (test) { function doSomething() { } }',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'do { function foo() {} } while (test)',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'function doSomething() { do { function somethingElse() { } } while (test); }',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: '(function() { if (test) { function doSomething() { } } }());',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'function foo() { { function bar() { } } }',
      errors: [{ messageId: 'moveDeclToRoot' }],
    },

    // blockScopedFunctions "disallow" — always reports regardless of strict mode
    {
      code: 'export {}; if (foo) function f(){}',
      options: ['functions', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; function bar() { if (foo) function f(){}; }',
      options: ['functions', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; while (test) { function doSomething() { } }',
      options: ['functions', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; do { function foo() {} } while (test)',
      options: ['functions', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; function doSomething() { do { function somethingElse() { } } while (test); }',
      options: ['functions', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; (function() { if (test) { function doSomething() { } } }());',
      options: ['functions', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },

    // "both" mode - var declarations in blocks
    {
      code: 'if (foo) { var a; }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'if (foo) var a;',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'function bar() { if (foo) var a; }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'if (foo) { var fn = function(){} }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'while (test) { var foo; }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'function doSomething() { if (test) { var foo = 42; } }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: '(function() { if (test) { var foo; } }());',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'const doSomething = () => { if (test) { var foo = 42; } }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },

    // Class method - var in block
    {
      code: 'class C { method() { if(test) { var foo; } } }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },

    // Class static block - var in nested block
    {
      code: 'class C { static { if (test) { var foo; } } }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    // Class static block - function with blockScopedFunctions "disallow"
    {
      code: 'class C { static { if (test) { function foo() {} } } }',
      options: ['both', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    // Class static block - deeply nested var
    {
      code: 'class C { static { if (test) { if (anotherTest) { var foo; } } } }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },

    // "both" + blockScopedFunctions "disallow"
    {
      code: 'if (foo) { var bar = 1; }',
      options: ['both', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; if (test) { function doSomething() { } }',
      options: ['both', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
    {
      code: 'export {}; function foo() { { function bar() { } } }',
      options: ['both', { blockScopedFunctions: 'disallow' }] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },

    // TypeScript-specific: var in nested block inside namespace
    {
      code: 'namespace Foo { if (test) { var x = 1; } }',
      options: ['both'] as any,
      errors: [{ messageId: 'moveDeclToRoot' }],
    },
  ],
});
