import { RuleTester } from '../rule-tester';

const ruleTester = new RuleTester();

ruleTester.run('strict', {
  valid: [
    // "never" mode
    { code: 'foo();', options: ['never'] as any },
    { code: 'function foo() { return; }', options: ['never'] as any },
    {
      code: 'var foo = function() { return; };',
      options: ['never'] as any,
    },
    { code: "foo(); 'use strict';", options: ['never'] as any },
    {
      code: "function foo() { bar(); 'use strict'; return; }",
      options: ['never'] as any,
    },
    {
      code: "var foo = function() { { 'use strict'; } return; };",
      options: ['never'] as any,
    },
    {
      code: "(function() { bar('use strict'); return; }());",
      options: ['never'] as any,
    },
    { code: 'var fn = x => 1;', options: ['never'] as any },
    { code: 'var fn = x => { return; };', options: ['never'] as any },

    // "global" mode
    { code: '// Intentionally empty', options: ['global'] as any },
    { code: "'use strict'; foo();", options: ['global'] as any },
    {
      code: "'use strict'; function foo() { return; }",
      options: ['global'] as any,
    },
    {
      code: "'use strict'; var foo = function() { return; };",
      options: ['global'] as any,
    },
    {
      code: "'use strict'; function foo() { bar(); 'use strict'; return; }",
      options: ['global'] as any,
    },
    {
      code: "'use strict'; function foo() { return function() { bar(); 'use strict'; return; }; }",
      options: ['global'] as any,
    },
    {
      code: "'use strict'; var foo = () => { return () => { bar(); 'use strict'; return; }; }",
      options: ['global'] as any,
    },

    // "function" mode
    {
      code: "function foo() { 'use strict'; return; }",
      options: ['function'] as any,
    },
    {
      code: "var foo = function() { 'use strict'; return; }",
      options: ['function'] as any,
    },
    {
      code: "function foo() { 'use strict'; return; } var bar = function() { 'use strict'; bar(); };",
      options: ['function'] as any,
    },
    {
      code: "var foo = function() { 'use strict'; function bar() { return; } bar(); };",
      options: ['function'] as any,
    },
    {
      code: "var foo = () => { 'use strict'; var bar = () => 1; bar(); };",
      options: ['function'] as any,
    },
    { code: 'class A { constructor() { } }', options: ['function'] as any },
    { code: 'class A { foo() { } }', options: ['function'] as any },
    {
      code: 'class A { foo() { function bar() { } } }',
      options: ['function'] as any,
    },
    {
      code: "(function() { 'use strict'; function foo(a = 0) { } }())",
      options: ['function'] as any,
    },

    // "safe" / default
    {
      code: "function foo() { 'use strict'; return; }",
      options: ['safe'] as any,
    },
    "function foo() { 'use strict'; return; }",

    // class static blocks have no directive prologue
    {
      code: "'use strict'; class C { static { foo; } }",
      options: ['global'] as any,
    },
    {
      code: "'use strict'; class C { static { 'use strict'; } }",
      options: ['global'] as any,
    },
    {
      code: "'use strict'; class C { static { 'use strict'; 'use strict'; } }",
      options: ['global'] as any,
    },
    { code: 'class C { static { foo; } }', options: ['function'] as any },
    {
      code: "class C { static { 'use strict'; } }",
      options: ['function'] as any,
    },
    {
      code: "class C { static { 'use strict'; 'use strict'; } }",
      options: ['function'] as any,
    },
    { code: 'class C { static { foo; } }', options: ['never'] as any },
    {
      code: "class C { static { 'use strict'; } }",
      options: ['never'] as any,
    },
    {
      code: "class C { static { 'use strict'; 'use strict'; } }",
      options: ['never'] as any,
    },

    // class heritage: function inside `extends` is NOT in the class body
    {
      code: "class Foo extends (function() { 'use strict'; return class {}; }()) {}",
      options: ['function'] as any,
    },

    // ambient / bodyless declarations (TypeScript)
    { code: 'declare function foo(): void;', options: ['function'] as any },
    { code: 'declare function foo(): void;', options: ['never'] as any },
    {
      code: 'abstract class A { abstract foo(): void; }',
      options: ['function'] as any,
    },
    {
      code: "function foo(): void; function foo(a: number): void; function foo(a?: number) { 'use strict'; }",
      options: ['function'] as any,
    },
  ],
  invalid: [
    // "never" mode
    {
      code: '"use strict"; foo();',
      options: ['never'] as any,
      errors: [{ messageId: 'never' }],
    },
    {
      code: "function foo() { 'use strict'; return; }",
      options: ['never'] as any,
      errors: [{ messageId: 'never' }],
    },
    {
      code: "var foo = function() { 'use strict'; return; };",
      options: ['never'] as any,
      errors: [{ messageId: 'never' }],
    },
    {
      code: "function foo() { return function() { 'use strict'; return; }; }",
      options: ['never'] as any,
      errors: [{ messageId: 'never' }],
    },
    {
      code: '\'use strict\'; function foo() { "use strict"; return; }',
      options: ['never'] as any,
      errors: [{ messageId: 'never' }, { messageId: 'never' }],
    },
    {
      code: '"use strict"; foo(); export {};',
      options: ['never'] as any,
      errors: [{ messageId: 'module' }],
    },

    // "global" mode
    {
      code: 'foo();',
      options: ['global'] as any,
      errors: [{ messageId: 'global' }],
    },
    {
      code: "function foo() { 'use strict'; return; }",
      options: ['global'] as any,
      errors: [{ messageId: 'global' }, { messageId: 'global' }],
    },
    {
      code: "var foo = function() { 'use strict'; return; }",
      options: ['global'] as any,
      errors: [{ messageId: 'global' }, { messageId: 'global' }],
    },
    {
      code: "var foo = () => { 'use strict'; return () => 1; }",
      options: ['global'] as any,
      errors: [{ messageId: 'global' }, { messageId: 'global' }],
    },
    {
      code: "'use strict'; function foo() { 'use strict'; return; }",
      options: ['global'] as any,
      errors: [{ messageId: 'global' }],
    },
    {
      code: "'use strict'; 'use strict'; foo();",
      options: ['global'] as any,
      errors: [{ messageId: 'multiple' }],
    },
    {
      code: "'use strict'; foo(); export {};",
      options: ['global'] as any,
      errors: [{ messageId: 'module' }],
    },

    // "function" mode
    {
      code: "'use strict'; foo();",
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },
    {
      code: "'use strict'; (function() { 'use strict'; return true; }());",
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },
    {
      code: "(function() { 'use strict'; function f() { 'use strict'; return } return true; }());",
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: '(function() { return true; }());',
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },
    {
      code: '(() => { return true; })();',
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },
    {
      code: "function foo() { 'use strict'; 'use strict'; return; }",
      options: ['function'] as any,
      errors: [{ messageId: 'multiple' }],
    },
    {
      code: "function foo() { return function() { 'use strict'; return; }; }",
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },
    {
      code: "function foo() { 'use strict'; return function() { 'use strict'; 'use strict'; return; }; }",
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessary' }, { messageId: 'multiple' }],
    },
    {
      code: 'var foo = () => { return; };',
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },

    // classes
    {
      code: 'class A { constructor() { "use strict"; } }',
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessaryInClasses' }],
    },
    {
      code: 'class A { foo() { "use strict"; } }',
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessaryInClasses' }],
    },
    {
      code: 'class A { foo() { function bar() { "use strict"; } } }',
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessaryInClasses' }],
    },
    {
      code: 'class A { field = () => { "use strict"; } }',
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessaryInClasses' }],
    },
    {
      code: 'class A { field = function() { "use strict"; } }',
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessaryInClasses' }],
    },

    // safe / default (= function in rslint script files)
    {
      code: "'use strict'; function foo() { return; }",
      options: ['safe'] as any,
      errors: [{ messageId: 'function' }, { messageId: 'function' }],
    },
    {
      code: "'use strict'; function foo() { return; }",
      errors: [{ messageId: 'function' }, { messageId: 'function' }],
    },
    {
      code: 'function foo() { return; }',
      errors: [{ messageId: 'function' }],
    },

    // non-simple parameter list
    {
      code: "function foo(a = 0) { 'use strict' }",
      options: ['never'] as any,
      errors: [{ messageId: 'nonSimpleParameterList' }],
    },
    {
      code: "function foo(a = 0) { 'use strict' }",
      options: ['global'] as any,
      errors: [
        { messageId: 'global' },
        { messageId: 'nonSimpleParameterList' },
      ],
    },
    {
      code: "function foo(a = 0) { 'use strict' }",
      options: ['function'] as any,
      errors: [{ messageId: 'nonSimpleParameterList' }],
    },
    {
      code: 'function foo(a = 0) { }',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    // wrap message: modifier / kind coverage
    {
      code: 'async function foo(a = 0) { }',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'function* gen(a = 0) { }',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'async function* gen(a = 0) { }',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: '(function(a = 0) { })();',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: '(function named(a = 0) { })();',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'var foo = (a = 0) => { };',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },
    {
      code: 'var foo = async (a = 0) => { };',
      options: ['function'] as any,
      errors: [{ messageId: 'wrap' }],
    },

    // functions inside class static blocks
    {
      code: "class C { static { function foo() { \n'use strict'; } } }",
      options: ['never'] as any,
      errors: [{ messageId: 'never' }],
    },
    {
      code: "function foo() {'use strict'; class C { static { function foo() { \n'use strict'; } } } }",
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessary' }],
    },
    {
      code: "class C { static { function foo() { \n'use strict'; } } }",
      options: ['function'] as any,
      errors: [{ messageId: 'unnecessaryInClasses' }],
    },

    // class heritage: function in extends without "use strict" still needs one
    // under "function" mode (it's effectively top-level, not in class body).
    {
      code: 'class Foo extends (function() { return class {}; }()) {}',
      options: ['function'] as any,
      errors: [{ messageId: 'function' }],
    },
  ],
});
