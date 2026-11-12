import { RuleTester } from '@rslint/rule-tester';
import { config } from '../config';

// Migrated from the existing suite. Each active case pins the full diagnostic and edits.
new RuleTester(config).run('no-inner-declarations', {
  valid: [
    {
      code: 'function doSomething() { }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'function doSomething() { function somethingElse() { } }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '(function() { function doSomething() { } }());',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { var fn = function() { }; }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { var fn = function expr() { }; }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'function decl() { var fn = function expr() { }; }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'function decl(arg) { var fn; if (arg) { fn = function() { }; } }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var x = {doSomething() {function doSomethingElse() {}}}',
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }',
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'function decl(arg) { var fn; if (arg) { fn = function expr() { }; } }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { var foo; }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { let x = 1; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { const x = 1; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { using x = 1; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2026,
        sourceType: 'script',
      },
    },
    {
      code: 'if (test) { await using x = 1; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2026,
        sourceType: 'script',
      },
    },
    {
      code: 'function doSomething() { while (test) { var foo; } }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var foo;',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'var foo = 42;',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'function doSomething() { var foo; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: '(function() { var foo; }());',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'foo(() => { function bar() { } });',
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'var fn = () => {var foo;}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'var x = {doSomething() {var foo;}}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'export var foo;',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'export function bar() {}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'export default function baz() {}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'exports.foo = () => {}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'exports.foo = function(){}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'module.exports = function foo(){}',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
    },
    {
      code: 'class C { method() { function foo() {} } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: 'class C { method() { var x; } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: 'class C { static { function foo() {} } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: 'class C { static { var x; } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: `'use strict'
 if (test) { function doSomething() { } }`,
      options: [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: `'use strict'
 if (test) { function doSomething() { } }`,
      options: ['functions'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: `function foo() {'use strict'
 if (test) { function doSomething() { } } }`,
      options: [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
    },
    {
      code: 'function foo() { { function bar() { } } }',
      options: [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'module',
      },
    },
    {
      code: 'class C { method() { if(test) { function somethingElse() { } } } }',
      options: [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
    {
      code: 'const C = class { method() { if(test) { function somethingElse() { } } } }',
      options: [
        'functions',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
    },
  ],
  invalid: [
    {
      code: 'if (test) { function doSomething() { } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'if (foo) var a; ',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to program root.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 16,
        },
      ],
    },
    {
      code: 'if (foo) /* some comments */ var a; ',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to program root.',
          line: 1,
          column: 30,
          endLine: 1,
          endColumn: 36,
        },
      ],
    },
    {
      code: 'if (foo){ function f(){ if(bar){ var a; } } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 44,
        },
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 34,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'if (foo) function f(){ if(bar) var a; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 10,
          endLine: 1,
          endColumn: 40,
        },
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 38,
        },
      ],
    },
    {
      code: 'if (foo) { var fn = function(){} } ',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to program root.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'if (foo)  function f(){} ',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 11,
          endLine: 1,
          endColumn: 25,
        },
      ],
    },
    {
      code: 'function bar() { if (foo) function f(){}; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to function body root.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'function bar() { if (foo) var a; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 33,
        },
      ],
    },
    {
      code: 'if (foo) { var a; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to program root.',
          line: 1,
          column: 12,
          endLine: 1,
          endColumn: 18,
        },
      ],
    },
    {
      code: 'function doSomething() { do { function somethingElse() { } } while (test); }',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to function body root.',
          line: 1,
          column: 31,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: '(function() { if (test) { function doSomething() { } } }());',
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to function body root.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 53,
        },
      ],
    },
    {
      code: 'while (test) { var foo; }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to program root.',
          line: 1,
          column: 16,
          endLine: 1,
          endColumn: 24,
        },
      ],
    },
    {
      code: 'function doSomething() { if (test) { var foo = 42; } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 38,
          endLine: 1,
          endColumn: 51,
        },
      ],
    },
    {
      code: '(function() { if (test) { var foo; } }());',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 27,
          endLine: 1,
          endColumn: 35,
        },
      ],
    },
    {
      code: 'const doSomething = () => { if (test) { var foo = 42; } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 41,
          endLine: 1,
          endColumn: 54,
        },
      ],
    },
    {
      code: 'class C { method() { if(test) { var foo; } } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2015,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to function body root.',
          line: 1,
          column: 33,
          endLine: 1,
          endColumn: 41,
        },
      ],
    },
    {
      code: 'class C { static { if (test) { var foo; } } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to class static block body root.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 40,
        },
      ],
    },
    {
      code: 'class C { static { if (test) { function foo() {} } } }',
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to class static block body root.',
          line: 1,
          column: 32,
          endLine: 1,
          endColumn: 49,
        },
      ],
    },
    {
      code: 'class C { static { if (test) { if (anotherTest) { var foo; } } } }',
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move variable declaration to class static block body root.',
          line: 1,
          column: 51,
          endLine: 1,
          endColumn: 59,
        },
      ],
    },
    {
      code: 'if (test) { function doSomething() { } }',
      options: [
        'both',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: 'if (test) { function doSomething() { } }',
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 13,
          endLine: 1,
          endColumn: 39,
        },
      ],
    },
    {
      code: `'use strict'
 if (test) { function doSomething() { } }`,
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 40,
        },
      ],
    },
    {
      code: `'use strict'
 if (test) { function doSomething() { } }`,
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 40,
        },
      ],
    },
    {
      code: `'use strict'
 if (test) { function doSomething() { } }`,
      options: [
        'both',
        {
          blockScopedFunctions: 'allow',
        },
      ],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 2,
          column: 14,
          endLine: 2,
          endColumn: 40,
        },
      ],
    },
    {
      code: `function foo() {'use strict'
 { function bar() { } } }`,
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to function body root.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      code: `function foo() {'use strict'
 { function bar() { } } }`,
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to function body root.',
          line: 2,
          column: 4,
          endLine: 2,
          endColumn: 22,
        },
      ],
    },
    {
      code: `function doSomething() { 'use strict'
 do { function somethingElse() { } } while (test); }`,
      options: [
        'both',
        {
          blockScopedFunctions: 'disallow',
        },
      ],
      languageOptions: {
        ecmaVersion: 5,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to function body root.',
          line: 2,
          column: 7,
          endLine: 2,
          endColumn: 35,
        },
      ],
    },
    {
      code: `{ function foo () {'use strict'
 console.log('foo called'); } }`,
      options: ['both'],
      languageOptions: {
        ecmaVersion: 2022,
        sourceType: 'script',
      },
      errors: [
        {
          messageId: 'moveDeclToRoot',
          message: 'Move function declaration to program root.',
          line: 1,
          column: 3,
          endLine: 2,
          endColumn: 30,
        },
      ],
    },
  ],
});
